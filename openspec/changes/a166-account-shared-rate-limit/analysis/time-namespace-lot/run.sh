#!/usr/bin/env bash
# a166 time-namespace 로트 1단계 — 실제 time namespace 재현(네트워크 0, 브로커 0, 제품 코드 무변경).
# 사용: run.sh <harness-binary> <ext4 위 빈 상위 디렉터리>
# 새 time namespace 는 비특권 `unshare --user --map-current-user --time` 로 만든다.
# 운영과 다른 점: 운영 컨테이너는 user namespace 가 호스트와 같다. 여기서는 새 user namespace 가 함께 생긴다
# (uid 1000 은 그대로 매핑). 제품 코드는 user namespace 를 읽지 않는다(rg 'ns/user' 0건).
set -euo pipefail
H=$1
D=$2
U=(unshare --user --map-current-user --time)
echo "kernel=$(uname -r) host_time_ns=$(readlink /proc/self/ns/time)"
for s in s1 s2 s3 s4; do mkdir -m 700 "$D/$s"; done

echo "== S1 대조군: 호스트 ns 에서 쓰고 호스트 ns 에서 읽음"
"$H" write "$D/s1/root"
"$H" read "$D/s1/root"

echo "== S2 실측 배치 모사: 호스트 ns 에서 쓰고, 새 time ns(offset 0) 에서 읽음"
"$H" write "$D/s2/root"
"${U[@]}" "$H" read "$D/s2/root" || echo "(rc=$?)"
echo "-- S2 뒤 호스트 ns 에서 다시 읽음(상태 손상 여부)"
"$H" read "$D/s2/root" || echo "(rc=$?)"

echo "== S3 컨테이너 재시작 모사: 새 time ns X 에서 쓰고, 또 다른 새 time ns Y 에서 읽음"
"${U[@]}" "$H" write "$D/s3/root"
"${U[@]}" "$H" read "$D/s3/root" || echo "(rc=$?)"

echo "== S4 역방향: 새 time ns 에서 쓰고 호스트 ns 에서 읽음"
"${U[@]}" "$H" write "$D/s4/root"
"$H" read "$D/s4/root" || echo "(rc=$?)"

echo "== S5 offset 비영 namespace 는 Snapshot 이 거절하는가(실제 커널)"
"${U[@]}" --boottime 100 "$H" stamp "$D/unused" || echo "(rc=$?)"
"${U[@]}" --monotonic 100 "$H" stamp "$D/unused" || echo "(rc=$?)"
echo "-- S5 대조: --fork 없이 unshare 한 셸 안에서 자식은 time_for_children 으로 들어간다"
"${U[@]}" --boottime 100 sh -c 'cat /proc/self/timens_offsets; readlink /proc/self/ns/time /proc/self/ns/time_for_children'
