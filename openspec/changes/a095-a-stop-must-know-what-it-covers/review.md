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

---

## 2라운드 (proposal-freeze, 2판) — **FAIL**

- **날짜**: 2026-09-25 · 측정 HEAD `634cf3c5`(브랜치 `feat/a112-four-family-runtime`) · base `ec29dc72`(HEAD 뒤 313커밋)
- **실행**: Opus Teammate(리뷰·기록만 — production 코드·테스트·proposal/design/specs 본문 무변경)
- **보이스 구성**: (A) 적대적 Eng — §4 손절 즉시성·부작용 예산·경합·fail-closed,
  (B) 소비자·폭발반경 — 운영자 화면·알림·원장·다른 change. 둘 다 Claude 서브에이전트, 서로의 결과와
  §1 판정을 보지 않게 격리(review.md 읽기 금지)·읽기 전용·`set -euo pipefail`·절대경로·toplevel 단언.
- **교차 보이스 미충족.** 두 보이스를 띄웠으나 **판정 기록 시점까지 결과가 도착하지 않았다**
  (둘 다 살아서 도구 호출 중이었다). Manager 지시로 보이스 결과 없이 **Teammate 단독 재검증**으로
  판정한다. 아래 발견은 전부 Teammate가 HEAD 코드·원장·AST로 직접 확인한 것이며 보이스 발견은 0건
  포함이다. 보이스 결과는 3판 리뷰 입력으로 넘긴다.
- **교차 모델**: `[codex-unavailable: Manager 지시]` — Codex는 사용자 계정으로 모델 호출을 하므로 쓰지 않았다.
  **a095 두 라운드 연속 미충족.**

### 2.3 게이트 명령 (판정은 rc, 파이프 없음)

| 명령 | rc | 비고 |
| --- | --- | --- |
| `openspec validate a095-a-stop-must-know-what-it-covers --strict --no-interactive` | **0** | valid |
| `python3 tools/logic-map/check_analysis.py --change a095-a-stop-must-know-what-it-covers` | **1** | ① AST 번들 4개 stale ② base가 313커밋 뒤라 working-tree 창에서 missing evidence 255건. **tasks 0.4의 "통과"는 HEAD에서 거짓**(그 라운드에는 참). 단 `tools/logic-map`은 다른 세션이 편집 중인 dirty 상태였다 — 도구 쪽 변화 몫은 가르지 않았다 |
| `sha256sum` 대조(AST 9개 `source_sha256`) | — | **MATCH 5 · STALE 4** (아래) |

**AST 산출물 stale 4개** — 구조(분기 수)는 같고 줄만 밀렸다(SHIFT). 좌표를 인용하는 문서 문장은 전부 낡았다.

| 번들 | 원인 커밋 | 문서 좌표 → HEAD |
| --- | --- | --- |
| `obs.Notifier.Notify` | a096/a097/a099·`a30eb35a` | `:107`/B1 `:111` → `:130`/B1 `:134` |
| `obs.Notifier.publishBestEffort` | 같음 | `:138`/B1 `:139` → `:161`/B1 `:162` |
| `obs.SeverityOf` | a099·a102 | `:309`/`:310` → `:347`/`:348`; `criticalEvents` `:279-298` → `:317-337`(여전히 18종); 이벤트 `:200` → `:238`; 인용 주석 `:190-194` → `:231-232`, `:212-217` → `:252-255` |
| `journal.resetExitStateForReadoptTx` | a111 `882a0b49` | `684-731` → `699-745` (호출자 여전히 1: `position_policy.go:145`) |

design D1의 `notifier.go:216-218`·`:283-285`·`:343`도 전부 다른 줄을 가리킨다. **재추출은 저자 몫이다(3판).**

### 2.4 §1.11 지시 여섯 항목의 반영 — 반영 기록(§2)을 믿지 않고 문서와 대조

| §1.11 | §2가 적은 반영 | 대조 결과 |
| --- | --- | --- |
| 1 총위험을 `baseline_price`로 · 값의 사본 전부 | 네 곳 모두 | **대체로 참.** 5행 산술 재계산 일치(3,200 · 19,128 · −7,220 · 5,060). 그러나 **475150 편입 수량 "3"이 틀렸다** — inst2의 편입 수량은 **2**(원장 `position_adoptions`; 3은 이미 닫힌 inst1). "3→32, 10.7배"는 실제 2→32(16배). proposal·D3·(tasks 5.1) 사본에 남았다(P2-1) |
| 2 `event.go` 주석 인용 + 명시적 뒤집기 | D1 새 절 | **부분.** 같은 이벤트에 대한 **셋째 주석** `exitwiring.go:87-89`(*"Normal, not critical: a person trading their own account by hand is not a malfunction, and a critical grade would make an undelivered alert about it block entries"*)를 지나쳤다. 반론의 정확한 시나리오(**transport 미설정·비활성**)에 답하지 않았다(P1-1) |
| 3 §0.3을 체류 시간으로 | tasks 6.2 재작성 | **부분.** 6.2만 고쳤다. **같은 약속의 사본**이 tasks 2.8(*"diff로 보인다"*)·안전 불변식 표 §4 행(*"6.2가 제출 경로에 새 호출 0개임을 보인다"*)·design D7 행에 그대로다. 선행 조건 a092는 그 속성을 약속하지 않는다(P0-4) |
| 4 `ExternalPositionFound` 오류 전파·latch | tasks 6.2a | **결정 없음.** 문제를 task로 옮겨 적었을 뿐 설계 결정이 없다. proposal Impact는 여전히 *"map에 한 줄. 함수 본문 무변화"*(P1-4) |
| 5 수량 억제 재발화 예산을 fold 수로 | 「2라운드 대상」 | **저자 미실행.** 이번 라운드가 쟀다(§2.7) — 그리고 재발화가 **outbox dedupe에 삼켜진다**는 것이 드러났다(P0-2) |
| 6 선후 관계 — a092 → a095, a091 합산 | tasks 재작성 | **부분.** *"논리 의존 없음/병합 순서만"*이 design `:124`·`:372`, proposal `:291`, tasks `:55`·`:166`에 살아 있고 tasks `:180`의 「합산」과 모순된다. proposal·design에 a092 언급 **0회**(P1-5) |

**정정의 단위가 또 값이 아니라 자리였다** — 1라운드와 같은 형태다(지시 1·3·6).

### 2.5 차단 — P0 4건

**P0-1 · FLM이 함수 경계에서 멈췄다 — R2-B2의 서사와 델타 SHALL이 거짓 전제 위에 있다**

`checkExternalIncrease`의 **유일한 호출자**는 `judgeHoldings`(`adoption.go:108-110`)이고
`if p.ExitEligible() { if p.Adopted() { d.checkExternalIncrease(ctx, p) } }`로 **편입된 포지션만** 부른다.
- FLM(`…checkexternalincrease/function-logic-map.md`)의 *"B2 — 없으면 조용히 반환한다. 엔진이 직접 연 포지션과
  미편입 보유가 여기로 온다"*는 **거짓**. 둘 다 이 함수에 도달하지 않는다.
- **010170**(adoption_id·entry_decision_id 모두 없음)은 `ExitEligible()`이 거짓이라 `alertUnmanaged`로 간다 —
  `exitloop.go:512-518`(무조건)과 `adoption.go:152-154`(fresh·비차단 시). 따라서 proposal·D5의
  *"R1만으로는 010170이 그대로 안 보인다"*도 **거짓**이다. R1만으로 010170은 critical·durable이 된다.
