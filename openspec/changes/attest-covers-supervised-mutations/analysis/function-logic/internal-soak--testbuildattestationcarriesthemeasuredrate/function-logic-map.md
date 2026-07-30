# Function Logic Map: `TestBuildAttestationCarriesTheMeasuredRate`

- Source: `internal/soak/attest_test.go`
- Function: `internal/soak/attest_test.go:TestBuildAttestationCarriesTheMeasuredRate`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`

이 change의 테스트. 아래 분기·호출은 ast.json에서 읽었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 테스트가 만든 값 | 이 파일 | t.Fatalf/t.Errorf로 실패를 보고한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 248) — `for i := range cycles {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `if` (line 253) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `if` (line 256) — `if a.RateLimitPerSecond <= 0 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `if` (line 259) — `if !strings.Contains(a.Notes, "note") {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `if` (line 262) — `if a.VerifiedBy != "tester" {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `threeCleanDays` | ast.json calls (line 247) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `cycles.StartedAt.Add` | ast.json calls (line 250) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `soak.BuildAttestation` | ast.json calls (line 252) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `soak.Summarize` | ast.json calls (line 252) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `criteria` | ast.json calls (line 252) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `soakStart.AddDate` | ast.json calls (line 252) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `t.Fatalf` | ast.json calls (line 254) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `t.Errorf` | ast.json calls (line 257) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Contains` | ast.json calls (line 259) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 테스트 지역 상태만 변경한다. 프로덕션 상태·브로커·계좌를 건드리지 않는다.

## Safety conclusion

- Safe edit boundary: 테스트 함수이므로 프로덕션 동작 경계가 없다.
- High-risk impact: no — 테스트 코드. httptest/fake 브로커만 쓰며 실계좌에 접근하지 않는다.
