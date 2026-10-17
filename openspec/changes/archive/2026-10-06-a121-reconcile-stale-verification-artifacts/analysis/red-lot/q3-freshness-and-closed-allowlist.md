# a121 RED 로트 — Q3 신선도 상수 제안 · CLOSED status 허용 목록 전사 결과

- 작성: 2026-10-05, Terra 팀메이트(RED 로트), 고정 사본 `2c6ef1ef`(FREEZE-APPROVE 좌표)
- 성격: **design 제안용 메모.** design·spec 문구 적용은 Manager 몫이다.
- 갱신(A-RED 리뷰 수리, 2026-10-05): Q3 = 15초·429 무재시도·allowlist 의 Q1 영수증 동승은 Manager 가 수용했다(착지 때 design 기록).
  그래서 골격은 Q3 상수를 세웠다 — `const reconcileFreshnessBoundValue = 15 * time.Second`,
  `var reconcileFreshnessBound = reconcileFreshnessBoundValue`(생산 파일의 다른 대입 0 은
  `TestReconcileFreshnessBoundIsTheApprovedConstantInProduction` 이 핀). 허용 목록은 여전히 빈 값(출처 부재).

## 1. Q3 — 두 번 읽기 신선도 한도 (design G1-6 · codex F7)

### 제안

```go
// reconcileFreshnessBound 는 첫 목록 읽기 시작부터 추가 승인 직전까지의 상한이다(Q3, 리뷰 승인 상수).
const reconcileFreshnessBoundValue = 15 * time.Second
```

이름 있는 상수 하나, 값 **15초**. 넘으면 `RefuseFreshnessExceeded`. 창의 시작은 첫 목록 읽기(승인 대기는 창 밖 — freeze 재검 P2),
끝은 추가 직전 재검사(F7).

### 근거 (영수증)

| 입력 | 값 | 출처 |
|---|---|---|
| 정상 경로의 요청 수 | 종목 조회 2 + 목록 3그룹 × 2회 = **8 GET**(각 그룹 1페이지) | design G1-6, `TestReconcileReadsTheGroupsInTheFixedOrder` |
| 공식 API 왕복 지연 | 77ms(DELETE)·124ms(modify)·최대 169ms/182ms(주문 왕복) | `verify-execution-capability/measurements.md` M41·M14·M35 |
| 버스트 한도 | 7요청/535ms 후 429, 지속 관측 9.67 req/s | 같은 문서 M8 |
| 429 벌칙 창 | 수십 초 단위(300ms 내 3연속 실패), 재시도 대기 15s·30s | M4, `internal/verifylive/retry.go:57-66` |
| 페이지 상한 | 그룹당 10페이지 × limit 100 | `steps.go:61`, design 「치르는 값」 |

- 정상 8 GET 은 순차로도 약 1~2초다(왕복 ≤ 200ms 가정). 15초는 그 7배 이상 여유이고, OPEN 그룹이 여러 페이지여도
  (그 경우 G1-1·G1-3 으로 어차피 거절되지만) 9 req/s 페이싱으로 수 초 안에 끝난다.
- **429 를 한 번이라도 만나면 창을 넘도록** 잡았다: 재시도 대기(15s)가 한 번만 끼어도 15초를 넘는다. 율속 제한에 걸린 읽기 쌍은
  "가까운 두 시점" 이라는 이 규칙의 전제(컷 토큰이 없으니 두 읽기 간격을 좁혀 창을 좁힌다)를 지키지 못하므로, 그때는 거절하고
  사람이 다시 돌리는 것이 보수 방향이다.
- CLOSED 그룹은 아래 §2 결과(허용 목록 비어 있음)로 행이 하나라도 있으면 거절되므로, 수락 경로의 CLOSED 는 사실상 1페이지다.

### 대안과 치르는 값

| 값 | 장점 | 비용 |
|---|---|---|
| 10s | 창이 더 좁다 | 느린 망·토큰 갱신(POST /oauth2/token) 1회가 끼면 거절될 수 있다 |
| **15s (제안)** | 정상 경로 여유 7배+, 429 1회면 거절 | 429 직후 재실행은 벌칙 창이 지나야 성공 |
| 30s | 429 재시도 1회를 흡수 | 재시도 대기 동안 브로커 상태가 바뀔 창이 두 배 — 이 규칙의 목적과 반대 방향 |

명령은 429 를 재시도하지 **않는** 쪽이 제안과 맞는다(재시도는 창을 소모만 한다). 이것은 design 에 없는 구현 선택이므로 같이 판정 바란다.

## 2. CLOSED status 허용 목록 — 전사 결과: **출처 부재**

design G1-2 codex F4 는 허용 목록을 "`verify-execution-capability` 영수증/골든에서 전사" 하라고 했다. 찾은 결과:

