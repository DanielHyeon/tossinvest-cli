# Function Logic Map: `TestEveryCardShowsWhatChangesAndWhen`

- Source: `internal/console/settings_card_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/settings_card_test.go` = `d7c163ee1591…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-26-a174-settings-startup-readiness-and-presets` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/settings_card_test.go` = `d7c163ee1591…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 일반 설정 카드 | save 전에 `card-preview` 설명 존재 | 각 탭 렌더 | 누락 시 시험 실패 |
| Guardian 한도 카드 | 프리셋 비교표가 저장 전 값·방향을 표시 | `/settings/daily`의 preset matrix | 표 부재 시 시험 실패 |
| 반영 시점 | 상시 탭 adoption·gate의 펼쳐진 한 줄 | 기존 `effect-timing` 계약 | 누락·접힘 시 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | 탭·카드 순회 | 없음 | 각 카드 검사 | 전체 탭 렌더 |
| B3 | 일반 카드의 preview 부재 | 없음 | `t.Errorf` | 일반 preview 유지 |
| B4 | limits 카드 | matrix 계약 검사 후 일반 preview 검사 제외 | 없음 | 프리셋 표 존재 |
| B5-B6 | effect-timing marker·class 부재 | 없음 | `t.Errorf` | never-folded timing |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fullSettingsHarness`, `h.get` | 모든 설정 탭의 GET 렌더 생성 | 시험 fixture, 재시도 없음 | AST calls |
| `cardsOn`, `body` | 카드 경계 추출 | 순수 문자열 추출 | AST calls |
| `strings.Contains`, `strings.Index` | 미리보기·반영 시점 표시 검사 | 부재면 시험 실패 | AST calls |

## State mutations and fallbacks

- 시험은 렌더 HTML만 읽으며 form submit을 수행하지 않음.
- Guardian 카드는 button 내부 속성이 아니라 표 자체를 preview 근거로 삼음.

## Safety conclusion

- Safe edit boundary: preview의 구조별 시험 분기만 수정함.
- High-risk impact: no. 저장·주문·토글 동작은 바꾸지 않음.
