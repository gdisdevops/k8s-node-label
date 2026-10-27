package spotdiscovery

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func SpotProviderFactory(ctx context.Context, provider string) (SpotDiscoveryInterface, error) {
	if provider == "aws" {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		ec2Client := ec2.NewFromConfig(cfg)
		return EC2SpotDiscovery{
			ec2Client: ec2Client,
		}, nil
	} else {
		return FalseSpotDiscovery{}, nil
	}
}
