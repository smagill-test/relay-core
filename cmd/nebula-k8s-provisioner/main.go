package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/puppetlabs/nebula-tasks/pkg/provisioning"
	"github.com/puppetlabs/nebula-tasks/pkg/provisioning/models"
	"github.com/puppetlabs/nebula-tasks/pkg/taskutil"
)

func main() {
	specURL := flag.String("spec-url", os.Getenv(taskutil.SpecURLEnvName), "url to fetch the spec from")
	specFile := flag.String("spec-file", "", "filepath to json formatted spec. overrides -spec-url.")

	flag.Parse()

	log.Println("provisioning k8s cluster")

	var spec models.K8sProvisionerSpec

	if *specFile == "" {
		planOpts := taskutil.DefaultPlanOptions{SpecURL: *specURL}

		if err := taskutil.PopulateSpecFromDefaultPlan(&spec, planOpts); err != nil {
			log.Fatal(err)
		}
	} else {
		f, err := os.Open(*specFile)
		if err != nil {
			log.Fatal(err)
		}

		decoder := taskutil.DefaultJSONSpecDecoder{}

		if err := decoder.DecodeSpec(f, &spec); err != nil {
			log.Fatal(err)
		}
	}

	manager, err := provisioning.NewK8sClusterManagerFromSpec(&spec)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: we need to figure out how to better provision a cluster and report readiness.
	// Currently we set a massively long timeout, which is a non-ideal solution.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	state, err := manager.Synchronize(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(state.Status)
}
