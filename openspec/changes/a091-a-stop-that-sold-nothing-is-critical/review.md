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

## freeze 재리뷰 2라운드 (2026-10-01, task 0.5) — **REJECT** (보이스 A · B REJECT, codex FAIL) · 분류만, 반영은 Manager 결정 뒤

- 대상: `git archive 64380cb6`(3판) → `/tmp/claude-1000/a091-r2-tree`, 전 보이스 read-only. Claude 보이스 A(적대 Eng) · B(소비자 · 폭발반경 · 시험)
  분리 실행(Opus, 각 독립 컨텍스트) · codex 교차 모델(v0.154.0, `-s read-only`, non-ephemeral, ~/.codex 접근 금지 + 위반 머리말 조항 —
  출력 머리말 위반 신고 없음. 실행 로그의 `~/.codex/skills/…` ERROR 한 줄은 codex 하네스 기동 시 스킬 적재이며 모델 명령이 아님; 명령 26건
  중 작업 디렉터리 밖 경로 0). 원문: `analysis/freeze-review/{claude-r2-voiceA.md, claude-r2-voiceB.md, codex-r2-prompt.md, codex-r2-output.md}`.
- 세 보이스가 **확인한 것**(참): 19 종 · class rule 셋 · `isProtective` 의 5 액션 정확한 이분(recovery 가 여섯째 액션을 만들지 않음 — codex) ·
  applyFloor 좌표와 0주 반환 두 자리 · `isZeroQuantity` 수치 비교 · D6 생산 출처 좌표 · 재시도 상수 · 1h · 5s · `Offer` 비차단 · 재기저화
  충실(83 줄 중 첫 줄만 치환) · `event_type` CHECK 없음 · 콘솔/httpapi 이름 필터 없음.

### 수렴 분류 (중복 병합)

