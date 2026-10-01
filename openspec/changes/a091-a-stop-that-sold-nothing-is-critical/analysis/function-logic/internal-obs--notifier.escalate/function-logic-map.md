# Function Logic Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go` (`425`–`446`)
- Qualified: `Notifier.escalate`
- AST evidence: `ast.json` (`source_sha256` 408579d504089072…) — 편집 뒤 `540aebe6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 4 · 반환 2

**편집.** **a091 편집(구현 로트, 5판 R3-3 · Manager 포함 승인)** — 두 로그 줄(실패 · 승격)에서 `FieldAccount` 원문을 뺐다. 판정 · 반환 · 원장 호출 무변경(변이 M13 · M13b · M13e CAUGHT).

**역할.** critical 전달 실패(또는 기록 실패)의 운영 모드 승격을 원장에 남긴다. 통지하지 않는다(전송 수단이 방금 실패).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.AccountRef` | 계좌 참조(생산 = 계좌번호, `interlock.go:680-684`) | 배선 | 빈 값 → 승격 없음(B1) |
| `e.Type` | 촉발 사건 종류 | 호출자 | 로그 필드 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `540aebe6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:426` `if n.Journal == nil \|\| strings.TrimSpace(n.AccountRef) == "" {` | 예 |
| B2 | switch | `:431` `switch {` | — |
| B3 | case | `:432` `case err != nil && n.Log != nil:` | 예 |
| B4 | case | `:438` `case changed && n.Log != nil:` | 예 |

Exact AST return positions: `427:3`, `445:2`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `strings.TrimSpace` | `:426` | 빈 계좌 판정 | 순수 |
| `n.Journal.EscalateOperatingMode` | `:429` | ENTRY_BLOCKED 승격 | 원장 트랜잭션(`busy_timeout` 5s · 연결 풀 대기 기한 없음), 원격 0 |
| `n.Log.Error` | `:434` | 승격 실패 로그 | 로그 한 줄 — 계좌 필드 없음(a091), 오류는 `MaskAccount` |
| `MaskAccount` | `:434` | 오류 속 계좌 가림 | 순수 |
| `string` | `:435` | — | 순수 |
| `n.Log.Warn` | `:440` | 승격 로그 | 로그 한 줄 — 계좌 필드 없음(a091) |

## State mutations and fallbacks

운영 모드 행(원장).

## Safety conclusion

- 호출자 셋(`notifier.go:238` · `:261` · `record_only.go:157`)이 공유하는 함수다 — a091 의 편집은 로그 필드 하나를 빼는 것뿐이고 판정 · 반환은 같다. 계좌는 한 알림기 하나이므로 필드를 빼도 줄의 뜻은 같다. High-risk: 아니오(로그), 단 모드 승격 경로에 있다.
