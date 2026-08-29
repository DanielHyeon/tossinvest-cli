# a098 review

## 1. 구현 후 독립 리뷰 (2026-08-13) — §6.4

- 대상: base `e6c4636a` → HEAD `5fd7b3f7` 전체 diff (33파일 · +5597 · −1)
- 보이스 둘, **병렬**, 각각 독립 프로세스(`codex exec`, read-only, high reasoning)
  - **A** — 적대적 staff engineer (안전 불변식 · 잠금 범위 · fail-closed · 소켓 · 종료 순서)
  - **B** — 적대적 reliability/QA (**단 하나의 질문**: 그 테스트가 실제로 증명하는가,
    아니면 깨진 구현에서도 통과하는가)
- High-risk 등급(진입 게이트 · 원장 · 운영자 표면)이므로 적대적 Eng 관점 필수 — 충족

### 판정

| 보이스 | 판정 | 발견 |
|---|---|---|
| A | **BLOCK** | BLOCKER 3 · CRITICAL 2 · HIGH 5 · MEDIUM 2 · 약한 테스트 10 |
| B | **BLOCK** | BLOCKER 5 · HIGH 5 · 테스트 전수 감사 40여 건 · 커버리지 공백 목록 |

**둘 다 BLOCK.** 그리고 **둘 다 자기 headline 을 틀렸다** — 아래 §1.2 가 실측이다.

> ⛔ **리뷰의 지적은 「어디를 볼지」이지 「무엇이 참인지」가 아니다.**
> 그래서 판정을 그대로 옮기지 않고, 판정을 가르는 주장마다 **뮤테이션을 돌렸다.**
> 이 절의 모든 CONFIRMED/REFUTED 는 추론이 아니라 그 실행 결과다.

### 1.1 두 보이스가 독립적으로 같은 것을 찾은 것 — 가장 강한 신호

| 수렴 | A | B | 내 판정 |
|---|---|---|---|
| 프로덕션 조립 경로가 **이름만** 검사된다 | 약한테스트 5 | P1 | **부분** — 이름 검사는 맞고, 결론은 틀렸다(§1.2 M2) |
| 무한 목록이 원장 연결 하나를 독점한다 | P1 | P9 | **부분 CONFIRMED** — 사실은 맞고 등급은 과장(§1.4) |
| 영구 실패는 정지로 안 세어 sender-down 이 안 걸린다 | P6 | P7 | **CONFIRMED**(§1.4) |
| 취소가 **publish 중일 때**는 아무도 안 잰다 | P3·약한테스트 3 | P5 | **CONFIRMED** — 공백 |
| publish 후 crash 는 임차 만료 뒤 중복을 낳는다 | P11 | 약한테스트 | **CONFIRMED, 의도된 at-least-once**(§1.5) |

### 1.2 ⛔⛔ 가장 무거운 발견 — B-P2, 그리고 그것을 **뮤테이션이 확인했다**

B: *「발송자 사망이 게이트를 잠근다」는 프로덕션 생성자를 통과해 검증되지 않는다.
R3·R15·R17 은 테스트가 손으로 지은 `a098DeliveryAux` 의 `OnStop` 을 쓴다.*

실측했다. `Context.AlertDeliverer`(`auxiliary.go:178-180`)에서 `OnStop` 을 지운다:

| 뮤테이션 | 패키지 | 결과 |
|---|---|---|
| **M1** — 프로덕션 `OnStop` 삭제 | engine · execgw · cmd/tossctl · obs | **네 개 전부 ok ⛔** |
| **M1b** — `OnStop` 은 두고 본문을 비움 | 같음 | **초록** (M1 과 같은 결과) |

**그 구현에서 배달 실행자는 죽고 진입 게이트는 열려 있다.** High-risk fail-closed
경로에서 아무도 안 잡는 뮤테이션이므로 「완료」가 아니다.

> ⛔ 처음 돌린 M1 은 `FAIL` 을 냈고 그것이 **테스트가 잡은 것처럼 보였다.**
> 실제로는 `gate`·`log` 가 미사용이 되어 `[build failed]` 였다.
> **build 실패는 테스트가 잡은 것이 아니다** — `_, _ = gate, log` 를 넣어 컴파일시킨
> 뒤에야 진짜 답(초록)이 나왔다. `missing-tool-reports-clean` 의 반대 방향 형태다.

