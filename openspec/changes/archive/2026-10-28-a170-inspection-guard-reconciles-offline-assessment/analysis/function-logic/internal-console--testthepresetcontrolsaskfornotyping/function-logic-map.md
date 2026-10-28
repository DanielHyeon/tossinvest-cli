# Function Logic Map: `TestThePresetControlsAskForNoTyping`

- Source: `internal/console/settings_limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/settings_limits_test.go` = `458b76e24765…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-26-a174-settings-startup-readiness-and-presets` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/settings_limits_test.go` = `458b76e24765…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 프리셋 form | 등록 tier마다 1개, hidden `tier`만 입력 | `/settings/daily` 렌더 | 누락·자유 입력·추가 확인이면 시험 실패 |
| 프리셋 열 | 다섯 한도와 실제 방향을 같은 tier 열에 표시 | `LimitTierCards().Preview` | 값·방향·전환 문구 불일치면 시험 실패 |
| 클릭 마찰 | `onsubmit` 확인·추가 입력 없음 | 렌더 HTML | 발견하면 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 첫 프리셋 form 시작점 없음 | 즉시 중단 | `t.Fatal` | 프리셋 control 존재 |
| B2 | form 닫힘 없음 | 즉시 중단 | `t.Fatal` | form 경계 |
| B3-B4 | 금지 입력 순회·발견 | 없음 | `t.Errorf` | one-click 유지 |
| B5 | inline `onsubmit` 존재 | 없음 | `t.Error` | 확인 단계 없음 |
| B6-B7 | 가짜 미리보기·가짜 방향 속성 존재 | 없음 | `t.Error` | form 속성 제거 |
| B8 이후 | tier별 표 열·전환 문구 불일치 | 없음 | `t.Errorf` | 실제 값·방향·문구 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `limitsHarness`, `h.page` | 빈 자격증명 시험 콘솔 렌더 | HTTP 시험 응답만 사용, 재시도 없음 | AST calls |
| `LimitTierCards` | 프리셋 표의 정본 Preview 수집 | 순수 뷰 계산, 오류 반환 없음 | `readiness.go` |
| `strings.Index`, `strings.Contains` | form·표 열 범위와 렌더 사실 확인 | 찾지 못하면 시험 실패 | AST calls |

## State mutations and fallbacks

- 시험은 GET 렌더만 읽고 저장·주문·토글을 호출하지 않음.
- failure는 시험 실패로만 표현하며 제품 상태를 바꾸지 않음.

## Safety conclusion

- Safe edit boundary: 프리셋 렌더 계약 단언만 수정함.
- High-risk impact: no. POST 경로·필드·hidden 값·클릭 수는 시험으로 보존함.
