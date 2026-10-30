#!/bin/sh
set -eu
ip link set lo up
readlink /proc/self/ns/net
ip -brief address
ip route show
printf 'mode=%s\n' "$1"
if [ "$1" = race ]; then
 GOTOOLCHAIN=go1.25.0 GOPROXY=off go test -race -tags tossos_testseams ./internal/accountbudget -run '^(TestA169Fence.*|TestObservedTransport(RateLimitAndPoison|BeginFaultClosesBody|FinishFaultPreservesResponse|PreSendReleaseError|CancelAfterBeginSettles|NoResponseAfterSent)|TestTransportOwnsFullBody|TestAttemptClassificationAfterDrain|TestProxySelectionRefusedBeforeAdmission|TestCancelBlockedTLSWriteDrains|TestCancelUnreadResponseReleasesOwnership|TestA166ObservationPostcommit(BeginFailureBecomesGap|FinishFailureNoDuplicate)|TestA166ConsumerSnapshotBoundary)$' -count=1 -v
 GOTOOLCHAIN=go1.25.0 GOPROXY=off go test -race -tags tossos_testseams ./internal/accountbudgetruntime -run '^(TestA169Fence(Official401TokenAndSticky|OfficialCachedAdoptAndMint|CrossProcessToken429)|TestA166WiringCensus(Baseline|RejectsBypasses|RejectsDirectorySymlink))$' -count=1 -v
else
 GOTOOLCHAIN=go1.25.0 GOPROXY=off go test ./internal/accountbudget -run '^(TestA169Fence.*|TestObservedTransport(RateLimitAndPoison|BeginFaultClosesBody|FinishFaultPreservesResponse|PreSendReleaseError|CancelAfterBeginSettles|NoResponseAfterSent)|TestTransportOwnsFullBody|TestAttemptClassificationAfterDrain|TestProxySelectionRefusedBeforeAdmission|TestCancelBlockedTLSWriteDrains|TestCancelUnreadResponseReleasesOwnership|TestA166ObservationPostcommit(BeginFailureBecomesGap|FinishFailureNoDuplicate)|TestA166ConsumerSnapshotBoundary)$' -count=1 -v
 GOTOOLCHAIN=go1.25.0 GOPROXY=off go test ./internal/accountbudgetruntime -run '^(TestA169Fence(Official401TokenAndSticky|OfficialCachedAdoptAndMint|CrossProcessToken429)|TestA166WiringCensus(Baseline|RejectsBypasses|RejectsDirectorySymlink))$' -count=1 -v
fi
