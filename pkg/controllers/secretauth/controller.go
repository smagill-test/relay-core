package secretauth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	tekv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tekclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	tekinformers "github.com/tektoncd/pipeline/pkg/client/informers/externalversions"
	tekv1informer "github.com/tektoncd/pipeline/pkg/client/informers/externalversions/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	nebulav1 "github.com/puppetlabs/nebula-tasks/pkg/apis/nebula.puppet.com/v1"
	"github.com/puppetlabs/nebula-tasks/pkg/config"
	"github.com/puppetlabs/nebula-tasks/pkg/data/secrets/vault"
	clientset "github.com/puppetlabs/nebula-tasks/pkg/generated/clientset/versioned"
	informers "github.com/puppetlabs/nebula-tasks/pkg/generated/informers/externalversions"
	sainformers "github.com/puppetlabs/nebula-tasks/pkg/generated/informers/externalversions/nebula.puppet.com/v1"
)

const (
	// default name for the workflow metadata api pod and service
	metadataServiceName = "workflow-metadata-api"
)

// Controller watches for nebulav1.SecretAuth resource changes.
// If a SecretAuth resource is created, the controller will create a service acccount + rbac
// for the namespace, then inform vault that that service account is allowed to access
// readonly secrets under a preconfigured path related to a nebula workflow. It will then
// spin up a pod running an instance of nebula-metadata-api that knows how to
// ask kubernetes for the service account token, that it will use to proxy secrets
// between the task pods and the vault server.
type Controller struct {
	kubeclient         kubernetes.Interface
	nebclient          clientset.Interface
	tekclient          tekclientset.Interface
	nebInformerFactory informers.SharedInformerFactory
	saInformer         sainformers.SecretAuthInformer
	saInformerSynced   cache.InformerSynced
	tekInformerFactory tekinformers.SharedInformerFactory
	plrInformer        tekv1informer.PipelineRunInformer
	plrInformerSynced  cache.InformerSynced
	saworker           *worker
	plrworker          *worker
	vaultClient        *vault.VaultAuth

	cfg *config.SecretAuthControllerConfig
}

// Run starts all required informers and spawns two worker goroutines
// that will pull resource objects off the workqueue. This method blocks
// until stopCh is closed or an earlier bootstrap call results in an error.
func (c *Controller) Run(numWorkers int, stopCh chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.saworker.shutdown()
	defer c.plrworker.shutdown()

	c.nebInformerFactory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, c.saInformerSynced); !ok {
		return fmt.Errorf("failed to wait for informer cache to sync")
	}

	c.tekInformerFactory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, c.plrInformerSynced); !ok {
		return fmt.Errorf("failed to wait for informer cache to sync")
	}

	c.saworker.run(numWorkers, stopCh)
	c.plrworker.run(numWorkers, stopCh)

	<-stopCh

	return nil
}

