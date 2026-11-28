#!/usr/bin/env bash
# a210 D3 측정 하네스 — 패키지마다 새 가짜 HOME 상자에서 시험을 돌려 상자 안에 생긴 파일을 셈.
# 사용: bash measure.sh <출력 디렉터리> <빌드 태그 또는 ""> <패키지>...
# 선택: A210_RUN=<정규식> 이면 go test -run 으로 시험을 좁힘(누수 시험을 가려낼 때만).
# 운영 파일을 건드리지 않기 위해 env -i 로 HOME 만 상자로 주고 XDG_* · TOSSOS_* 는 넣지 않음.
# GOCACHE · GOMODCACHE · GOPATH 는 실제 경로로 명시해 빌드 캐시가 상자에 생기지 않게 함.
# GOTOOLCHAIN=local · GOPROXY=off 는 측정 중 네트워크 다운로드를 막기 위함.
# -timeout 45m 은 Makefile test-seams 와 같은 값 — go test 기본 10 분으로는 journal 이 끝나기 전에 죽음.
# files_tossos 는 go 툴체인 자신이 쓰는 .config/go/telemetry/ 를 뺀 파일 수 — 시험 코드가 만든 파일만 셈.
set -euo pipefail
out=$1; tags=$2; shift 2
repo=$(git rev-parse --show-toplevel)
gocache=$(go env GOCACHE); gomod=$(go env GOMODCACHE); gopath=$(go env GOPATH)
parent=${A210_BOX_PARENT:?A210_BOX_PARENT 가 필요함}
mkdir -p "$out"
echo "# head=$(git -C "$repo" rev-parse HEAD) dirty=$(git -C "$repo" status --porcelain -- '*.go' | wc -l) tags=${tags:-none} at=$(date -u +%FT%TZ)" >> "$out/summary.txt"
for pkg in "$@"; do
  name=$(printf '%s' "$pkg" | tr / _)${tags:+.$tags}
  box=$(mktemp -d "$parent/a210-box-$name-XXXXXX")
  # 측정 전 상자 목록(빈 상자여야 함)
  find "$box" -mindepth 1 -printf '%P\t%s\n' | sort > "$out/$name.before"
  set +e
  (cd "$repo" && env -i HOME="$box" PATH="$PATH" GOCACHE="$gocache" GOMODCACHE="$gomod" GOPATH="$gopath" \
     GOFLAGS= GOTOOLCHAIN=local GOPROXY=off \
     go test -count=1 -timeout 45m ${tags:+-tags "$tags"} ${A210_RUN:+-run "$A210_RUN"} "./$pkg") > "$out/$name.log" 2>&1
  rc=$?
  set -e
  # 측정 뒤 상자 목록 — 경로와 크기만 남김
  find "$box" -mindepth 1 -printf '%P\t%s\n' | sort > "$out/$name.after"
  files=$(find "$box" -mindepth 1 -type f | wc -l)
  files_tossos=$(find "$box" -mindepth 1 -type f -not -path "$box/.config/go/telemetry/*" | wc -l)
  printf '%s\ttags=%s\trc=%s\tfiles=%s\tfiles_tossos=%s\n' "$pkg" "${tags:-none}" "$rc" "$files" "$files_tossos" | tee -a "$out/summary.txt"
  # 상자에 audit 가 생겼으면 action · setting 이름만 셈(값은 남기지 않음 — 시험 자료여도 같은 규칙)
  if [ -f "$box/.local/share/tossos/audit.log" ]; then
    python3 -I -c 'import json,sys,collections
c=collections.Counter((r["action"],r["setting"]) for r in map(json.loads,open(sys.argv[1],"rb")))
[print(n,a,s,sep="\t") for (a,s),n in sorted(c.items())]' "$box/.local/share/tossos/audit.log" > "$out/$name.audit-settings"
  fi
  # 정리는 이 하네스가 만든 상자 한 개만, 정확한 경로로 함
  rm -rf -- "$box"
done