- B2가 실제로 받는 것은 **adoption_id는 있는데 JOIN 행이 없는 무결성 결함 또는 DB 오류**다
  (`journal/adoption.go:294-311`). 이것을 `alertUnmanaged`로 보내면 **exit state를 가진 포지션을 「무보호」로
  거짓 보고**한다.
- 진짜 구멍은 다른 곳이다: **엔진이 직접 연 포지션(entry decision)의 수량 증가는 어디서도 검사되지 않는다.**
  R2는 그것을 덮지 않는데 engine-safety 델타는 *"보호 상태가 고정한 수량보다 실제 보유 수량이 많으면
  critical"*로 **모든** 보호 포지션을 약속한다 — 델타가 설계보다 넓다.

**P0-2 · R2의 재알림은 outbox dedupe가 삼킨다 — 설계가 이벤트 키를 다루지 않는다**

base 이후 착지한 a096의 재알림 창이 R2를 무력화한다.
- `checkExternalIncrease`의 Key는 `exit.position_unmanaged|grown|<posID>` — **수량이 없다**(`adoption.go:457`).
- 같은 키의 전달된 행은 `DefaultRemindAfter = 1h`(`notifier.go:59`) 안에서 `claimOwed`가 not-owed →
  `ClaimSettled` → **전송도 본문 갱신도 없다**(`outbox.go:294-346`, `notifier.go:285-293`). 본문 UPDATE는 재무장
  (rearm) 때만 일어난다(`outbox.go:334-344`).
- **원장 실측**(inst2, 편입 수량 2): 증가 폴드 00:08:58(3) · 00:21:33(4) · 00:30:57(5) · 00:54:58(8) ·
  03:32:39(26) · 03:42:02(32). 키가 그대로면 전송되는 것은 3과 26 두 건이고 **「32」는 한 번도 나가지 않는다.**
  080220의 6건은 17분 안에 4건, 2분 안에 2건이다. 델타가 적은 *"한 번만 우는 보고는 32주를 말한 적이 없다"*가
  R2 후에도 그대로 성립한다.

**P0-3 · exit-policy 델타의 SHALL 전제가 HEAD에서 거짓이다 — 1라운드 정정 값의 사본이 남았다**

2판은 「유효 손절 = `baseline_price`」로 정정했다. 그런데 델타 `specs/exit-policy/spec.md`의
*"유효 손절가를 갱신하는 쓰기 경로가 하나임이 유지되어야 한다(SHALL — 오늘 그 경로는 재편입 하나뿐이고
운영자 행동에서만 불린다)"*, design D4 (1), issues I1 (1)은 옛 정의(네 기준 컬럼)로 쓰여 있다.
`baseline_price`의 writer는 **셋**이다:

| writer | 자리 |
| --- | --- |
| 래칫 판정 `recordExitJudgementTx` | `internal/journal/exit_state.go:553`·`:563` (`judgement.Baseline = effective.Line.CurrentProtection` `:541`) |
| 관측 갱신 | `internal/journal/exit_observation_refresh.go:137` |
| 재편입 `resetExitStateForReadoptTx` | `internal/journal/apply_hook.go:718-720` |

그리고 「하향 거부」는 이미 있다 — `notBelow("baseline", …)`(`exit_state.go:491`), snapshot 경로는
`SelectRecoverySnapshot`의 monotone 선택. 475150이 57,900으로 올라간 것 자체가 첫째 writer의 작동이다.
**델타가 SHALL로 적는 현재 상태가 거짓이면, 그 SHALL은 만족 불가이거나 무의미하다.**

**P0-4 · §0.3/§4 — critical 승격이 exit 루프의 손절 판정 앞에 최악 54s 체류를 넣는다. 선행 조건 a092는 그것을 약속하지 않는다**

- `ExitObserver.workingSet`이 `!p.ExitEligible()`인 보유마다 `o.alertUnmanaged`를 부른다(`exitloop.go:518`)
  — `observe`(`:441`)와 `judge` 루프(`:451-468`) **앞**이다.
- critical이면 `notifyCritical` → `claimAndDeliver`가 `n.mu`를 잡고(`notifier.go:254-255`) **같은
  goroutine에서** `deliver`의 재시도 예산 전체를 돈다(`:309`, `:420-574`).
- **최악 체류는 코드 상수로 54s다**(도출값): `AlertDeliveryBound` = 3 × (10s publish timeout + 5s busy) +
  2 × 2s + 5s release (`alert_lease.go:39-67`, `DefaultPublishTimeout` `:17`, `journal.DefaultBusyTimeout`
  `journal.go:36`). `Ntfy`는 `notifications.go:101`에서 Timeout 없이 만들어져 10s 기본을 쓴다.
- 무보호 보유가 N개면 첫 사이클에서 **N × 54s가 직렬로** 쌓인다. `c.Notifier`는 reconcile driver와
  공유되므로(`exitwiring.go:341-342`, `reconcileloop.go:367-368`) 상대편 deliver를 기다리면 **최대 +54s**.
- 비교 기준: exit 관측 주기 5s(`exitloop.go:98`), 청산 지연 상한 30s(`:116`), 관측 두절 60s(`:105`).
- **tasks 선후 관계는 「a092 → a095」라고 적지만** a092 20판 머리 블록은 *"8.7 … 발송이 exit goroutine에
  남는다 — 그대로 열림"*이고 기록 경로 선택(D0.3d 3번, 세 안)을 **고르지 않았다.** a095가 필요로 하는 속성 —
  「exit goroutine의 critical Notify는 deliver 없이 유계 시간에 반환한다」 — 을 a092가 **약속하지 않는다.**
  순서만 적은 의존은 이 체류를 막지 못한다.
- 같은 약속의 옛 사본(tasks 2.8, 안전 불변식 §4 행, design D7 「총위험 계산이 손절을 늦추는가」 행)은 여전히
  *"새 호출 0개를 diff로 보인다"* — 1라운드가 구조적으로 볼 수 없다고 한 그 약속이다.

### 2.6 P1 6건 (freeze 전 수정)

**P1-1 · 반론에 대한 답이 반론의 시나리오를 비켜 간다.** D1 (2) *"수동 매매마다 성립하지 않는다 — 억제는
상태를 따른다"*는 ① 재시작마다 메모리 latch가 재무장되고(`adoption.go:385-387` 주석 스스로), ② 수동으로 산
**새 종목**은 새 포지션 ID라 새 알림이며, ③ `ExternalPositionFound`는 latch가 없어서 막히지 않는다.
D1 (3)은 *"transport가 죽은 동안"*만 다룬다. `notifications.enabled=false`면 Publisher가 nil이고
(`notifications.go:83-88`), `deliver`는 즉시 `"no notification publisher is configured"`로 실패하고
(`notifier.go:429-431`) Gate.Block + `ModeEntryBlocked`로 승격한다. **알림을 끈 엔진은 수동 보유가 하나라도
있으면 진입이 막힌다** — 인용한 주석(`event.go:252-255`)이 말하는 바로 그 시나리오다.