| # | 심각도(최고) | 제기 | 내용 | 처분 후보 |
|---|---|---|---|---|
| R2-1 | **P0** | A · B · codex | **알림 꺼진(기본) 엔진 · 전송 부재에서 보호 0주가 진입 차단 + ENTRY_BLOCKED 승격을 부른다**(기본값 off `a074_notifications_test.go:28-30` · nil publisher `notifications.go:83-87` · Notifier 는 늘 생성 `gateway.go:327` · 배달 실행자는 publisher 부재를 실패로 셈 `alertdelivery.go:316-326` → 3 사이클 뒤 래치 · 승격 `:446-495`). a095 정본(`engine-safety/spec.md:1845-1851`)은 같은 모양의 사실을 알림 off 에서 critical 로 매기지 않았다(`adoption.go:451` `NotificationsEnabled`). 3판은 a095 를 선례로 들며 그 게이트를 뺐고 비용을 적지 않았다. 불변식 3(토글 OFF = upstream)과도 닿는다 | **사용자/Manager 결정 Q1**: (a) a095 식 게이트(알림 켜짐일 때만 critical, 「꺼짐」과 「켜짐 · 전송 실패」를 구별 — codex) 또는 (b) 무게이트 + 비용(래치 ≈3 배달 사이클 · durable ENTRY_BLOCKED · ack + mode-release) 명시 |
| R2-2 | **P0** | A · B | **exit-policy 정본 충돌** — `exit-policy/spec.md:62` 「캡 발생은 알림된다(SHALL — 일반 등급 알림이다 …)」. 보호 0주도 캡이므로 아카이브 뒤 두 정본이 어긋난다. delta 는 engine-safety 하나 | exit-policy MODIFIED delta 추가(보호 0주는 engine-safety 가 등급을 정한다 + Scenario) — 기계적 |
| R2-3 | **P0** | codex · A | **계좌 원문 로그** — 유지 · 핀하는 B2 `logErr` 가 `obs.FieldAccount, o.opts.AccountRef`(`exitloop.go:1838`, AccountRef = 계좌번호 `interlock.go:680-684`, 로거 비가림 `log.go:113-117`)를 쓰고, 새로 닿는 기록 실패 → `escalate` 로그도 원문(`notifier.go:432-444`) | 편집하는 줄에서 가림(`obs.MaskAccount` 선례) + sentinel 계좌 시험(B2 · 기록 실패 · 승격 성공/실패) |
| R2-4 | **P0** | codex · A | **§0.3 D5 불완전** — 기록 실패 시 `escalate` 가 exit goroutine 에서 동기로 원장 트랜잭션 하나 더(`record_only.go:136-157` · `notifier.go:425-449`), 종료 시 ctx 취소가 B2 로 가(`retry.go:365-366` ClassCanceled) 취소된 ctx 로 `RecordAlert` → 가짜 래치 · 승격 시도(A), `Acknowledge` 가 `n.mu` 아래 임의 백로그를 처리(`notifier.go:957-981`), 뒤쪽 포지션의 시세 신선도(`exitloop.go:477-495`) | 몫을 항목별로 이름 붙임(로그 · 잠금 대기 · 기록 · 실패 승격) + `context.Canceled`/`DeadlineExceeded` 는 0주 보고 제외(또는 시험) + 뒤쪽 보호 포지션 포함 시험 |
| R2-5 | P1 | B | **8/2 동기 사건의 성격** — 결말 `ADJUSTMENT_CLOSED` 는 엔진 밖 종결(`converge.go:95-98`)이고, 수동 매도 · 외부 주문이면 대사 블록 아래 `reconcileFloor` 가 보유 0(`exitwiring.go:216-220`) 또는 매도가능 0 을 낸다 — a091 은 주인이 이미 판 포지션에 「손절이 한 주도 나가지 않았다」 critical 을 낸다. `EventExitPositionClosedExternally` 는 바로 그 이유로 normal(`event.go` 주석). 8/2 의 한정 bound 가 `sellable` 이었다는 점은 외부 매도 주문의 잠금과 맞다(추론 — 원장 미열람) | **결정 Q2**: 보유 0 으로 인한 하한 0(보호할 것 없음)을 매도가능 0 · 낡음 · 오류와 가를지. Why/근거 문장 정정. 8/2 원장 재열람(읽기 전용, 사람 또는 승인 범위) |
| R2-6 | P1 | A · codex | **「원인은 세부 정보로 구분」 SHALL 과 공유 키** — PENDING 행은 첫 내용 유지(`outbox.go:285-340` · a097), 배달은 Title/Body 만(`alertdelivery.go:324-332`), `RecordOnly` 로그는 필드를 지움(`record_only.go:51`). 원인이 바뀌면 행 · 푸시는 첫 원인만 말한다. a095 정본이 같은 함정을 명명(`spec.md:1851-1852`) | **결정 Q3**: 원인별 키 vs 「에피소드 첫 원인(시각 포함, 한국어 Body) + 관측마다 안전한 로그」 명문화 — 두 순서 시험 |
| R2-7 | P2 | A · B · codex | **재생 5.1 · 5.2 의 전제** — 13회 안에서 행은 PENDING 이라 재알림 창이 판정하지 않고, 8/2 조건(publisher 없음)의 실제 결말은 배달 실행자의 `alert_undelivered` 매 사이클(`alertdelivery.go:212-262, 316-322`) + 래치 + 승격이다. D5 의 「발생원이 없다」는 누락에 의한 거짓. 기존 `newExitHarness` 는 가짜 알림 수집기(`exitloop_test.go:122-126, 236`) — outbox 를 못 본다 | 5.1 을 팔 셋(정상 전송 · 실패 전송 · publisher 없음) + 정착 행 1h 경계 별도 사례로, 하네스 = 실제 `RecordOnly` + 원장 + 배달 실행자(내보내기 훅) |
| R2-8 | P2 | B · codex | **시험 계획 구멍** — 익절 3 액션 · 보호 2 액션 전수 표 시험(현재 「익절」 단수) · B2 반환 회귀의 관측 가능 단언(레벨 해제 · 재발의 — 기존 시험은 「제출 없음」만) · 로그 캡처 하네스(H2) · a085 문구 규칙(한국어 · 이름(코드) · 계좌 없음, `engine-safety/spec.md:801-813`) · B2 원문 오류가 Title/Body 에 안 들어감 · 4.1 을 금지 문구 하나가 아니라 참인 0주 문장 단언으로 | tasks 2.x~5.x 재작성 |
| R2-9 | P2 | A | **알림 전수 표 누락** — `EventOperatingMode`(critical) 통지 경로: `exitloop.go:882-883` Announcer · `exit_unobserved.go:229-230` · exit Retrier 401 강화(`exit_record_only.go:15-21`, `retry.go:360-364`) — 마지막은 `applyFloor` 의 하한 읽기에서 닿고, 같은 호출에서 새 critical 과 겹친다 | 표를 「알림 경로에 닿는 경로」 단위로(a092 정본 방법) |
| R2-10 | P2 | B | **소비자 조사 불완전** — 배달 실행자는 「발송만」이 아니라 래치 · 승격(issues 문장 거짓) · `tossctl engine alerts ack`(`engine_alerts.go:74-100`) · mode-release · 콘솔 알림 꺼짐 안내(`settings_notifications.go:169`) · `docs/operations.md:514-520` 런북(새 종류 절 없음) | 표 보강 + 런북 Impact |
| R2-11 | P2 | A · codex | **수치 과장** — 「≈98s」 는 상한이 아님(`tm.mu` · 토큰 캐시 파일 I/O 기한 없음 `token.go:61-78,109-125,172-223`; 풀 대기 `SetMaxOpenConns(1)` `journal.go:174` 은 busy_timeout 밖), 여섯 요청 산식은 실현 불가(둘째 refresh 는 첫째가 adopt 일 때만 `client.go:344-360`), 호출 좌표 `:204`/`:227` 는 호출 자리가 아님, 「수 자릿수 크다」 비교는 근거가 안 됨 | 가정 붙인 HTTP 추정으로 강등 · 실현 가능한 요청 열 열거 · 비교 문장 삭제 |
| R2-12 | P3 | 전원 | **낡은 · 거짓 문장** — tasks 6.3 「브로커에 닿지 않는다」 · 6.1 「upstream 650」 · 3.7 보호 한정 누락(3.4a 와 모순) · 6.2 「시점 무변화를 diff 로」 · proposal `:1446` 잔존 · `notifier.go:139-141` 문장(base 에선 다른 코드) · issues 열린 질문 M1 현재형 · applyFloor 번들 「유일한 자리」(0 투영 보호 액션도 있음 — `ladder.go:446` · `snapshot.go:136-144`, 정확히는 「확정 하한이 손절을 0 으로 깎는 유일한 자리」) · B4 · B6(오류 반환)의 범위 미명시 | 문서 정정 |

