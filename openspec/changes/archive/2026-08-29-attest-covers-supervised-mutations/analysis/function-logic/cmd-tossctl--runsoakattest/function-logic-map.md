# Function Logic Map: `runSoakAttest`

- Source: `cmd/tossctl/soak.go`
- Function: `cmd/tossctl/soak.go:runSoakAttest`
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
| B1 | `if` (line 472) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `if` (line 475) — `if opts.validity > 0 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `if` (line 480) — `if base := soakSurveyedBase(root); base != "" {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `if` (line 486) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `if` (line 491) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B6 | `if` (line 497) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B7 | `if` (line 500) — `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B8 | `if` (line 503) — `if err := attest.Save(path, attestation); err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B9 | `range` (line 516) — `for _, p := range attestation.SupervisedBy {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B10 | `if` (line 518) — `if market == "" {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B11 | `if` (line 526) — `if len(missing) > 0 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B12 | `range` (line 529) — `for _, e := range missing {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loadSoakSummary` | ast.json calls (line 471) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `soakSurveyedBase` | ast.json calls (line 480) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.TrimSpace` | ast.json calls (line 481) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `UTC` | ast.json calls (line 484) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `time.Now` | ast.json calls (line 484) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `supervisedProofs` | ast.json calls (line 485) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `soak.BuildAttestation` | ast.json calls (line 490) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `resolveSoakAttestationPath` | ast.json calls (line 496) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `os.MkdirAll` | ast.json calls (line 500) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `filepath.Dir` | ast.json calls (line 500) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `fmt.Errorf` | ast.json calls (line 501) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attest.Save` | ast.json calls (line 503) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `cmd.OutOrStdout` | ast.json calls (line 507) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `fmt.Fprintf` | ast.json calls (line 508) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attest.Mask` | ast.json calls (line 510) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attestation.ExpiresAt.Format` | ast.json calls (line 512) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Join` | ast.json calls (line 513) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `Format` | ast.json calls (line 522) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `p.At.UTC` | ast.json calls (line 522) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attestation.MissingEndpoints` | ast.json calls (line 525) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `soak.LiveOnlyEndpoints` | ast.json calls (line 525) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `len` | ast.json calls (line 526) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `fmt.Fprintln` | ast.json calls (line 527) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 테스트 지역 상태만 변경한다. 프로덕션 상태·브로커·계좌를 건드리지 않는다.

## Safety conclusion

- Safe edit boundary: 테스트 함수이므로 프로덕션 동작 경계가 없다.
- High-risk impact: no — 테스트 코드. httptest/fake 브로커만 쓰며 실계좌에 접근하지 않는다.
