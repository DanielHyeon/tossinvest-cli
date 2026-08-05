# Function Logic Map: `newDriverHarness`

- Source: `internal/app/engine/reconcileloop_test.go` (lines 121–186)
- AST evidence: `ast.json` (`source_sha256: d53e58de88a1444c84d97750f68023fe8878894fc04eb5bc2026ade205132eae`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드.

## What it does

드라이버 테스트 하네스. a085는 InstrumentNames를 하네스와 옵션에 배선한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (129) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (133) `if` — if err := j.SetApplyHooks(journal.ApplyHooks | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (177) `if` — if mutate != nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (181) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 122, 'column': 2}, 'text': 't.Helper'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 123, 'column': 9}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 124, 'column': 12}, 'text': 'journal.Open'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 124, 'column': 25}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 125, 'column': 13}, 'text': 'filepath.Join'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 125, 'column': 27}, 'text': 't.TempDir'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 127, 'column': 13}, 'text': 'journal.FixedFSProber'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 130, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 132, 'column': 25}, 'text': 'j.Close'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 132, 'column': 2}, 'text': 't.Cleanup'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 133, 'column': 12}, 'text': 'j.SetApplyHooks'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 136, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
