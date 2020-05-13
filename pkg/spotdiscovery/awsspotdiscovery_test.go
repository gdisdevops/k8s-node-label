package spotdiscovery

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var WorkerNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-worker-node",
	},
	Spec: v1.NodeSpec{
		ProviderID: "aws:///eu-central-1/i-123qwe123",
	},
}
var SpotWorkerNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-spot-node",
	},
	Spec: v1.NodeSpec{
		ProviderID: "aws:///eu-central-1/i-123asd132",
	},
}
var UnManagedNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-unmanaged-node",
	},
	Spec: v1.NodeSpec{},
}

func TestIsSpotShouldReturnFalseForNonSpotInstance(t *testing.T) {
	spot := EC2SpotDiscovery{
		ec2Client: &MockEC2Client{},
	}

	response := spot.IsSpotInstance(WorkerNode)
	if response == true {
		t.Errorf("Expected no spot response for test-worker-node, but has spot response")
	}
}

func TestIsSpotShouldReturnTrueForSpotInstance(t *testing.T) {
	spot := EC2SpotDiscovery{
		ec2Client: &MockEC2Client{},
	}

	response := spot.IsSpotInstance(SpotWorkerNode)
	if response == false {
		t.Errorf("Expected spot response for test-spot-node, but no response available")
	}
}

func TestIsSpotShouldReturnFalseForNonProviderManagedInstance(t *testing.T) {
	spot := EC2SpotDiscovery{
		ec2Client: &MockEC2Client{},
	}

	response := spot.IsSpotInstance(UnManagedNode)
	if response == true {
		t.Errorf("Expected no spot response for test-worker-node, but has spot response")
	}
}

type MockEC2Client struct {
	ec2iface.EC2API
}

func (c *MockEC2Client) DescribeSpotInstanceRequests(in *ec2.DescribeSpotInstanceRequestsInput) (*ec2.DescribeSpotInstanceRequestsOutput, error) {
	spotInstances := []*ec2.SpotInstanceRequest{}
	if len(in.Filters) == 1 && len(in.Filters[0].Values) == 1 {
		instanceID := in.Filters[0].Values[0]

		if *instanceID == "i-123asd132" || *instanceID == "i-123uzu123" {
			spotInstances = append(spotInstances, &ec2.SpotInstanceRequest{
				InstanceId: instanceID,
			})
		}
	}

	return &ec2.DescribeSpotInstanceRequestsOutput{
		SpotInstanceRequests: spotInstances,
	}, nil
}
