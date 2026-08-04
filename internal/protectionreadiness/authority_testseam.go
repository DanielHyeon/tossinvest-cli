//go:build tossos_testseams

package protectionreadiness

// ReadinessSnapshotForTest seals explicit verdicts only in repository test
// binaries. Production snapshots still come exclusively from the signed
// provider.
func ReadinessSnapshotForTest(kr, us Verdict) ReadinessSnapshot {
	snapshot := ReadinessSnapshot{release: ReadinessRelease, kr: kr, us: us}
	snapshot.krSeal = marketVerdictSeal(snapshot.release, snapshot.kr)
	snapshot.usSeal = marketVerdictSeal(snapshot.release, snapshot.us)
	snapshot.seal = readinessSnapshotSeal(snapshot)
	return snapshot
}
