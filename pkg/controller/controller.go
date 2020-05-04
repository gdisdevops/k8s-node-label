package controller

import (
	"context"
	"time"

	"github.com/daspawnw/k8s-master-label/pkg/common"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type NodeController struct {
	client               kubernetes.Interface
	Controller           cache.Controller
	includeAlphaLabel    bool
	excludeLoadBalancing bool
	excludeEviction      bool
}

const (
	AlphaExcludeLoadBalancerLabel = "alpha.service-controller.kubernetes.io/exclude-balancer"
	ExcludeLoadBalancerLabel      = "node.kubernetes.io/exclude-from-external-load-balancers"
	ExcludeDisruptionLabel        = "node.kubernetes.io/exclude-disruption"
	NodeRoleMasterLabel           = "node-role.kubernetes.io/master"
)

func NewNodeController(client kubernetes.Interface, excludeLoadBalancing bool, includeAlphaLabel bool, excludeEviction bool) NodeController {
	c := NodeController{
		client:               client,
		includeAlphaLabel:    includeAlphaLabel,
		excludeLoadBalancing: excludeLoadBalancing,
		excludeEviction:      excludeEviction,
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
	if !isMasterNode(node) {
		log.Debugf("Node %s is not master", node.Name)
		return
	}

	if !isAlreadyMarked(node) {
		log.Infof("Mark node %s because it's a master node", node.Name)
		nodeCopy := common.CopyNodeObj(node)
		addLabels(nodeCopy, c.includeAlphaLabel, c.excludeLoadBalancing, c.excludeEviction)

		_, err := c.client.CoreV1().Nodes().Update(context.TODO(), nodeCopy, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("Failed to mark node %s with error: %v", node.Name, err)
		}
	} else {
		log.Debugf("Skip node %s because it's already marked", node.Name)
	}
}

// for details which labelss are recommended please see:
// * https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/2019-07-16-node-role-label-use.md
func addLabels(node *v1.Node, includeAlphaLabel bool, excludeLoadBalancing bool, excludeEviction bool) {
	node.Labels[NodeRoleMasterLabel] = ""

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

func isAlreadyMarked(node *v1.Node) bool {
	if node.Labels != nil {
		if _, ok := node.Labels[NodeRoleMasterLabel]; ok {
			return true
		}
	}

	return false
}

func isMasterNode(node *v1.Node) bool {
	for _, t := range node.Spec.Taints {
		if t.Key == NodeRoleMasterLabel {
			return true
		}
	}

	return false
}
