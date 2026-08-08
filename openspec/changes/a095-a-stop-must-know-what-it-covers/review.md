# a095 · review

## 1라운드 (gstack plan-eng-review) — **FAIL**

### 1.0 교차 모델 — **미충족** (a095 1라운드)

Codex를 outside voice로 돌렸고 **사용량 한도로 출력 0바이트**였다
(`ERROR: You've hit your usage limit … try again at Aug 8th, 2026 12:36 PM`).
gstack 폴백대로 Claude 서브에이전트를 돌렸다 — **fresh context이지 다른 모델이 아니다.**

**a092 여섯 라운드 + a094 두 라운드 + a095 = 아홉 라운드 연속 미충족.**

### 1.1 증거 사슬에서 성립한 것 (먼저 적는다)

- AST 산출물 9개가 **문서보다 먼저**다(13:37 vs 13:49~13:54). a094 1라운드 §1.6-8이
  잡았던 순서 역전이 여기서는 없다
- `criticalEvents` 18종 · `EventExitPositionUnmanaged` 미등재 · `SeverityOf` 분기 1개 ·
  `Notify` B1 `:111` · `publishBestEffort` B1 `:139` — **전부 실물과 일치**
- `alert_outbox` 13행 전부 critical · `exit.position_unmanaged` 0행 — **실측 재확인**
- `checkExternalIncrease` 분기 3개 중 B1·B2 미진입 — **커버리지 프로파일로 재현**
- `resetExitStateForReadoptTx`가 유일 reset writer이고 운영자 행동에서만 불린다 — 확인
- `escalate`가 막는 것은 **신규 진입뿐**이다(`notifier.go:229` `ModeEntryBlocked`) —
  D7의 주장이 맞다

### 1.2 차단 P0 — **핵심 증거표가 틀린 컬럼으로 계산됐다. 대표 사례를 18배 과대보고한다**

R3의 공식은 `(평단 − **유효손절**) × 수량`인데, 측정은 `initial_stop`을 썼다.

**원장 실측 (2026-08-07)** — `baseline_price`가 유효 손절이다:

| 종목 | 평단 | `initial_stop` | **`baseline_price`(유효)** | a095가 적은 총위험 | **실제** | 배수 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| **475150** | 58,000 | 56,163 | **57,900** | **58,784** | **3,200** | **18.4×** |
| 080220 | 79,000 | 77,406 | 77,406 | 19,128 | 19,128 | 1.0× |
| **272210** | 68,100 | 66,154 | **69,905** | 7,784 | **−7,220** | 부호 반전 |
| 066570 | 181,600 | 176,540 | 176,540 | 5,060 | 5,060 | 1.0× |
| **TSLA** | 300.01 | 305.55 | **326.97** | ≈0 | 음수 | — |

**5행 중 3행이 틀렸다.** 475150은 이미 본전으로 승격돼 있었고(a094 proposal이 같은
원장에서 *"손절선 56,163 → 57,900(본전)"*을 인용한다) 272210은 손절이 진입가 **위**에
있어 이익이 고정된 상태다 — a095는 그것을 *"물타기 방향이라 −2.86%"*로 적었다.

**과소보고를 없애겠다는 change가 자기 대표 사례를 18배 과대보고한다.** 그리고 같은
475150 행이 `R/주 1,737`(= entry−initial_stop)과 총위험 58,784(= avg−initial_stop 기반)를
**한 줄에 함께** 싣는다 — 주당 수가 두 개고 둘 다 유효 손절이 아니다.

**정정의 단위는 좌표가 아니라 값이다.** `initial_stop`을 「유효 손절가」로 쓴 것이
proposal · design D3 · tasks 3.1 · issues I5에 **전부 복제됐다.**

**추가 미다룸**: 유효 손절이 평단 위면 총위험이 **음수**다. tasks 3.4는 「못 읽는 경우」만
덮고 「음수」를 덮지 않는다.

### 1.3 차단 P0 — **코드가 이 등급을 normal로 둔 이유를 문서가 인용조차 하지 않았다**

