//go:build tossos_testseams

package strategyrouter

import "time"

// ProductionRouteAuthorityForTest creates a sealed production-shaped route
// authority only in explicit repository test-seam binaries.
func ProductionRouteAuthorityForTest(key OwnerKey, horizon Horizon, laneID, laneVersion, evidenceDigest, configDigest string, evaluatedAt time.Time) (ProductionRouteAuthority, error) {
	request, err := StrategyflowRouteFixture(key, horizon, laneID, laneVersion, evidenceDigest, configDigest, evaluatedAt)
	if err != nil {
		return ProductionRouteAuthority{}, err
	}
	return ProductionRouteAuthority{request: request, manifestDigest: "sha256:test-manifest", ownerDigest: request.Snapshot.Digest}, nil
}

// ProductionRouteBatchAuthorityForTest copies explicit sealed authorities into
// a production-shaped batch for engine integration tests.
func ProductionRouteBatchAuthorityForTest(manifestDigest string, authorities ...ProductionRouteAuthority) ProductionRouteBatchAuthority {
	values := make(map[string]ProductionRouteAuthority, len(authorities))
	for _, authority := range authorities {
		request := authority.Request()
		if request.Key.Symbol == "" {
			return ProductionRouteBatchAuthority{}
		}
		authority.manifestDigest = manifestDigest
		values[request.Key.Symbol] = authority
	}
	return ProductionRouteBatchAuthority{values: values, manifestDigest: manifestDigest}
}