### Manager 결정 요청

- **Q1 (R2-1)** 알림 꺼진 엔진에서 보호 0주의 등급 — (a) a095 식 게이트 / (b) 무게이트 + 비용 명시. a095 의 게이트가 사용자 결정 (2)였으므로 사용자 확인이 필요할 수 있다.
- **Q2 (R2-5)** 보유 0(엔진 밖 종결)으로 인한 하한 0 을 critical 에서 뺄지 — 8/2 원장 재열람(읽기 전용) 여부 포함.
- **Q3 (R2-6)** 원인 계약 — 원인별 키 vs 첫 원인 유지 명문화.
- 나머지(R2-2 · R2-3 · R2-4 · R2-7~R2-12)는 결정 없이 4판에서 반영 가능.

### Manager 판정 (2026-10-01) — Q1 · Q2 · Q3 (원문 요지)

- **Q1 = (a) 게이트, a095 문자 그대로.** 근거: a095 게이트는 정본(engine-safety 「등급은 사실이 정한다」 + enabled 조건) — 정본 선례를 따르는 데
  새 사용자 결정 불요, 거부권 보고만(Manager 가 사용자에게). 불변식 3: 알림 꺼짐(기본값) 엔진이 보낼 수 없는 critical 로 ENTRY_BLOCKED 로 가는 것은
  OFF = upstream 위반 방향. codex 세부 구분 채택: 게이트는 **enabled 플래그만** — 「켜짐 + transport 실패」의 래치는 의도된 a092 의미론이라 게이트
  대상 아님(design 에 명문화). 꺼짐일 때 사실은 normal 로 남는다.