`internal/obs/event.go:190-194`, 대상 이벤트 자신에 대해:

> *"Normal — **somebody trading their own account by hand is not a malfunction.**"*

그리고 형제 이벤트(`:212-217`)가 기전을 적는다:

> *"Normal, for the same reason the fold is: a person selling their own shares is not a
> malfunction, and **grading it critical would mean an engine with no alert transport
> configured stops opening positions every time its owner takes a profit by hand.**"*

**D6·D7 어디에도 이 두 주석이 없다.** D1은 `SeverityOf`의 주석
(*"Genuinely critical conditions are named in the table above"*)만 인용하고, **그 표가 이
이벤트를 뺀 이유를 적어 둔 자리를 지나쳤다.**

기전은 실재한다: `deliver` 실패 → `Gate.Block`(`notifier.go:283-285`) → `escalate` →
`EscalateOperatingMode(ModeTriggerCriticalAlertUndelivered)` → `ENTRY_BLOCKED`,
`Acknowledge`로만 해제(`notifier.go:343`).

**`.claude/CLAUDE.md`상 침묵한 생략이다.** 등급을 올리려면 이 결정을 **명시적으로 뒤집고
그 근거를 적어야** 한다 — 지나친 것과 뒤집은 것은 다르다.

### 1.4 차단 — R1이 관측 주기에 **동기 정지**를 넣는다. 그리고 tasks 6.2는 그것을 볼 수 없다

`alertUnmanaged`는 `workingSet` 안 `exitloop.go:515`에서 불린다 — **`observe`(`:443`)와
포지션 판정(`:453-465`) 전이다.**

critical로 올리면 `publishBestEffort`(fire-and-forget)에서 `notifyCritical` → `deliver`로
옮겨 가고, `deliver`는 **`n.mu`를 잡은 채 재시도 예산 전체를 돈다**(`notifier.go:241`,
`:251-272`). 예산은 `DefaultCriticalAttempts = 3`(`:45`) · `DefaultRetryDelay = 2s`(`:48`).

**tasks 7.2가 배포 즉시 ≥2건이 운다고 예측한다.** 그동안 손절이 평가되지 않는다.
`c.Notifier`는 공유 인스턴스이므로(`exitwiring.go:341-342`) ReconcileDriver 쪽 알림도
같은 뮤텍스에서 exit 루프를 막는다.

**tasks 6.2는 *"제출 경로에 새 호출이 0개임을 diff로 보인다"*고 약속한다.
호출 수 diff는 이것을 구조적으로 볼 수 없다** — 호출이 늘지 않고 **기존 호출의 등급이
바뀔** 뿐이다. a094 1라운드 차단 7과 같은 형태의 약속이다.

### 1.5 차단 — R1이 reconcile의 오류 표면을 바꾼다. *"map에 한 줄, 함수 본문 무변화"*가 아니다

`Notify`는 **critical outbox 쓰기 실패에서만** 오류를 반환한다(`notifier.go:104-107`).
normal은 항상 nil이다.

`notifierAlerter.ExternalPositionFound`(`exitwiring.go:98-103`)는 그 오류를 그대로
reconcile의 `alertErrs`로 넘기고 `Run`이 `errors.Join(alertErrs...)`를 반환한다
(`internal/reconcile/external.go:278-295`).

**즉 `EnqueueAlert` 실패가 이제 외부 포지션 대사를 실패시킨다 — 종전에는 불가능했다.**

겹쳐서: 네 발신 자리 중 **이 자리만 latch가 없다.** 나머지 셋은 메모리 latch를 갖는데
(`exitloop.go:1496`·`adoption.go:393`·`:442`) `ExternalPositionFound`는 **적용되는 fold마다**
발화한다(`external.go:277-281`). **issues I3는 넷을 「같은 사실」로 묶고 이 차이를 적지
않았다.**

### 1.6 차단 — R2가 R1이 durable하게 만든 자리의 유일한 rate limit을 제거한다

