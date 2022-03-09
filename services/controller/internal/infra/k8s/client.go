package k8s

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func BuildInternalClient() kubernetes.Interface {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("unable to get In cluster config, error %v", err)
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("unable to build client from config, error %v", err)
	}

	return client
}

func BuildExternalClient() kubernetes.Interface {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("unable to get cluster config from flags, error %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to build client from config, error %v", err)
	}

	return client
}
