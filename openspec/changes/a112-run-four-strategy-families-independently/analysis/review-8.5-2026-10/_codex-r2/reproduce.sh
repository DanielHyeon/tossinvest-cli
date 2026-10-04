#!/bin/bash
set -euo pipefail
ROOT=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-85-r2
cd "$ROOT"
for label in A-pre A-post; do
 REVIEW_OUT="$ROOT/review-r2/$label.replay.jsonl" bash review-r2/run-test.sh "$label" ./internal/breakoutlane -run '^TestReviewA' -count=1 -v
done
cmp review-r2/A-pre.replay.jsonl review-r2/A-post.replay.jsonl
for label in B-pre B-post; do
 REVIEW_OUT="$ROOT/review-r2/$label.replay.txt" bash review-r2/run-test.sh "$label" ./internal/strategyrouter -run '^TestReviewB1' -count=1 -v
 bash review-r2/run-test.sh "$label" -tags tossos_testseams ./internal/app/engine -run '^TestReviewB2(NilGetenv|CancelledAndDelayed|ClosedDownstream)$' -count=1 -v
 bash review-r2/run-test.sh "$label" -overlay "$ROOT/review-r2/$label-io-overlay.json" -tags tossos_testseams ./internal/app/engine -run '^TestReviewB2ProductionReadDelay$' -count=1 -v
 for branch in B11 B14; do
 REVIEW_FORCED_BRANCH="$branch" bash review-r2/run-test.sh "$label" -overlay "$ROOT/review-r2/$label-$branch-overlay.json" -tags tossos_testseams ./internal/app/engine -run '^TestReviewB2ClosedDownstream$' -count=1 -v
 done
done
cmp review-r2/B-pre.replay.txt review-r2/B-post.replay.txt
REVIEW_OUT="$ROOT/review-r2/A-golden-drift.replay.jsonl" bash review-r2/run-test.sh A-golden-drift ./internal/breakoutlane -count=1
