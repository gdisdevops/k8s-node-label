package spotdiscovery

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func SpotProviderFactory(provider string) SpotDiscoveryInterface {
	if provider == "aws" {
		awsSession := session.New()
		awsConfig := &aws.Config{}
		ec2Client := ec2.New(awsSession, awsConfig)
		return EC2SpotDiscovery{
			ec2Client: ec2Client,
		}
	} else {
		return FalseSpotDiscovery{}
	}
}
