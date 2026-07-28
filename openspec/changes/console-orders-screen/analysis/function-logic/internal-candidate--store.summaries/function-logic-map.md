# Function Logic Map: `Store.Summaries`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. 후보 하나를 읽는 데 세 번 묻지 않으려고 존재한다 — 효율만이 아니라, 스캔이 두 컬럼의 writer이고 각각을 **한 번만** 쓰려면 어느 후보가 아직 비어 있는지 알아야 하기 때문이다. 후보마다 묻는 배선은 tick마다 심볼당 IMMEDIATE 트랜잭션 2개이고, 그 테이블이 바로 D16이 원장의 파일시스템을 채울 수 있다고 지목한 테이블이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `at` | 상태 판정 순간 | 호출자 인자 | `Candidates`와 같은 계약 — 상태는 스윕이 아니라 읽기에서 유도한다 |
| 16개 컬럼 | 한 질의 | `candidates` | 십진값·순위는 NULL 가능. 십진값은 TEXT이고 NULL은 '원천이 그 필드를 나르지 않았다'이다. 0 기본값을 가진 수치 컬럼이었다면 '원천이 0이라고 말했다'와 '아무도 재지 않았다'가 합쳐지고, D10은 그 차이가 veto까지 살아남는 것 위에 서 있다. |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 질의 에러 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B2 | 행 순회 | — | — | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` |
| B3 | `rows.Scan` 에러 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B4 | `first_seen_at` 파싱 실패 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B5 | `last_seen_at` 파싱 실패 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B6 | `cooled.Valid` | `c.CooledAt` 설정 | — | `TestACycleSweepsTheSummariesItsOwnRetentionHasOrphaned` |
| B7 | `cooled_at` 파싱 실패 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B8 | `decodeBaseline` 에러 | 없음 | `nil, parseErr` | 직접 테스트 없음 |
| B9 | `decodeFirstRank` 에러 | 없음 | `nil, parseErr` | 직접 테스트 없음 |
| B10 | `rows.Err()` | 없음 | `nil, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryContext` | 16컬럼 전량, `ORDER BY market, symbol` | 단일 문장 | ast.json calls |
| `parseStamp` | 세 타임스탬프 | 실패는 즉시 에러 — 조용히 제로 시각으로 흡수하지 않는다 | `stamp`는 고정폭(`2006-01-02T15:04:05.000000000Z07:00`)이다. TEXT 컬럼이라 모든 `<`·`>=`·`ORDER BY`가 BINARY 문자열 비교이고, `time.RFC3339Nano`는 소수부 뒤 0을 떼어내 텍스트 순서가 시간 순서가 아니게 만든다 — §1 리뷰가 실행으로 재현한 P0다(`TestTheStoredInstantOrdersLikeTime`). |
| `stateAt` | 상태 유도(암묵 냉각 포함) | 순수 | `store.go:1669` |
| `decodeBaseline` / `decodeFirstRank` | 두 저장된 사실 | NULL은 미기록, 읽을 수 없는 순간은 에러 | 같은 파일 |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기 전용이다.
- 파싱 실패는 **부분 결과를 돌려주지 않는다**. 한 행이라도 읽을 수 없으면 `nil, err`다. 절반의 요약으로 `recordFirsts`가 '이미 기록됨'을 오판하는 것보다 낫다.
- `Summary`는 `Candidate`를 embed하고 두 저장된 사실을 옆에 단다. 화면과 스캔이 같은 읽기를 쓴다.

## Safety conclusion

- Safe edit boundary: 파싱 실패를 행 건너뛰기로 바꾸는 것(누락된 요약은 `recordFirsts`에서 재기록으로 이어진다), `at`을 시계로 바꾸는 것은 금지
- High-risk impact: no — 읽기 전용. 결과가 `recordFirsts`의 '아직 안 쓴 컬럼' 판정 입력이므로 누락 방향만 위험하다.
