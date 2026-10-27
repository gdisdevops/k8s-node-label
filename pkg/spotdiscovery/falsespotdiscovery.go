package spotdiscovery

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type FalseSpotDiscovery struct{}

func (FalseSpotDiscovery) IsSpotInstance(ctx context.Context, node *v1.Node) bool {
	return false
}
