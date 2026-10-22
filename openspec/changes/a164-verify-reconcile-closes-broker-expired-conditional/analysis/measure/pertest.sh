#!/usr/bin/env bash
# 시험 하나씩 coverprofile 을 떠서 분기 본문 도달을 잰다(BTM 근거).
set -u
R=/mnt/D/Axipient/workspace/TossOS-a164-impl
O=${PERTEST_OUT:-/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/1711ae2d-4cc3-48b0-97b1-c0fa1bb7c4da/scratchpad/pertest}
cd "$R"
run() { # pkg tags files...
  local pkg=$1 tags=$2; shift 2
  for f in "$@"; do
    for t in $(/usr/bin/grep -ho '^func Test[A-Za-z0-9_]*' $f | cut -c6-); do
      go test $tags "$pkg" -count=1 -run "^${t}\$" -coverprofile="$O/$(echo $pkg | tr / _)__$t.out" > "$O/$t.log" 2>&1 || echo "FAIL $pkg $t" >> "$O/failures.txt"
    done
  done
}
run ./internal/verifylive/ "" internal/verifylive/reconcile_*_test.go internal/verifylive/record_filled_test.go internal/verifylive/record_replay_test.go internal/verifylive/report_test.go
run ./internal/official/ "" internal/official/reconcile_*_test.go
run ./cmd/tossctl/ "-tags tossos_testseams" cmd/tossctl/verify_reconcile*_test.go
echo done > "$O/DONE"
