# a133 design — 밖으로 나가는 알림에는 계좌가 없다

기준: base 는 `base-commit.txt`. 줄 번호는 base 의 것이다. 분기 주장은 이 change 의 AST 번들(`analysis/function-logic/*/ast.json`)을 근거로 한다.
규칙은 a131 의 `internal/redact`(정본: `openspec/changes/archive/2026-10-13-a131-no-account-in-any-log/design.md` D0 · D2)를 **그대로** 쓴다 — 사본을 만들지 않는다.

## D0. 무엇이 「계좌 정보」인가

a131 D0 과 같다: 등록된 계좌 원문 · 등록 숫자열 꼴(8자리 이상) · 마스크 꼴 · 소금 없는 저엔트로피 유도 id(`pos-`·`rec-`·`mode-`·`adj-` + 24 hex,
`corr-` + 32 hex). 판정 함수는 `redact.String`(값) · `redact.IsAccountKey`(필드 이름) 둘이다.
**규칙 수리 하나(구현 리뷰 I-1).** a131 의 `redact.String` 은 유도 id 둘이 구분자 한 글자로 붙으면(`pos-A pos-B`, `%v` 로 펼친 목록) 둘째를 놓쳤다 — 뒤 경계 그룹이 구분자를 먹기 때문. 규칙 하나를 유지하려고 사본을 만들지 않고 `redact.String` 안에서 고정점까지 되풀이하도록 고쳤다(로그 · 기록 · 알림이 같이 고쳐짐). 편집 전 FLM: `analysis/function-logic/internal-redact--string/`.

운영자에게 필요한 것은 남는다(규칙이 건드리지 않는 꼴): 종목 코드 · 종목 이름 · 사유 코드 · 시장 · 수량 · 사건 유형 · intent/attempt id
(`execgw.randomID` = 접두 없는 128비트 hex — 유도 id 꼴의 접두가 없고, 등록 숫자열과 같을 수 없음) · 운영 모드 이름 · 행위자.

## D1. 외부 발행 경로 전수

**전송 수단.** 비시험 코드에서 `obs.Publisher`(`Publish(context.Context, obs.Notification) error`)를 구현하는 타입은 `*obs.Ntfy`(`ntfy.go:85`) **하나**다
(`rg 'func \([^)]*\) Publish\('` — 다른 하나 `httpapi.(*Stream).Publish` 는 로컬 SSE 스트림이고 시그니처가 다르다). 생산 구성 자리는 둘:
`engine.resolveNotificationPublisher`(`notifications.go:101`, 엔진 프로세스의 유일한 전송 수단) · `publishNotificationTest`(`notificationsettings.go:151`, 콘솔 「알림 테스트」 버튼).
알림이 아닌 외부 호출(브로커 API · 갱신 확인 · 릴리스 · 모니터 프로브 · push 수신)은 사건 내용을 싣지 않으므로 범위 밖이다.
**httpapi SSE(범위 밖 — 리뷰 F-5).** `httpapi.(*Stream).Publish` 는 시그니처만 다른 게 아니라 역할이 다르다: 모바일 앱이 인증해 읽는 조회 API 의 스냅숏 푸시다(`cmd/tossctl/httpapi.go:166-195` · `:732`). 그 스냅숏에 `AccountLabel`(마스크)과 `PositionID`(`pos-…`)가 실린다(`httpapi_reader.go:184` · `:381`). 알림이 아니고 마스크 표기는 그 API 를 만든 앞선 change(a051 계열)의 표시 결정이므로 a133 은 바꾸지 않는다 — Manager 에게 후속 후보로 올린다(완료 보고).

**발행 호출 자리 7** (CodeGraph `callers Publish` + CGC `analyze callers Publish` + HEAD grep 대조 — `analysis/code-context/`):

| # | 자리 | 무엇을 보냄 | 계좌가 실리는 길(base) |
|---|---|---|---|
| P1 | `Notifier.publishBestEffort` `notifier.go:183` | `notificationFor(e)` — Title · Body + **Fields 펼침** | `Fields[account]` · `position_id` · 감싼 오류 문구 |
| P2 | `Notifier.deliver` `notifier.go:489`(critical 실시간) | 같음 | 같음 — 운영 원장의 payload 계좌 4 행 · pos- 12 행이 이 길로 나갔을 수 있음 |
| P3 | `NormalRelay.publish` `normal_relay.go:108`(a092 일반 등급 이관) | 같음 | 같음 |
| P4 | `alertDeliverer.deliverOne` `alertdelivery.go:327`(outbox 재전송 — 엔진 생산) | 저장 행의 Title · Body(필드 없음) | 저장된 본문 안 문구 |
| P5 | `Notifier.Flush` `notifier.go:884`(outbox 재전송 — 생산 배선 0, 시험 전용) | 같음 | 같음 |
| P6 | `Heartbeat.Beat` `heartbeat.go:67`(생산 배선 0) | 고정 Title · `Status()` | `Status` 문구 |
| P7 | `publishNotificationTest` `notificationsettings.go:152` | 고정 문구 | 없음(고정) |

