# a135 design — attestation 만료 · 범위 상실 사전 알림

## Context

- 인터록: 게이트 ON 기동은 `verifyGate`(`internal/app/engine/interlock.go:478-552`) 의 B4 `:501`(`attest.Load` 실패) · B6 `:518`
  (`att.Verify(now, accountRef, RequiredEndpoints())` 실패)에서 거절된다(AST: `analysis/code-context/internal-app-engine--verifygate.ast.json`).
  읽는 파일은 `AutomationStatus.AttestationFile`(`attestationPath`: 설정값 또는 `<config>/capability-attestation.json`).
- 실행 중 엔진은 attestation 을 다시 읽지 않는다. 그래서 만료 · 범위 상실은 **다음 재시작**에서야 드러난다.
- `obs.EventCredentialExpiring`(critical, 등급표 등재)은 선언만 있고 생산 발행자가 0(`grep` 비시험: `event.go` 선언 · 등급표뿐).
- 콘솔 attestation 칸(a063, `internal/console/data.go` `readAttestation`)은 남은 시간 · 72시간 경고 · 누락 endpoint 를 이미 보여 준다.
- a063(열림)은 갱신 단위의 프로필 정렬 · 거절 가시화 · 설치를 맡는다. 이 change 는 그 범위(단위 · 갱신 명령 · soak 기준)를 건드리지 않는다.
- a134(병행)는 엔진 **밖** 생존 감시. 이 change 는 엔진 **안** — 살아 있는 엔진이 다음 재시작의 실패를 미리 말한다. 엔진이 죽어
  있으면 이 감시도 없고, 그것은 a134 가 알린다. 둘은 겹치지 않는다.
- 운영 실측은 `analysis/verify/live-attestation.md`(만료 10-23T01:39Z · POST 포함 · 갱신 거절 중 · 감독 증거 10-09 기간 밖).

## Goals / Non-Goals

Goals: 재시작 전에 운영자가 (a) 만료 임박, (b) 만료, (c) 갱신이 필요 범위를 잃은 것, (d) 파일이 무효인 것을 ntfy 로 안다.
엔진 · 손절 루프 · 진입 판정에 영향 0.

Non-Goals: 갱신 실행 · 예약 · 기준 변경 · 단위 설치(a063), 감독 검증 실행(사람), 콘솔 변경(a063 이 이미 표시), 인터록 의미 변경,
「갱신하면 POST 를 잃을 것」의 사전 예측(D6).

## Decisions

### D1. 판정은 인터록의 술어 그대로 — `judgeAttestation`

`attest.Load` + `Attestation.Verify(now, Context.AccountRef, RequiredEndpoints())` 를 그대로 부르고 오류의 sentinel 로 단계를 고른다:

| Verify 결과 | 단계 | 등급 |
|---|---|---|
| Load 실패(ErrMissing · ErrMalformed · ErrFormatTooNew) · ErrIncomplete | `unreadable` | critical |
| ErrExpired | `expired` | critical |
| ErrAccountMismatch | `account_mismatch` | critical |
| ErrEndpointNotAttested | `scope_lost`(누락 = `MissingEndpoints(RequiredEndpoints())`) | critical |
| 통과, D-N(N = 남은 시간 하루 단위 **올림**) N ≥ 8 | `healthy` | 없음 |
| 통과, 3 ≤ N ≤ 7 | `notice` | normal |
| 통과, 1 ≤ N ≤ 2 | `imminent` | critical |

(구현 리뷰 P3 반영: `ErrIncomplete` 는 `unreadable` 과 갈라 `incomplete` 단계 — 「파일이 없다」는 문구가 「endpoint 칸이 비었다」에 붙지 않게.)

근거: 술어를 다시 쓰면 두 판정이 갈린다(「판정이 둘이면 반증이 죽는다」). 올림인 이유: 「7일 남음」을 D-7 로 읽고, 7일 1초는 아직 D-8.
Verify 의 순서(필수 칸 → 만료 → 계좌 → endpoint)를 따르므로 만료 + 범위 상실이 겹치면 `expired` 로 보고된다(둘 다 재시작 거절 —
고칠 것은 갱신 하나).

인터록의 다른 조항(Guardian · 한도 · 정책 · 게이트웨이)은 판정하지 않는다 — 실행 중에 바뀌지 않는 조립 사실이거나 설정이다.

### D2. 알림 — 타입으로 등급, 단계마다 한 번

- normal 단계 → 새 `obs.EventCredentialExpiryNotice`(`broker.credential_expiry_notice`, 등급표 밖).
- critical 단계 → 기존 `obs.EventCredentialExpiring`(등급표 안 — outbox · 재시도 · 전달 실패 판정을 그대로 탐).
- **알림이 꺼진 엔진**(`Config.Engine.Notifications.Enabled=false`)은 critical 단계도 normal 타입으로 기록만 한다. 보낼 수 없는
  critical 행은 전달 실패로 진입을 잠그므로(engine-safety 「등급화된 알림」), 알림을 끈 엔진이 attestation 때문에 진입을 멈추는 일을
  만들지 않는다 — a091(`zeroIsCritical`) · a095 와 같은 원천 · 같은 규칙.
