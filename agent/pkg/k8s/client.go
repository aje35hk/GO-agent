package k8s

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// InitClient initializes the Kubernetes client using in-cluster configuration.
func InitClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Failed to create in-cluster config: %v", err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Failed to create clientset: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to Kubernetes cluster")
	return clientset, nil
}