**P1-2 · 정본과 충돌한다 — ADDED만으로는 부족하다.** 정본 `openspec/specs/exit-policy/spec.md:81`:
*"`adoption.enabled`의 기본값은 false이며(SHALL — … **false에서의 동작은 무관리 보유 알림을 포함한 기존
동작과 동일하다**)"*. R1은 false에서도 무관리 보유 알림의 등급과 그 결과(진입 차단)를 바꾼다. 델타가 이
문장을 MODIFIED하지 않으면 gate가 모순을 정본에 넣는다(안전 불변식 3 · WORKFLOW §0.2). 운영자가 **의도적으로**
`exclude_symbols`에 넣은 종목(`adoption.go:407-408` *"deliberately left unprotected"*)도 critical이 된다 —
이 교환도 적혀 있지 않다.

**P1-3 · 발신 네 자리가 「같은 사실」이 아니다(D7·I3의 판단이 틀렸다).** `notifierAlerter.ExternalPositionFound`
(`exitwiring.go:103-120`)는 `ExitEligible=true`인 **편입된(보호 중인) 포지션**의 재대사 폴드에도
*"관리 대상은 아니다 … 손절·익절이 자동으로 걸려 있지 않다"*로 발화한다(필드 `exit_eligible`만 참).
R1 뒤에는 **보호 중인 포지션에 대한 거짓 critical**이 된다. 또 세 자리(`adoption.go:419`,
`exitloop.go:1608`, `exitwiring.go:105-106`)가 같은 Key `exit.position_unmanaged|<posID>`를 써서 outbox의
**한 행으로 합쳐진다** — 원인 구분(델타 *"원인은 세부 정보로 구분해 담아야 한다"*)이 행 단위로 사라진다.
tasks 1.11은 미실행이다.

**P1-4 · `ExternalPositionFound`의 오류 전파에 결정이 없다.** 실측: ingest 오류는 `cycle.Err`로 기록되고
사이클은 **계속된다**(`reconcileloop.go:420-455`) — 1라운드 §1.5의 *"대사를 실패시킨다"*는 과장이다.
그러나 `note`의 연속 실패 계수에 들어가고(`reconcileloop.go:281-296`), 정본 engine-safety `:205`는 reconcile
연속 5주기 실패에 critical + ENTRY_BLOCKED를 건다. a095가 새 결합을 만드는지·받아들이는지 설계가 답해야 한다.
또 원장상 편입 후 증가는 **converge 경로**(kind `UNKNOWN`, `converge.go:213`)로 들어왔다 —
`ExternalPositionFound`는 EXTERNAL 폴드(0→N, `external.go:226`)에서만 운다. §1.6이 이 자리의 반복 발화로
적은 475150 예는 이 자리에 해당하지 않는다.

**P1-5 · 선후 관계의 옛 값이 다섯 자리에 살아 있다**(§2.4 지시 6). 또한 a091은 `EventExitProposalCapped`를
통째로 올리는 것을 **기각**하고(a091 proposal `:108`) 0주 경로만 critical로 올린다 — proposal·design의
*"a091은 `EventExitProposalCapped`를 등재한다"*도 틀렸다.

**P1-6 · 측정 시점이 지났다 — 동기 포지션이 전부 닫혔다.** 원장 재조회(2026-09-25, 읽기 전용):
010170 inst2 CLOSED 2026-08-13 · 475150 inst2·080220·066570 CLOSED 2026-08-11 · 272210 inst2 CLOSED
2026-08-13. **현재 OPEN은 TSLA 먼지 1건뿐.** tasks 7.2(*"6건 중 최소 2건 즉시 critical"*), 7.3, issues I5,
proposal 「지금 열린 것들」은 **지금 거짓**이다. `alert_outbox`는 현재 16행이 전부 critical이고
`exit.position_unmanaged`는 여전히 **0행**이다 — 결함 계열은 남아 있다. 동기는 유지되지만 예측·배포 서술은
측정 시각을 달고 다시 써야 한다.

### 2.7 §2.2 미완 항목 — 이번 라운드가 잰 것

| 항목 | 결과 | 성격 |
| --- | --- | --- |
| 수량 억제의 fold 수 재계수(§1.6) | 원장 전체에서 **편입 후 수량 증가 폴드 38건 / 포지션 15개**. 포지션당 최대 7(272210 inst5), 6(080220 · 466100 inst6 · 475150 inst2). 오늘 코드는 프로세스당 1건(≤15 + 재시작), R2 후 이론상 최대 38건 — **단 P0-2 때문에 실제 전송은 훨씬 적다** | **실측** (journal `mode=ro`, `position_adjustments` ⋈ `position_adoptions`, `created_at ≥ observed_at` · `new > prev` · `new > adopted`) |
| `deliver` 1회 체류 | 최악 **54s** (`DefaultAlertDeliveryBound`) | **코드 상수로 도출** — 실 운영 체류는 **미측정**(실 POST는 사람 승인 사안) |

### 2.8 비차단 P2

- **P2-1** 475150 편입 수량 3 → 실제 2 (§2.4 지시 1 행)
- **P2-2** proposal *"손절은 전부 `entry_price` 대비 정확히 −3.00%다"* — `initial_stop`에 대한 진술이다. 유효 손절은 5개 중 3개가 −3%가 아니다
- **P2-3** proposal *"그때 커지는 것은 주당 위험이 아니라 총위험"* vs issues I1 *"그때 커지는 것은 주당 위험"* — 서로 반대
- **P2-4** I2 보강: 272210 inst4 `avg_price = 81922.222222` — 브로커가 평단을 실제로 갱신하는 사례가 있다. 「stale일 수 있다」의 반대 사례도 보고 설계에 넣을 것
- **P2-5** 델타가 요구하지 않는 task(3.4 「미측정」, 3.4a 「고정된 이익」)가 있다 — 요구사항이면 델타로, 아니면 tasks에서 빼기

### 2.9 재검증 표 (Teammate 단독, HEAD `634cf3c5`)

| # | 주장(출처) | 판정 | 근거 |
| --- | --- | --- | --- |
| 1 | 5행 총위험 3,200/19,128/−7,220/5,060 (proposal·D3·3.1·I5) | 참 | 산술 재계산 |
| 2 | 475150 편입 3주, 10.7배 (proposal·D3) | **거짓** | `position_adoptions` inst2 = 2 |
| 3 | `criticalEvents` 18종·미등재 | 참 | `event.go:317-337` |
| 4 | 좌표 `:107/:111/:138/:139/:309/:310/:279-298/:200` | **거짓(stale)** | §2.3 표 |
| 5 | B2가 미편입·엔진 개설 포지션을 삼킨다 (proposal ⑤·D2·D5·FLM) | **거짓** | 호출자 가드 `adoption.go:108-110` |
| 6 | R1만으로 010170은 안 보인다 (proposal·D5) | **거짓** | `exitloop.go:512-518` |
| 7 | R2 재알림이 늘어난 만큼 보인다 (D2) | **거짓** | 키에 수량 없음 + 1h 재알림 창 |
| 8 | 유효 손절 writer는 재편입 하나 (델타 exit-policy·D4·I1) | **거짓** | writer 셋, §2.5 P0-3 |
| 9 | `resetExitStateForReadoptTx` 호출자 1 | 참 | `position_policy.go:145` |
| 10 | `alertUnmanaged`가 observe 앞 (tasks 선후·6.2) | 참 | `exitloop.go:518` vs `:441` |
| 11 | deliver가 `n.mu`를 잡고 재시도 예산을 돈다 | 참 (최악 54s) | `notifier.go:254-309`, `alert_lease.go:39-67` |
| 12 | a092가 선행 조건으로 체류를 유계로 만든다 (tasks 선후) | **거짓(미약속)** | a092 tasks 20판 머리 블록 8.7 「열림」 |
| 13 | 넷 다 같은 사실 (D1·D7·I3) | **거짓** | `exitwiring.go:103-120` 보호 중 포지션에도 발화 |
| 14 | a091과 논리 의존 없음 (design `:124`·`:372`, proposal `:291`, tasks `:55`·`:166`) | **거짓** | tasks `:180` 자신의 「합산」 |
| 15 | `ModeEntryBlocked`는 신규 진입만 막는다 (D1 (3)) | 참 | `notifier.go:397` 문구 |
| 16 | 교환은 「transport가 죽은 동안」뿐 (D1 (3)) | **거짓(불완전)** | 비활성 → Publisher nil → 즉시 실패 |
| 17 | `ExternalPositionFound` 오류가 대사를 실패시킨다 (§1.5) | **부분 거짓** | 사이클은 계속, 건강 계수만 |
| 18 | `ExternalPositionFound`가 475150 증가 폴드마다 운다 (§1.6) | **거짓** | 증가는 converge(UNKNOWN) 경로 |
| 19 | 열린 포지션 6건 · 010170 무보호 (proposal·7.2·7.3·I5) | **거짓(지금)** | 원장: OPEN = TSLA 1건 |
| 20 | outbox에 `exit.position_unmanaged` 0행 | 참 | 16행 전부 critical, 해당 0 |
| 21 | `check_analysis` 통과 (tasks 0.4) | **거짓(HEAD)** | rc=1 |

