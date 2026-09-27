# Function Logic Map: `strategyFamilyGate.installed`

- Source: `internal/app/engine/strategy_family_activation.go` (79-81)
- Function: `strategyFamilyGate.installed` in package `engine`
- File SHA-256: `230cc4c84bc3ff2bec2b98caaed10cdec7fd18e46181e3ce995899bbb0ee3492`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (post-edit).
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 뒤 (태스크 8.7.2, 2026-09-27).** 편집 전 번들은 커밋 `c1d1e295`(validate 는 `cb378a63`) 에 있다 — 이 파일은 GREEN 뒤 현재 소스의 AST 와 측정이다.

반환식 하나: `gate.rolledBack || gate.activation.Verified()`. **선언된 시장의 관문은 언제나 선다** — 검증된 활성화(승격 판정)든
`rolledBack`(넷 다 OFF)이든. 편집 전의 `&& len(gate.lanes) != 0` 는 적대 리뷰 A 가 연 잠재 fail-open(검증된 활성화 + 레인 nil →
기존 경로 → 사람이 끈 가족 선택)이라 지웠다. 레인이 없으면 `admit` 이 주인 없는 제안으로 REFUSED 를 낸다.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite (and the strategyrouter tagged suite for the two router bundles, same flags): `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 80:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
_(AST 분기 0 — 반환식 하나. 행동은 아래 Inputs 의 불리언 식이 전부다.)_



## Calls and live bindings

| Callee expression | Position |
|---|---|
| `gate.activation.Verified` | 80:28 |

## State mutations and fallbacks

없음. 읽기 전용 술어다.

## Safety conclusion

- 되돌림 경로 census(`TestTheRollbackPathOnlyReads`)가 이 함수의 호출을 `gate.activation.Verified` 하나로 못 박는다.
- 변이: M10(되돌림을 안 봄)·M11(8.7.1 의 레인 조건 복원)·M11b(되돌림에도 레인 조건) 전부 CAUGHT (`analysis/harness/a872_mutate.result.txt`).
- High-risk impact: yes.