### 1.3 같은 형태의 구멍 둘을 더 찾았다 (B-P3 · B-P10, 둘 다 실측)

| 뮤테이션 | 무엇을 바꾸나 | 결과 |
|---|---|---|
| **M3** | `deliverOne` 이 보내는 알림의 `Title`·`Body` 를 **빈 문자열**로 | **초록 ⛔** |
| **M4** | `AlertOperations.Acknowledge` 가 `ids...` 를 **버린다** | **초록 ⛔** |

- **M3** — 「배달됐다」는 원장의 상태고 운영자가 받는 것은 **내용**이다. 둘을 안 가르는
  테스트는 제목도 본문도 빈 정지 실패 통보를 배달로 센다.
- **M4** — a098 의 승인 테스트는 **전부** id 를 안 넘긴다(= 전체 승인). 그래서 한 건
  승인이 backlog 전체 승인이 되어도 안 보인다. 걸쇠가 풀리고 원장에는 그 사람 이름이
  남는다 — 감사 기록이 거짓이 되는 형태다.

### 1.4 ⛔ 두 보이스의 headline 은 **둘 다 틀렸다**

| 보이스의 결론 | 실측 | 판정 |
|---|---|---|
| B: *「프로덕션이 무력한 실행자를 등록해도 통과한다」* | **M2** — 프로덕션 `Run` 을 `<-ctx.Done()` 으로 무력화 | **FAIL** — R4·R4b 가 죽인다. **REFUTED** |
| A: *「무한 목록이 정지 판단을 지연시킨다」(BLOCKER)* | 목록 비용은 §3.2 측정으로 100행 0.449ms → 1만 행 ≈ 45ms. 운영자가 눌러야 일어나고, 서버가 5초에서 끊는다 | **사실은 맞고 등급은 과장.** BLOCKER 아님 |

B 는 `cmd/tossctl/a098_the_engine_registers_a_sender_test.go` 만 보고 「아무도 안 
지나간다」고 결론지었다. 실제로는 `a098_the_backlog_does_not_delay_protection_test.go:143`
이 **프로덕션 생성자로** 실행자를 세우고 그것이 실제로 보내기를 요구한다 — 그 자리를
지나가라고 지난 세션에 일부러 그렇게 쓴 것이고(뮤테이션 NN), 그것이 여기서 값을 했다.

### 1.5 CONFIRMED — 실재하나 a098 이 만든 것이 아니거나 후속인 것

| # | 발견 | 판정 |
|---|---|---|
| A-4 | 배달 성공 후 **승인 없이** 재시작하면 미배달 걸쇠가 풀린다 (`UndeliveredCount` 는 PENDING 만 센다) | **CONFIRMED, 회귀 아님.** a098 이전엔 `restoreAlertEntryLatch` 자체가 없어 재시작이 **항상** 걸쇠를 지웠다. a098 은 구멍을 **좁혔다**. 완전히 닫으려면 `acknowledged` 개념이 필요 — 후속 |
| A-6·B-7 | `Run` 은 cycle 오류를 **영원히 기록하고 계속한다**. 반환·패닉만 `OnStop` 에 닿는다 | **CONFIRMED.** sender-down 걸쇠는 *실행자가 반환하는가*를 키로 삼고 *진전이 있는가*는 안 본다. 영구 실패 발송자는 게이트에서 건강한 것과 구분 불가 — 후속 |
| A-7 | 독행(poison row) 열 개가 새 알림을 **영구히 굶긴다**. `ORDER BY id LIMIT 10`, 시도 상한 없음, backoff 없음 (`outbox.go:517-523`) | **CONFIRMED.** 완화: 동기 경로가 새 알림을 직접 보내고 outbox 는 대체 경로다. 회귀 아님(이전엔 아무도 안 뺐다). dead-letter 필요 — 후속 |
| A-11 | publish 성공 후 crash → 임차 만료(81초) 뒤 **재발송** | **CONFIRMED, 의도된 at-least-once.** 다만 `alertdelivery.go:32-40` 의 *"releasing here would let the next cycle send"* 는 과장이다 — 붙잡음은 중복을 **막는 게 아니라 81초 미룬다**. 문서 정확성 문제 |
| A-2 | 알림 OFF 여도 실행자가 무조건 등록된다(`cmd/tossctl/engine.go:389`). `Publisher == nil` 가지가 행마다 claim+release **두 번 쓰고** 진전 없이 매 주기 반복 | **CONFIRMED.** 불변식 3 은 안 건드린다(동기 경로·게이트 무관). 소음과 낭비다. 값싼 수정: 청구 **전에** publisher 를 본다 — 후속 |
| B-9 | 목록은 무한인데 서버는 5초에서 끊고(`alert_control_transport_unix.go:136`) 클라이언트는 1MiB 에서 자른다 | **CONFIRMED, 자기모순.** 클라이언트 주석은 *"무한이라 5초보다 넉넉하게 준다"*(30초)고 적는데 **서버가 5초다.** 「전부 보여 준다」가 세 곳에서 거짓이 될 수 있다 — 후속 |
| B-4 | 두 발송자 테스트가 원장 핸들 **하나**를 쓴다. `SetMaxOpenConns(1)` 이라 청구는 필연적으로 순차다 | **부분 CONFIRMED, a098 것이 아니다.** 겹침은 **publisher 안에서** 만들어지므로 실행자의 청구는 실제로 원장에 닿는다(§1.6). 다만 **핸들 둘을 여는 테스트는 저장소에 없다** — `a099_claim_excludes_the_second_sender_test.go` 도 자기 핸들을 안 연다. 임차는 a099 소유 → **a099 후속** |