- **Q2 = 배제 승인 + 8/2 원장 재독 승인(읽기 전용).** 보유 0(외부 종결) 원인의 0주는 정본 ClosedExternally = normal 과 같은 사실 — 원인 배제하고
  detail 에 담는다. 8/2 재독으로 「추론」을 측정으로.
- **Q3 = 에피소드 단일 키 유지 + 계약 문구 정정.** 원인별 키는 a092 에피소드 의미론과 충돌. 델타 SHALL 을 배달 가능한 것으로: 「에피소드 첫 원인
  (시각 포함)이 행 본문에, 각 관측의 원인은 구조화 로그 줄로」. 정정 경위 기록.
- R2-2 · R2-3 · R2-4 · R2-7~R2-12 는 결정 불요 — 4판 적용. 4판 → 재리뷰(보이스 표적 재확인 + codex 같은 세션) → 합본.

## 4판 (2026-10-01) — 2라운드 반영 · 판정 아님

| 항목 | 반영 | 자리 |
|---|---|---|
| R2-1 · Q1 | 알림 켜짐 게이트(생산 배선이 로드된 설정으로 덮음) · 「켜짐 + 전송 실패」는 게이트 밖(의도된 래치) · 불변식 3 문장 · 꺼짐에서 남는 흔적(옛 종류 normal · 로그, outbox 행 없음) | design D1 · engine-safety delta 문단 2 · Scenario 「알림이 꺼진 엔진」 · tasks 3.2a · 3.2b · 5.1 (iv) |
| R2-2 | exit-policy MODIFIED 「관측 경로와 fail-safe」 — 캡 문장 뒤 예외 한 문장 + Scenario. 정본 36 줄 중 치환 1(그 문장이 든 줄) | `specs/exit-policy/spec.md` |
| R2-3 | B2 오류 줄: 계좌 필드 없이 `MaskAccount` · 새 알림 계좌 없음 · 카나리 시험. `Notifier.escalate` 의 `FieldAccount` 원문은 **이름 붙인 잔여**(모든 critical 경로 공유 · 사람 결정 큐 「계좌 가림 설계」) | design D8 · tasks 3.8 |
| R2-4 | 루프 몫 이름 둘(「0주 기록」 · 「0주 기록 실패 승격」) · `n.mu` 보유자(`Acknowledge`) · 연결 풀 · 뒤쪽 포지션의 시세 수명 15s · 종료 취소 배제(호출자 ctx) · 래치 겹치지 않음 · 미실측 명시 + 실측 과제 | design D5 · D3 ④ · tasks 3.3b · 5.3 |
| R2-5 · Q2 | 8/2 원장 재독(측정): 보유 5 · 매도가능 0 · 엔진 밖 매도 10 → 5 → 2 → 0 — **8/2 는 보유 0 이 아니다**. 보유 0 판정 = `Bound == FloorBoundHoldings ∧ "0"`(동치 논증 · 표 시험, riskcalc 무편집) | design 「8/2 원장 재독」 · D3 ③ · tasks 3.3a · delta 근거 문단 |
| R2-6 · Q3 | 단일 키 · 행 본문 = 첫 원인 + 시각(한국어) · 관측마다 = 로그(`logEvent` 가 본문을 detail 로 씀 — `notifier.go:168-170`) · 원문 오류 본문 금지 | design D7 · delta 문단 4 · Scenario 「원인이 바뀌는 에피소드」 · tasks 3.10 |
| R2-7 | 재생 네 팔(정상 · 실패 · publisher 없음 · 꺼짐) + 배달 실행자 + 정착 행 1h 경계, 결과는 잰 수로 | tasks 5.1 · 5.1a · 5.2 |
| R2-8 | 5 액션 표(`Orderable()` 대조) · 관측 가능한 반환 회귀(레벨 해제 · 재발의) · 로그 캡처 하네스 · a085 문구 · 원문 오류 없음 · 참인 문장 단언 | tasks 3.4 · 3.5 · 3.7 · 4.1 |
| R2-9 | 범위 표를 경로 단위로(모드 통지 셋 추가) · `applyFloor` 한 호출의 기록 둘 겹침 시험 | design 범위 표 · tasks 3.9 |
| R2-10 | 소비자 표 보강(배달 실행자 판정 · `alerts ack` · `mode-release` · 콘솔 안내 · 런북) · 런북 과제 | issues · tasks 6.8 |
| R2-11 | 「≈98s」 삭제 → 실현 가능한 요청 열 + 「가정 붙은 HTTP 추정, 상한 아님」 · 호출 좌표 `:207` · `:231` · 비교 문장 삭제 | design D5 · 번들 calls |
| R2-12 | tasks 6.1 · 6.2 · 6.3 다시 씀 · 3.7 보호 한정(표 시험 3.4 가 익절을 가름) · proposal `:1446` · `notifier.go:139-141` 정정 · issues M1 현재형 취소선 · 번들 「확정 하한이 … 유일한 자리」 · B4/B6 · 0 투영 범위 밖 명시 | tasks · proposal · issues · design D3 · 번들 |