**참 7 · 거짓 14**(부분·stale 포함). 보이스 발견의 재검증은 0건(미도착).

### 2.10 3판이 받는 것

1. **FLM을 호출자까지 다시 뽑는다** — `judgeHoldings`·`ExitObserver.workingSet`·`notifierAlerter.ExternalPositionFound`의 AST를 더하고 stale 4개를 HEAD로 재추출한 뒤 base를 재고정한다. R2-B2를 폐기하거나 「무결성 결함」으로 다시 정의한다. **엔진 개설 포지션의 수량 증가 미검사**를 범위에 넣을지 결정하고, 델타 SHALL을 설계 범위와 같게 좁히거나 넓힌다
2. **R2에 이벤트 키 설계를 넣는다** — 수량을 키에 넣을지, 재무장 규칙을 쓸지 정하고, 475150 원장 순서(3·4·5·8·26·32)를 재생하는 시험으로 「32가 나간다」를 못 박는다
3. **exit-policy 델타의 「writer 하나」 전제를 고친다** — 유효 손절의 writer 셋과 이미 있는 하향 거부를 인용하고, 래칫 선행 조건 셋을 참인 문장으로 다시 쓴다(D4·I1 사본 포함, 값 단위로)
4. **§4 체류의 소유자를 정한다** — ① a092가 「exit goroutine의 critical Notify는 deliver 없이 유계 반환」을 **약속하고 착지한 뒤에만** R1을 착지(그 속성을 a095 tasks의 착수 조건·시험으로 인용), 또는 ② a095가 자체 완화(exit 루프 발신 자리를 normal로 두고 reconcile 쪽만 critical로 하거나, 발신을 judge 뒤로 옮기기)를 설계한다. 옛 약속 사본(tasks 2.8, §4 행, D7)을 제거한다
5. 셋째 주석 `exitwiring.go:87-89` 인용, **알림 비활성·미설정 엔진**과 `exclude_symbols` 종목에 대한 교환의 명시적 결정(P1-1)
6. 정본 exit-policy `:81`을 MODIFIED로 다루거나 R1을 `adoption.enabled`·알림 설정과 정합하게 좁힌다(P1-2)
7. 네 자리의 문구·키 분리(`ExternalPositionFound`가 편입된 포지션에서 「관리 대상 아님」이라 말하지 않게) — tasks 1.11을 실행해 증거로(P1-3)
8. `ExternalPositionFound` 오류 전파의 결정(P1-4), 선후 관계 옛 값 다섯 자리 제거와 a091 서술 정정(P1-5)
9. 원장 서술·배포 예측에 측정 시각을 달고 다시 쓴다(P1-6, P2-1~P2-5)
10. 3판 리뷰는 **두 보이스 결과를 실제로 받아** 교차 보이스를 충족한다

### 2.11 사용자 결정 대기

1. **a095의 착지를 a092에 묶을지.** a092가 「exit goroutine에서 동기 deliver 없음」을 약속·착지할 때까지 a095 R1을 동결할지, 아니면 a095가 자체 완화를 설계할지(3판 입력 4)
2. **알림 비활성·미설정 엔진과 `adoption.enabled=false`(기본값) 엔진에서 무관리 보유가 ENTRY_BLOCKED를 부르는 교환을 받아들일지.** 받아들이면 정본 exit-policy `:81`의 「false = 기존 동작」 SHALL을 바꾸는 결정이다(안전 불변식 3)
3. **범위를 옮길지.** 동기 포지션이 전부 닫힌 지금, 실측으로 드러난 진짜 구멍 — **엔진이 직접 연 포지션의 수량 증가 미검사**와 R2의 키 설계 — 으로 초점을 옮길지

### 2.12 이번 라운드가 하지 않은 것 (침묵한 생략 아님)

- 보이스 A·B 결과 수합 — 기록 시점에 미도착. 교차 보이스 미충족
- 교차 모델 — `[codex-unavailable: Manager 지시]`
- AST 재추출·문서 수정 — 저자 몫(3판). 이번 라운드는 판정과 지시만 남긴다
- `go test`·`make sdd-sync`·`make gate` — 문서 리뷰 단계, Go diff 0
- `deliver`의 실 운영 체류 — 54s는 도출값이다

### 2.13 착지 기록 — 이 절은 남의 커밋에 실려 들어갔다

위 2라운드 절(§2.3~§2.12)과 tasks.md 0.4·0.5 주석은 **`7582d1ee`**
(`docs(a092): 20라운드 적대 리뷰 — BLOCK …`, 2026-09-25 13:42:16)에 착지했다. 이 워크트리의 인덱스를
병행 세션과 공유하고 있어서, Teammate가 `git add <a095 두 파일>`로 스테이징한 직후 그 세션의 커밋이
같은 인덱스를 커밋했다. 내용은 Teammate가 쓴 그대로다(`git show 7582d1ee -- openspec/changes/a095-…`).
이력은 다시 쓰지 않는다. 이 주석을 경로 한정 커밋(`git commit -- <경로>`)으로 남겨 a095 라벨을 붙인다.
**교훈**: 공유 인덱스에서는 `git add` 다음 `git commit`이 원자적이지 않다 — 경로 한정 커밋을 쓸 것.

## 2.14 보이스 A(적대적 Eng) 결과 — 지연 도착, 2라운드 판정 뒤 (2026-09-25)

> 2라운드 판정(§2.3~§2.12)은 보이스 결과 없이 단독 재검증으로 냈다. 그 뒤 보이스 A 의 보고가
> Manager(tossos-d6)에게 도착했다. **3판 리뷰의 입력**으로 원문 그대로 보존한다(Manager 재검증 전 —
> 아래 주장은 리뷰어의 것이며 코드·원장 재검증은 3판 저자 몫이다). 보이스 B 는 미도착.

범위: 읽기 전용. 저장소 쓰기 0, 다른 review.md 미열람, 운영 원장·네트워크 접근 0. 코드 근거는 HEAD `634cf3c5`, base `ec29dc72`(2026-08-05).

### P0: 동결을 막는다 (안전 위반 또는 거짓 전제)