### 1.6 REFUTED — 확인해 보니 아닌 것

| 발견 | 왜 아닌가 |
|---|---|
| A-8 *「ntfy 오류 문자열이 secret 을 원장에 적는다」* | `notifier.go:494`·`:781` 이 **이미** `err.Error()` 를 그대로 넘긴다. §5.2 가 `internal/obs` diff **0파일**을 쟀으므로 base 에도 있다. a098 은 세 번째 호출자일 뿐 — 회귀 아님(저장소 차원 과제로는 유효) |
| A-12 *「핸들러가 도는 중에 원장이 닫힌다」(순서)* | `defer` 는 LIFO 다. `ectx.Close()` 가 **먼저**(217), `alertControl.Close()` 가 **나중**(285) 등록되므로 실행은 소켓 → 원장 순이다. **순서는 맞다.** 잔여는 `Close` 의 2초 상한뿐 |
| A-5 *「ack 재생 공격」* | 토큰은 프로세스마다 새로 나고 소켓은 같은 UID 0600 이다. 재생할 수 있는 자는 **새 요청도 만들 수 있다** — A-9 로 흡수, 독립 발견 아님 |
| A-9·A-10 *「같은 UID 면 운영자 이름을 위조할 수 있다」·「못 본 알림을 승인한다」* | 사실이고, 저장소의 **기존 신뢰 모델**이다(control socket 셋 모두 같다). A-10 은 불변식 8 과 감사 의미 사이의 설계 긴장이며 `alertops.go:21-31` 이 이미 그 근거를 적는다 |
| B-1 headline | §1.4 — M2 가 죽는다 |

### 1.7 내가 따로 확인한 것 (세 번째 표본)

- **불변식 3 — 기계로 확인.** `ReasonAlertSenderDown` 은 진입만 막는다.
  `Gateway.checkEntry`(`execgw/gateway.go:853-859`)가 `!plan.raisesExposure` 면
  **즉시 nil** 이고, `CheckEntryFor` 는 그 한 곳에서만 불린다. **정지·청산은 게이트를
  안 지나간다.** 두 보이스 모두 이 확인을 안 했다.
- **R19 는 강한 테스트다.** 핸들 하나를 쓰지만 겹침을 **publisher 안에서** 만든다 —
  동기 경로가 청구를 끝내고(연결 반납) publish 에서 멈춘 사이 실행자의 청구가 원장에
  닿는다. `a099 R1`(핸들 공유로 순차만 증명)과 다른 형태다. 양방향이고 합계 1을 센다.
- **D0.10 경계 문장은 여전히 낡았다** — *"a098 은 `Flush` 를 얼마나 자주 부르는가를
  정한다"*. a098 은 `Flush` 를 안 부른다(D1.1). §6.5 에 적었고 a092 인계 목록에 있다.

## 2. 1라운드 처리 (2026-08-13)

### 2.1 닫은 것 — 새 파일 하나, 프로덕션 변경 0줄

