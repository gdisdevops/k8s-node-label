package spotdiscovery

import v1 "k8s.io/api/core/v1"

type FalseSpotDiscovery struct{}

func (FalseSpotDiscovery) IsSpotInstance(node *v1.Node) bool {
	return false
}
