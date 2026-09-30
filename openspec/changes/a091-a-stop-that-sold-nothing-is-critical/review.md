# a091 proposal-freeze 리뷰

- **날짜**: 2026-08-06
- **대상**: proposal / design / tasks / specs(engine-safety) / **analysis 산출물**, base `ec29dc72`
- **위험 등급**: High-risk → 적대적 Eng 필수
- **보이스**: Claude Eng(적대적) + Claude 소비자·폭발반경 렌즈, 둘 다 독립 실행 ·
  Codex `[codex-unavailable]` 한도 소진(2026-08-08 회복)
- **판정**: **FREEZE 거부**
- 두 보이스의 지적은 Manager가 전부 코드·원장으로 재검증했다.

## 새로운 것 — FLM 산출물은 정확했다

두 보이스 모두 `analysis/function-logic/`의 AST·행번호·분기표가 **HEAD와 전 행 일치**함을
확인했다(`ast.json`의 `source_sha256`까지). 다섯 라운드 만에 처음으로 **근거 자체는
틀리지 않았다.**

거부 사유는 다른 데 있다 — **FLM이 `applyFloor` 안에서 멈췄고, 승격이 실제로 부르는
경로(`o.alert → Notify → notifyCritical → deliver`)를 따라가지 않았다.**

## C1 (차단) — 이 change는 손절을 지연시킨다. 나는 반환값만 봤다

`proposal.md`와 `design.md`가 "반환값을 안 바꾸니 §0.3 무관"이라고 썼다. 반환값은 안 바뀐다.
**바뀌는 것은 `o.alert`가 돌아오기까지 걸리는 시간이다.**

| | normal (현재) | critical (승격 후) |
| --- | --- | --- |
| 경로 | `publishBestEffort` (`notifier.go:138-150`) | `notifyCritical` (`:153`) → `deliver` (`:238`) |
| 예산 | publish 1회, 상한 10s | **3회 × 10s + 2회 × 2s ≈ 34s**, `n.mu` 보유 |

- `DefaultCriticalAttempts = 3` (`notifier.go:45`), `DefaultRetryDelay = 2s` (`:48`)
- ntfy 기본 `Timeout` 10s (`ntfy.go:72-73`)
- `o.alert`는 **동기**다 (`exitloop.go:1600-1607`)
- `ObserveOnce`의 포지션 순회는 **순차**다 (`:453`, 주기 5초 `:97`)

그래서 한 포지션의 0주 캡이 **같은 사이클 뒤쪽 포지션들의 손절 제출을 최대 34초 민다.**
8/2처럼 RECONCILE이 3분 지속되면 매 사이클 반복된다.

내 FLM의 calls 표에 이 칸이 있었다 — **"Error/timeout/retry contract"**. 나는 거기에
"삼킴"이라고 적고 예산을 적지 않았다. **산출물이 물어본 것을 안 채웠다.**

### 그리고 이것은 a091의 문제가 아니라 HEAD의 문제다

exit 루프의 critical 알림 4종이 **이미 같은 성질로 돈다**.

| 줄 | 이벤트 | 위치 |
| --- | --- | --- |
| `:781` | `EventExitObservationOutage` | `checkOutage` — `ObserveOnce` 안 |
| `:1527` | `EventExitJudgementRefused` | `judgeLadder` 안 |
| `:1551` | `EventExitProposalRefused` | `submit` 안 |
| `:1581` | `EventExitLiquidationDelayed` | `record`·`submit` 안 |

**publisher가 붙어 있고 느리면 손절 관측 루프가 알림 하나당 최대 34초 멈춘다 — 지금.**
a091은 그 성질을 **빈도가 훨씬 높은 경로**로 확장할 뿐이다.
`proposal_refused`는 브로커 거부 때만 돌지만 0주 캡은 RECONCILE 중 **매 사이클** 돈다.

이것이 이 리뷰에서 나온 가장 큰 발견이고, **a091보다 먼저 서야 한다.**

## C2 (차단) — 동기가 된 실측 주장이 거짓이다

`proposal.md`가 "8/5엔 5번 다 받았고 8/2엔 한 번도 못 받았다 — **차이는 등급 하나다**"라고
썼다. 거짓이다. 차이는 **transport의 유무**다.

- `alert_outbox` id 1~9 (2026-07-31 ~ 08-04T13:29): 전부 `critical`·`PENDING`·**`attempts=0`**.
  `attempts`가 0으로 남는 분기는 `notifier.go:252` `if n.Publisher == nil { break }` 하나뿐
- `engine.log`: `2026-08-01T19:31:19`과 `2026-08-03T09:03:45`에
  `"no notification publisher is configured"` — **8/2를 양쪽에서 끼고 있다**