일곱 모두 `Ntfy.Publish` 하나로 모인다. Tags 를 채우는 생산 자리는 0(모든 알림의 태그 = 사건 유형 하나). 요청 URL 은 `base + "/" + topic` 뿐이고
쿼리가 없다 — 서버와 토픽은 설정값이며 콘솔이 만드는 토픽은 기계가 만든 무작위 값이다(a075 `NewTopic`). 환경 변수 `TOSSCTL_NTFY_TOPIC` 은 설정 토픽을 덮으며 운영자가 고른 임의 문자열일 수 있다(`notifications.go:89-92`) — URL 은 일부러 치환하지 않으므로 운영자가 고른 토픽 이름은 운영자의 책임이다(잔재, 리뷰 F-11). 헤더는 Title · Tags · Priority · Content-Type · Authorization(설정 토큰).

## D2. 치환 지점 — 전송 경계 한 곳(`Ntfy.Publish` 의 첫 문장)

`Ntfy.Publish` 는 첫 문장에서 `msg = redactForTransport(msg)` 를 하고, 그 뒤의 모든 바이트(본문 · Title 헤더 · Tags 헤더)는 치환된 값에서 만든다.
새 leaf `redactForTransport`(`internal/obs/a133_transport_redact.go`)는 Title · Body · Tags 의 각 원소 · Type(태그 헤더의 첫 원소)을 `redact.String` 으로 바꾼 사본을 돌려준다 —
순수 함수, 오류 · 패닉 경로 없음(`redact.String` 은 정규식 치환뿐).

**왜 여기인가.**
- 이 함수는 프로세스에서 사건 내용이 네트워크로 나가는 **유일한** 자리다(D1). P1~P7 이 모두 거치므로 「경로마다 사본」이 0 이다.
- 호출자 쪽(P1~P7)에 두면 일곱 사본이고, 재전송(P4 · P5)은 `notificationFor` 를 안 지나므로 그 함수 안에 두면 저장 본문이 그대로 나간다.
- 구성 자리(`resolveNotificationPublisher`)에서 감싸는 발행자를 두면 P7(콘솔이 `&obs.Ntfy{}` 를 직접 만듦)이 빠진다.
- 토픽 · 서버 · 토큰은 치환하지 않는다 — 사건 내용이 아니라 설정이고, 바꾸면 전달이 깨진다(토픽은 비밀 경로). 시험이 「URL = 서버 + /토픽, 쿼리 없음」을 고정한다.

**구조 핀(새 전송 수단 대비).** go/parser 시험이 고정한다: ① 비시험 `cmd/` · `internal/` 에서 `Publish(context.Context, obs.Notification) error` 메서드를 가진 타입은 `Ntfy` 하나 — `obs` 를 별칭으로 import 한 꼴 포함, `obs.Notification` 의 타입 별칭 · 정의 타입은 금지
② `Ntfy.Publish` 의 첫 문장이 `redactForTransport` 호출이고 `http.NewRequestWithContext` 보다 앞 ③ `renderFields` 가 `redact.IsAccountKey` 를 부름 ④ `redactForTransport` 가 `redact.String` 을 부름 ⑤ `internal/obs` 안에서 `net/http` 를 import 하는 파일은 `ntfy.go` 하나.
핀의 범위(리뷰 F-2): `obs.Publisher` 를 채우는 생산 타입과 obs 패키지 안의 전송. obs 밖에서 알림 문구를 http 로 직접 보내는 코드는 핀 밖이고 spec 시나리오도 이 범위로 좁혔다. 문자열 칸 완전성(새 `Notification` 칸이 생기면 실패)은 리플렉션 시험이 잰다(리뷰 F-3).
새 발행자를 추가하는 사람은 ①에서 실패하고, 그 발행자도 같은 함수를 부르게 고친 뒤 목록을 늘린다.

## D3. 필드 펼침에서 계좌 키 버림 — `renderFields`

