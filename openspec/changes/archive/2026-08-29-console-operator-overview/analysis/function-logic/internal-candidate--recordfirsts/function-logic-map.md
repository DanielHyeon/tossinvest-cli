# Function Logic Map: `recordFirsts`

- Source: `internal/candidate/scan.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. issues.md 4가 §5로 넘긴 배선 — `Collect`가 `NoteFirstPrice`/`NoteFirstRank`를 부르지 않아 `extended`가 영구히 `NO_BASELINE`, `seen_late`가 영구히 `NO_FIRST_SIGHTING`이던 것 — 을 여기서 갚는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `promoted` | 이번 pass에서 승격까지 성공한 심볼 | `Collect` | 비면 저장소를 아예 읽지 않는다 |
| `priced` / `ranked` | 심볼 → 패널 순서상 처음 그 값을 나른 관측 | `Collect` | 없는 심볼은 건너뛴다. 빈 문자열 가격과 rank 0은 애초에 들어오지 않는다 |
| `at` | 이번 스캔의 순간 | 호출자 인자 | `Summaries`의 상태 판정 기준이자 write의 순간이 아니다 — write에는 관측 자신의 `ObservedAt`이 간다 |
| `result` | 포인터 — 카운터와 `Rejected`를 여기에 쓴다 | 호출자 | nil이면 패닉이지만 호출자가 항상 `&result`를 준다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `len(promoted) == 0` | 없음 | `nil` — 저장소 읽기도 생략 | `TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone`(간접) |
| B2 | `store.Summaries` 에러 | 없음 | `err` — pass 중단 | 직접 테스트 없음 |
| B3 | 요약 순회 | `needPrice`/`needRank` 지역 map | — | `TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank` |
| B4 | `s.Market != market` | 없음 | `continue` — 다른 시장의 요약은 무시 | 직접 테스트 없음 |
| B5 | 승격 심볼 순회 | 저장소 write | — | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` |
| B6 | 가격 관측이 있고 기준선이 아직 없음 | `NoteFirstPrice` **write** | — | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` |
| B7 | `NoteFirstPrice` 에러 | `result.Rejected` 추가 | 계속 — 이미 승격된 후보다 | 직접 테스트 없음 |
| B8 | else — 성공 | `result.FirstPrices++` | — | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` |
| B9 | 순위 관측이 없거나 이미 기록됨 | 없음 | `continue` | `TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank`, `TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone` |
| B10 | `NoteFirstRank` 결과 분기 | — | — | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` |
| B11 | 에러 | `result.Rejected` 추가 | 계속 | 직접 테스트 없음 |
| B12 | `stored.Recorded()` | `result.FirstRanks++` | — | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `store.Summaries` | 어떤 후보가 아직 두 컬럼을 비워 뒀는지 **한 번에** 읽는다 | 실패는 pass 중단 | 심볼마다 묻는 순진한 배선은 tick마다 심볼당 IMMEDIATE 트랜잭션 2개다(D16) |
| `store.NoteFirstPrice` | 기준선 1회 기록 | 실패는 `Rejected`, 이미 있으면 기존 값을 돌려준다(idempotent) | `store.go` |
| `store.NoteFirstRank` | 최초 목격 순위 1회 기록 | 창 밖 관측은 **에러가 아니고 저장도 안 된다** | `store.go`, `veto.go:nearFirstSighting` |
| `s.Baseline.Recorded()` / `s.FirstRank.Recorded()` | 부재와 0을 가르는 술어 | 순수 | `Recorded()`는 `Price != ""`, `Rank>0 && Total>0` |

## State mutations and fallbacks

- 저장소 write 2종: `candidates.first_price/at/source`와 `candidates.first_rank/_total/_at/_source`. 둘 다 해당 컬럼이 NULL일 때만, 각 삶에 한 번.
- 요약을 **승격 뒤에** 읽는 것이 계약이다. `Promote`는 만료된 후보가 다시 교차하면 두 컬럼을 NULL로 되돌리므로(D1), 루프 앞에서 모은 집합은 방금 초기화된 삶을 '이미 기록됨'으로 읽고 새 삶이 죽은 삶의 기준선으로 달린다.
- 부분 실패: 한 심볼의 first-write 실패는 `Rejected`에 이름을 남기고 나머지 심볼은 계속 간다. 그 후보는 승격은 되었고 기준선은 없는 상태로 남으며, 그것이 `extended`가 영구 미측정이 된다는 뜻이라 세는 값이 있다.
- 가격 write 실패 뒤에도 같은 심볼의 순위 write는 시도된다 — 둘은 독립된 사실이다.

## Safety conclusion

- Safe edit boundary: 요약 읽기를 승격 앞으로 옮기는 것, 저장소의 NULL 가드에만 기대어 `needPrice`/`needRank` 선별을 없애는 것(D16의 write 예산), 창 밖 순위를 에러로 만드는 것 모두 금지
- High-risk impact: no — 발굴 저장소 write만 한다. 다만 이 두 컬럼이 `extended`와 `seen_late` 두 veto의 **내구성 있는 절반**이고, 여기서 안 쓰면 두 veto가 조용히 영구 미측정이 된다. D10 덕분에 미측정은 통과가 아니지만, 화면의 기본 독해를 '대부분 미확인'으로 만드는 쪽이다.
