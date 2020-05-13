package spotdiscovery

import v1 "k8s.io/api/core/v1"

type SpotDiscoveryInterface interface {
	IsSpotInstance(node *v1.Node) bool
}