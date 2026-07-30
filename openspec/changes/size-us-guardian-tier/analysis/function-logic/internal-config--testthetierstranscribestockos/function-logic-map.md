# Function Logic Map: `TestTheTiersTranscribeStockOS`

- Source: `internal/config/limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `stockOSTiers` | StockOS `risk_profiles.py`의 네 프로파일 | 손으로 전사한 리터럴 | 레지스트리와 다르면 `t.Errorf` |
| `tossOSMeasuredTiers` | StockOS 대응이 없는 행 | design D1 + measurements.md M49 | 레지스트리와 다르면 `t.Errorf` |
| `GuardianTiers()` | 두 코퍼스의 합집합 | 현재 HEAD 레지스트리 | 개수 불일치는 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 등록 티어 수 != 두 코퍼스 합 | 없음 | `t.Fatalf` (즉시 중단) | 자기 자신 |
| B2 | 등록 티어 순회 | 없음 | 없음 | 자기 자신 |
| B3 | 라벨이 빈 문자열 | 없음 | `t.Errorf` | 자기 자신 |
| B4 | 전사 코퍼스에 없는 ID | 없음 | B5로 위임 | 자기 자신 |
| B5 | 실측 코퍼스에 있음 | 없음 | `continue` (유도 테스트가 받는다) | `TestTheUSTierMatchesItsRecordedDerivation` |
| B6 | 전사 값 불일치 | 없음 | `t.Errorf` | 자기 자신 |
| B7 | 실측 코퍼스 순회 | 없음 | 없음 | 자기 자신 |
| B8 | 실측 티어가 미등록 | 없음 | `t.Errorf` | 자기 자신 |
| B9 | 실측 값 불일치 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GuardianTiers` | 레지스트리 사본 | 오류 없음 | CodeGraph + AST |
| `GuardianTierByID` | 실측 티어 조회 | 미등록은 `ok=false` | CodeGraph + AST |

## State mutations and fallbacks

- 상태 변경 없음. 순수 비교다.
- 이 change의 편집은 코퍼스를 **둘로 쪼갠 것**이다. 이전에는 map 하나였고 등록 티어가 거기 없으면 곧바로 실패했다. 쪼갠 뒤에도 그 canary는 살아 있다(B4→B5→`t.Errorf`): 두 코퍼스 **어디에도** 없는 티어는 여전히 실패한다.
- fallback 없음. `continue`(B5)는 판정을 건너뛰는 것이 아니라 B7 루프로 넘기는 것이고, B7이 같은 값을 다시 검사한다.

## Safety conclusion

- Safe edit boundary: 두 코퍼스 리터럴. 레지스트리를 고치면 여기도 손으로 고쳐야 하고, 그것이 이 테스트의 목적이다 — 레지스트리에서 유도하면 아무것도 검사하지 않는다.
- High-risk impact: yes — 티어 추가는 상한을 올린다. 이 테스트가 "근거 없이 추가된 행"을 막는 첫 관문이다.
