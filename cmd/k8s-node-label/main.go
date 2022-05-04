package main

import (
	"flag"

	"github.com/daspawnw/k8s-node-label/pkg/common"
	"github.com/daspawnw/k8s-node-label/pkg/controller"
	"github.com/daspawnw/k8s-node-label/pkg/spotdiscovery"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	kubeconfig := flag.String("kube-config", "", "Path to a kubeconfig file")
	excludeNodeFromLoadbalancer := flag.Bool("exclude-loadbalancer", false, "Exclude Master nodes from loadbalancer label")
	alphaFlags := flag.Bool("alpha-flags", false, "Include alpha labels")
	excludeEviction := flag.Bool("exclude-evication", false, "Exclude Master node from eviction in case node is not-ready")
	controlPlaneTaint := flag.String("control-plane-taint", "node-role.kubernetes.io/control-plane", "Override default taint for control-plane nodes")
	controlPlaneLegacyLabel := flag.Bool("control-plane-legacy-label", false, "Enable legacy controlPlane label: \"node-role.kubernetes.io/master\"")
	provider := flag.String("provider", "", "Select a provider for spot instance detection, available values: (aws)")
	verbose := flag.Bool("v", false, "Print verbose log messages")
	customRoleLabel := flag.String("custom-role-label", "", "Add additional \"node-role.kubernetes.io/VALUE\" labels equal to this label's value")
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

	spotProvider := spotdiscovery.SpotProviderFactory(*provider)
	controller.NewNodeController(client, spotProvider, *excludeNodeFromLoadbalancer, *alphaFlags, *excludeEviction, *controlPlaneTaint, *controlPlaneLegacyLabel, *customRoleLabel).Controller.Run(wait.NeverStop)
}
