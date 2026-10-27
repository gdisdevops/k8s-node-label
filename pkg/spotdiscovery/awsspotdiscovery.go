package spotdiscovery

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

// EC2API is the subset of *ec2.Client used by EC2SpotDiscovery, allowing it to be mocked in tests.
type EC2API interface {
	DescribeSpotInstanceRequests(ctx context.Context, params *ec2.DescribeSpotInstanceRequestsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSpotInstanceRequestsOutput, error)
}

type EC2SpotDiscovery struct {
	ec2Client EC2API
}

func (d EC2SpotDiscovery) IsSpotInstance(ctx context.Context, node *v1.Node) bool {
	instanceID := receiveInstanceID(node)
	if instanceID != nil {
		input := ec2.DescribeSpotInstanceRequestsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("instance-id"),
					Values: []string{*instanceID},
				},
			},
		}
		spotRequest, err := d.ec2Client.DescribeSpotInstanceRequests(ctx, &input)
		if err != nil {
			log.Errorf("Failed to detect if node %s spot instance request in ec2 with error: %v", node.Name, err)
			return false
		}

		return len(spotRequest.SpotInstanceRequests) == 1
	}
	return false
}

func receiveInstanceID(node *v1.Node) *string {
	r, _ := regexp.Compile(`.*?:[\/]{2,3}.*?\/(.*)$`)
	matches := r.FindStringSubmatch(node.Spec.ProviderID)
	if len(matches) == 2 {
		return &matches[1]
	}

	return nil
}
