package main

import (
	"context"
	"flag"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"os"
	"time"

	"github.com/daspawnw/k8s-node-label/pkg/common"
	"github.com/daspawnw/k8s-node-label/pkg/controller"
	"github.com/daspawnw/k8s-node-label/pkg/spotdiscovery"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	const NAMESPACE_FILE = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	kubeconfig := flag.String("kube-config", "", "Path to a kubeconfig file")
	excludeNodeFromLoadbalancer := flag.Bool("exclude-loadbalancer", false, "Exclude Master nodes from loadbalancer label")
	alphaFlags := flag.Bool("alpha-flags", false, "Include alpha labels")
	excludeEviction := flag.Bool("exclude-evication", false, "Exclude Master node from eviction in case node is not-ready")
	controlPlaneTaint := flag.String("control-plane-taint", "node-role.kubernetes.io/control-plane", "Override default taint for control-plane nodes")
	controlPlaneLegacyLabel := flag.Bool("control-plane-legacy-label", false, "Enable legacy controlPlane label: \"node-role.kubernetes.io/master\"")
	provider := flag.String("provider", "", "Select a provider for spot instance detection, available values: (aws)")
	verbose := flag.Bool("v", false, "Print verbose log messages")
	customRoleLabel := flag.String("custom-role-label", "", "Add additional \"node-role.kubernetes.io/VALUE\" labels equal to this label's value")
	// leases
	leaseId := flag.String("id", uuid.New().String(), "Lease holder identity name")
	leaseLockName := flag.String("lease-lock-name", "k8s-node-label", "Lease lock resource name")
	defaultNs := getCurrentNamespace(NAMESPACE_FILE)
	leaseLockNamespace := flag.String("lease-lock-namespace", defaultNs, "Lease lock resource namespace")

	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if len(*leaseLockNamespace) == 0 {
		log.Fatal("Flag lease-lock-namespace is not set and default value is not available")
		os.Exit(1)
	}

	client, err := common.ClientSet(*kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client %v", err)
		os.Exit(1)
	}

	spotProvider, err := spotdiscovery.SpotProviderFactory(*provider)
	if err != nil {
		log.Fatalf("can't get spot provider client: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	leaseLock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      *leaseLockName,
			Namespace: *leaseLockNamespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: *leaseId,
		},
	}

	// start the leader election code loop
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: leaseLock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// we're notified when we start - this is where you would
				// usually put your code
				// run(ctx)
				controller.NewNodeController(client, spotProvider, *excludeNodeFromLoadbalancer, *alphaFlags, *excludeEviction, *controlPlaneTaint, *controlPlaneLegacyLabel, *customRoleLabel).Controller.Run(wait.NeverStop)
			},
			OnStoppedLeading: func() {
				// we can do cleanup here
				log.Infof("leader lost: %s", *leaseId)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == *leaseId {
					// I just got the lock
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
	})
}

func getCurrentNamespace(path string) string {
	_, err := os.Stat(path)
	if err != nil {
		log.Warnf("Can't find namespace file under %s", path)
		return ""
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		log.Warnf("Can't read file %s correctly", path)
	}
	return string(contents)
}