// processSingleItem is responsible for creating all the resouces required for
// secret handling and authentication.
// TODO break this logic out into smaller chunks... especially the calls to the vault api
func (c *Controller) processSingleItem(key string) error {
	klog.Infof("syncing SecretAuth %s", key)
	defer klog.Infof("done syncing SecretAuth %s", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	sa, err := c.nebclient.NebulaV1().SecretAuths(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// if anything fails while creating resources, the status object will not be filled out
	// and saved. this means that if any of the keys are empty, we haven't created resources yet.
	if sa.Status.ServiceAccount != "" {
		klog.Infof("resources have already been created %s", key)
		return nil
	}

	var (
		saccount  *corev1.ServiceAccount
		role      *rbacv1.Role
		binding   *rbacv1.RoleBinding
		pod       *corev1.Pod
		service   *corev1.Service
		configMap *corev1.ConfigMap
	)

	saccount, err = createServiceAccount(c.kubeclient, sa)
	if err != nil {
		return err
	}

	klog.Infof("writing vault readonly access policy %s", sa.Spec.WorkflowID)
	// now we let vault know about the service account
	if err := c.vaultClient.WritePolicy(namespace, sa.Spec.WorkflowID); err != nil {
		return err
	}

	klog.Infof("enabling vault access for workflow service account %s", sa.Spec.WorkflowID)
	if err := c.vaultClient.WriteRole(namespace, saccount.GetName(), namespace); err != nil {
		return err
	}

	role, binding, err = createRBAC(c.kubeclient, sa)
	if err != nil {
		return err
	}

	pod, err = createMetadataAPIPod(
		c.kubeclient,
		c.cfg.MetadataServiceImage,
		saccount,
		sa,
		c.vaultClient.Address(),
		c.vaultClient.EngineMount(),
	)
	if err != nil {
		return err
	}

	service, err = createMetadataAPIService(c.kubeclient, sa)
	if err != nil {
		return err
	}

	configMap, err = createWorkflowConfigMap(c.kubeclient, service, sa)
	if err != nil {
		return err
	}

	klog.Infof("waiting for metadata service to become ready %s", sa.Spec.WorkflowID)

	// This waits for a Modified watch event on a service's Endpoint object.
	// When this event is received, it will check it's addresses to see if there's
	// pods that are ready to be served.
	if err := c.waitForEndpoint(service); err != nil {
		return err
	}

	// Because of a possible race condition bug in the kernel network stack, there's a very
	// tiny window of time where packets will get dropped if you try to make requests to the ports
	// that are supposed to be forwarded to underlying pods. This unfortunately happens quite frequently
	// since we exec task pods from Tekton very quickly. This function will make GET requests in a loop
	// to the readiness endpoint of the pod (via the service dns) to make sure it actually gets a 200
	// response before setting the status object on SecretAuth resources.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := c.waitForServiceToActuallyBeUp(ctx, service); err != nil {
		return err
	}

	klog.Infof("metadata service is ready %s", sa.Spec.WorkflowID)

	saCopy := sa.DeepCopy()
	saCopy.Status.MetadataServicePod = pod.GetName()
	saCopy.Status.MetadataServiceService = service.GetName()
	saCopy.Status.ServiceAccount = saccount.GetName()
	saCopy.Status.ConfigMap = configMap.GetName()
	saCopy.Status.Role = role.GetName()
	saCopy.Status.RoleBinding = binding.GetName()
	saCopy.Status.VaultPolicy = namespace
	saCopy.Status.VaultAuthRole = namespace

	klog.Info("updating secretauth resource status for ", sa.Spec.WorkflowID)
	saCopy, err = c.nebclient.NebulaV1().SecretAuths(namespace).Update(saCopy)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) waitForEndpoint(service *corev1.Service) error {
	var (
		conditionMet bool
		timeout      = int64(30)
	)

	endpoints, err := c.kubeclient.CoreV1().Endpoints(service.GetNamespace()).Get(service.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	listOptions := metav1.SingleObject(endpoints.ObjectMeta)
	listOptions.TimeoutSeconds = &timeout

	watcher, err := c.kubeclient.CoreV1().Endpoints(endpoints.GetNamespace()).Watch(listOptions)
	if err != nil {
		return err
	}

eventLoop:
	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Modified:
			endpoints := event.Object.(*corev1.Endpoints)

			if endpoints.Subsets != nil && len(endpoints.Subsets) > 0 {
				for _, subset := range endpoints.Subsets {
					if subset.Addresses != nil && len(subset.Addresses) > 0 {
						watcher.Stop()
						conditionMet = true

						break eventLoop
					}
				}
			}
		}
	}

	if !conditionMet {
		return fmt.Errorf("timeout occurred while waiting for the metadata service to be ready")
	}

	return nil
}

func (c *Controller) waitForServiceToActuallyBeUp(ctx context.Context, service *corev1.Service) error {
	u := fmt.Sprintf("http://%s.%s.svc.cluster.local/healthz", service.GetName(), service.GetNamespace())

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout occurred while waiting for the metadata service to be ready")
		default:
			<-time.After(time.Millisecond * 750)

			resp, err := http.Get(u)
			if err != nil {
				klog.Infof("got an error when probing the metadata api %s", err)

				continue
			}

			if resp.StatusCode != http.StatusOK {
				klog.Infof("got an invalid status code when probing the metadata api %d", resp.StatusCode)

				continue
			}

			return nil
		}
	}

	return nil
}

func (c *Controller) enqueueSecretAuth(obj interface{}) {
	sa := obj.(*nebulav1.SecretAuth)

	key, err := cache.MetaNamespaceKeyFunc(sa)
	if err != nil {
		utilruntime.HandleError(err)

		return
	}

	c.saworker.add(key)
}

func (c *Controller) enqueuePipelineRunChange(old, obj interface{}) {
	// old is ignored because we only care about the current state
	plr := obj.(*tekv1alpha1.PipelineRun)

	key, err := cache.MetaNamespaceKeyFunc(plr)
	if err != nil {
		utilruntime.HandleError(err)

		return
	}

	c.plrworker.add(key)
}

