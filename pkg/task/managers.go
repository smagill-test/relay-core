package task

import (
	"context"
	"fmt"

	"github.com/puppetlabs/nebula-tasks/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PreconfiguredMetadataManager is a task metadata manager that can be
// pre-populated for use in development and testing.
type PreconfiguredMetadataManager struct {
	tasks map[string]*Metadata
}

func (m PreconfiguredMetadataManager) GetByIP(ctx context.Context, ip string) (*Metadata, errors.Error) {
	if m.tasks == nil {
		return nil, errors.NewTaskNotFoundForIP(ip)
	}

	if task, ok := m.tasks[ip]; ok {
		return task, nil
	}

	return nil, errors.NewTaskNotFoundForIP(ip)
}

func NewPreconfiguredMetadataManager(tasks map[string]*Metadata) *PreconfiguredMetadataManager {
	return &PreconfiguredMetadataManager{tasks: tasks}
}

// KubernetesMetadataManager provides metadata about a task by introspecting
// the Kubernetes pod it runs in using regular resource apis and kube
// clients.
type KubernetesMetadataManager struct {
	kubeclient kubernetes.Interface
	namespace  string
}

func (mm *KubernetesMetadataManager) GetByIP(ctx context.Context, ip string) (*Metadata, errors.Error) {
	listOpts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("status.podIP=%s", ip),
	}

	pods, err := mm.kubeclient.CoreV1().Pods(mm.namespace).List(listOpts)
	if err != nil {
		return nil, errors.NewKubernetesPodLookupError().WithCause(err)
	}

	if len(pods.Items) < 1 {
		return nil, errors.NewTaskNotFoundForIP(ip)
	}

	// TODO fine tune this: this should theoretically never return more than 1 pod (if it does,
	// then out network fabric has some serious issues), but we should figure out how to handle
	// this scenario.
	pod := pods.Items[0]

	labels := pod.GetLabels()

	return &Metadata{Name: labels["task-name"], ID: labels["task-ip"]}, nil
}

func NewKubernetesMetadataManager(kubeclient kubernetes.Interface, namespace string) *KubernetesMetadataManager {
	return &KubernetesMetadataManager{
		kubeclient: kubeclient,
		namespace:  namespace,
	}
}

type PreconfiguredSpecManager struct {
	specs map[string]string
}

func (sm PreconfiguredSpecManager) GetByTaskID(ctx context.Context, taskID string) (string, errors.Error) {
	if sm.specs == nil {
		return "", errors.NewTaskSpecNotFoundForID(taskID)
	}

	if _, ok := sm.specs[taskID]; !ok {
		return "", errors.NewTaskSpecNotFoundForID(taskID)
	}

	return sm.specs[taskID], nil
}

func NewPreconfiguredSpecManager(specs map[string]string) *PreconfiguredSpecManager {
	return &PreconfiguredSpecManager{specs: specs}
}

type KubernetesSpecManager struct {
	kubeclient kubernetes.Interface
	namespace  string
}

func (sm KubernetesSpecManager) GetByTaskID(ctx context.Context, taskID string) (string, errors.Error) {
	configMap, err := sm.kubeclient.CoreV1().ConfigMaps(sm.namespace).Get(taskID, metav1.GetOptions{})
	if nil != err {
		if kerrors.IsNotFound(err) {
			return "", errors.NewTaskSpecNotFoundForID(taskID).WithCause(err)
		}

		return "", errors.NewTaskSpecLookupError().WithCause(err)
	}

	if _, ok := configMap.Data["spec.json"]; !ok {
		return "", errors.NewTaskSpecNotFoundForID(taskID)
	}

	return configMap.Data["spec.json"], nil
}

func NewKubernetesSpecManager(kubeclient kubernetes.Interface, namespace string) *KubernetesSpecManager {
	return &KubernetesSpecManager{
		kubeclient: kubeclient,
		namespace:  namespace,
	}
}