`notificationFor` 는 `Fields` 를 `k: v` 줄로 본문 끝에 붙인다(`renderFields`, base `notifier.go:1028-1040`, AST 분기 B1 키 모으기 range · B2 쓰기 range).
경계(D2)에서는 이미 문자열이라 필드 **이름**을 모른다 — `account: [account]` 처럼 값만 바뀐 줄이 남고, 등록되지 않은 계좌(다른 프로세스 · 시험 · 계좌 전환)는
값조차 안 바뀐다. 그래서 이름을 아는 마지막 자리인 `renderFields` 의 B1 에서 `redact.IsAccountKey(k)` 인 키를 건너뛴다(a131 `redact.Handler` 의 계좌 키 버림과 같은 함수).
`position_id` 는 계좌 키가 아니므로 줄은 남고 값이 경계에서 `pos-[redacted]` 가 된다(a131 의 로그 결정과 같음 — 종목 · 수량 · 시장으로 운영자가 포지션을 식별).

규칙 적용은 두 자리지만 규칙은 하나다: 이름 판정은 `redact.IsAccountKey`, 값 치환은 `redact.String`. 어느 쪽도 규칙을 베끼지 않는다.

## D4. outbox 저장분 — 저장 시점에는 치환하지 않는다

| 안 | 내용 | 판정 |
|---|---|---|
| 저장 시점 치환(`notifyCritical` · `recordCritical` 이 만드는 `journal.Alert`) | Title · Body · Payload 를 저장 전에 치환 | **기각** |
| 경계만(D2) | 저장은 그대로, 나갈 때 치환 | **채택** |

근거:
1. **저장된 outbox 는 그대로 나가지 않는다.** 재전송 경로 둘(P4 · P5)은 저장 행의 Title · Body 로 `Notification` 을 만들어 `Ntfy.Publish` 를 부른다 —
   경계를 지난다. Payload · event_key 는 어떤 경로도 전송하지 않는다(P4 · P5 는 Title · Body 만 씀). 시험이 P4 · P5 를 httptest ntfy 로 잰다(D6).
2. **저장 시점은 완전한 지점이 아니다.** outbox 쓰는 자리는 obs 밖에도 있다 — `execgw.(*Gateway).parkAlert`(`replay.go:553`, `EnqueueAlert` 직접, payload 에 `account`).
   저장 시점을 보장 지점으로 삼으려면 원장 계층(`journal.recordAlertTx`)을 바꿔야 하고 그것은 intent journal(High-risk)이다. 경계는 모든 발송자를 본다.
3. **원장은 로컬 정본 상태다**(a131 「범위 밖」: `journal.db`). `event_key` 는 계좌를 신원으로 담는다(`OperatingModeEventKey` = `operating_mode:<계좌>:…`) —
   치환하면 서로 다른 계좌 · 포지션의 조건이 한 키로 합쳐져 알림이 억제된다(a096 중복 제거). 그래서 원장을 「계좌 없음」으로 만들 수 없고,
   Title · Body 만 치환하는 것은 원장의 계좌를 없애지 못하면서 운영자가 로컬에서 읽는 내용만 줄인다. payload 를 읽는 생산 코드는 0
   (`engine alerts` 목록도 제목 · 본문 · payload 를 싣지 않음 — `engine_alerts.go:59`).
4. **이미 쌓인 행**(실측 `analysis/verify/outbox-census.md`): 16 행 전부 DELIVERED, PENDING 0. title · body 에 계좌 0. 정착 행이 다시 나가는 길은
   재무장뿐이고 재무장은 내용을 **새 값으로 덮는다**(`recordAlertTx` UPDATE). 따라서 옛 행의 저장 내용이 재전송될 길이 없다 — 처리할 것 없음.
   payload 계좌 4 · pos- 12 와 event_key 계좌 2 · pos- 12 는 로컬 정본에 남는다(전송 안 됨).

**기각한 다른 안(리뷰 F-6).** ①을 저장 시점이 아니라 경계 설계 안에서 읽기 전용으로 닫는 길도 있다 — 기동 때 원장에 있는 모든 `account_ref`(포지션 · intent · 운영 모드)를 `redact.Register` 한다(원장 쓰기 · High-risk 편집 없음). Manager 결정으로 채택하지 않았고(YAGNI — PENDING 0 · 드문 사람 작업), 필요해지면 이것이 후속 안이다.