**A-1 · §4 손절 즉시성. R1이 늘리는 것은 호출 수가 아니라 exit goroutine이 관측 전에 머무는 시간이다. 최악 체류가 60초 두절 계약을 넘고, 넘어도 감지되지 않는다.**
- 경로: `ExitObserver.ObserveOnce` → `workingSet`(`exitloop.go:426`) → `!p.ExitEligible()` → `o.alertUnmanaged`(`:518`). 이 호출은 `o.observe`(`:441`)와 `judge`(`:465`)보다 **앞**에 있다. 이어서 `o.alert` → `Notify`(`notifier.go:130`) → `notifyCritical` → `claimAndDeliver`로 가고, 여기서 `n.mu`가 `:254`부터 deliver 전체를 덮는다.
- Notifier는 인스턴스 하나를 공유한다. `gateway.go:323` 하나를 exit 관측(`exitwiring.go:342`), reconcile(`reconcileloop.go:368`), ingest·converge(`gateway.go:332-338`)가 함께 쓴다.
- 상수 기준 최악 체류 (publish가 모두 타임아웃, 원장 쓰기마다 busy 대기): claim 5s(`journal.go:36`) · deliver 3×(10s `alert_lease.go:17` + 5s) + 2×2s(`notifier.go:48`) + 5s release = 54s(`alert_lease.go:52-54`, 자기 주석 "54s") · escalate 5s(잠금 밖, 같은 goroutine) → 합계 호출 1건당 약 **64s**. 다른 goroutine이 쥔 `n.mu`를 기다리는 시간(건당 ≤59s)은 여기에 더 붙는다.
- 원장이 정상이고 네트워크만 블랙홀이어도 3×10+2×2 = 34s. 같은 key의 PENDING 행은 발신자마다 예산 전체를 다시 쓴다(B-2 참조).
- 오늘은 `publishBestEffort` 1회, 10s, 잠금 없음.
- 64s는 `DefaultExitObservationOutage = 60s`(`exitloop.go:105`)를 미관측 포지션 **1건**만으로 넘는다. `checkOutage`는 `:422`·`:444`, 즉 관측 실패 경로에서만 불린다. 이 정체는 끝나면 관측이 성공하고 `lastObserved`가 초기화되므로 **조용히 지나간다**.
- 계획이 기대는 선행 조건은 약속된 속성이 아니다. tasks `:172-182`는 「a095는 a092 뒤」를 세운다. 그러나 a092 design D0.4 `:389-394`는 *"관측 goroutine은 그 뮤텍스를 계속 쥔다 … outbox 트랜잭션에는 여전히 기한이 없다 … 이것을 상한이라고 부르지 않는다"*; a092 tasks 20판 머리 블록 `:60-67`은 20라운드 **FAIL/BLOCK**, P0 2건 열림, 기록 경로 가·나·다 **미선택**. 이 의존은 proposal·design·spec 어디에도 없고, 막는 task도 없다.
- 계획 내부도 모순된다. tasks 6.2 `:125-129`는 "호출 수 diff로는 볼 수 없다"고 적는데, 같은 파일의 안전표 §4 `:194`는 여전히 "6.2가 제출 경로에 새 호출 0개임을 보인다"고 적는다. 6.2에는 합격 기준(상한 수치)도 없고 "실측"만 있다.
- **요구:** ① a092 착지와 체류 상한 수치를 spec SHALL과 task 게이트로 둔다. ② 또는 R1을 exit goroutine의 관측 전 발신 자리에서 떼어낸다. ③ 60s 두절 계약과의 관계를 명시한다.

**A-2 · 거짓 전제: 「수량이 늘면 증가분에 손절이 안 걸려 있다」.** 늘어난 수량은 이미 보호받고 있다.
- 청산 수량은 현재 보유에서 나온다. `snapshotContext`의 `RemainingQuantity: m.position.Quantity`(`exitloop.go:1087`) → `canonicalSnapshotContext`(`snapshot.go:263`) → `ProjectWholeShares(quantity, ratio)`(`:139`) → `submit`(`exitloop.go:1184,1301,1345`).
- 이벤트 본문도 스스로 그렇게 말한다: *"늘어난 수량은 원래 수량 기준으로 산정된 손절의 보호를 받는다"*(`adoption.go:461`).
- 틀린 문장: engine-safety 델타 `:14-20`(*"일부에 손절이 걸려 있지 않다"*) · design D1 `:142`·D7 `:369`, issues I3 `:103`(*"넷 다 같은 사실"*).
- 늘어나는 것은 무보호가 아니라 **총위험**이다. 이것을 `exit.position_unmanaged` critical로 알리면 운영자에게 거짓을 말하게 된다.
- **요구:** 수량 증가는 별도 이벤트 종류로 분리하고, 등급 근거를 다시 쓴다.

**A-3 · 거짓 전제: R2의 B2가 010170과 엔진이 직접 연 포지션을 삼킨다는 주장.**
- `checkExternalIncrease`는 `p.ExitEligible() && p.Adopted()`일 때만 불린다(`adoption.go:108-110`, 호출자 가드).
- `positions.adoption_id`는 `REFERENCES position_adoptions(id)`이고(`journal/adoption.go:92`) `foreign_keys(on)`이다(`journal.go:223`). 따라서 B2(`:446`)가 도달하는 것은 사실상 DB 오류나 ctx 취소뿐이다.
- 010170(편입 기록도 진입 결정도 없음)은 exit 루프에서 조건 없이 `alertUnmanaged`에 이미 도달한다(`exitloop.go:512-519`). 따라서 *"R1만으로는 010170이 안 보인다"*(proposal `:256-257`, design `:342-343`)와 D5 그림 `:327-329`는 틀렸다.
- 엔진이 직접 연 포지션에 수동 매수가 붙으면 수량 증가 검사가 **아예 없다**(`Adopted()`=false). R2도 그 구멍을 닫지 못한다.
- 델타 SHALL `:28-29`(*"그런 보유는 보호 상태가 없는 경우로 보고"*)를 구현하면, exit state가 있어 **보호받는** 엔진 포지션을 무보호라고 critical로 알리게 된다.
- FLM `checkexternalincrease/function-logic-map.md` Inputs 표의 B2 행도 같은 거짓을 담고 있다.
- **요구:** B2를 고치는 것을 철회한다. 엔진 개설 포지션의 수량 증가를 범위에 넣을지 명시적으로 결정한다.

**A-4 · 거짓 전제: 「유효 손절가(`baseline_price`)의 쓰기 경로는 하나(`resetExitStateForReadoptTx`)」.**
- 주석의 "four guarded columns"는 `taken_ratio_total, pending_action, pending_level, pending_intent_id`다(`apply_hook.go:503-505`). 손절 컬럼이 아니다.
- HEAD에서 `baseline_price`를 쓰는 자리는 **최소 4곳**: `exit_state.go:553`·`:563`(매 판정; 래칫이 475150을 57,900으로 올린 경로) · `exit_observation_refresh.go:136-137` · `apply_hook.go:718-720` · INSERT `exit_state.go:181`.
- 이 전제 위에 선 문장: exit-policy 델타 `:45-46`의 SHALL(*"오늘 그 경로는 재편입 하나뿐"*)은 오늘부터 거짓 · design D3 `:234-235`, D4(1) `:277-284`, issues I1(1) `:21-29`, tasks 3.5.
- tasks 3.2의 *"`policy_*`에 어떤 쓰기도 일어나지 않는다 — 구조로 고정"*은 기존 판정 경로와 충돌한다. `exit_state.go:563`과 refresh `:136`이 매번 `policy_id/version/digest`를 쓴다.
- 덧붙여, 계획이 모범으로 든 재편입 경로 자체가 `baseline`을 새 합성 손절로 **낮출 수 있다**.
- **요구:** 쓰기 경로를 AST로 열거한 뒤 SHALL을 다시 쓴다.

