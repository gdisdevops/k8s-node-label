package spotdiscovery

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func SpotProviderFactory(provider string) (SpotDiscoveryInterface, error) {
	if provider == "aws" {
		awsSession, err := session.NewSession()
		if err != nil {
			return nil, err
		}
		awsConfig := &aws.Config{}
		ec2Client := ec2.New(awsSession, awsConfig)
		return EC2SpotDiscovery{
			ec2Client: ec2Client,
		}, nil
	} else {
		return FalseSpotDiscovery{}, nil
	}
}