- 중복 방지: 신원(발급 시각 | 만료 시각) × 단계 키를 메모리에 기록. 날수 단계는 `notice:7` · `imminent:2`, 재시작이 거절될
  상태는 `expired:<KST 날짜>` 처럼 날짜를 붙인다. 신원이 바뀌면(갱신 · 교체 · 사라짐) 기록을 비운다. 그래서 D-7 ~ D-1 은 하루 한 번,
  만료 · 범위 상실 · 계좌 불일치 · 무효는 **고쳐질 때까지 하루 한 번**(구현 리뷰 P3 — 한 번 알리고 침묵하면 다른 지속 critical 과 다름).
  Notify 가 오류(critical outbox 기록 실패)면 보낸 것으로 세지 않아 다음 판정이 다시 보낸다.
- outbox `Key` 는 `attestation-watch:<단계 키>:<신원>` — 단계마다 다른 행이라 알림기의 「같은 조건 이미 보냄」이 다음 단계를 삼키지 않는다.
- 재시작 뒤 메모리 기록은 사라진다 → 현재 단계를 한 번 더 알린다(critical 은 outbox 의 같은 키 + `RemindAfter` 1시간이 다시
  거른다). 재시작당 최대 한 번이라 받아들인다(원장 스키마를 늘리지 않음).
- 문구: 한국어 고정(정본 a085), 계좌 · 파일 경로 · 원문 오류 없음. 필드: `reason`(단계 코드) · `expires_at` · `days_left` ·
  `missing_endpoints` · `supervised_proof_at`(endpoint 별 최신 감독 증거 중 가장 이른 것 — 가장 먼저 기간을 벗어날 POST 근거,
  운영자가 감독 검증을 언제 다시 해야 하는지 읽는 칸; 리뷰 P3 반영).
  a133 전송 경계는 `Ntfy.Publish` 에서 그대로 적용된다.

### D3. 어디서 도는가 — 엔진의 보조 실행자(선택), 타이머는 아님

| 후보 | 장점 | 단점 |
|---|---|---|
| **엔진 보조 실행자** | 인터록이 읽은 **그 파일 · 그 계좌**로 판정, 알림기 · outbox · a133 경계가 이미 있음, 설치 변경 0(이미지 교체만) | 엔진이 죽어 있으면 없음(→ a134) |
| attest 타이머 실행 뒤 보고 | 갱신 직후 즉시 | 호스트 `tossctl` 에 ntfy 자격 증명 배선 필요(a134 가 여는 입구에 의존), systemd 단위 변경 = 사람 승인, 엔진 계좌와의 대조 불가 |

엔진을 고른다. 갱신 직후의 범위 상실은 1시간 판정 주기 안에 잡힌다(타이머 6시간보다 짧음). 타이머 쪽 보고는 a063 의
`--record-renewal-status` + 콘솔이 이미 맡고, a134 가 호스트 알림 입구를 내면 후속으로 검토한다(이 change 밖).

구조(`internal/app/engine/attestation_watch.go`):

- `Context.AttestationWatch(clk) AuxiliaryExecutor` — 이름 `attestation-watch`, `StopEvent=EventAttestationWatchFailed`(새 normal),
  **`OnStop` 없음**(래치 없음). `Runtime.Auxiliary` 이므로 정지가 감독 루프를 내리지 않는다(auxiliary.go 계약).
- `Run`: 게이트 꺼짐이면 아무것도 읽지 않고 취소를 기다린다(일찍 돌아가면 런타임이 정지로 기록하므로). 켜짐이면 기동 직후 한 번 +
  `clock.Sleep(1h)` 반복, 취소면 `ctx.Err()`.
- 회수 경계는 **이 파일의 코드에만**(`prepare` — 읽기 · 판정 · 문구). 판정 하나의 패닉은 로그 한 줄(`check_panicked`)로 끝나고
  다음 판정이 돈다.
- **알림기 호출은 회수 경계 밖**(구현 리뷰 P1). `Notifier.claimAndDeliver` 는 `n.mu` 를 defer 없이 놓으므로 그 안의 패닉을 회수하면
  잠금이 영구히 잡혀, 같은 잠금을 쓰는 exit 관측기의 `alert` · `Acknowledge` · `Flush` 가 영원히 막힌다(리뷰어 overlay 프로브로 재현).
  보조 실행자의 `runAuxiliaryBody` 도 회수하므로 checkOnce 의 회수만 좁혀서는 안 된다. 그래서 Notify 는 **별도 goroutine** 에서 부르고
  끝날 때까지 기다린다 — 그 패닉은 어떤 회수에도 닿지 않아 프로세스를 내린다(감독 루프가 부를 때와 같은 결과, 그리고 a134 가 그
  죽음을 알린다). 살아 있는 채 손절 알림이 멈춘 엔진보다 낫다. 자식 프로세스 시험 `TestA135NotifyPanicIsNotRecovered` 가 고정.
  근본 수리(`claimAndDeliver` 의 패닉 안전 잠금 해제)는 알림 경로(High-risk 인접) 기존 함수 편집이라 이 change 밖 후속으로 둔다.
