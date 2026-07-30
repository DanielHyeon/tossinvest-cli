# Function Logic Map: `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire`

- Source: `cmd/tossctl/soak_test.go`
- Function: `cmd/tossctl/soak_test.go:TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire`
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
| B1 | `if` (line 682) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `if` (line 685) — `if err := rec.Append(verifylive.Entry{` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `if` (line 698) — `if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest"); err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `if` (line 702) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `range` (line 705) — `for _, e := range a.Endpoints {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B6 | `if` (line 706) — `if strings.Contains(e, "conditional-orders") {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B7 | `range` (line 710) — `for _, p := range a.SupervisedBy {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B8 | `if` (line 711) — `if strings.HasPrefix(p.Endpoint, "GET ") {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `testenv.Isolate` | ast.json calls (line 675) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `seedQualifyingRecord` | ast.json calls (line 676) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `filepath.Join` | ast.json calls (line 676) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `verifylive.RecordFileName` | ast.json calls (line 677) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `Add` | ast.json calls (line 678) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `UTC` | ast.json calls (line 678) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `time.Now` | ast.json calls (line 678) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `seedSupervisedRecord` | ast.json calls (line 679) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attest.Mask` | ast.json calls (line 679) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `verifylive.OpenRecorder` | ast.json calls (line 681) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `t.Fatalf` | ast.json calls (line 683) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `rec.Append` | ast.json calls (line 685) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `rec.Close` | ast.json calls (line 696) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `runCLI` | ast.json calls (line 698) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attest.Load` | ast.json calls (line 701) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Contains` | ast.json calls (line 706) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `t.Errorf` | ast.json calls (line 707) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.HasPrefix` | ast.json calls (line 711) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 테스트 지역 상태만 변경한다. 프로덕션 상태·브로커·계좌를 건드리지 않는다.

## Safety conclusion

- Safe edit boundary: 테스트 함수이므로 프로덕션 동작 경계가 없다.
- High-risk impact: no — 테스트 코드. httptest/fake 브로커만 쓰며 실계좌에 접근하지 않는다.
