#!/usr/bin/env bash
# a164 FLM 재추출 한 판: AST 재추출 → 시험별 coverage(이미 뜬 프로필 디렉터리) → map 생성 → 재추출 절 덧붙임.
set -euo pipefail
R=/mnt/D/Axipient/workspace/TossOS-a164-impl
S=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/1711ae2d-4cc3-48b0-97b1-c0fa1bb7c4da/scratchpad
FL=$R/openspec/changes/a164-verify-reconcile-closes-broker-expired-conditional/analysis/function-logic
PROFILES=$1; LABEL=$2
cd "$R"
for d in "$FL"/*/; do
  f=$(python3 -c "import json;a=json.load(open('$d/ast.json'));print(a.get('file',''))" 2>/dev/null || true)
  name=$(basename "$d")
  if [ -z "$f" ]; then f=$(python3 -c "import json;print(json.load(open('$S/old-$name.ast.json'))['file'])"); fi
  fn=$(python3 -c "import json;a=json.load(open('$S/old-$name.ast.json'));print((a.get('receiver')+'.' if a.get('receiver') else '')+a['function'])")
  GOFLAGS=-trimpath GOCACHE=$S/gocache go run ./tools/logic-map --file "$f" --func "$fn" > "$d/ast.json"
  python3 tools/logic-map/risk_pattern_report.py "$f" --output "$d/risk-pattern-report.md" > /dev/null
done
python3 -I $S/tools/btm_cov.py "$PROFILES" "$FL" > $S/btm-cov-current.json
python3 -I $S/tools/gen_flm.py $S/btm-cov-current.json $S/tools/notes.json "$LABEL" > /dev/null
python3 -I $S/tools/append_reextract.py "$S" "$FL"
