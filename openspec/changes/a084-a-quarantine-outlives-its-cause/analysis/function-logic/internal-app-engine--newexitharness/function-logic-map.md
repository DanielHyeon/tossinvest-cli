# Function Logic Map: `newExitHarness`

- Source: `internal/app/engine/exitloop_test.go` (lines 177–246)
- AST evidence: `ast.json` (`source_sha256: df13bc9e9772d9d2806fe1a9dc1393191b1382e241b78eb0b933e46dde218541`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 하네스. 주문·손절·원장 판정 경로가 아니다.

## What it does

exit loop 테스트 하네스. a084는 journal 파일 경로를 하네스에 노출한다 — 각인 없는 격리 행은 저널 API가 만들 수 없고 만들어서도 안 되므로, 테스트가 파일을 직접 열어 v29 시절의 행 모양을 재현한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (186) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (190) `if` — if err := j.SetApplyHooks(journal.ApplyHooks | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (205) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (235) `if` — if mutate != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (239) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 178, 'column': 2}, 'text': 't.Helper'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 179, 'column': 9}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 180, 'column': 12}, 'text': 'filepath.Join'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 180, 'column': 26}, 'text': 't.TempDir'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 181, 'column': 12}, 'text': 'journal.Open'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 181, 'column': 25}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 184, 'column': 13}, 'text': 'journal.FixedFSProber'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 187, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 189, 'column': 25}, 'text': 'j.Close'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 189, 'column': 2}, 'text': 't.Cleanup'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 190, 'column': 12}, 'text': 'j.SetApplyHooks'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 193, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 198, 'column': 10}, 'text': 'execgw.NewEntryGate'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 201, 'column': 19}, 'text': 'execgw.NewRiskGuardian'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만.

## Safety conclusion

- Safe edit boundary: journal 파일 경로를 구조체에 기록하는 것뿐. 하네스가 조립하는 엔진 배선은 그대로다.
- High-risk impact: no — 테스트 하네스다.
