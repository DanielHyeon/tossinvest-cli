#!/bin/sh
set -eu
ip link set lo up
printf 'namespace: '
readlink /proc/self/ns/net
ip -brief address
ip route show
export GOFLAGS=-buildvcs=false
regex='^(TestA166ExactExitOrigin|TestA166OriginRejectsForeignRowsAndBound|TestA166OriginDBErrorReturned|TestA166ReplayOriginClassification|TestA166ReplayOriginDBErrorPropagates|TestA166OrderableActionCoverage|TestA166RequiredReplayUnknownOriginFallsBack|TestA166ReplayAdmissionRefusalPreservesInDoubtJournal|TestA166Official401TokenInheritsCritical|TestReplay409DoesNotConsumeTheCap|TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome|TestA166ExitObservePricesCritical|TestA166ExitProposalPriority|TestA166FullTakeProfitCancelIsNormal|TestA166GatewayReplayBinding|TestA166FlattenBootstrapPriority|TestA166FlattenBootstrapUsesBorrowedBudget)$'
go test "$@" ./internal/journal ./internal/execgw ./internal/app/engine ./cmd/tossctl -run "$regex" -count=1 -timeout=90s