- 알림 배선 커밋 `e540668f`는 **2026-08-04**다. 8/2보다 이틀 뒤

**8/2에 이 이벤트가 critical이었어도 운영자는 0회 받는다.** 실제로 일어났을 일은
outbox 1행 + `alert_undelivered` ERROR 13줄 + 게이트 13회 래치다.

**이 change의 실이익은 "운영자 호출"이 아니라 "원장에 흔적이 남는 것" 하나다.**
그 이익만으로도 change는 성립한다 — 문제는 과장이다.
`[미측정]`이라고 쓴 항목도 측정 가능했고 답은 "없었다"였다.

## 그 외 확인된 지적

| # | 내용 |
| --- | --- |
| H1 | D1의 A/B 이분법이 선택지 C(이벤트별 override)를 빠뜨렸다. B가 옳은 **진짜 근거**는 `measurement_test.go:47-54`의 class rule — subject 단위로 `CriticalEvents()`를 훑는 규칙이 per-call override로 무력화된다. design이 그것을 인용하지 않았다 |
| H2 | `design`은 `EventExitProposalCapped`를 "부분 캡 전용"으로 좁히는데 `tasks 3.7`은 B2의 `logErr(EventExitProposalCapped, …)`(`:1412`)를 유지한다. B2는 **0주**다 — 같은 사건이 로그와 알림에서 다른 종류가 된다 |
| H3 | 13회는 outbox 1행으로 접히지만 **발송은 13번 나가고**, 2회차부터 `MarkAlertDelivered`가 `state=PENDING`에 걸려 **`alert_undelivered` ERROR 12줄**이 남는다. tasks 5.2는 `attempts=1`만 알고 이 12줄을 모른다 |
| M1 | FLM의 "0주 경로는 둘"은 결론만 맞고 근거가 없다. `isZeroQuantity`(`:1657-1664`)는 `""`와 파싱 실패도 0으로 보고 **정확히 `"0"` 문자열 비교**라 `"0.0"`·`" 0"`은 통과한다. FLM Inputs 표가 `quantity`의 non-zero 불변식과 그 출처를 인용하지 않았다 |
| M2 | D4(문구)의 실측 근거가 없다. 8/2 로그 본문은 **영어**이고, 한국어 "일부만 나갔다" Title은 8/2 **이후**에 들어왔다. 그 거짓 문구는 운영자에게 전달된 적이 없다 — 고쳐도 좋지만 8/2 증거로 배치하면 안 된다 |
| M3 | FLM calls 표가 `ast.json`의 10개 중 5개만 적고 "AST"로 표기했다. 결론은 유효 |
| M4 | spec delta는 ADDED보다 **MODIFIED**가 맞다. base `engine-safety`「등급화된 알림」이 critical 부류를 괄호로 열거하고 있고, `exit-policy` `:62`에 이미 "캡 발생은 알림된다(SHALL)"가 있다. 마지막 SHALL NOT은 구현 서술이고 그 Scenario의 WHEN("이 요구사항이 적용된다")은 트리거가 없어 테스트 불가 |
| M5 | Impact 누락: `internal/obs/a091_*_test.go`(신규, **등록 누락을 잡는 유일한 장치**), `internal/app/engine/exitloop_test.go`, `event.go`의 종류 주석, `issues.md`(미존재), `docs/pm/generated/` 3종, `cmd/tossctl/engine_assembly.go:31-35`의 stale 주석 |

## 검증에서 살아남은 것

| 주장 | 증거 |
| --- | --- |
| **`applyFloor` FLM 분기표 전 행 정확** | `:1403-1447` 대조, branches 6·returns 7·defers 0, `source_sha256` 일치 |
| **`SeverityOf` FLM 정확** | `event.go:309-314`, `criticalEvents` **18종**, 기본값 normal |
| codegraph가 `SeverityOf` production 호출자 2개를 놓친다 | `log.go:186`, `notifier.go:108` — 도구 한계 기록이 사실 |
| 등급은 종류에만 붙어 있다(현행) | `obs.Event`에 severity 필드 없음, per-call override 부재 |
| `EventExitProposalCapped`는 한 번도 critical이었던 적이 없다 | `git log -S"EventExitProposalCapped:"` → 0건 |
| **8/2 실측 숫자 전부 사실** | `exit_events` id 141~172, `STOP_LOSS_LADDER` **정확히 13회**, `ADJUSTMENT_CLOSED` id 174. 로그 `proposal_capped` 13줄 전부 `severity:normal`·`quantity:"0"`. `alert_outbox` 관련 행 **영구 0건**. 경로는 tail `:1446`(B2 아님) — 주장대로 |
| D2 성립 | `submit:1237-1239`에 `proposal` 스코프 존재. `isProtective`={BaselineBreach, LadderStop}는 `Action.Orderable()` 5종의 **정확한 이분** — 누락된 보호 액션도 익절 오분류도 없다 |
| 새 종류가 깨는 소비자 없음 | `CriticalEvents()` 호출자는 테스트 2개(집합·개수 미고정), console/httpapi에 이벤트명 필터 없음, `alert_outbox.event_type`은 CHECK 없는 자유 문자열, rename 규칙은 `execgw ReasonCode` 소유 |
| 스키마 무변경 | `outbox.go:51` `event_type TEXT NOT NULL`, CHECK·인덱스 없음 |
| 범위 완결성 | exit 루프 알림 6종 전수 대조. normal 2종 중 `EventExitPositionUnmanaged`는 **운영자가 선택한 정상 상태**라 승격 대상 아님 |
| 게이트 위생 | `base-commit.txt` = `ec29dc72` = HEAD, `ast.json` 두 개 해시 일치 |