`d.grown`을 수량 키로 바꾸면 **수량 증가마다** critical 1건이다. 475150은 3→32주를
브로커 보고 fold **여러 번**에 걸쳐 옮겼다. tasks 7.2의 「최소 2건」은 **포지션 수를 세고
fold 수를 세지 않는다.** 각 재발화가 outbox 1행 + §1.4의 뮤텍스 점유다.

### 1.7 선후 관계가 틀렸다

- **a095는 a092보다 먼저 갈 수 없다.** a092가 존재하는 이유가 손절 경로의 동기 알림
  체류를 유계로 만드는 것이다. a095 R1은 **그 경로를 타는 알림의 모집단을 늘린다.**
  proposal의 *"a092와 겹치지 않는다 — 알림 체류"*는 **거꾸로다. a092가 a095의 선행 조건이다**
- **a091 + a095는 합산된다.** 둘 다 같은 `criticalEvents` map에 행을 더한다. 두 계획 다
  이것을 「텍스트 충돌」이라 부르는데, 실제로는 **직렬화된 `deliver` 예산에 두 종류가 더
  들어오는 것**이다. **어느 계획도 그 합을 소유하지 않는다**
- **a089의 분모가 a094 R2 아래에서 움직인다**(a094 판정 참조)

### 1.8 좌표 오류

| 문서 | 실물 |
| --- | --- |
| proposal: 동결 주석 `adoption.go:437-441` | 주석은 `437-440`, `441`은 함수 시그니처 |
| D2: ``alertUnmanaged`(`adoption.go:392-436`)` | 함수는 `392-432`, `433-436`은 `checkExternalIncrease` doc |

### 1.9 매니저가 독립으로 찾은 것

- **R2의 수량 억제에 래칫 구멍**: 32 → 20 → 32면 저장값이 32라 `32 > 32`가 거짓이라
  **재발화하지 않는다.** 시험 목록에 없다
- **`alertUnmanaged` 사유 분기 6개는 전부 진입**하고 why-matrix가 옳다 — 안 바꾸는 결정은
  유지 가능하다

### 1.10 막힌 시도 — 설계가 실제로 막은 것

1. **`escalate`가 손절을 막는가** → **아니다.** `ModeEntryBlocked`,
   *"new entries are blocked"*(`notifier.go:229-233`). D7의 주장이 맞다
2. **`resetExitStateForReadoptTx`의 호출자가 하나인가** → **그렇다**
   (`position_policy.go:145`). D4가 자동 갱신을 미룬 근거는 성립한다
3. **`EvaluateLadder`의 모든 선이 `entry_price`에서 나오는가** → **그렇다**
   (`:358` `percentOf`, `:387`·`:503-509` `lockPrice`). D4의 파급 논거는 성립한다
4. **AST가 문서보다 먼저인가** → **그렇다**(mtime 13:37 vs 13:49)

### 1.11 2라운드가 받는 것

**FAIL. 차단 6건.**

1. **총위험을 `baseline_price`(유효 손절)로 다시 계산하고 5행 표를 전부 고친다.**
   475150 = 3,200, 272210 = 음수. **값의 사본을 proposal·design D3·tasks 3.1·issues I5에서
   같이 고친다** — 좌표만 고치면 틀린 값이 살아남는다. **음수 총위험 처리도 정한다**
2. **`event.go:190-194`·`:212-217`을 인용하고 그 결정을 명시적으로 뒤집는다.**
   *"alert transport 없는 엔진이 소유자의 수동 매매마다 진입을 멈춘다"*는 반론에
   답해야 한다 — 지나친 것으로는 안 된다
3. **§0.3을 호출 수가 아니라 체류 시간으로 다시 약속한다.** `deliver`가 `n.mu`를 잡는
   구간을 재고, `alertUnmanaged`가 `observe` **전에** 있다는 사실을 다룬다.
   **a092의 결과에 의존한다면 그것을 선행 조건으로 적는다**
4. **`ExternalPositionFound`의 오류 전파와 latch 부재를 다룬다**(§1.5)
5. **수량 억제의 재발화 예산을 fold 수 기준으로 다시 센다**(§1.6·§1.9)
6. **선후 관계를 고친다** — a092 → a095, a091과의 합산