func (c *Controller) processPipelineRunChange(key string) error {
	klog.Infof("syncing PipelineRun change %s", key)
	defer klog.Infof("done syncing PipelineRun change %s", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	plr, err := c.tekclient.TektonV1alpha1().PipelineRuns(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// TODO if the pipeline run isn't found, then we will still need to clean up SecretAuth
		// resources, but the business logic for this still needs to be defined
		return nil
	}
	if err != nil {
		return err
	}

	if plr.IsDone() {
		sas, err := c.nebclient.NebulaV1().SecretAuths(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		core := c.kubeclient.CoreV1()
		rbac := c.kubeclient.RbacV1()
		sac := c.nebclient.NebulaV1().SecretAuths(namespace)
		opts := &metav1.DeleteOptions{}

		for _, sa := range sas.Items {
			klog.Infof("deleting resources created by %s", sa.GetName())

			err = core.Pods(namespace).Delete(sa.Status.MetadataServicePod, opts)
			if err != nil {
				return err
			}

			err = core.Services(namespace).Delete(sa.Status.MetadataServiceService, opts)
			if err != nil {
				return err
			}

			err = core.ServiceAccounts(namespace).Delete(sa.Status.ServiceAccount, opts)
			if err != nil {
				return err
			}

			err = core.ConfigMaps(namespace).Delete(sa.Status.ConfigMap, opts)
			if err != nil {
				return err
			}

			err = rbac.RoleBindings(namespace).Delete(sa.Status.RoleBinding, opts)
			if err != nil {
				return err
			}

			err = rbac.Roles(namespace).Delete(sa.Status.Role, opts)
			if err != nil {
				return err
			}

			err = c.vaultClient.DeleteRole(sa.Status.VaultAuthRole)
			if err != nil {
				return err
			}

			err = c.vaultClient.DeletePolicy(sa.Status.VaultPolicy)
			if err != nil {
				return err
			}

			if err := sac.Delete(sa.GetName(), opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewController(cfg *config.SecretAuthControllerConfig, vaultClient *vault.VaultAuth) (*Controller, error) {
	kcfg, err := clientcmd.BuildConfigFromFlags(cfg.KubeMasterURL, cfg.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kc, err := kubernetes.NewForConfig(kcfg)
	if err != nil {
		return nil, err
	}

	nebclient, err := clientset.NewForConfig(kcfg)
	if err != nil {
		return nil, err
	}

	tekclient, err := tekclientset.NewForConfig(kcfg)
	if err != nil {
		return nil, err
	}

	nebInformerFactory := informers.NewSharedInformerFactory(nebclient, time.Second*30)
	saInformer := nebInformerFactory.Nebula().V1().SecretAuths()

	tekInformerFactory := tekinformers.NewSharedInformerFactory(tekclient, time.Second*30)
	plrInformer := tekInformerFactory.Tekton().V1alpha1().PipelineRuns()

	c := &Controller{
		kubeclient:         kc,
		nebclient:          nebclient,
		tekclient:          tekclient,
		nebInformerFactory: nebInformerFactory,
		saInformer:         saInformer,
		saInformerSynced:   saInformer.Informer().HasSynced,
		tekInformerFactory: tekInformerFactory,
		plrInformer:        plrInformer,
		plrInformerSynced:  plrInformer.Informer().HasSynced,
		vaultClient:        vaultClient,
		cfg:                cfg,
	}

	c.saworker = newWorker("SecretAuths", (*c).processSingleItem, defaultMaxRetries)
	c.plrworker = newWorker("PipelineRuns", (*c).processPipelineRunChange, defaultMaxRetries)

	saInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueSecretAuth,
	})

	plrInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.enqueuePipelineRunChange,
	})

	return c, nil
}

func createServiceAccount(kc kubernetes.Interface, sa *nebulav1.SecretAuth) (*corev1.ServiceAccount, error) {
	saccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getName(sa, ""),
			Namespace: sa.GetNamespace(),
			Labels:    getLabels(sa, nil),
		},
	}

	klog.Infof("creating service account %s", sa.Spec.WorkflowID)
	saccount, err := kc.CoreV1().ServiceAccounts(sa.GetNamespace()).Create(saccount)
	if errors.IsAlreadyExists(err) {
		saccount, err = kc.CoreV1().ServiceAccounts(sa.GetNamespace()).Get(getName(sa, ""), metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}

	return saccount, nil
}

func createRBAC(kc kubernetes.Interface, sa *nebulav1.SecretAuth) (*rbacv1.Role, *rbacv1.RoleBinding, error) {
	var err error

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getName(sa, ""),
			Namespace: sa.GetNamespace(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"list", "watch", "get"},
			},
		},
	}

	binding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getName(sa, ""),
			Namespace: sa.GetNamespace(),
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     getName(sa, ""),
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      getName(sa, ""),
				Kind:      "ServiceAccount",
				Namespace: sa.GetNamespace(),
			},
		},
	}

	klog.Infof("creating role %s", sa.Spec.WorkflowID)
	role, err = kc.RbacV1().Roles(sa.GetNamespace()).Create(role)
	if errors.IsAlreadyExists(err) {
		role, err = kc.RbacV1().Roles(sa.GetNamespace()).Get(getName(sa, ""), metav1.GetOptions{})
	}
	if err != nil {
		return nil, nil, err
	}

	klog.Infof("creating role binding %s", sa.Spec.WorkflowID)
	binding, err = kc.RbacV1().RoleBindings(sa.GetNamespace()).Create(binding)
	if errors.IsAlreadyExists(err) {
		binding, err = kc.RbacV1().RoleBindings(sa.GetNamespace()).Get(getName(sa, ""), metav1.GetOptions{})
	}
	if err != nil {
		return nil, nil, err
	}

	return role, binding, nil
}

