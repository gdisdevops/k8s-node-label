package controller

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

var MasterNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-master-node",
	},
	Spec: v1.NodeSpec{
		Taints: []v1.Taint{
			{
				Key:    NodeRoleMasterLabel,
				Effect: "NoSchedule",
			},
		},
	},
}
var WorkerNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-worker-node",
	},
	Spec: v1.NodeSpec{},
}

func TestHandlerShouldSetNodeRoleMaster(t *testing.T) {
	node := MasterNode

	clientset := fake.NewSimpleClientset(node)

	c := NewNodeController(clientset, false, false, false)
	c.handler(node)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-master-node", metav1.GetOptions{})

	if _, ok := foundNode.Labels[NodeRoleMasterLabel]; !ok {
		t.Errorf("Expected label %s on node %s, but was not assigned", NodeRoleMasterLabel, "test-master-node")
	}
}

func TestHandlerShouldNotSetRoleIfNotMaster(t *testing.T) {
	node := WorkerNode

	clientset := fake.NewSimpleClientset(node)
	c := NewNodeController(clientset, false, false, false)
	c.handler(node)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-worker-node", metav1.GetOptions{})
	if _, ok := foundNode.Labels[NodeRoleMasterLabel]; ok {
		t.Errorf("Expected no label %s on node %s, but was assigned", NodeRoleMasterLabel, "test-worker-node")
	}
}

func TestHandlerShouldPreventMasterFromLoadbalancing(t *testing.T) {
	node := MasterNode

	clientset := fake.NewSimpleClientset(node)
	c := NewNodeController(clientset, true, true, false)
	c.handler(node)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-master-node", metav1.GetOptions{})
	if val, ok := foundNode.Labels[ExcludeLoadBalancerLabel]; ok {
		if val != "true" {
			t.Errorf("Expected label %s value 'true', but value was %s", ExcludeLoadBalancerLabel, val)
		}
	} else {
		t.Errorf("Expected label %s on node, but was not assigned", ExcludeLoadBalancerLabel)
	}

	if val, ok := foundNode.Labels[AlphaExcludeLoadBalancerLabel]; ok {
		if val != "true" {
			t.Errorf("Expected label %s value 'true', but value was %s", AlphaExcludeLoadBalancerLabel, val)
		}
	} else {
		t.Errorf("Expected label %s on node, but was not assiged", AlphaExcludeLoadBalancerLabel)
	}
}

func TestHandlerShouldExcludeNodeFromEviction(t *testing.T) {
	node := MasterNode

	clientset := fake.NewSimpleClientset(node)
	c := NewNodeController(clientset, false, false, true)
	c.handler(node)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-master-node", metav1.GetOptions{})
	if val, ok := foundNode.Labels[ExcludeDisruptionLabel]; ok {
		if val != "true" {
			t.Errorf("Expected label %s value 'true', but value was %s", ExcludeDisruptionLabel, val)
		}
	} else {
		t.Errorf("Expected label %s on node, but was not assigned", ExcludeDisruptionLabel)
	}
}