- 종료 중(`ctx.Err()`)에는 기록을 시작하지 않는다(리뷰 P3 — 끊긴 기록이 메모리 래치 · 오류 줄을 남기지 않게).
- 상태(신원 · 보낸 키)는 Run goroutine 하나만 만지므로 잠금 없음. 알림기 호출은 `Notifier.Notify` — 엔진의 다른 비-exit 호출자와
  같은 입구(전송은 알림기 잠금 밖, a092 단위 ③).
- nil `*obs.Notifier` 를 인터페이스에 넣지 않는다(nil 이 아닌 인터페이스 → 패닉). 알림기가 없으면 로그만.

### D4. 엔진에 주는 영향

- 브로커 호출 0, 파일 읽기 1회/시간. `Verified` · `EntryPermitted` · 게이트 · 운영 모드 쓰기 0.
- 유일한 간접 효과: critical 알림이 전달 한도를 넘겨 실패하면 알림기의 기존 규칙대로 진입이 잠기고 운영 모드가 ENTRY_BLOCKED 로
  승격된다. 그것은 engine-safety 가 「자격증명 만료 임박」에 정한 그대로이며 청산 · 손절에는 닿지 않는다. 알림이 꺼진 엔진은 D2 로 제외.

### D5. 콘솔

a063 의 attestation 칸이 이미 `ExpiryHorizon`(남은 시간) · 72시간 경고 · 누락 endpoint 를 그린다. 남은 날 표시는 추가하지 않는다
(콘솔 무편집). 72시간 경고와 이 change 의 D-2 critical 은 같은 구간을 다른 표면에서 말한다 — 기준 하나로 합치는 것은 후속 후보.

### D6. 「갱신하면 POST 를 잃는다」 예측은 하지 않음

감독 증거의 나이와 soak 유효 기간을 엔진에서 다시 계산하면 `acceptSupervised` 의 규칙 사본이 된다. 대신 `supervised_proof_at` 를
필드로 싣고, 실제로 범위를 잃은 파일이 쓰이면 `scope_lost` 로 알린다. 지금 운영 상태(증거가 10-09 에 기간 밖)는 이 change 의 실측으로
Manager 에게 보고한다(`analysis/verify/live-attestation.md`).

## Function Logic Map

- 새 파일 `attestation_watch.go` 의 함수는 base 에 없는 새 함수다 → `not-applicable`(새 leaf).
- `internal/obs/event.go`: 상수 선언 2개 추가, 함수 · 등급표 무편집 → `not-applicable`.
- 기존 함수 편집: `cmd/tossctl/engine.go` `engineRuntime`(보조 실행자 목록 한 원소) · 시험 함수
  `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(기대 집합). 게이트가 수정된 기존 함수로 계산하므로 `tools/logic-map` 번들을
  만든다(분기 변화 0). **`cmd/tossctl/engine.go` 는 a134 가 같은 파일(`newEngineCmd`)을 편집 중이라 이 배선 한 줄은 a134 커밋 뒤에 넣는다**
  (tasks 1.3).
- 근거로 인용한 기존 함수 분기(`verifyGate` · `acceptSupervised` · `BuildAttestation`)는 편집하지 않고 AST 열거만 `analysis/code-context/` 에 둔다.

## Risks / Trade-offs

- **알림 피로** — D-7 ~ D-3 매일 normal 1건, D-2 · D-1 critical 각 1건. 재시작마다 현재 단계 1건 추가 가능. 받아들임.
- **판정 주기 1시간** — 만료 순간과 알림 사이 최대 1시간. 만료 전 D-1 critical 이 이미 나가므로 받아들임.
- **같은 구간 두 기준**(콘솔 72시간 · 알림 D-2) — D5.
- **병행 change 기준 커밋** — base `421bf53c` 이후 a134 의 기존 함수 편집(`newEngineCmd`)이 같은 워크트리 diff 에 섞인다.
  Manager 게이트에서 귀속을 가른다(memory 「쌓은 change 는 게이트를 깬다」).

## 배포

이미지 재빌드 + 엔진 재시작(사람, KST 05:00~09:00 창). 현재 attestation 은 10-23T01:39Z 까지 유효하므로 그 전 재시작은 무장한다.
D-7 경계가 10-16T01:39Z 이므로 그 전에 배포해야 D-7 통지가 나간다. 형식 · 스키마 변경 없음.