**A-5 · 거짓 전제: 「평단은 브로커가 원가를 안 주면 이어받을 수 있다」.** 수량 수렴 경로에서 평단은 **결정적으로** 낡는다.
- `ConvergeQuantities`는 늘 `NewAvgPrice: ""`를 보낸다(`converge.go:219-221`). 이것이 `firstNonEmpty`(`position_adjustments.go:312`)를 거쳐 `UPDATE positions … avg_price`(`:347-349`)가 된다.
- 외부 수량 변화가 평단을 갱신하는 경로는 없다. `avg_price` 쓰기는 이 자리와 체결 투영(`position_projection.go:322`)뿐이다.
- 결과: R3의 총위험은 R3가 겨냥한 바로 그 경우(수량 증가)에 **설계상** fold 시점의 평단으로 계산된다 · 475150의 "3,200"과 proposal `:20-58`의 물타기/불타기 표는 낡은 값에서 나왔다 · 사용자 보고 *"추가 구매하면 평균단가가 낮아지는데 반영하지 않는다"*는 원장에서 **문자 그대로 참**이다. proposal `:18` "절반은 반대"는 잘못된 진단이다.
- 틀린 문장: proposal `:282-285`, design `:246-247`, issues I2 `:79-83`(*"원장만으로는 가를 수 없다"*. 코드로는 가를 수 있다), exit-policy 델타 `:17-19`.
- tasks 3.3의 조건부 표기는 출처 컬럼이 없어 구현할 수 없다. issues I2가 스키마 변경을 범위 밖으로 둔다.
- 브로커 평단은 스냅샷에 있다(`Holding.CostBasisRaw`, `adoption.go:333`).
- **요구:** 평단의 출처를 다시 정하거나, R3를 보류하고 Why를 다시 쓴다.

### P1: 동결 전에 고쳐야 한다

**B-1 · R2 재알림은 전송되지 않는다.** key가 `…|grown|<id>`로 고정(`adoption.go:457`). DELIVERED 뒤 1h(`DefaultRemindAfter`, `notifier.go:59`) 안이면 `recordAlertTx`/`claimOwed`가 not owed → `ClaimSettled`(`notifier.go:284-293`)가 되어 전송도, 행 갱신도 없다. 재무장은 `outbox.go:295-345`에만 있다. `Notify`는 nil을 반환하므로 R2의 메모리 map은 새 수량으로 전진한다 — 창 안에서 늘어난 수량은 다음 증가나 재시작 전까지 **영영 전송되지 않는다**. 델타 `:24-26`과 Scenario `:49-51`은 정본 engine-safety `:674-676`의 SHALL NOT(창 안 재전송 금지)과 충돌. 계획 문서에서 a096·a098·a099·remind·event_key·claim·lease를 grep하면 **0건** — 계획의 Notifier 모형은 base(a096 이전) 상태다.

**B-2 · 발신 자리 셋이 event key를 공유한다.** `adoption.go:419`, `exitloop.go:1608`, `exitwiring.go:105`가 모두 `exit.position_unmanaged|<id>`를 쓴다. outbox 행은 하나이고, 먼저 온 내용만 전달된다. PENDING이면 발신자마다 예산을 따로 소진한다(A-1 누적). tasks 1.11은 문구만 대조한다.

**B-3 · fail-closed의 대가를 덜 열거했다.** design D1(2) `:58-62`의 「수동 매매마다 성립하지 않는다」는 거짓: `exitwiring.go:104`는 latch 없이 적용된 fold마다 발화(`external.go:278-281`) — 새 종목 수동 매수 1건마다 critical 1건 · R2는 수량이 늘 때마다 critical · exit 루프는 전이 상태를 조용히 두는 가드가 없다(`exitloop.go:512-519`, reconcile 쪽 `adoption.go:115-121`과 대조) — adoption 정상 편입에서도 편입 전 한 사이클 동안 critical · Publisher가 nil이면 즉시 `Gate.Block`과 모드 승격(`notifier.go:429-431,570-572,228`), 재시작 뒤 1h 지나면 다시 잠근다 · `adoption.enabled`=OFF 경로의 동작도 바뀐다(진입 게이트)인데 6.4는 "무도입"으로 비켜 간다(§3) · 등급을 normal로 고정하는 기존 시험 `exitloop_test.go:508-510`이 계획에 없다.

**B-4 · 델타와 설계가 모순된다.** 델타 `:24`의 SHALL NOT(*"프로세스 수명당 한 번만 보고해서는 안 된다"*)과 D6 `:361`/tasks 2.5(`d.unmanaged` 프로세스당 1회 유지)가 충돌. D2 `:202-204`의 *"outbox 재시도가 반복을 대신 책임진다"*는 거짓 — latch가 새 `Notify`를 막으므로 remind가 발화하지 않는다.

**B-5 · FLM이 함수 경계에서 멈췄고, 일부는 낡았다.** `source_sha256`이 HEAD와 다른 번들 9개 중 4개(severityof, notify, publishbesteffort, resetexitstateforreadopttx) — 본문은 base와 HEAD가 같아 **SHIFT**. 줄 인용은 모두 낡았다: Notify B1 :111→:134 · SeverityOf :309→:347 · criticalEvents :279-298→:317-337 · 이벤트 선언 :200→:238 · exitloop alertUnmanaged 정의 :1501→:1607 · 호출 :515→:518 · observe :443→:441. 안전 주장이 실제로 기대는 함수 가운데 `notifyCritical`은 base→HEAD **DIFF**(38→56줄). `claimAndDeliver`, `deliver`, `recordAlertTx`, `judgeHoldings`(호출자 가드), `ExitObserver.workingSet`, `IngestExternalPositions`, `ConvergeQuantities`, 판정의 baseline 쓰기는 번들이 **0개**. D1 사슬, D7, 6.2a, D4(1)은 손으로 읽은 주장이다.

**B-6 · SHALL 대상이 불명확하다.** exit-policy `:10`의 *"보고는 현재 총위험을 담아야"*는 어느 보고인지 경계가 없다. engine-safety `:14`의 *"보호 상태가 고정한 수량"*은 `exit_states`에 수량 컬럼이 없어 가리킬 대상이 없다(`core_domain.go:203-235`).

**B-7 · a091과의 합산을 아무도 소유하지 않는다.** tasks `:180-182`는 "먼저 가는 쪽이 측정"뿐. a091이 착지하면 `applyFloor`의 알림(`exitloop.go:1536`)이 `IssueReduction`(`:1357`) **앞**에서 critical `Notify`가 되고, a095의 reconcile goroutine 전달이 쥔 `n.mu`(건당 ≤59s)를 기다린다 → 청산 주문 제출 자체가 늦어진다.

### P2: 차단하지 않는다

