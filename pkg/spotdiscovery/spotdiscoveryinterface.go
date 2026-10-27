package spotdiscovery

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type SpotDiscoveryInterface interface {
	IsSpotInstance(ctx context.Context, node *v1.Node) bool
}
