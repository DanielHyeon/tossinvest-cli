# Function Logic Map: `coolAbsent`

- Source: `internal/candidate/scan.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

본문이 바뀐 기존 함수다. 세 번째 인자가 `panel`에서 `heard`로 바뀌었다 — 일정이 넘긴 원천을 '없어진 원천'으로 읽어 냉각을 허용하던 §5 리뷰 P1-1의 수리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `covered` | 이번 스캔에서 어떤 응답 원천이든 올린 심볼 집합 | `Collect` | 여기 있으면 냉각 대상이 아니다 |
| `responded` | 응답했고 **행을 나른** 원천 | `Collect`(빈 200은 제외) | 지지자 중 하나라도 여기 없으면 냉각하지 않는다 |
| `heard` | 이번 pass에서 들을 수 있었던 원천 = 패널 + not-asked | `Collect` | 지지자가 전부 여기 밖이면(설정에서 사라진 원천) 냉각을 막지 않는다 |
| `at` | 냉각 시각 | 호출자 인자 | `Store.Cool`이 마지막 관측보다 이른 시각을 거부한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `store.Candidates` 에러 | 없음 | `0, err` | 직접 테스트 없음 |
| B2 | 저장된 후보 순회 | — | — | `TestASymbolThatLeavesEveryListCools` |
| B3 | 다른 시장이거나 활성이 아니거나 이번에 올라옴 | 없음 | `continue` | `TestASymbolThatLeavesEveryListCools` |
| B4 | `!coverageAnswered(c.Sources, responded, heard)` | 없음 | `continue` — 증거 없이는 냉각하지 않는다 | `TestAScanDoesNotCoolASymbolItDidNotLookFor`, `TestOneSurvivingSupporterIsNotEnoughToCool`, `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` |
| B5 | `store.Cool` 에러 | 지금까지의 `cooled` | `cooled, err` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `store.Candidates` | 시장의 저장된 후보 전량 | 실패 시 즉시 반환. §2가 남긴 미해결: 시장·상태를 SQL에서 걸러야 한다 | `review.md` §2 '수용하지 않은 것' |
| `coverageAnswered` | 증거 판정 | 순수 함수 | 같은 파일 |
| `store.Cool` | `cooled_at` 기록 | 역행 시각은 저장소가 거부 | `store.go:Cool` |

## State mutations and fallbacks

- 저장소 write는 `Store.Cool` 하나 — `candidates.cooled_at`만 건드리고 `first_seen_at`·`first_price`·`first_rank`는 손대지 않는다.
- 부분 실패: 한 심볼의 `Cool`이 실패하면 그때까지 냉각된 수와 함께 에러가 올라가고, 나머지 심볼은 냉각되지 않는다. 냉각은 삭제가 아니므로 남은 것은 다음 pass가 다시 만난다.
- 냉각은 삭제하지 않는다. TTL 안의 재진입이 조인할 행이 남아 있어야 D1의 우회가 막힌다.

## Safety conclusion

- Safe edit boundary: `heard`를 `responded`나 `seen`으로 되돌리거나, `coverageAnswered`의 `present > 0` 요구를 없애는 것은 금지
- High-risk impact: no — 읽기 전용 발굴의 수명 관리다. 주문 경로에 닿지 않는다. 다만 잘못된 냉각은 30분 뒤 만료로 `first_seen_at`을 지우고, 그것이 이 패키지가 하는 유일한 주장이다. 파괴 방향으로만 위험하다.
