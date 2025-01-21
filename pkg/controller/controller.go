package controller

import (
	"context"
	"fmt"
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
	customRoleLabel         string
	karpenterEnabled        bool
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
	NodeUninitialziedTaint        = "node.cloudprovider.kubernetes.io/uninitialized"
	NodeKarpenterManagedLabelKey  = "karpenter.sh/nodepool"
	NodeKarpenterLabel            = "node-role.kubernetes.io/karpenter"
)

func NewNodeController(client kubernetes.Interface, spotInstanceDiscovery spotdiscovery.SpotDiscoveryInterface, excludeLoadBalancing bool, includeAlphaLabel bool, excludeEviction bool, controlPlaneTaint string, controlPlaneLegacyLabel bool, customRoleLabel string, karpenterEnabled bool) NodeController {
	c := NodeController{
		client:                  client,
		includeAlphaLabel:       includeAlphaLabel,
		excludeLoadBalancing:    excludeLoadBalancing,
		excludeEviction:         excludeEviction,
		spotInstanceDiscovery:   spotInstanceDiscovery,
		controlPlaneTaint:       controlPlaneTaint,
		controlPlaneLegacyLabel: controlPlaneLegacyLabel,
		customRoleLabel:         customRoleLabel,
		karpenterEnabled:        karpenterEnabled,
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
	if c.isNodeInitialized(node) {
		c.markNode(node)
	} else {
		log.Warnf("Node %s was not yet initialzied by cloud controller.", node.Name)
	}
}

func (c NodeController) markNode(node *v1.Node) {
	nodeCopy := common.CopyNodeObj(node)
	nodeChanged := false

	if c.customRoleLabel != "" {
		customRoleLabelValue, err := c.getCustomRoleLabelValue(node)
		if err == nil {
			if !isAlreadyMarkedWithCustomLabel(node, customRoleLabelValue) {
				log.Infof("Mark node %s with custom role label %s", node.Name, customRoleLabel(customRoleLabelValue))
				addCustomRole(nodeCopy, customRoleLabelValue)
				nodeChanged = true
			}
		} else {
			log.Debugf("Node %s doesn't have custom label: %s", node.Name, c.customRoleLabel)
		}
	}

	if c.isWorkerNode(node) && !isAlreadyMarkedWorkerNode(node) {
		log.Infof("Mark worker node %s", node.Name)
		addWorkerLabels(nodeCopy, c.spotInstanceDiscovery.IsSpotInstance(node))
		nodeChanged = true
	} else if c.isControlPlaneNode(node) {
		if !isAlreadyMarkedControlPlane(node) || (c.controlPlaneLegacyLabel && !isAlreadyMarkedMaster(node)) {
			log.Infof("Mark master node %s", node.Name)
			addControlPlaneLabels(nodeCopy, c.includeAlphaLabel, c.excludeLoadBalancing, c.excludeEviction, c.spotInstanceDiscovery.IsSpotInstance(node), c.controlPlaneLegacyLabel)
			nodeChanged = true
		}
	}

	if c.karpenterEnabled && isNodeManagedByKarpenter(node) {
		log.Infof("Mark node %s with karpenter role label %s", node.Name, NodeKarpenterLabel)
		addKarpenterLabel(nodeCopy)
		nodeChanged = true
	}

	if nodeChanged {
		_, err := c.client.CoreV1().Nodes().Update(context.TODO(), nodeCopy, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("Failed to mark node %s with error: %v", node.Name, err)
		}
	} else {
		log.Debugf("Skip node %s because it's already marked", node.Name)
	}
}

func (c NodeController) getCustomRoleLabelValue(node *v1.Node) (string, error) {
	if node.Labels != nil {
		if label, ok := node.Labels[c.customRoleLabel]; ok {
			return label, nil
		}
	}

	return "", fmt.Errorf("Node %s doesn't have %s label", node.Name, c.customRoleLabel)
}

func addCustomRole(node *v1.Node, role string) {
	node.Labels[customRoleLabel(role)] = ""
}

func addWorkerLabels(node *v1.Node, isSpot bool) {
	if isSpot {
		node.Labels[NodeRoleSpotWorkerLabel] = ""
	} else {
		node.Labels[NodeRoleWorkerLabel] = ""
	}
}

func addControlPlaneLabels(node *v1.Node, includeAlphaLabel bool, excludeLoadBalancing bool, excludeEviction bool, isSpot bool, useLegacyMasterLabel bool) {
	if isSpot {
		if useLegacyMasterLabel {
			node.Labels[NodeRoleSpotMasterLabel] = ""
		}
		node.Labels[NodeRoleSpotControlPlaneLabel] = ""

	} else {
		if useLegacyMasterLabel {
			node.Labels[NodeRoleMasterLabel] = ""
		}
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

// Deprecated. Will be removed in future release
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

func isAlreadyMarkedWithCustomLabel(node *v1.Node, customRoleLabelValue string) bool {
	if node.Labels != nil {
		if _, ok := node.Labels[customRoleLabel(customRoleLabelValue)]; ok {
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

func customRoleLabel(role string) string {
	return fmt.Sprintf("node-role.kubernetes.io/%s", role)
}

func (c NodeController) isNodeInitialized(node *v1.Node) bool {
	for i := range node.Spec.Taints {
		if node.Spec.Taints[i].Key == NodeUninitialziedTaint {
			return false
		}
	}
	return true
}

func isNodeManagedByKarpenter(node *v1.Node) bool {
	if node.Labels == nil {
		return false
	}
	_, ok := node.Labels[NodeKarpenterManagedLabelKey]

	return ok
}

func addKarpenterLabel(node *v1.Node) {
	node.Labels[NodeKarpenterLabel] = ""
}

func isAlreadyMarkedKarpenterNode(node *v1.Node) bool {
	if node.Labels == nil {
		return false
	}

	_, ok := node.Labels[NodeKarpenterLabel]
	return ok
}
