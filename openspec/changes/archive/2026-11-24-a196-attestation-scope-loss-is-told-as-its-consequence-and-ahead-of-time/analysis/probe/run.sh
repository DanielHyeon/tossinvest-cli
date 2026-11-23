#!/usr/bin/env bash
# a196 1단계 overlay 프로브 — 저장소 파일을 편집하지 않고 패키지에 가상 시험 파일을 얹어 잰다.
# 사용: bash openspec/changes/<change>/analysis/probe/run.sh  (저장소 루트에서)
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(git -C "$here" rev-parse --show-toplevel)"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
cp "$here/engine_probe_test.go.txt" "$tmp/engine_probe_test.go"
cp "$here/soak_probe_test.go.txt" "$tmp/soak_probe_test.go"
printf '{"Replace":{"%s/internal/app/engine/zz_a196_probe_internal_test.go":"%s/engine_probe_test.go","%s/internal/soak/zz_a196_probe_internal_test.go":"%s/soak_probe_test.go"}}' \
  "$root" "$tmp" "$root" "$tmp" > "$tmp/overlay.json"
cd "$root"
export GOFLAGS=-trimpath
go test -count=1 -overlay "$tmp/overlay.json" -run '^TestZZA196Probe$' -v ./internal/app/engine
go test -count=1 -overlay "$tmp/overlay.json" -run '^TestZZA196SoakProbe$' -v ./internal/soak
test -z "$(git status --porcelain -- internal)" || { echo "probe touched the tree" >&2; exit 1; }