`internal/app/engine/a098_the_production_executor_carries_its_stop_handler_test.go`

| 테스트 | 죽이는 뮤테이션 |
|---|---|
| `TestTheProductionDelivererCarriesItsStopHandler` | **M1**(처리기 삭제) · **M1b**(빈 처리기) |
| `TestTheDeliveredNotificationCarriesTheRecordedAlert` | **M3**(제목·본문 비움) |
| `TestAcknowledgingOneAlertLeavesTheOthers` | **M4**(부분 승인 → 전체 승인) |

RED 증거는 뮤테이션 자체다 — 넷 다 새 테스트 아래에서 **죽는 것을 실행해서 봤다**.

```text
M1  → 프로덕션 실행자에 정지 처리기가 없다 — 발송자가 죽어도 진입이 열려 있다
M1b → 정지 뒤에도 critical_alert_sender_down 가 안 걸렸다: map[]
M3  → 나간 알림의 제목 = "", 원장의 제목 = "STOP_UNSERVED: 005930"
M4  → 한 건 승인 뒤 남은 미배달 = 0, want 1
```

새 파일에 넣은 이유는 `tossos-logic-map-scope-creep` 이다 — 기존 파일에 새 함수를
넣으면 그 파일의 **바로 위 함수**까지 logic-map 기준이 옮겨 간다.
프로덕션 코드는 **한 줄도 안 고쳤다.** 세 뮤테이션 전부 *증거*의 결함이지 *구현*의
결함이 아니다.

**GREEN 재측정: 2868 passed / 5 패키지 / exit 0** (engine · execgw · cmd/tossctl ·
obs · journal). gofmt 는 `$(go env GOROOT)/bin/gofmt` 로 돌려 새 파일 위반 0.

### 2.2 후속으로 넘긴 것 — 이 change 에서 안 고친다, 그리고 그것을 적는다

프로덕션 동작 변경이고 각각 Function Logic Map 이 먼저 필요하다. a098 의 scope 는
「밀린 것을 보내는 주체를 만든다」이고 아래는 그 주체의 **운영 품질**이다.

| # | 무엇 | 어디로 |
|---|---|---|
| F1 | 알림 OFF 일 때 claim **전에** publisher 를 본다 (A-2) | 새 change |
| F2 | 진전 없는 발송자를 sender-down 으로 센다 (A-6·B-7) | 새 change |
| F3 | 독행 dead-letter / 시도 상한 (A-7) | 새 change |
| F4 | 목록 상한과 서버 5초의 자기모순 해소 (B-9) | 새 change |
| F5 | 배달 성공 후 재시작이 승인을 건너뛴다 (A-4) | 새 change (`acknowledged` 필요) |
| F6 | 핸들 둘로 임차 배제를 재는 테스트 (B-4) | **a099** 후속 |
| F7 | `alertdelivery.go:32-40` 의 「막는다」→「81초 미룬다」 (A-11) | 문서, 이 표로 갈음 |
| F8 | publish 중 취소를 재는 테스트 (A-3·B-5) | F2 와 같은 change |

> ⛔ **왜 지금 안 고치나.** F1~F5 는 전부 `cycle`·`deliverOne`·`Run` 의 **내부 분기**를
> 바꾼다. `.claude/CLAUDE.md` 는 기존 함수 내부를 고치면 Function Logic Map 과 Branch
> Test Map 을 **먼저** 만들라고 하고 High-risk 는 면제가 없다. 리뷰 끝물에 그것을
> 건너뛰고 넣는 것이 `flm-before-claiming-not-before-editing` 이 적은 실패다.
> 침묵한 생략이 아니라 **여기 적은 이월**이다.

### 2.3 판정

**BLOCK 두 건은 해소됐다.** 두 보이스의 headline 은 실측으로 각각 REFUTED·강등됐고,
그 아래에서 진짜였던 증거 구멍 셋(M1·M3·M4)은 닫혔다. 남은 CONFIRMED 는 전부
(a) 회귀가 아니고 (b) 불변식 3 을 안 건드리며 (c) §2.2 에 이름이 적혔다.

불변식 3 은 §1.7 에서 **기계로** 확인했다 — 정지·청산은 진입 게이트를 안 지나간다.

**a098 은 §6.3 gate 로 갈 수 있다.**