### 1.12 확인하지 못한 것 (침묵한 생략 아님)

- **교차 모델 미충족** — Codex 사용량 한도(2026-08-08 12:36 복구)
- `go test ./...` 미실행 (이번 판은 문서만, Go diff 0)
- `make sdd-sync`·`make gate` 미실행 (`mutating: true` — 사람이 승인)
- `deliver` 1회의 **실 체류 시간** 미측정 — publisher 타임아웃을 재지 않았다.
  구조(뮤텍스 × 3회 × 2초 대기)는 확인했고 **총 시간은 미측정**이다
- a087·a089·a091·a092 delta 본문 미대조

---

## 2판 — 1라운드 지시의 반영 (판정 아님)

**판정이 아니라 반영 기록이다.** 2라운드 리뷰는 아직 돌지 않았다.

| §1.11 지시 | 반영 |
| --- | --- |
| 총위험을 `baseline_price`로 재계산 + **값의 사본 전부** | proposal 실측표 · design D3 표 · tasks 3.1 · issues I5 **네 곳 모두**. 475150 **58,784 → 3,200** |
| 음수 총위험 처리 | tasks **3.4a** — 「위험」이 아니라 **「고정된 이익」**으로 보고. 272210(−7,220)·TSLA |
| `event.go:190-194`·`:212-217`을 인용하고 **명시적으로 뒤집는다** | design D1 새 절 — 두 주석 원문 인용 + `notifyCritical → deliver → escalate → ModeEntryBlocked` 사슬 + **뒤집는 근거 셋** |
| §0.3을 체류 시간으로 다시 약속 | tasks **6.2** 재작성 — *"호출 수 diff는 구조적으로 볼 수 없다"*를 명시 |
| `ExternalPositionFound`의 오류 전파와 latch 부재 | tasks **6.2a** |
| 수량 억제의 재발화 예산 | tasks 7.2에 유지 · fold 수 기준 재계수는 2라운드 대상 |
| 선후 관계 — a092 → a095, a091과 합산 | tasks 「선후 관계」 재작성. **1판의 *"a092와 겹치지 않는다"*가 거꾸로였음을 적었다** |

### 2.1 정정이 서사를 바꾼 곳 (지시에 없던 것)

**`baseline_price`로 다시 재니 「손절선이 얼어붙어 있다」가 틀렸다.**
열린 다섯 중 **셋**(475150·272210·TSLA)은 래칫이 이미 손절을 올렸다.

| 종목 | 평단 | 유효손절 | 유효손절/평단 | 승격 |
| --- | ---: | ---: | ---: | --- |
| 066570 | 181,600 | 176,540 | −2.79% | 아니오 |
| 080220 | 79,000 | 77,406 | −2.02% | 아니오 |
| 475150 | 58,000 | 57,900 | **−0.17%** | **예(본전)** |
| 272210 | 68,100 | 69,905 | **+2.65%** | **예** |
| TSLA | 300.01 | 326.97 | **+8.99%** | **예** |

**얼어붙은 것은 손절선이 아니라 `entry_price`** — 모든 레벨이 파생되는 기준점이다.
proposal의 「불타기 방향」 절과 `issues.md` I1을 그에 맞게 고쳤다:
**초판이 걱정한 475150의 −3.17%는 존재하지 않는다.**

**남는 결함은 총위험의 무상한 증가와 보고 부재이며, 그것은 정정 뒤에도 그대로다.**

### 2.2 아직 안 한 것

- **2라운드 리뷰 미실행**
- **교차 모델 미충족** — Codex 사용량 한도(2026-08-08 12:36 복구)
- 수량 억제의 fold 수 기준 재계수(§1.6)는 tasks에 반영만 하고 **수를 다시 세지 않았다**
- `deliver` 1회의 **실 체류 시간** 미측정 (구조만 확인)
- FLM·AST 재생성 안 함 — a095는 이번 판에서 대상 함수가 늘지 않았다
