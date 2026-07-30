# Function Logic Map: `Store.PruneExpiredCandidates`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. D11은 보존을 둘로 나눠 놓고 두 번째 절반(만료된 후보 요약 정리)에 집행자를 주지 않았다 — §1·§2 리뷰가 둘 다 지적했으나 §4의 컬럼이 들어오기 전에 배선하면 spec이 남기라고 한 데이터를 실패 없이 파괴하므로 순서 가드로 미뤘던 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `at` | 0이 아닌 판정 순간 | 호출자(watch 사이클이 `store.Now()`) | 0이면 에러 — 순간 없는 스윕은 판정할 수 없다 |
| `grace` | 0 이하면 `DefaultRawRetention`(48h) | 호출자 | **0이 '유예 없음'이 되지 않는다.** 미설정 필드가 경계를 끄는 것이 이 패키지가 반복해서 만난 실패다 |
| `s.cooling` / `s.staleness` | 저장소의 TTL | `Open` | 컷오프를 Go에서 계산해 소수부를 strftime에 잃지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `grace <= 0` | 없음 | `DefaultRawRetention` 적용 | `TestAnAbsentGracePeriodIsTheDefaultAndNotNoGraceAtAll` |
| B2 | `at.IsZero()` | 없음 | `0, error` | 직접 테스트 없음 |
| B3 | `ExecContext` 에러 | 없음 | `0, wrap` | 직접 테스트 없음 |
| B4 | `RowsAffected` 에러 | 삭제는 이미 커밋 | `0, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.ExecContext` | `DELETE FROM candidates WHERE (cooled_at IS NOT NULL AND cooled_at <= ?) OR (cooled_at IS NULL AND last_seen_at <= ?)` | 단일 문장 | 두 분기가 필요한 이유는 `stateAt`이 만료를 컬럼이 아니라 유도로 정하기 때문이다 |
| `at.Add` / `cooledCutoff.Add` | 컷오프 산술 | — | Go에서 계산해 경계를 정확히 유지 |
| `stamp` | 고정폭 직렬화 — 사전식 순서 = 시간 순서 | — | `stamp`는 고정폭(`2006-01-02T15:04:05.000000000Z07:00`)이다. TEXT 컬럼이라 모든 `<`·`>=`·`ORDER BY`가 BINARY 문자열 비교이고, `time.RFC3339Nano`는 소수부 뒤 0을 떼어내 텍스트 순서가 시간 순서가 아니게 만든다 — §1 리뷰가 실행으로 재현한 P0다(`TestTheStoredInstantOrdersLikeTime`). |

## State mutations and fallbacks

- `candidates` 행을 삭제한다. 이 패키지에서 요약을 **지우는** 유일한 문장이다.
- 삭제가 수명 사건이 아니다. D1은 이미 만료 후 다음 교차가 새 `first_seen_at`을 가진 새 후보라고 정했고 `Promote`가 만료된 행의 모든 컬럼을 초기화한다. 행을 없애는 것은 같은 상태에 더 짧은 길로 가는 것이다.
- 유예가 없으면 `first_seen_at`이 그것을 설명하는 이틀치 원 관측이 아직 디스크에 있는 동안 사라진다. D3의 '늦게 본 비율' 회계가 셀 수 없게 되는 방향이다.
- 부분 실패는 없다 — 단일 DELETE다. 실패하면 아무 요약도 지워지지 않는다.

## Safety conclusion

- Safe edit boundary: `grace` 기본값 제거, `cooled_at IS NULL` 분기 제거(암묵 냉각한 후보 — 스캔이 죽어 남긴 바로 그 후보 — 에 영원히 닿지 못하거나 살아 있는 후보를 지운다), 원 관측 보존 창보다 짧은 유예는 모두 금지
- High-risk impact: no — 발굴 저장소의 파생 요약 삭제다. 주문·원장·손절 어디에도 닿지 않는다. 다만 이 함수는 이 패키지에서 **`first_seen_at`을 지울 수 있는 유일한 코드**이고, 그것이 이 change의 유일한 주장이다. 실수하면 조용히 증거가 사라지는 방향이라 경계 테스트가 양쪽에서 걸려 있다.
