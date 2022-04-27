package controller

import (
	"context"
	"time"

	"github.com/daspawnw/k8s-node-label/pkg/common"
	"github.com/daspawnw/k8s-node-label/pkg/spotdiscovery"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type NodeController struct {
	client                  kubernetes.Interface
	Controller              cache.Controller
	includeAlphaLabel       bool
	excludeLoadBalancing    bool
	excludeEviction         bool
	spotInstanceDiscovery   spotdiscovery.SpotDiscoveryInterface
	controlPlaneTaint       string
	controlPlaneLegacyLabel bool
}

const (
	AlphaExcludeLoadBalancerLabel = "alpha.service-controller.kubernetes.io/exclude-balancer"
	ExcludeLoadBalancerLabel      = "node.kubernetes.io/exclude-from-external-load-balancers"
	ExcludeDisruptionLabel        = "node.kubernetes.io/exclude-disruption"
	NodeRoleMasterLabel           = "node-role.kubernetes.io/master"
	NodeRoleSpotMasterLabel       = "node-role.kubernetes.io/spot-master"
	NodeRoleControlPlaneLabel     = "node-role.kubernetes.io/control-plane"
	NodeRoleSpotControlPlaneLabel = "node-role.kubernetes.io/spot-control-plane"
	NodeRoleWorkerLabel           = "node-role.kubernetes.io/worker"
	NodeRoleSpotWorkerLabel       = "node-role.kubernetes.io/spot-worker"
)

func NewNodeController(client kubernetes.Interface, spotInstanceDiscovery spotdiscovery.SpotDiscoveryInterface, excludeLoadBalancing bool, includeAlphaLabel bool, excludeEviction bool, controlPlaneTaint string, controlPlaneLegacyLabel bool) NodeController {
	c := NodeController{
		client:                  client,
		includeAlphaLabel:       includeAlphaLabel,
		excludeLoadBalancing:    excludeLoadBalancing,
		excludeEviction:         excludeEviction,
		spotInstanceDiscovery:   spotInstanceDiscovery,
		controlPlaneTaint:       controlPlaneTaint,
		controlPlaneLegacyLabel: controlPlaneLegacyLabel,
	}

	nodeListWatcher := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"nodes",
		v1.NamespaceAll,
		fields.Everything())

	_, controller := cache.NewInformer(nodeListWatcher,
		&v1.Node{},
		60*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handler,
			UpdateFunc: func(old, new interface{}) { c.handler(new) },
		},
	)

	c.Controller = controller

	return c
}

func (c NodeController) handler(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		return
	}
	log.Debugf("Received handler event for node %s", node.Name)
	c.markNode(node)
}

func (c NodeController) markNode(node *v1.Node) {
	nodeCopy := common.CopyNodeObj(node)

	if c.isWorkerNode(node) && !isAlreadyMarkedWorkerNode(node) {
		log.Infof("Mark worker node %s", node.Name)
		addWorkerLabels(nodeCopy, c.spotInstanceDiscovery.IsSpotInstance(node))
	} else if c.isControlPlaneNode(node) {
		if c.controlPlaneLegacyLabel && !isAlreadyMarkedMaster(node) {
			log.Infof("Mark master node %s", node.Name)
			addMasterLabels(nodeCopy, c.includeAlphaLabel, c.excludeLoadBalancing, c.excludeEviction, c.spotInstanceDiscovery.IsSpotInstance(node))
		}
		if !isAlreadyMarkedControlPlane(node) {
			log.Infof("Mark control-plane node %s", node.Name)
			addControlPlaneLabels(nodeCopy, c.includeAlphaLabel, c.excludeLoadBalancing, c.excludeEviction, c.spotInstanceDiscovery.IsSpotInstance(node))
		}
	} else {
		log.Debugf("Skip node %s because it's already marked", node.Name)
		return
	}

	_, err := c.client.CoreV1().Nodes().Update(context.TODO(), nodeCopy, metav1.UpdateOptions{})
	if err != nil {
		log.Errorf("Failed to mark node %s with error: %v", node.Name, err)
	}
}

func addWorkerLabels(node *v1.Node, isSpot bool) {
	if isSpot {
		node.Labels[NodeRoleSpotWorkerLabel] = ""
	} else {
		node.Labels[NodeRoleWorkerLabel] = ""
	}
}

//Deprecated
// for details which labels are recommended please see:
// * https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md
func addMasterLabels(node *v1.Node, includeAlphaLabel bool, excludeLoadBalancing bool, excludeEviction bool, isSpot bool) {
	if isSpot {
		node.Labels[NodeRoleSpotMasterLabel] = ""
	} else {
		node.Labels[NodeRoleMasterLabel] = ""
	}

	if excludeEviction == true {
		node.Labels[ExcludeDisruptionLabel] = "true"
	}

	if excludeLoadBalancing == true {
		node.Labels[ExcludeLoadBalancerLabel] = "true"

		if includeAlphaLabel == true {
			node.Labels[AlphaExcludeLoadBalancerLabel] = "true"
		}
	}
}

func addControlPlaneLabels(node *v1.Node, includeAlphaLabel bool, excludeLoadBalancing bool, excludeEviction bool, isSpot bool) {
	if isSpot {
		node.Labels[NodeRoleSpotControlPlaneLabel] = ""
	} else {
		node.Labels[NodeRoleControlPlaneLabel] = ""
	}

	if excludeEviction == true {
		node.Labels[ExcludeDisruptionLabel] = "true"
	}

	if excludeLoadBalancing == true {
		node.Labels[ExcludeLoadBalancerLabel] = "true"

		if includeAlphaLabel == true {
			node.Labels[AlphaExcludeLoadBalancerLabel] = "true"
		}
	}
}

//Deprecated. Will be removed in future release
func isAlreadyMarkedMaster(node *v1.Node) bool {
	if node.Labels != nil {
		if _, ok := node.Labels[NodeRoleMasterLabel]; ok {
			return true
		}

		if _, ok := node.Labels[NodeRoleSpotMasterLabel]; ok {
			return true
		}
	}

	return false
}

func isAlreadyMarkedControlPlane(node *v1.Node) bool {
	if node.Labels != nil {
		if _, ok := node.Labels[NodeRoleControlPlaneLabel]; ok {
			return true
		}

		if _, ok := node.Labels[NodeRoleSpotControlPlaneLabel]; ok {
			return true
		}
	}

	return false
}

func isAlreadyMarkedWorkerNode(node *v1.Node) bool {
	if node.Labels != nil {
		if _, ok := node.Labels[NodeRoleWorkerLabel]; ok {
			return true
		}

		if _, ok := node.Labels[NodeRoleSpotWorkerLabel]; ok {
			return true
		}
	}

	return false
}

func (c NodeController) isControlPlaneNode(node *v1.Node) bool {
	for _, t := range node.Spec.Taints {
		if t.Key == c.controlPlaneTaint {
			return true
		}
	}

	return false
}

func (c NodeController) isWorkerNode(node *v1.Node) bool {
	return !c.isControlPlaneNode(node)
}
