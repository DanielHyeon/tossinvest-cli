#!/usr/bin/env bash
# 자식 갈래만 직접 띄워 수명·종료 코드·stderr 를 잰다 (a202 RED/GREEN 공용).
# 사용: child-probe.sh <test-binary> <observe-seconds>
set -u
bin=$1; secs=$2
root=$(mktemp -d /tmp/a202-child-XXXXXX)
read -r dev ino < <(stat -c '%d %i' "$root")
echo "git HEAD (cwd): $(git rev-parse HEAD 2>/dev/null || echo n/a)"
echo "binary: $bin  sha256: $(sha256sum "$bin" | cut -c1-16)"
echo "root: $root dev=$dev ino=$ino"
start=$(date +%s.%N)
SHADOWSESSION_CHILD=1 SHADOWSESSION_ROOT="$root" SHADOWSESSION_ID="$dev $ino" \
  "$bin" -test.run='^TestChildLockAndCrash$' >"$root.out" 2>"$root.err" &
pid=$!
alive=yes
for i in $(seq 1 $((secs*10))); do
  if ! kill -0 $pid 2>/dev/null; then alive=no; break; fi
  sleep 0.1
done
if [ $alive = yes ]; then
  echo "after ${secs}s: child pid $pid STILL ALIVE"
  kill -9 $pid; wait $pid; rc=$?; echo "killed by probe, wait rc=$rc"
else
  wait $pid; rc=$?
  end=$(date +%s.%N)
  echo "child exited by itself after $(echo "$end - $start" | bc) s, exit code=$rc"
fi
echo "--- stdout"; cat "$root.out"
echo "--- stderr (first 15 lines)"; head -15 "$root.err"
rm -rf "$root" "$root.out" "$root.err"
