# Function Logic Map: `Tracker.Observe`

- Source: `internal/reconcile/mismatch.go` (lines 486–654)
- AST evidence: `ast.json` (`source_sha256: 49fcb007cab394f144230db5b1bf330b95e2303dfef48648edc806501edd2a57`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 진입 차단 상태기계. 면제 없음.

이 map은 a083의 것을 복사하지 않고 HEAD에서 다시 만들었다. 네 번째 독립 리뷰가
a083 복사본이 *수정 전* 코드를 현재 동작으로 서술하고 있음을 잡았다 — 특히
"`t.adjusted = nil` — 결함의 지점 — credit을 무조건 전부 버린다"는 행. HEAD에는
그 문장이 없다.

## What it does

한 번의 비교(`Diff`)를 받아 진입 차단을 세우거나 풀고, 결과를 저널에 기록하고,
실행 게이트를 동기화한다. a083이 바꾼 것은 두 가지다.

1. **해제 자격.** 조정 credit은 그 조정이 계산된 비교보다 *엄격히 나중인* 비교의
   관측으로만 소비된다(`creditUsableBy`). 같은 비교는 재조회가 아니다 — 그 동등성이
   a083 결함의 전부였다.
2. **credit 수명.** 개정 3 이전에는 시계 비교(`Block.Since` 대 credit 스탬프)로
   staleness를 재려 했는데, `Block.Since`는 사이클 *끝*에 쓰이는 `entered_at`이고
   credit은 스냅샷 읽기가 *시작된* `Diff.AsOf`를 담는다. 실계좌에서는 그 사이에 사이클
   전체가 들어가므로 평범한 수렴→차단→동의 흐름이 영영 해제되지 않았고, 아래 폐기 규칙이
   credit까지 버렸다. a083 프로덕션 결함의 재현이다. 지금은 시계를 전혀 보지 않고
   **차단의 존재**(`hasBlockFor`)로 수명을 묶는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff.AsOf` | RFC3339, 스냅샷 읽기가 시작된 시각 | 대사 비교기 | 파싱 불가면 `creditUsableBy`가 false — 차단 유지(보수) |
| `diff.BlocksEntry()` | bool | `compare.go` — 수량 불일치와 누락 주문만 센다 | 외부 보유는 세지 않는다. 그래서 `symbolsInDispute`가 따로 필요하다 |
| `t.adjusted` | `map[string]string` (심볼 → credit 비교 스탬프) | `Tracker.AdjustmentApplied`가 쓴다 | 비어 있으면 아무 차단도 풀리지 않는다 |
| `t.blocks` | `map[string]Block`, nil 허용 | tracker 소유 | nil이면 492행에서 초기화 — 영 값 tracker가 사용 가능 |
| `t.failures` / `maxFailures` | 0 이상 | tracker + config | 한도 도달 시 영구 차단으로 승격 |
| `t.mu` | 491행 Lock, 650행 Unlock | tracker | 전 구간이 하나의 임계구역 |

## Branches and early returns

조기 반환이 없다. 단일 임계구역을 끝까지 지나 항상 `(Outcome, persistErr)`로 나간다 —
저널 쓰기가 실패해도 게이트 동기화(649)를 건너뛰지 않기 위해서다.
분기별 조건과 **측정된** 커버리지는 `branch-test-map.md`에 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `symbolsInDispute(diff)` (501) | 이견 심볼 집합. 해제 판정과 credit 폐기가 **같은** 집합을 써야 한다 | 순수 함수 | AST `calls` |
| `creditUsableBy` (520, 625) | credit이 이 비교로 소비 가능한지 | 파싱 실패는 false — 차단 유지 | AST `calls` |
| `hasBlockFor` (639) | 이 credit이 풀 수 있는 차단이 아직 있는지 | 순수 조회. 시계 비교 없음 | AST `calls` |
| `blocksFor(diff, t.AccountRef, now)` (544) | 이번 비교가 만드는 차단 | 순수 | AST `calls` |
| `t.persist(ctx, toPersist)` (594) | 저널 write-through | 실패해도 진행. `persistErr`를 그대로 반환하고 차단은 메모리에 남아 다음 관측에서 재시도 | AST `calls` |
| `t.syncGate` (583, 649) | 실행 게이트 동기화 | persist 앞뒤로 두 번. 앞의 호출이 보수적 집합을 먼저 반영해 느린 저널이 게이트를 열어두지 못하게 한다 | AST `calls` |
| `hasPermanentQuantityAccountBlock` (606) | 영구 차단 여부 재계산 | 순수 | AST `calls` |

## State mutations and fallbacks

- `t.blocks` — 추가(546), durable 확정(597), 권위 반영(600), 해제 삭제(603).
- `t.adjusted` — 640행의 `delete`가 유일한 폐기 지점. 세 조건 중 하나면 버린다:
  이 비교가 그 심볼에 이견이거나, 해제가 이미 커밋됐거나, 풀 수 있는 차단이 더 없다.
  **durable 쓰기가 실패한 경우는 버리지 않는다** — 차단이 남아 있으므로 `hasBlockFor`가
  참이고, 저장 실패가 이미 번 해제를 앗아가면 안 된다.
- `t.failures`(510/543), `t.permanent`(606), `t.lastAt`(495), `t.observed`(496).
- `out.AwaitingAdjustment` — 해제 후보였으나 credit이 없거나(525) 아직 이견인(537) 차단.
  a085가 이 개수를 사이클 로그의 `awaiting`으로 내보낸다.
- Fallback: `persistErr`가 있어도 게이트는 동기화되고 Outcome은 반환된다.

## Safety conclusion

- Safe edit boundary: 해제 자격(520·528)과 credit 폐기(639). 셋 다 "차단을 더 오래 유지"가
  보수 방향이고, 반대 방향의 실수는 엔진이 설명할 수 없는 노출에 진입을 다시 연다.
- High-risk impact: yes — 진입 차단의 유일한 상태기계다. 개정 3이 되돌린 시계 비교를
  다시 넣으면 실계좌 타이밍에서 차단이 영구화된다. 가짜 시계 테스트는 경계값에 앉아
  통과하므로, 그 회귀는 −10분 어긋난 시계로 드라이버를 끝까지 돌려야 잡힌다.
