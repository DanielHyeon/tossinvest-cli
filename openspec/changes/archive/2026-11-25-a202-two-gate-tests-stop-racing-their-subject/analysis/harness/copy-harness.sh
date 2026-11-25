#!/usr/bin/env bash
# a202 사본 하네스 — 실제 worktree 는 건드리지 않고, 사본에만 변이를 넣어 시험을 돌림.
# 사용: A202_SCRATCH=<scratch dir> copy-harness.sh <source-worktree> <variant> <go test args...>
# variant: none | prod-sleep | prod-noclose | ss-parent-delay
set -euo pipefail
src=$(cd "$1" && pwd -P); variant=$2; shift 2
copy=$(mktemp -d "${A202_SCRATCH:?A202_SCRATCH 필요}/a202-copy-$$-XXXX")
trap 'rm -rf "$copy"' EXIT
rsync -a --exclude=.git --exclude=openspec "$src"/ "$copy"/
git -C "$copy" init -q
git -C "$copy" add -A
git -C "$copy" -c user.name=a202-harness -c user.email=a202@invalid commit -qm copy
top=$(git -C "$copy" rev-parse --show-toplevel)
# 사본임을 단언함 — 실제 저장소로 떨어지면 즉시 멈춤
[ "$top" = "$copy" ] || { echo "ABORT: toplevel=$top != copy=$copy"; exit 90; }
[ "$top" != "$src" ] || { echo "ABORT: copy is source"; exit 91; }
echo "source=$src source-HEAD=$(git -C "$src" rev-parse HEAD)"
echo "copy=$copy toplevel=$top copy-start-sha=$(git -C "$copy" rev-parse HEAD)"
echo "variant=$variant"
python3 - "$copy" "$variant" <<'PY'
import sys, pathlib
copy, variant = sys.argv[1], sys.argv[2]
def sub(rel, old, new):
    p = pathlib.Path(copy, rel); s = p.read_text()
    n = s.count(old)
    if n != 1:
        sys.exit(f"ABORT: {rel}: anchor count {n} != 1")
    p.write_text(s.replace(old, new))
prod = "cmd/tossctl/httpapi_strategy_attach.go"
anchor = "\ta.mu.Unlock()\n\tcloseEvictedStrategyReader(evicted)\n"
if variant == "none":
    pass
elif variant == "prod-sleep":
    # unlock 과 close 사이의 창을 50ms 로 벌림 — 부하 속 스케줄링 지연의 재현
    sub(prod, anchor, "\ta.mu.Unlock()\n\ttime.Sleep(50 * time.Millisecond)\n\tcloseEvictedStrategyReader(evicted)\n")
elif variant == "prod-noclose":
    # 양성 대조: 닫기 호출을 지움 — 시험이 여전히 닫힘을 재는지 확인
    sub(prod, anchor, "\ta.mu.Unlock()\n\t_ = evicted\n")
elif variant == "ss-parent-delay":
    # 부모가 READY 를 받은 뒤 Begin 까지 300ms 늦음 — 부하 속 부모 지연의 재현
    t = "internal/shadowsession/session_linux_test.go"
    s = pathlib.Path(copy, t).read_text()
    if '\t"time"\n' not in s:
        sub(t, '\t"testing"\n', '\t"testing"\n\t"time"\n')
    sub(t, '\tif _, err := Begin(context.Background(), r, "2026-11-06"); !errors.Is(err, ErrBusy) {\n',
           '\ttime.Sleep(300 * time.Millisecond)\n\tif _, err := Begin(context.Background(), r, "2026-11-06"); !errors.Is(err, ErrBusy) {\n')
else:
    sys.exit(f"ABORT: unknown variant {variant}")
PY
echo "--- copy diff (variant)"
git -C "$copy" --no-pager diff
echo "--- go test $*"
cd "$copy"
set +e
go test "$@"
rc=$?
set -e
echo "go test rc=$rc"