**정정 기록 — Q1 의 「원장 흔적 유지」**: 알림 꺼진 엔진의 옛 종류는 normal 이라 outbox 행을 만들지 않는다. 그 엔진에 남는 흔적은 구조화 로그 줄과
(이관 버퍼가 받으면) 최선 발송이다. 이 change 의 원장 흔적 이익은 **알림 켜진 엔진**에 한정된다(design D1). Manager 판정 문구의 뜻을 로그 흔적으로 읽었다.

**정정 기록 — Q3 의 SHALL**: 3판 delta 의 「원인은 세부 정보로 구분해 담아야 한다」는 미전달 행이 첫 내용을 유지하고(a097) 배달이 제목 · 본문만
보내므로 관측마다의 원인을 행에 담을 수 없는 문장이었다(2라운드 보이스 A #3 · codex #5). 4판은 그것을 「행 본문 = 에피소드 첫 원인 + 시각, 관측마다의
원인 = 구조화 로그 줄」로 바꿨다 — 둘 다 base 코드가 이미 배달할 수 있는 것이다(`outbox.go:285-345` · `notifier.go:168-170`).

**남는 판단 하나(Manager 보고)**: 8/2 는 보유 0 이 아니라 운영자 자신의 매도 주문이 매도가능을 잡은 경우였다 — Q2 의 배제에 들지 않으므로 알림 켜진
엔진에서는 critical 이다. 엔진은 외부 매도 주문이 잡은 매도가능 0 과 그 밖의 매도가능 0(브로커 보류 등)을 가를 입력이 없다(하한 계산은 로컬
미체결 매도만 안다). 4판은 그것을 critical 로 둔다(손절이 나가지 못한 사실은 참).

## freeze 재리뷰 3라운드 (2026-10-01, 표적 재확인) — 보이스 A REJECT(좁음) · 보이스 B APPROVE · codex FAIL

- 대상: `git archive 3c9bf7b0`(4판) → `/tmp/claude-1000/a091-r3-tree`, read-only. 보이스 A · B 는 2라운드 에이전트 재개(문맥 유지), codex 는 **같은 세션**
  (`01a0f3b4-…`) 재개(`-c sandbox_mode="read-only"`). 원문: `analysis/freeze-review/{claude-r3-voiceA.md, claude-r3-voiceB.md, codex-r3-prompt.md, codex-r3-output.md}`.
- **codex 머리말 `ACCESS-VIOLATION`**: 비교 명령의 셸 프로세스 치환이 `/dev/fd/63` · `/dev/fd/62` 를 읽었다(트리 밖 경로 = 파이프 fd). ~/.codex 접근 신고 없음. 조항대로 자진 신고 —
  내용상 무해(파일 시스템 밖 파이프)로 기록.
- 2라운드 대비 닫힘: R2-1(알림 꺼짐) · R2-2(exit-policy) · R2-6(원인 계약) · R2-7(재생) · R2-10(소비자) · R2-11(수치) · R2-12(낡은 문장) — 세 보이스 수렴.

| # | 심각도(최고) | 제기 | 내용 | 5판 처분 |
|---|---|---|---|---|
| R3-1 | P1 | A · B · codex | D3 ③ 「⟺」 거짓 — (⇐)는 매도가능 신선 · 로컬 유효일 때만. 새는 칸: 매도가능 조회 실패(B2) · 보유 스냅숏 낡음(`Now` 가 두 조회 뒤, 한계 10s) · 로컬 오류 | 한 방향 정확 + 조건부 반대 방향으로 정정, 새는 칸 셋은 **critical 유지(과보고 방향)** · 이름 붙인 잔여, 3.3a 칸 추가 |
| R3-2 | P1 | codex(A: 건전, 경합 P3) | `ctx.Err()` 만으로는 원인 출처를 못 가린다(진짜 오류 뒤 취소 · 판정 뒤 기록 전 취소) | 억제 = 하한 오류가 `context.Canceled` ∧ ctx 끝남(출처), 기록은 `context.WithoutCancel` — 판정 뒤 취소도 가짜 래치 0, 3.3b 네 팔 |
| R3-3 | **P0** | codex(A: P2) | 계좌 — 기록 실패 시 `o.alert` → `logErr` 원문(`exitloop.go:1818-1838`) · `Notifier.escalate` 두 줄 원문은 명명만으로 불변식 불충족 | a091 보고는 `o.alert` 를 거치지 않고 가린 실패 줄 · `escalate` 두 줄의 `FieldAccount` 제거(번들 `notifier.escalate` 신설) · 카나리 확장 |
| R3-4 | **P0** | codex(A: PARTIAL) | §0.3 — 수락 기준 · 배정 없음 | a092 `alertLoopShare` 750ms 대입 배정(근거 356.1ms 상위집합) + 수락 (i) 세 칸 × 두 몫 최악 ≤ 750ms(넘으면 멈춤) (ii) 뒤쪽 보호 포지션 750ms 지연 주입 시험(5.4) · 실패 경로 로그 두 줄 |
| R3-5 | P1 | codex · B | 알림 꺼짐 B2 에 normal Notify 를 더하면 base 와 다름 | 꺼짐 B2 = 알림 0 · 로그 한 줄(3.2a 호출 수 단언) |
| R3-6 | P2 | codex · A | 원장 재독이 추론을 측정으로 적음 · 「4분」은 10분 25초 · 한정 항 Sellable 은 로컬 매도를 가리지 못함 | 측정/추론 분리, 엔진 intent 0 건(읽기 전용 조회)을 로컬 매도 0 의 근거로, 시간 정정 |
| R3-7 | P3 | A · B | delta ¶1 열거 무조건 · exit-policy 새 Scenario 「잔여 pending 유지」 거짓(0주는 해제) · 3.4 `Orderable()` 는 술어 · 3.2b 하네스 라벨 · 5.1 훅 불요 · 콘솔 안내 문장 · ntfy `Tags` · B2 가림 범위 | 전부 반영 |

## 5판 (2026-10-01) — 3라운드 반영 · 판정 아님

위 표의 「5판 처분」 열이 전부다. 편집: design(header · 원장 재독 · D1 꺼짐 B2 · D3 ③ · ④ · D5 배정 · 수락 · 취소 · D8), engine-safety delta(열거 조건 · 취소 출처 ·
꺼짐 B2 · 로그 계좌 SHALL NOT · 근거 문장), exit-policy delta(Scenario THEN · 제외 문구), tasks(3.0 · 3.2a · 3.2b · 3.3a · 3.3b · 3.4 · 3.8 · 5.1 · 5.3 · 5.4 · 6.8),
issues(콘솔 안내 · ntfy), 번들 `internal-obs--notifier.escalate` 신설(base AST · 커버리지).

**범위 확장 하나 — Manager 보고**: R3-3 처분으로 a091 이 `internal/obs/notifier.go` `Notifier.escalate` 의 로그 두 줄을 편집한다(필드 하나 제거, 판정 무변경). a092 · a124
공유 경로이고 모든 critical 기록 실패 · 동기 발송 실패의 승격 로그가 바뀐다(계좌 필드가 빠짐). Manager 상임 지시(계좌 원문 로그 금지)와 codex P0 에 따른 것이며, 원하면
이 편집을 별도 change 로 떼어 a091 의 선행으로 둘 수 있다.
