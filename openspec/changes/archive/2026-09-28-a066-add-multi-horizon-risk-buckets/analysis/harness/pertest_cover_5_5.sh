#!/usr/bin/env bash
# a066 5.5 — q_final 진입 경로(commitFresh·CommitRiskBucketAdmission·RevalidateQFinalAdmission)를
# 지나는 시험을 하나씩 돌려 journal 패키지 문장 커버리지 프로필을 시험마다 남김.
# branch_coverage_rows.py 의 per-test-dir 입력을 만드는 용도임.
#
# 사용: pertest_cover_5_5.sh <out-dir> [extra-go-test-tags]
#   extra tags 는 execgw 태그 시험에 더할 태그(예: a066_red_5_5)임.
set -euo pipefail

out=${1:?out dir}
extra=${2:-}
root=$(git rev-parse --show-toplevel)
cd "$root"
mkdir -p "$out"

# 시험 이름은 파일에서 유도함 — 손으로 고른 목록은 새 시험을 빠뜨림.
journal_files=(internal/journal/risk_bucket_admission_test.go internal/journal/risk_bucket_fill_test.go
	internal/journal/risk_bucket_issuance_test.go internal/journal/risk_bucket_owner_test.go
	internal/journal/risk_bucket_reference_time_test.go internal/journal/strategy_first_leg_atomic_test.go
	internal/journal/strategy_weekly_first_leg_test.go internal/journal/risk_bucket_entry_loss_lock_test.go
	internal/journal/strategy_dispatch_runtime_test.go internal/journal/strategy_first_leg_v26_migration_test.go)
execgw_plain=(internal/execgw/riskguardian_qfinal_test.go)
execgw_seams=(internal/execgw/riskguardian_account_base_testseam_test.go internal/execgw/strategy_gateway_account_base_test.go)
for f in internal/execgw/a066_*_test.go; do
	[ -f "$f" ] && execgw_plain+=("$f")
done

names() { grep -h -o -E '^func (Test[A-Za-z0-9_]+)\(t \*testing\.T\)' "$@" | sed -E 's/^func (Test[A-Za-z0-9_]+).*/\1/' | sort -u; }

jobs=()
while read -r n; do jobs+=("journal||./internal/journal|$n"); done < <(names "${journal_files[@]}")
while read -r n; do jobs+=("execgw|$extra|./internal/execgw|$n"); done < <(names "${execgw_plain[@]}")
seam_tags="tossos_testseams${extra:+,$extra}"
while read -r n; do jobs+=("execgw|$seam_tags|./internal/execgw|$n"); done < <(names "${execgw_seams[@]}")

printf '%s\n' "${jobs[@]}" > "$out/jobs.txt"
run_one() {
	IFS='|' read -r label tags pkg name <<<"$1"
	local tagflag=()
	[ -n "$tags" ] && tagflag=(-tags "$tags")
	if go test -count=1 "${tagflag[@]}" -run "^${name}\$" -coverpkg ./internal/journal \
		-coverprofile "$2/${name}.out" "$pkg" >"$2/${name}.log" 2>&1; then
		echo "PASS ${name}"
	else
		echo "FAIL ${name}"
	fi
}
export -f run_one
printf '%s\n' "${jobs[@]}" | xargs -P 8 -I{} bash -c 'run_one "$1" "$2"' _ {} "$out" > "$out/results.txt"
echo "jobs=$(wc -l <"$out/jobs.txt") pass=$(grep -c '^PASS' "$out/results.txt" || true) fail=$(grep -c '^FAIL' "$out/results.txt" || true)"