**잔재(명명).** ① 경계의 원문 · 숫자열 규칙은 **등록된** 계좌만 안다. 엔진은 계좌 해석(`Build` 1단계 `resolveAccountRef` → `accountsLocked` → `redact.Register`)이
원장 열기와 전달기 구성보다 앞이라 같은 프로세스의 알림은 덮인다. 그러나 자격 증명을 **다른 계좌**로 바꾼 뒤 재시작했고 옛 계좌의 PENDING 행이 본문에 옛 계좌 원문을
담고 있으면, 새 프로세스는 옛 계좌를 등록하지 않았으므로 그 원문이 나간다(마스크 꼴 · 유도 id 는 등록 없이 덮임). 실측 PENDING 0 · 본문 계좌 0 이고, 저장 시점
치환으로 막으려면 2 의 원장 편집이 필요하다. **Manager 결정(2026-10-13): 잔재로 수용** — 지금 PENDING 0 이고 계좌 자격 증명 교체는 드문 사람 작업이다. 코드로 막지 않고(YAGNI) 운영 절차 한 줄로 둔다: 「계좌 자격 증명을 바꾸기 전에 `alert_outbox` 의 PENDING 이 0 인지 확인하고, 아니면 먼저 전달 · 정리한다」(배포 절 · tasks 게이트 뒤 절차). ② 브로커 응답 본문이 감싼 오류 문구에 등록 전 계좌가 실리는 경우(a131 잔재 ①)와 같다.

## D5. 운영 의미 · 전달 판정 · 손절 즉시성 불변

- `Ntfy.Publish` 의 분기 B1~B9(AST)는 그대로다 — 치환은 B1 앞 대입 한 문장이고 오류를 만들지 않는다. 같은 서버 응답에 같은 오류 · nil 을 돌려주므로
  outbox 정산(`MarkAlertDelivered` · `MarkAlertAttemptFailed`) · 시도 수 · a124 한도(3) 판정 · 진입 래치 · 모드 승격 · 임차 · 승인 해제의 입력이 같다.
- `renderFields` 는 반환 문자열에서 계좌 키 줄만 빠진다. 호출자 `notificationFor` 의 분기(필드 유무)는 바뀌지 않는다. `eventKey`(중복 제거)는 `renderFields` 를 쓰지 않는다.
- 잠금: 치환은 `Ntfy.Publish` 안 — 모든 호출자에서 알림기 뮤텍스 밖(a092 단위 ③ · a124 D7)이다.
- 청산 · 손절을 늦추지 않는 근거는 치환의 위치가 아니라 a092 구조다(리뷰 F-7 정정): exit 루프는 기록 전용 입구(`RecordCritical` · `RecordOperatingMode` — 전송 없음)와 일반 등급 이관(`NormalRelay` — 별도 실행자)만 쓰고, 동기 `Notify` 의 P1(`publishBestEffort`)은 그것을 부른 쪽에서 돌지만 exit 경로가 아니다. 치환은 이미 네트워크 왕복을 하는 호출에 CPU 몇 마이크로초를 더할 뿐이다.
- 분기 불변의 근거(리뷰 F-8): `redact.String("") == ""` 이고 치환 결과는 비지 않으므로(대체 문자열이 비어 있지 않음) B5(`title != ""`)의 참 · 거짓이 같다. 대체 문자열 `[account]` · `pos-[redacted]` 는 ASCII 이고 제어 문자가 없어 Title 헤더가 무효가 되지 않는다(B4 요청 구성 실패 경로 불변). 등록 참조가 9자보다 짧으면 본문이 길어질 수 있으나 생산 계좌는 11자리이고 길이는 판정 입력이 아니다.
- 비용: 알림 하나당 문자열 넷 정도의 정규식 치환 — 네트워크 왕복(기본 시한 수 초)에 비해 무시할 수 있다.

## D6. 시험 전략

센티널(리뷰 F-4 반영 — 계좌 키 필드 다섯 모두 **등록 안 된** 값이라 값 치환이 필드 버림 누락을 가리지 못함): 등록 원문 `987-65-SENTINEL` · 등록 숫자열 `98765432107` · 그 마스크 `*******2107` · 유도 id `pos-<24 hex>` · `mode-<24 hex>` ·
등록 안 된 계좌 값(필드 버림 전용) `UNREGISTERED-ACCT-77`.

1. **경계(obs, httptest ntfy):** `Ntfy.Publish` 에 센티널을 Title · Body · Tags · Type 에 심고 서버가 받은 본문 · 모든 헤더 값 · URL 에 없음. URL 은 `/<topic>` 이고 쿼리 없음.
   종목 · 사유 코드 · attempt id 는 남음.