## 판정과 다음

**FREEZE 거부.** 차단 2건(C1·C2).

**C1이 순서를 바꾼다.** exit 루프의 critical 알림이 손절 관측을 최대 34초 막는 것은
**HEAD의 결함**이고 a091이 만든 것이 아니다. 그 위에 빈도가 높은 새 경로를 얹으면
안전을 개선하려는 change가 §0.3을 후퇴시킨다.

권고:

- **a092(신설·선행)** — exit 루프의 critical 알림을 관측 경로에서 떼어낸다.
  durable enqueue는 동기로 두되(원장 보장) 발송은 루프 밖으로. 기존 critical 4종이 대상
- **a091(수정 후 재제출)** — Why를 "durable row"로 다시 쓰고 C2의 과장·`[미측정]` 삭제,
  H2(B2의 로그 종류) 일치, M4(MODIFIED로 전환), M5(Impact 보강), `issues.md` 생성.
  C1은 a092가 해소한 뒤 발효
- **a090·a089·a087** — 종전 순서 유지

## 다섯 번째 — 형태가 바뀌었다

| 라운드 | 틀린 방식 |
| --- | --- |
| a087 초안 | 실측 1건에서 브로커 성질을 단정 |
| a087 교체본 | 선례를 조건 없이 일반화 + 호출 사슬 미추적 |
| a089 초판 | 로그 침묵에서 사건을 추론 |
| a089 재작성본 | 흔적이 남는 경로만 조회, 안 남는 경로 누락 |
| **a091** | **산출물은 정확했으나 그 산출물이 물어본 칸("timeout/retry contract")을 안 채웠고, 부작용의 호출 사슬을 안 따라갔다** |

앞의 넷은 증거가 틀렸다. 이번엔 **증거는 맞고 읽기가 얕았다.** FLM은 함수 경계에서
멈추지만 §0.3은 경계를 넘는다 — 반환값이 아니라 **부작용의 예산**을 따라가야 한다.
다음 FLM은 calls 표의 timeout/retry 칸을 반드시 수치로 채운다.

---

# 2차 판 — 재freeze (2026-10-01~)

## 착수 승인 기록

- **Manager 배정 원문 (2026-10-01)**: 「다음 배정: a091-a-stop-that-sold-nothing-is-critical 재freeze → (통과 시) 구현. … 발효 조건(C1 = a092 완주)이
  a092 아카이브로 성립했다. … 0.0 a092 아카이브 확인 → 0.1 base 재고정(WORKFLOW 절차·영수증·승인 참조는 형제 로트 규격). 0.2 spec delta 재기저화 …
  0.3 FLM 재검증 … calls 표의 timeout/retry 칸을 수치로 … 0.4 validate → 0.5 freeze 적대 재리뷰(분리 보이스, codex 는 슬롯 요청제) → 0.6 check_analysis.
  freeze 통과 보고 → 내 승인 → 구현」. 승인 참조: 사용자 상임 오케스트레이션 지시(구현은 Teammate 위임, Manager 스케줄링). 실행: 팀메이트(Opus).

## 0.0 발효 조건 C1 — a092 아카이브 확인

- `openspec/changes/archive/2026-09-30-a092-an-alert-does-not-hold-the-stop/` 실재, 아카이브 커밋 `75d138b5`(2026-09-30 17:55 KST)가 새 base `b30318d6` 의
  조상이다(`git merge-base --is-ancestor`). 정본 engine-safety 「등급화된 알림」 은 a092 문단(기록 입구 · 배달 실행자 · 재알림 창)을 담은 판이다 — 0.2 가 그 위에 재기저화한다.

