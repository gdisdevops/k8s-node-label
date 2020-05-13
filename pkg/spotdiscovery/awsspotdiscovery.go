package spotdiscovery

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type EC2SpotDiscovery struct {
	ec2Client ec2iface.EC2API
}

func (d EC2SpotDiscovery) IsSpotInstance(node *v1.Node) bool {
	instanceID := receiveInstanceID(node)
	if instanceID != nil {
		input := ec2.DescribeSpotInstanceRequestsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance-id"),
					Values: []*string{aws.String(*instanceID)},
				},
			},
		}
		spotRequest, err := d.ec2Client.DescribeSpotInstanceRequests(&input)
		if err != nil {
			log.Errorf("Failed to detect if node %s spot instance request in ec2 with error: %v", node.Name, err)
			return false
		}

		return len(spotRequest.SpotInstanceRequests) == 1
	}
	return false
}

func receiveInstanceID(node *v1.Node) *string {
	r, _ := regexp.Compile(".*?:[\\/]{2,3}.*?\\/(.*)$")
	matches := r.FindStringSubmatch(node.Spec.ProviderID)
	if matches != nil && len(matches) == 2 {
		return &matches[1]
	}

	return nil
}