2. **경로별(받은 바이트):** P1(일반 등급 Notify) · P2(critical Notify) · P3(NormalRelay) · P5(Flush — 저장 행 재전송) · P4(엔진 `alertDeliverer` — 저장 행 재전송,
   `internal/app/engine` 내부 시험) · P7(`publishNotificationTest` — 고정 문구, 토픽 경로 확인)을 실제 `*obs.Ntfy` + httptest 로 몰아 센티널 없음.
   P2 는 `Fields` 에 `account`(등록 안 된 값) · `position_id` · `symbol` · `reason` · `attempt_id` 를 심어 계좌 키 줄 없음 · `position_id: pos-[redacted]` · 나머지 그대로.
3. **전달 판정 불변:** 서버 503 이면 `Publish` 가 오류를 돌려주고 critical 3회 실패 뒤 진입 게이트가 `ReasonAlertUndelivered` 로 잠김(센티널을 심은 사건으로) · 서버 200 이면 행이 DELIVERED.
4. **구조(go/parser):** D2 의 ①②③.
5. 기존 시험 전부 green(`internal/obs` · `internal/app/engine` · `cmd/tossctl` · `internal/redact` · `internal/flatten`). 기대값을 바꾼 기존 시험은 목록으로 남긴다(바뀌면 그 시험 함수도 FLM 대상).

변이(사본 · 무변이 대조군 GREEN 먼저, `analysis/mutation/`): M1 치환 우회(`Ntfy.Publish` 의 치환 문장 삭제) · M2 발행 경로 하나 누락(치환을 경계에서 `notificationFor` 로 옮김 —
재전송 P4 · P5 가 샘) · M3 필드 버림 삭제 · M4 Title 만 치환 안 함 · M5 Tags/Type 치환 안 함 · M6 치환된 사본 대신 원본으로 요청 구성 · M7 치환 없는 두 번째 생산 발행자 추가(구조 핀) · M9 좁힌 필드 버림(`account` 하나만 버림, 리뷰 F-4) · M10 obs 안 새 http 파일(구조 핀 ⑤) · M11 새 `Notification` 문자열 칸(리플렉션 시험).
청산 · 손절을 막지 않음은 기존 a092 시험이 다시 증명한다(리뷰 F-13): `internal/obs` 의 `a092_record_only_test.go` · `a092_lock_scope_test.go` · `a092_normal_relay_test.go`, `internal/app/engine` 의 `a092_exit_cycle_records_only_test.go` · `a124_judging_does_not_delay_protection_test.go` · `a098_the_backlog_does_not_delay_protection_test.go` — 2.2 배터리에 포함.

## 배포 (범위 밖 — 사람 절차)

- a129 · a131 · a130 과 **한 번에**(사용자 결정). a133 은 `internal/obs` 의 바이트만 바꾼다 — 원장 스키마 · 기록 형식 · attestation · 설정 변경 0 이라 **새 순서 의존이 없다**.
- a131 의 「호스트 `tossctl`(tossos-attest.timer)과 이미지를 같은 창에 함께 교체」는 그대로 따른다. a133 의 변화가 실리는 프로세스는 엔진(컨테이너)과 콘솔(알림 테스트 버튼)이며,
  호스트 attest 는 알림을 보내지 않는다 — 함께 교체에 무엇을 더하지도 막지도 않는다.
- 충돌: a133 편집 파일은 `internal/obs/{ntfy,notifier}.go` + 새 obs 파일 · 새 시험 파일뿐이다. a130(`internal/journal/trade_outcomes.go` · `apply_hook.go` · 콘솔 portfolio 템플릿 · backfill 도구)과
  파일 겹침 0. a131 은 이미 HEAD(`internal/redact` 를 쓰므로 a131 이 먼저 있어야 함 — 같은 이미지에 함께 실리므로 충족).
- 교체 뒤 확인(사람): 콘솔 「알림 테스트」 한 번 — 도착하면 전송 경로가 산다. 다음 critical 알림의 본문에 `account:` 줄이 없고 포지션 id 가 `pos-[redacted]` 임을 휴대폰에서 확인.
- 운영 절차(D4 ①, Manager 결정): 계좌 자격 증명을 바꾸기 전에 `alert_outbox` PENDING 0 을 확인한다 — 아니면 먼저 전달 · 정리(승인)한다.
- 롤백: 이미지 핀을 되돌리면 알림 본문이 base 와 같아진다(원장 · 형식 영향 없음).
