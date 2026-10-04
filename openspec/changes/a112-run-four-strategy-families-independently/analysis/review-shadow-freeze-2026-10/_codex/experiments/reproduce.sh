#!/usr/bin/env bash
set -euo pipefail
cd /tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow
GOCACHE="$PWD/.gocache" GOTMPDIR="$PWD/.gotmp" GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local TMPDIR="$PWD/.gotmp" go test -tags tossos_testseams ./internal/strategyworker ./internal/strategyrouter ./review-experiments ./internal/strategyprojectionrpc ./internal/app/engine -run '^(TestReview.*|TestARestartAfterAnObservedPromotionComesBackOffOffUnobservedWithNothingWritten|TestTheRuntimeVocabularyIsExactlyUnobservedUntilAShadowLotExtendsIt|TestTheEncoderGuardCountsEveryWayToNameThePackage)$' -count=1 -v
GOCACHE="$PWD/.gocache" GOTMPDIR="$PWD/.gotmp" GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local TMPDIR="$PWD/.gotmp" go test ./tools/logic-map -run '^TestReviewShadowEvidence$' -count=1 -v