| 찾은 곳 | 결과 |
|---|---|
| `openspec/changes/verify-execution-capability/measurements.md` · `issues.md` · `tasks.md` · `specs/` | 조건주문 status 관측은 **`WATCHING` 뿐**(M12 등록 직후, M18·M39 재시작 후, M19 정정 후). `COMPLETED`·`EXPIRED` 문자열은 이 change 디렉터리 전체에 **0회**(`grep -rn` 실측). M17·M36 은 목록 **필터** 값(`OPEN`/`CLOSED`)만 측정. M20·M41 은 취소 뒤 by-id 읽기 실패만 측정 — 취소된 조건주문이 CLOSED 목록에 무슨 status 로 남는지(남는지조차)는 미측정 |
| 골든 디렉터리(`analysis/goldens/`) | 저장소에 하나뿐(`archive/2026-10-04-a112-…/analysis/goldens`) — 전략 계약용, 조건주문 무관 |
| 실기록 기반 시험(`record_replay_test.go` 등) | 조건주문 CLOSED 목록 status 를 기록하는 관측 키 없음(`steps.go` 의 `conditional.status.*` 는 살아 있는 by-id 읽기) |

**측정이 아닌 출처(전사 대상 아님 — 참고로만):**

- `docs/migration/openapi.latest.json:940-949` — 그룹 status 문서 enum `WATCHING · PAUSED · ORDERING · ORDERED · COMPLETED · EXPIRED`
  (leg enum 은 `HOLDING`·`CANCELED` 추가, `:720-731`).
- `internal/official/conditional_reads.go:158-164` — 위 문서를 옮긴 오류 문구(CLOSED = {COMPLETED, EXPIRED}).
- `internal/protectionofficial/gateway.go:312` — 종결 판정에 `COMPLETED · EXPIRED · CANCELLED · CANCELED` 를 쓴다(문서 enum 에 없는
  두 철자 포함, 측정 근거 표기 없음).

**처리(지어내지 않음):** 골격의 허용 목록은 비어 있고, `TestReconcileClosedStatusAllowlistStaysEmptyWithoutASource` 가 출처 없이
채우는 것을 막는다. 전사 대조 시험 `TestReconcileClosedStatusAllowlistIsATranscribedGolden` 은 이 사유로 skip 한다.

**귀결(숨기지 않는다):** 허용 목록이 비어 있으면 CLOSED 행이 **하나라도** 있는 심볼은 `RefuseClosedStatus` 로 영구 거절된다 —
대상이 아닌 다른 조건주문의 만료 이력 하나로도. design 「치르는 값」 의 "COMPLETED 이력이면 영구 거절" 보다 넓다(EXPIRED 이력도).
a063 심볼에 CLOSED 이력이 있으면 Q1 과 별개로 이 경로가 닫힌다.

**Manager 판정 필요(제안):** (a) 빈 목록 수용(측정 전까지 CLOSED 이력 있는 심볼은 거절), (b) Q1 사람 실측(장중·조회 전용)이
CLOSED 목록 status 어휘도 영수증으로 남기게 하고 그것을 전사 — Q1 측정이 어차피 CLOSED 를 읽으므로 같은 세션에서 가능,
(c) 문서 enum 을 출처로 인정(design F4 문구 "측정·골든" 과 충돌 — 문서 개정 필요). RED 로트 권고는 (b).

## 3. 계수 정정 (A-RED 리뷰 P2-1)

첫 보고의 "거절 코드 49개" 는 틀렸다 — 실측은 **45**(수리 전). 수리 뒤 계수(고정 사본 `2c6ef1ef` 위 작업 트리, 2026-10-05):

| 항목 | 수리 전 | 수리 후 | 잰 방법 |
|---|---|---|---|
| 거절 코드 | 45 | **46**(`RefuseRecordFormat` 추가 — P1-4 format_version 2) | `grep -c 'ReconcileRefusalCode = "' internal/verifylive/reconcile.go` |
| 시험이 참조하지 않는 거절 코드 | 3(RecordUnreadable·NoCredentials·AccountsRead) | **0** | 코드별로 새 시험 파일 전체에서 `\bName\b` 계수 |
| RED FAIL(최상위) | 74 = verifylive 55 + official 6 + cmd 13(tagged, 무태그 10 포함) | **82** = verifylive 58 + official 6 + cmd 18(tagged, 무태그 13 포함) | `go test -v` 의 `^--- FAIL` |
| 새 최상위 시험 함수 | — | 98(verifylive 72 · official 7 · cmd 19) | `grep -c '^func Test'` |

이름 일관(같은 실측): 새 비시험 파일의 공개 식별자는 전부 `Reconcile` 계열이다 — `Reconcile*`·`DecodeReconcile*`·
`ErrReconcileSchema`·`KindReconcile`·`StepReconcile`·`ObservationReconcileBasis`·`ReconcileGroup*`·`Refuse*`(코드)·
`SetReconcilePoliciesForTest`(seam). `A121`/`a121` 은 식별자로 쓰이지 않는다(주석·문자열에만 22회).