## 0.1 base 재고정 `ec29dc72` → `b30318d6` (WORKFLOW 「사람 승인 base 재고정」)

- **조건 ① 귀속 실측**: 옛 base `ec29dc72..b30318d6` 은 696+ 커밋. a091 디렉터리를 만진 비병합 커밋 = `989ab031`(2판 재작성, `.go` 0) · `a30eb35a`(`.go` 6).
  `a30eb35a` 는 a091 디렉터리를 **처음 들인 묶음 커밋**(「fix(safety): bound alerts and plan exit hardening」 — a087 · a089 · a091 · a092 · a094 · a095 · a096 문서를
  함께 추가)이고 그 `.go` 6 은 전부 a096 몫이다: `internal/journal/a096_claim_for_delivery_test.go` · `a096b_round2_test.go` · `internal/obs/a096_one_send_per_condition_test.go` ·
  `a096b_round2_test.go` · `internal/journal/outbox.go` · `internal/obs/notifier.go` — 뒤 둘의 편집 함수는 a096 번들(아카이브 `2026-08-29-a096-one-condition-is-one-alert`:
  `journal.claimalertfordelivery` · `enqueuealert` · `markalertdelivered` · `claimowed` · `notifier.deliver` · `notify` · `notifycritical` · `claimanddeliver` · `flush` · `acknowledge`)이 덮는다.
  **a091 자기 Go 편집 0**(a091 은 구현 전이다). renumber 이력 없음(`git log --follow` 두 커밋).
- **옛 base 에서의 판정**: 깨끗한 격리 워크트리 @ `b30318d6` 에서 `check_analysis` — required **325**(전부 base 뒤 형제 착지 몫 — a091 Go 0) · 번들 2 stale
  (`exitloop.go` · `event.go`). 원문 `analysis/rebase/check-oldbase-ec29dc72-at-b30318d6.txt`.
- **조건 ② 승인**: 위 「착수 승인 기록」 의 Manager 배정 원문(0.1 base 재고정 지시).
- **조건 ③**: 다음 커밋이 `base-commit.txt` 한 파일.

## 0.2 spec delta 재기저화 (2026-10-01)

- 정본 `openspec/specs/engine-safety/spec.md` 「등급화된 알림」(base `b30318d6`, 125 줄 — a092 아카이브 뒤 판, a095 는 이 요구를
  건드리지 않았다: 정본 이력 `75d138b5` a092 · `78e70d62` a095 는 engine-safety 에 **다른 요구 둘**을 더함)을 기계 복사하고
  a091 몫 셋만 더했다: ① 첫 문단 열거에 「보호 청산의 0주 제출」 ② 첫 문단 뒤 0주 문단 넷 ③ 끝에 Scenario 넷.
- 대조(파이썬, 줄 단위): 정본 비공백 83 줄 중 delta 에 없는 것 = **1**(열거 첫 줄 — 의도한 치환), delta 에만 있는 줄 28(a091 몫).
  `openspec validate --strict` 통과.

## 0.3 FLM 재검증 (2026-10-01)

- 깨끗한 격리 워크트리 @ `b30318d6` 에서 `go run ./tools/logic-map` 로 AST 14, 같은 커밋의 `go test -coverprofile`(engine ·
  obs · riskcalc · reconcile · exitpolicy)로 진입 실측, 산문 · Test 열 · 수치는 하네스 `analysis/harness/write_bundles.py`.
- 편집 대상 3(`applyFloor` · `submit` · `SeverityOf`), 알림 경로 2(`ExitObserver.alert` · `RecordOnly.Notify`), M1 사슬 9.
- **calls 표의 timeout/retry 칸을 수치로 채웠다(첫 리뷰 방법 교훈)**: `ConfirmedFloor` 의 RECONCILE 읽기(Query 2 · 3시도 · 대기
  400/800ms ±25% · 예산 8s 는 대기만 · HTTP 15s · 401 refresh ≤2 → 한 Query 최악 ≈ 98s) · critical 기록(`busy_timeout` 5s ·
  `n.mu` 기한 없음 · 원격 0) · 재알림 창 1h.
- 3판에서 드러난 2판의 틀린 문장 셋: 「`applyFloor` 는 브로커에 닿지 않는다」(§0.4 — RECONCILE 에서 읽는다) · 「`isZeroQuantity`
  는 정확히 `"0"` 비교」(M1 — base 는 수치 비교) · 「`EventExitProposalCapped` 는 부분 캡 전용」(D1 — 익절 0주가 남는다).
  전부 3판 문서에서 정정했다.
- engine 커버리지 실행은 `-trimpath` 때문에 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다(무관한 AST 핀) — 프로파일은 유효.