func createMetadataAPIPod(kc kubernetes.Interface, image string, saccount *corev1.ServiceAccount,
	sa *nebulav1.SecretAuth, vaultAddr, vaultEngineMount string) (*corev1.Pod, error) {

	name := getName(sa, metadataServiceName)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sa.GetNamespace(),
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            name,
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command: []string{
						"/usr/bin/nebula-metadata-api",
						"-bind-addr",
						":7000",
						"-vault-addr",
						vaultAddr,
						"-vault-role",
						sa.GetNamespace(),
						"-workflow-id",
						sa.Spec.WorkflowID,
						"-vault-engine-mount",
						vaultEngineMount,
						"-namespace",
						sa.GetNamespace(),
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 7000,
						},
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/healthz",
								Port: intstr.FromInt(7000),
							},
						},
					},
				},
			},
			ServiceAccountName: saccount.GetName(),
			RestartPolicy:      corev1.RestartPolicyOnFailure,
		},
	}

	klog.Infof("creating metadata service pod %s", sa.Spec.WorkflowID)

	pod, err := kc.CoreV1().Pods(sa.GetNamespace()).Create(pod)
	if errors.IsAlreadyExists(err) {
		pod, err = kc.CoreV1().Pods(sa.GetNamespace()).Get(name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func createMetadataAPIService(kc kubernetes.Interface, sa *nebulav1.SecretAuth) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getName(sa, metadataServiceName),
			Namespace: sa.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(7000),
				},
			},
			Selector: map[string]string{
				"app": getName(sa, metadataServiceName),
			},
		},
	}

	klog.Infof("creating pod service %s", sa.Spec.WorkflowID)

	service, err := kc.CoreV1().Services(sa.GetNamespace()).Create(service)
	if errors.IsAlreadyExists(err) {
		service, err = kc.CoreV1().Services(sa.GetNamespace()).Get(getName(sa, metadataServiceName), metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}

	return service, nil
}

func createWorkflowConfigMap(kc kubernetes.Interface, service *corev1.Service, sa *nebulav1.SecretAuth) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getName(sa, ""),
			Namespace: sa.GetNamespace(),
		},
		Data: map[string]string{
			"metadata-api-url": fmt.Sprintf("http://%s.%s.svc.cluster.local", service.GetName(), sa.GetNamespace()),
		},
	}

	klog.Infof("creating config map %s", sa.Spec.WorkflowID)
	configMap, err := kc.CoreV1().ConfigMaps(sa.GetNamespace()).Create(configMap)
	if errors.IsAlreadyExists(err) {
		configMap, err = kc.CoreV1().ConfigMaps(sa.GetNamespace()).Get(getName(sa, ""), metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}

	return configMap, nil
}

func getName(sa *nebulav1.SecretAuth, name string) string {
	prefix := "workflow-run"

	if name == "" {
		return fmt.Sprintf("%s-%s", prefix, sa.Spec.WorkflowRunID)
	}

	return fmt.Sprintf("%s-%s-%s", prefix, sa.Spec.WorkflowRunID, name)
}

func getLabels(sa *nebulav1.SecretAuth, additional map[string]string) map[string]string {
	labels := map[string]string{
		"workflow-run-id": sa.Spec.WorkflowRunID,
		"workflow-id":     sa.Spec.WorkflowID,
	}

	if additional != nil {
		for k, v := range additional {
			labels[k] = v
		}
	}

	return labels
}