- **C-1 · outbox 실측은 낡았고 결론을 뒷받침하지 못한다.** 실측은 2026-08-07로 a096·a098·a099 이전. `attempts=0`인 9/13행은 시도조차 안 된 행(전달 성공도 `attempts`를 올린다, `outbox.go:453`). DELIVERED 행 수가 없으므로 「호출」은 미입증. 운영 원장을 열 수 없어 **미확인**.
- **C-2 · 배포 직후 critical 수를 적게 셌다.** tasks 7.2는 "최소 2건"이라 적지만 080220(2→12)도 수량 증가 대상이라 최소 3건.
- **C-3 · 관리 중인 포지션에 무보호 문구가 나갈 수 있다.** `exitwiring.go:107-110`의 본문은 `ExitEligible`와 무관하게 "손절·익절이 자동으로 걸려 있지 않다"고 적는다. 도달 가능성 **미확인**.

**권고: FAIL.** P0 5건(§4 체류 1건 + 거짓 전제 4건). 거짓 전제 넷은 R1·R2·R3와 델타 SHALL 셋의 근거를 무너뜨린다.

## 2.15 보이스 B(소비자·폭발반경 렌즈) 결과 — 지연 도착 (2026-09-25)

> §2.14 와 같은 지위 — 2라운드 판정 뒤 도착, 3판 입력, Manager 재검증 전. 기준 HEAD `634cf3c5`, base `ec29dc72`.
> 운영 원장 미열람(원장 수치의 현재성은 전부 미확인). review.md 미열람. **두 보이스가 서로 보지 않고 같은 거짓 전제
> 셋(B2 도달 불가 · 증가분은 보호받는다 · `baseline_price` writer 는 하나가 아니다)에 수렴했다.**

### P0 — freeze 차단

**B-P0-1 · 기본 설정 엔진은 무관리 보유 하나만 있어도 신규 진입이 영구 차단된다.** R1이 들어가면 알림 기본값(off)과 `adoption.enabled=false`(기본값)인 엔진이 손으로 산 보유 하나를 발견하는 순간 durable `ENTRY_BLOCKED`로 올라간다. `exclude_symbols`에 의도적으로 넣은 종목도 똑같이 막힌다. 이 상태를 운영자가 풀 production 표면을 찾지 못했다.
- 증거: `internal/config/notifications.go:31-34` "the zero value is `Enabled: false` … it wires no publisher" · `notifier.go:429-431` `if n.Publisher == nil { lastErr = …; break }`(재시도 없음) → `:570-572` `Gate.Block` → `:223-228` `owed && !sent → n.escalate` → `:382` `EscalateOperatingMode(…CriticalAlertUndelivered)` · `adoption.go:389-391` "It fires regardless of `adoption.enabled`" · `adoption.go:127-129` exclude 종목도 `unmanaged`, 사유 `:408` "deliberately left unprotected" · 해제: `Acknowledge`(`notifier.go:840-880`)는 `ReasonAlertUndelivered` 래치만 푼다, 운영 모드 완화는 OPERATOR·승인·audit 필요(`operating_mode.go:102-114`), 비테스트 `TransitionOperatingMode` 호출자는 `EscalateOperatingMode`(`:504`) 하나, CLI `engine_alerts.go:233-236` "⚠ 진입은 아직 막혀 있다".
- 정본 exit-policy "`adoption.enabled`의 기본값은 false이며(SHALL — false에서의 동작은 무관리 보유 알림을 포함한 기존 동작과 동일하다)" 와 충돌, 안전 불변식 3(토글 OFF 동등성)과 충돌. 같은 결정을 명시한 선례: `exitwiring.go:87-89`, a091 `design.md:70-73`.
- 틀린 문장: design:66, 69(최악을 「transport 가 죽은 동안」으로만) · design:47, tasks:155(「`Acknowledge`로만 해제」— 거짓) · tasks:136-137(6.4 「토글을 도입하지 않는다」— 기존 토글 둘의 OFF 동작 변경을 다루지 않음) · exclude/include 는 계획 전체에 0회.
- 요구: 등급을 이벤트 종류가 아니라 **사실**로 가를 것. 운영자가 선택한 상태(off·exclude)는 critical 에서 뺄 것. 또는 정본 exit-policy SHALL 을 MODIFIED 로 바꾸고 사람 승인. ENTRY_BLOCKED 를 푸는 실제 경로 명시.

**B-P0-2 · engine-safety 델타가 엔진이 직접 연 모든 포지션에 대해 거짓 critical 을 요구한다. R2-B2 의 전제도 거짓이다.**
- 델타 `specs/engine-safety/spec.md:28-29` "편입 기록이 없다는 이유로 수량 증가 검사를 건너뛰어서는 안 된다 … 보호 상태가 없는 경우로 보고되어야 한다(SHALL)", 시나리오 `:53-55`. 엔진 진입 포지션은 편입 기록이 없지만 exit_state 와 손절이 있다(정본 exit-policy t0). 문자 그대로 구현하면 엔진이 연 모든 포지션에 거짓 critical, P0-1 과 겹치면 엔진이 자기 진입을 스스로 막는다.
- 호출 조건 `adoption.go:108-111` `if p.ExitEligible() { if p.Adopted() { d.checkExternalIncrease(ctx, p) } … }`; B2 `:446` 은 `adoption_id` 있는 포지션에서만, `REFERENCES position_adoptions(id)`(`journal/adoption.go:92`) + `foreign_keys(on)`(`journal.go:223`) → 「기록 없음」(`ErrAdoptionNotFound`)은 구조적으로 도달 불가. 010170 은 이미 `adoption.go:135-137 → :152-155` `alertUnmanaged` 와 `exitloop.go:512-518` 로 간다.
- 틀린 문장: proposal:167-168 · :219-220 · :256-257, design:342-343(「R1만으로는 … 010170은 여전히 안 보인다」— 거짓, R1 만으로도 critical) · design:183-184.
- 요구: R2-B2 와 해당 SHALL·시나리오 삭제 또는 엔진 진입 포지션 명시 제외. 010170 이 오늘 왜 조용한지 코드로 재확인.

**B-P0-3 · 「넷 다 같은 사실」이 거짓 — 수량 증가 알림의 증가분은 손절 보호를 받는다.** `adoption.go:461` "늘어난 수량은 원래 수량 기준으로 산정된 손절의 보호를 받는다" · 청산 수량은 현재 수량(`exitloop.go:1087` `RemainingQuantity: m.position.Quantity`) · 정본 「편입 후 외부 포지션 … 전량 청산이 RISK_REDUCING 의도로 발의」. 커지는 것은 R 의 크기이지 덮지 못하는 수량이 아니다. 틀린 문장: design:142 · :369, issues:103 · 델타 engine-safety:16-17 · D1(3) · D1(2)(「수동 매매를 아무리 많이 해도 한 번도 더 울지 않는다」는 R2 와 모순 — `event.go:252-255` 가 경고한 기전이 매수 방향으로 재현). 요구: grown 알림 별도 이벤트 종류·등급 별도 판단, tasks 1.11 을 freeze **전에**(결과는 이미 「다르다」).

### P1 — freeze 전 수정 필요

**B-P1-4 · R2 의 「다시 보고」는 정본 재알림 창 SHALL NOT 에 막힌다.** 정본 engine-safety `spec.md:674-676` · `DefaultRemindAfter = time.Hour`(`notifier.go:59`) · `outbox.go` `claimOwed` DELIVERED/ACKNOWLEDGED 창 안 → `ClaimSettled` · 키 `…|grown|<pid>`(`adoption.go:457`) 수량 무관 → 12주 알림 후 1시간 안에 32주면 미전송. 틀린 문장: 델타 engine-safety:24-26, design D2:166-170, tasks 2.4(가짜 notifier 로는 통과). a096~a099 는 계획 전체 0회. 요구: 키 설계를 정본 SHALL NOT 과 대조해 결정, 키를 늘리면 행마다 ack·차단이 붙는 비용도 적을 것.

