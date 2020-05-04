package main

import (
	"flag"

	"github.com/daspawnw/k8s-master-label/pkg/common"
	"github.com/daspawnw/k8s-master-label/pkg/controller"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	kubeconfig := flag.String("kube-config", "", "Path to a kubeconfig file")
	excludeNodeFromLoadbalancer := flag.Bool("exclude-loadbalancer", false, "Exclude Master nodes from loadbalancer label")
	alphaFlags := flag.Bool("alpha-flags", false, "Include alpha labels")
	excludeEviction := flag.Bool("exclude-evication", false, "Exclude Master node from eviction in case node is not-ready")
	verbose := flag.Bool("v", false, "Print verbose log messages")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	client, err := common.ClientSet(*kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client %v", err)
	}
	controller.NewNodeController(client, *excludeNodeFromLoadbalancer, *alphaFlags, *excludeEviction).Controller.Run(wait.NeverStop)
}
