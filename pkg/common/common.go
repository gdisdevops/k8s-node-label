package common

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func CopyNodeObj(node *v1.Node) *v1.Node {
	nodeCopy := node.DeepCopy()

	if nodeCopy.Labels == nil {
		nodeCopy.Labels = make(map[string]string)
	}

	return nodeCopy
}

func ClientSet(kubeconfig string) (kubernetes.Interface, error) {
	var config *rest.Config
	if kubeconfig != "" {
		log.Debug("Use kubeconfig provided by commandline flag")
		conf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		config = conf
	} else {
		log.Debug("Use in-cluster k8s configuration")
		conf, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		config = conf
	}

	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}