**B-P1-5 · 두 발신 자리가 같은 키를 써서 사유 행렬 SHALL 이 퇴행한다.** `exitloop.go:1608`·`adoption.go:419` 둘 다 `exit.position_unmanaged|<pid>`; exit 루프(5초)가 일반 문구로 행을 선점 → 대사 쪽 사유 행렬 본문(exclude·include 실패)은 창 안 `ClaimSettled` 미전송. 정본 exit-policy 사유 행렬 SHALL/SHALL NOT 과 충돌. 틀린 문장: design D2 :191-199. 요구: 두 자리의 키·문구 정리, dedupe 결과를 시험으로.

**B-P1-6 · exit 루프 자리는 전이 상태 무알림을 지키지 않는다.** `exitloop.go:512-518` 에 RECONCILE·신선도 판정 없음(대사 쪽 `adoption.go:115-121` 은 "transition states: silent"). `adopt()` 가 시세 실패면 빈 집합(`adoption.go:179-186`) → `:143-150` 후보 전원 `alertUnmanaged`. 편입 성공 뒤에도 PENDING "관리하지 않는다" 행을 a098 배달 루프가 계속 보낸다 → `event.go:241-247` 계약 위반. 요구: 편입 예정·전이 상태의 critical 동작과 편입 뒤 outbox 행 처분을 설계에.

**B-P1-7 · 네 번째 발신 자리(`exitwiring.go:104`)는 production 에서 도달 불가 — tasks 6.2a 전제 거짓.** `IngestExternalPositions` 의 production 호출자는 `reconcileloop.go:424` 하나이고 `:338` `d.ingest.Alert = nil`. 도달 가능한 자리의 오류는 삼켜진다(`reconcileloop.go:552-559`, `exitloop.go:1706-1712`). 어댑터 본문은 `exit_eligible` 을 무시하고 항상 "손절·익절이 자동으로 걸려 있지 않다"(`external.go:85-91` 주석과 어긋남). 틀린 문장: tasks 6.2a(:130-134), proposal:207-208, design:133-140 「발신 자리 4곳」. 요구: 도달 가능성 기준 재계수, 6.2a 삭제/재작성.

**B-P1-8 · 손절 경로 체류와 a092 의존이 문서마다 반대로.** `workingSet`(`exitloop.go:493`) 안 `:518` 동기 `Notify` 가 `observe`(`:441`)보다 먼저; `claimAndDeliver` 는 `n.mu` 잡은 채(`:254-255`) 최대 3×10s(`DefaultPublishTimeout`, `alert_lease.go:17`)+2×2s = 34s → 재시작 직후 transport 죽어 있으면 무관리 N개 × 34s 동안 첫 손절 관측 지연. 대사 goroutine critical 도 같은 `n.mu` 경합. a092 미착지. 틀린 문장: proposal·design 은 a092 선행 미선언(tasks:172-174 만) · design D7:370 · tasks 안전표 :194 ↔ 6.2(:125-127) 모순. 요구: proposal 에 a092 착지를 hard gate 로, 안전표 정정.

**B-P1-9 · a091 과 논리 충돌 — 「논리 의존 없음」 거짓.** a091 `design.md:68` 「`EventExitPositionUnmanaged`는 승격하지 않는다」(근거 `:70-73`); a091 미착지(미체크 30); a091 이 선택지 B 로 착지하면 표가 19종 → a095 tasks 2.2a 「19종(18+1)」 하드코딩이 병합 순서에 따라 틀린다. 틀린 문장: design:124·:372, proposal:290-291 ↔ tasks:180-182 모순; proposal:210-213 은 a091 핵심 반대 근거(게이트 영구 잠금) 미대응. 요구: a091 반대 결정 인용·뒤집는 근거, 19 는 +1 로.

**B-P1-10 · 인용 줄번호·AST 산출물이 HEAD 기준 낡았다.** `source_sha256` 불일치: notifier `d5b3…`→`0bc7…`, event `acf3…`→`7732…`, apply_hook `88af…`→`459f…`. HEAD 위치: event.go `:200→:238`, `:279-298→:317-337`(18종 유지), `:309→:347`, `:190-194→:228-237`, `:212-217→:248-256`; notifier.go `:107/:111→:130/:134`, `:138/:139/:142→:161/:162/:165`, `:216-218→:378-398`, `:283-285→:570-572`, `:343→:840`; exitloop.go `:1501→:1607`, `:515→:518`, `:443→:441`; apply_hook.go `:684→:699`. design D1 도식(:40-48)에 a096 재알림 창·a097 claim 실패 래치·a098 배달 루프·a099 lease·10초 publish timeout 없음. 요구: base 재고정·FLM 재생성.

**B-P1-11 · exit-policy 델타 「유효 손절 쓰기 경로는 하나」 SHALL 거짓.** `baseline_price` 는 래칫(`exit_state.go:553`, `:563`)·관측 refresh(`exit_observation_refresh.go:137`)가 쓴다; `apply_hook.go:694` 는 "the only **reset** writer for the four guarded execution-time columns". 계획이 인용한 475150 의 57,900 도 래칫이 쓴 값. 틀린 문장: 델타 exit-policy:45-46, proposal:240-243, design D4(1):277-281, issues I1(1) — 이대로 SHALL 이면 정본 Baseline Ratchet 과 모순. 요구: 「쓰기」→「reset」 정정, SHALL 문언 수정.

**B-P1-12 · spec 델타와 tasks 불일치.** (a) 평단 불확실성 표기: 델타 exit-policy:17-19·시나리오 :25-27 SHALL ↔ issues I2:81-87 「원장만으로는 가를 수 없다 … 스키마 변경 필요」— 구현·시험 불가한 SHALL. (b) 총위험 보고 범위: Requirement 1 은 모든 보고로 읽히는데 tasks 는 3.1 하나만. (c) 「보호 상태가 고정한 수량」(engine-safety:14): exit_states 에 initial_quantity 없음(`core_domain.go:206-223`), task 없음.

### P2 — 비차단

- **B-P2-13:** 「로그도 없다」(proposal:140, design:104)는 오해 — `Notify` 가 등급 판정 전에 `logEvent`(base `:109`, HEAD `:132`), 콘솔 `console/portfolio.go:39-47` 이 `관리 외(미편입)` 표시. 「durable outbox 와 재시도가 없다」로 적을 것.
- **B-P2-14:** 「transport 가 있다 — 4/13행 전달 시도」만으로는 성공 불명(`attempts>0` 은 전부 실패여도 성립). 원장 현재성 미확인.
- **B-P2-15:** tasks 7.2 예측은 7주 전 원장 기반(미확인); 080220(2→12) 도 grown; exit 루프·대사 이중 발화 누락.
- **B-P2-16:** D6 「outbox 와 재시도가 반복을 대신 책임진다」는 PENDING 행에만 참 — 전달된 행은 프로세스 래치로 재관측 없음.

**권고: FAIL.** P0 셋은 거짓 전제이면서 안전 폭발반경(기본 설정 엔진의 진입 영구 차단 · 보호 중 포지션에 거짓 critical · 정본 SHALL 둘과 충돌). 등급을 이벤트 종류에서 **사실 단위**로 다시 설계하기 전에는 freeze 할 수 없다.
