# issues — verify-execution-capability

> WORKFLOW 예외 경로 §분류: ① blocking ② safe local ③ editorial. 아래는 구현 중 발생한 편차와
> Manager 판단이 필요한 항목의 기록이다.

## 2026-07-26 · task 1.5 (`tossctl verify run`) 구현

### ② safe local — `official.ModifyConditionalOrderRef` 신설

**배경**: 2.5 는 조건주문 정정이 "신규 ID 발급·기존 ID 무효화"인지 실측하도록 요구한다.
기존 `official.Client.ModifyConditionalOrder` 는 응답 본문을 버려서(`postAcct(..., nil)`)
새 `conditionalOrderId` 를 알 수 없고, 목록 재조회로 유추하면 조건주문이 2건 이상일 때 추측이
된다.

**조치**: 기존 메서드를 **건드리지 않고** `ModifyConditionalOrderRef` 를 추가했다
(`internal/official/conditional_writes.go`). 기존 호출부(hybrid, app/engine)는 `nil` out 을
넘겨 null result 도 성공으로 처리하는데, 그 동작을 바꾸면 라이브 mutation 의 실패 표면이 넓어진다.
기존 동작 보존 테스트(`TestModifyConditionalOrderStillToleratesANullResult`)를 함께 추가했다.

`internal/app/engine/seal_test.go` 의 `officialMutators` 목록에 새 메서드를 추가했다 —
그 파일의 주석이 "새 mutator 는 의식적으로 추가하라"고 요구한다.

근거: `POST /api/v1/conditional-orders/{id}/modify` 설명, `docs/migration/openapi.latest.json`.

### ② safe local — `--redo <step-id>` 플래그 추가 (설계에 없던 표면)

**배경**: `read-fixtures`(2.1)는 계좌의 기존 주문 이력에서 status enum 을 수집한다. 주문 이력이
없는 계좌에서는 skip 된다. 그런데 이후 단계들이 주문을 만들어 이력을 남기므로, 되돌아갈 방법이
없으면 신규 계좌에서 2.1 이 영구히 미측정으로 남는다.

**조치**: `--redo read-fixtures` 로 기록에 판정이 있는 단계를 다시 돌릴 수 있게 했다. 재실행해도
mutation 은 여전히 매번 typed confirmation 을 거친다. skip 사유 문구가 이 플래그를 직접 안내한다.

### ② safe local — `--max-sell-quantity`(기본 1)와 "최소 오버셀"

**배경**: 2.2 의 매도 경계는 부분/전량/보유초과 3가지다. "전량"과 "보유초과"는 정의상 최소 수량
1주 규칙과 충돌한다.

**조치**:

- 전량 매도는 보유수량이 `--max-sell-quantity`(기본 1주) 이하일 때만 실제로 제출한다. 그보다 큰
  보유에서는 주문을 내지 않고 `sell.boundary.full_accepted = unverified` 로 정직하게 기록한다.
  누군가의 전체 포지션을 시장 위에 걸어두는 것은 이 도구가 할 일이 아니다.
- 보유초과는 `sellable + 1`주 — 가능한 가장 작은 초과 — 로 제출한다. 시장가 대비 멀리 떨어진
  지정가 매도이고, 브로커가 **수락하면** 즉시 취소한 뒤 단계를 FAIL 로 기록한다(청산 예약 공식이
  브로커의 경계 검사에 의존할 수 없다는 뜻이므로).

### ② safe local — 취소도 typed confirmation 을 받는다 (flatten 과 다름)

flatten 은 §0.3 에 따라 취소 앞에 프롬프트를 두지 않는다(취소는 노출을 줄이기만 하므로). 이 도구는
취소 자체가 측정 대상(2.2 의 cancel 경로, 2.5 의 조건주문 취소)이라 확인을 받는다. 프롬프트에
"direction: this REDUCES exposure" 를 명시해 위험 방향을 오해하지 않도록 했다.

### ① Manager 판단 필요 — 보유 없는 계좌에서 2.5 전체가 미측정으로 남는다

**사실**: SINGLE+MARKET 손절은 SELL 이므로 보유가 필요하다. 설계 지시는 "never buy-to-create a
holding automatically; skip-with-reason if none" 이고 그대로 구현했다. 결과적으로 보유가 0인
계좌에서는 `conditional-register / sellable-reserved / conditional-persist /
conditional-modify / conditional-cancel` 이 전부 skip 되고, **2.5·2.6 의 입력이 하나도 생기지
않는다**. attestation 의 조건주문 endpoint 집합도 비므로 게이트는 계속 닫힌다(의도된 방향).

**검토했으나 구현하지 않은 대안**: `--conditional-buy-fallback` — SINGLE + LIMIT + BUY,
triggerPrice 를 시장가보다 훨씬 **위**로(발동 안 됨), orderPrice 를 훨씬 **아래**로(발동해도 체결
불가). 보유 없이 조건주문 등록·조회·존속·정정·취소 endpoint 를 전부 검증할 수 있고 체결 위험은
이중으로 막힌다. 다만 (a) 설계 문구에 없고, (b) 사용자 계좌에 조건부 매수를 만드는 일이라 임의로
넣지 않았다.

**현재 안내**: skip 사유가 "KR 종목 1주 이상을 보유한 뒤 `--resume`(또는 `--holding-symbol`)"
을 지시한다. 사용자 협조 절차에 이 선행 조건을 넣을지, 위 fallback 을 opt-in 으로 추가할지 Manager
판단이 필요하다.

### 측정 불가로 확정된 항목 (report 가 `unverified` 로 노출)

| 항목 | 사유 |
|---|---|
| `idempotency.key_scope` (2.7 계좌 스코프) | 자격증명 1세트 = 계좌 1개. 키가 계좌를 넘는지 관측할 수단이 없다 |
| `conditional.trigger_observed` / `triggered_order_id_exposed` / latency (2.5) | 발동에는 시장이 가격까지 와야 한다. 체결을 의도한 주문을 이 도구는 내지 않는다 — [별도 세션 — 시장 조건 필요] |
| `idempotency.ttl_window_closed` (2.7 유효 창) | `--include-ttl-edge` 옵트인 전에는 미측정. 의도적 이중 주문 절차이므로 기본 생략 |
| US 시장 전체 | mutation 단계는 KR 심볼만 받는다. amend 의 quantity 필수 여부·일일 상하한가 모두 KR 규칙이고, US 규칙으로 대체 검증하면 잘못된 근거가 된다 |
| 정규장 밖 동작 (2.5) | 등록 시각의 KST 세션 라벨만 기록한다. 세션별 반복 실행은 사용자가 시간대를 골라 재실행해야 한다 |

### 2.9 (실측 비용) 는 이 도구만으로는 채워지지 않는다

모든 주문이 체결 불가 가격이므로 `execution.commission` 은 항상 null 이다. `costs` 단계는
`costs.collected = false` 와 "no verification order filled" 사유를 기록한다 — 0원이라고 쓰지
않는다. 2.9 를 채우려면 실제로 체결되는 주문이 필요하고, 그것은 이 도구의 안전 전제(체결 불가
가격) 밖이다. Manager 결정 필요: 사용자의 평소 거래 체결 내역에서 수집할지, 별도 절차를 둘지.

## Manager 판정 (1.5 검증 도구, 2026-07-26)

- 안전 기제 승인: mutate.go 단일 mutation 파일(AST 증명), --yes/--force/env 우회 부재(테스트+소스 가드), 미체결가 산정 불가 시 추측 대신 거부, 잔여물 명명 오류. ModifyConditionalOrderRef additive 확인.
- **보유 없는 계좌의 2.5(--conditional-buy-fallback)**: 보류 — 사용자가 KR 1주 보유 시 불필요한 작업. 사용자가 skip을 실제로 만나면 그때 옵트인 플래그로 구현(이중 차단 설계는 승인됨을 기록). 단 BUY 폴백으로도 매도가능수량 예약 의미(2.8 후반)는 미측정으로 남음을 report에 명시할 것.
- **2.9 비용 실측**: 이 도구로 불가(미체결가 설계상 commission 없음) — 정직한 경로는 실제 체결뿐. 판정: tracer 실전 1주 왕복(사용자 승인 트랙)의 체결에서 수집하거나 사용자의 임의 실거래 1건에서 수집. 그때까지 비용 모델은 보수 placeholder 유지(§0.9 정합).

## Manager 판정 (배치 승인 전환, 2026-07-26)

- 사용자 결정에 따라 승인 모델을 run당 1회 배치로 전환(커밋 1df778a). 이 파일 상단의 "재실행해도 mutation은 매번 typed confirmation" · "취소도 typed confirmation" 서술은 구모델 기준 — 현행은 배치 승인이 기본, `--confirm-each` 옵트인. 계획 밖 mutation은 ErrOutsidePlan으로 전체 중단(전송 0건), mutate.go의 gate 경유는 AST 정적 가드로 증명.
- 배치 거부 시 읽기 전용 단계도 중단: 계약 문언대로 승인 — 읽기 증거는 --confirm-each로 도달 가능.
- sell-boundary 수량을 계획 시점 SellableQuantity 실측으로 표기(실패 시 blind 승인 대신 단계 제외): 승인.

## Manager 판정 (웹 콘솔, 2026-07-26)

- 사용자 결정: 검증을 웹 화면에서 수행. 승인 채널을 TTY→localhost 웹 폼으로 확장하되 등가성 조건을 태스크 1.6에 명시(토큰+CSRF+nonce 타이핑, runner 레일 무변경, CLI 비대화 승인 경로 신설 금지). TTY의 실질은 "사람의 의도적 타이핑"이며 웹 폼 타이핑은 같은 강도 — 자동화 차단의 실체는 양쪽 모두 결의된 사용자를 막지 못하고, 지키는 것은 사고·에이전트의 무의도 mutation이다. P4 웹 데몬 아키텍처와는 별개의 임시 운영자 표면임을 명시(단일 사용자·루프백 한정).

## 2026-07-26 · task 1.6 (`tossctl console`) 구현

### ⓪ seam 변경 없음 — `internal/verifylive`는 한 줄도 바뀌지 않았다

설계 지시는 "Confirmer seam이 주입 가능하지 않으면 unexported 생성자를 만들라"였으나, 확인 결과
`verifylive.Options.Confirm` / `.ConfirmBatch` 는 이미 exported 필드이고 `verifylive.New` 가
공개 생성자다. 콘솔은 자신의 `BatchConfirmer` 를 넘기는 것으로 끝이며, verifylive 패키지 파일은
**추가·수정 0건**이다. 계획 인가(`Plan.Authorises`)·상한·취소·`ErrOutsidePlan` 전부 무변경.

주입 가능성이 넓다는 점은 대신 `cmd/tossctl` 쪽 정적 가드로 좁혔다 — `console_test.go` 가
(a) `verify.go` 의 runner 가 여전히 `terminalConfirmer` / `terminalBatchConfirmer` 를 쓰는지 AST로,
(b) `internal/console` 를 import 하는 파일이 `console.go` 하나뿐인지 소스 워크로 검증한다.

### ② safe local — 오승인 처리: 틀린 nonce는 재시도가 아니라 중단

TTY 계약은 "Type the confirmation string to approve, anything else to abort" 다. 웹에서 재입력을
허용하면 5분 창이 5분치 추측으로 바뀌므로 같은 계약을 택했다: 틀리면 `ErrRefused` 를 runner에
전달해 실행이 끝나고(전송 0건), 페이지가 다시 시작하라고 안내한다. 승인 거부는 record에
`KindApproval`·`refused` 로 남고 `StepCount` 는 0이라 다음 시도를 막지 않는다.

### ② safe local — 콘솔은 항상 "이어하기"이고, 프로세스당 검증 1회다

- `verify run` 의 "기록이 있으면 `--resume` 없이는 거부" 가드는 콘솔에 옮기지 않았다. 콘솔에는
  잊을 수 있는 플래그가 없고, runner 자체가 판정이 끝난 단계를 건너뛰며 plan에서 사유와 함께
  제외하므로 기록을 이어가는 것이 유일하게 안전한 기본값이다.
- 대신 **한 프로세스에서 단계를 밟은 검증은 1회로 제한**했다(`Console.spent`). 조건주문 존속은
  등록한 프로세스가 죽은 뒤에만 관측되므로, 이 경계는 안내문이 아니라 거부로 구현했다. 재시작한
  콘솔은 기록을 읽어 `awaiting-restart` 를 감지하고 버튼 라벨을 "이어하기"로 바꾼다.

### ② safe local — `--confirm-each` 미제공, per-mutation confirmer는 거부값

태스크 지시대로 웹은 배치 승인 전용이다. `verifylive.Options.Confirm` 은 nil을 거부하므로
`consoleMutationConfirmer()` 를 넣었고, 이 함수는 항상 `ErrNotATerminal` 을 돌려준다 — 나중에
누군가 `ConfirmEach` 를 켜더라도 승인이 아니라 거부로 실패한다. `console.go` 가 `ConfirmEach:` 를
쓰지 않는 것도 테스트가 확인한다.

### ② safe local — `consoleProbeSymbol` 은 `verify run --symbol` 기본값의 복제다

`verify.go` 는 라이브 주문 경로라 이 태스크 범위에서 건드리지 않았다. 대신 상수를 콘솔 쪽에
따로 두고, 두 값이 어긋나면 실패하는 테스트를 뒀다(`TestConsoleProbeSymbolMatchesVerifyRunsDefault`).

### ② safe local — 기존 테스트 파일 1건 수정: `help_convention_test.go`

`TestMutatingAnnotationOnTradeCommands` 는 `mutating=true` 인 leaf 커맨드를 화이트리스트로
강제한다. `tossctl console` 은 runner를 통해 실주문을 낼 수 있으므로 정직하게
`mutating=true` 로 표기했고, 그 목록에 사유 주석과 함께 추가했다. 프로덕션 동작 변경 없음.

### 남은 위험 · 범위 밖

| 항목 | 상태 |
|---|---|
| HTTP(평문) | 루프백 전용이라 TLS 없음. 세션 토큰은 URL→쿠키 1회 교환 후 주소창에서 제거된다 |
| 세션 토큰이 터미널 스크롤백에 남는다 | 설계상 그렇다 — "터미널 점유 = 인증"이 인증 모델이다. 토큰은 프로세스마다 새로 발급된다 |
| 게이트 ON | 콘솔 범위 밖(태스크 명시). 라우트도 없고 대시보드가 그렇게 적는다 |
| 진행 스트리밍 | SSE 아님 — 2초 meta refresh. 단, nonce 폼이 떠 있는 동안에는 새로고침하지 않는다(입력 유실 방지) |
| `cmd/tossctl` 의 TestMain에는 testenv 가드가 없다 | 기존 상태 유지. 신규 `internal/console` 테스트에는 설치했다 |

## Manager 판정 (콘솔 1.6, 2026-07-26)

- 편차 5건 전부 승인: seam 변경 0(기존 exported Options.Confirm 활용 — "콘솔만 배선 가능" 보증을 AST 가드로 이전한 판단 타당), 틀린 nonce=전체 중단(TTY 계약 그대로 — 5분 창을 추측 5분으로 만들지 않음), 콘솔=항상 resume 모델(플래그가 없으니 잊을 플래그도 없음, 프로세스당 1회 실행 상한이 존속 경계 보존), help_convention 화이트리스트 additive 1건(mutating=true 정직 표기), probe 심볼 중복은 drift 테스트로 고정.
- Pre-Edit 게이트 검토 통과 확인(§0.7 — 게이트 토글 라우트 부재 포함).

## Manager 판정 (검증 1차 실행, 2026-07-26 일요일)

실측 기록은 [measurements.md](measurements.md). 판정:

- **레일 실전 검증 성립**: 오승인 거부(전송 0건)·재시작 간 resume·plan digest 고정 — 계약대로 동작함을 실계좌 증거로 확인(M2·M3). 도구 결함 아님.
- **측정 0건의 원인 3가지 전부 환경**: 일요일 휴장(M1), account-seq 429(M4), 보유 0(M5). 그런데 `fail`/`skipped`가 terminal이라 **콘솔(always-resume, redo 부재)에서는 재측정이 불가** — 콘솔 단독 운용이라는 사용자 결정과 충돌하는 갭. task 1.7 발주.
- **`--conditional-buy-fallback` 최종 기각** (앞선 보류 판정의 발동 조건 "사용자가 skip을 실제로 만나면"이 성립했으나): BUY 폴백으로 측정 가능한 것은 조건주문 등록·조회·존속·정정·취소 endpoint뿐이고, **2c의 임계 입력인 SELL측**(SINGLE+MARKET 손절 실동작·sell-boundary·sellable 예약 의미, 2.5 후반·2.6·2.8)은 보유 없이는 어떤 폴백으로도 측정 불가. long-only 제품은 모든 포지션에 SELL 보호가 필요하므로 SELL측 미측정 = 게이트 영구 폐쇄. 따라서 **KR 종목 1주 보유는 폴백으로 우회할 수 없는 선행 조건**이고, 1주를 보유하면 폴백은 불필요해진다(보류 판정의 자체 논리). 사용자 협조 절차에 "1회, 임의 KR 종목 1주 수동 매수"를 선행 조건으로 확정한다 — 도구는 계속 매수하지 않는다(레일 유지).
- **1.7 범위 통제**: 재측정은 fail·skipped 단계에 한정하고 반드시 **새 배치 승인(신규 nonce)** 경유 — 비대화 승인 경로 신설 금지 유지. 장시간 경고는 advisory(정규장 달력을 하드 레일로 박지 않음 — 주문 접수 창은 [미측정]). 429 대응은 읽기 전용 단계만 재시도, mutation 자동 재시도 금지.

## 2026-07-26 · task 1.7 (재측정 경로·실행 강건성) 구현

### ① 재측정 — 라우트를 늘리지 않고 mode로 분기

`/verify/start`에 `mode=redo` 폼 값을 더했다. 새 라우트(`/verify/redo`)를 만들면
`static_test.go`의 두 게이트 표(세션 전수·CSRF 전수)와 라우트 수 하한을 함께 고쳐야 하고,
"승인 문을 하나로 유지한다"는 1.6의 성질이 약해진다. 라우트 표는 무변경(7개)이다.

**대상 집합은 폼이 아니라 기록에서 계산한다.** `handleStart`는 `c.redoSet()`으로
`capability-verify.jsonl`을 다시 읽어 `verifylive.RedoSet`을 부른다. 폼에 단계 ID를 넣어도
무시된다 — 런타임 테스트(`TestTheRedoSetComesFromTheRecordAndNotFromTheForm`)와 소스 가드
(`pages.go`에 `verifylive.StepID(` 변환·`r.PostForm[` 접근 금지) 양쪽으로 고정했다. 폼이
단계를 지명할 수 있으면 이미 pass한 단계에 두 번째 실주문을 겨눌 수 있기 때문이다.

`refused`는 대상이 아니다(태스크 문언이 fail·skipped로 한정). 어차피 콘솔은 배치 승인 전용이라
단계 수준 `refused`를 만들 수 없다 — 배치 거부는 `KindApproval` 라인이고 단계 판정이 아니다.

### ② advisory — Go 코드가 읽지 못하게 막았다

`verifylive.KRSessionAdvisory`는 판정을, 템플릿은 문장을 담당한다. `static_test.go`에
**data.go·templates.go 밖에서 `.Outside`·`KRSessionAdvisory(`를 언급하면 실패**하는 가드를
넣었다 — advisory가 조건문이 되는 순간 미측정 달력이 레일이 되기 때문이다.

`steps.go`의 `sessionLabel()`이 같은 창(평일 09:00–15:30 KST)을 쓰고 있었으므로 hours.go로
옮겨 정의를 하나로 만들었다. 출력 문자열은 기존과 동일하고 테스트가 그 동일성을 고정한다.

advisory 문장은 한국어다 — 유일한 소비자가 한국어 콘솔 화면이다. 실측 코드
`order-hours-closed`와 `HTTP 422`는 원문 그대로 인용해 증거와 대조 가능하게 뒀다.

### ③ 429 — 셋 다 했고, 하나는 지시보다 더 갔다

**account-seq**: 지시는 "run당 1회 캐시"였다. 확인해보니 `official.Client`는 이미 성공 후
캐시한다 — 진짜 낭비는 **`/api/v1/accounts`를 두 번 읽는 것**이었다. `buildVerifyBroker`가
계좌 표기용으로 1회 읽고 seq를 버리면, 각 단계의 첫 account-scoped GET이 lazy 해석으로 또
읽는다. 그 두 번째가 M4의 429다. 이제 첫 조회의 seq로 `WithAccountSeq` 클라이언트를 구성해
**lazy 해석이 아예 일어나지 않는다**(토큰은 디스크 캐시라 재교환 없음).

부수 효과 1건 — 표기용 계좌와 헤더 계좌를 **같은 엔트리**에서 뽑도록 바꿨다. 기존에는
표기는 DisplayName이 빈 것을 건너뛰며 스캔하고 헤더는 `accts[0]`을 썼다. 자격증명 1세트 =
계좌 1개인 현 환경에서는 차이가 없지만, 다계좌에서는 "A를 측정했다고 적고 B를 측정"할 수
있는 잠재 불일치였다. ID가 숫자가 아니면 seq=0으로 두어 기존 lazy 경로로 되돌아간다.

부수 효과 2건 — **이 계정 조회 자체도 429면 재시도한다**. 태스크 문언은 "읽기 전용 단계"지만
이 호출은 Runner가 생기기도 전, 사람에게 아무것도 묻기 전에 run 전체를 죽이는 지점이고
M4에서 실제로 죽은 호출이다. 정책 상수는 `verifylive.ReadRetryExtraAttempts` /
`ReadRetryBackoff`로 export해 정의를 한 곳에만 뒀다.

**읽기 재시도**: `readRetry`(제네릭 헬퍼)가 `official.ErrRateLimited`에만, 15s·30s로 2회 더
보낸다. **모든 시도가 개별 `Call`로 기록에 남는다** — 1.3 retry matrix의 입력이 시도 로그
자체이므로, 성공한 시도만 남기면 재시도가 존재하는 이유를 기록이 감춘다.

`plan.go`의 sellable 조회도 같은 경로를 탄다(단계가 아직 없으므로 로깅은 생략). 이 읽기가
429로 죽으면 매도측 배치 전체가 조용히 plan에서 빠지는데, 그게 M4의 실제 피해다.

**mutation 재시도 없음**: transport 데코레이터가 아니라 읽기 경로만 부르는 헬퍼로 만든 이유다.
`TestAMutationIsNeverRetried`가 429 하에서 `PlaceOrder` 전송 1건·대기 0초를 고정한다.

**run 전역 재시도 예산은 두지 않았다** — 호출당 2회로만 제한했다. 최악의 경우 429가 지속되면
읽기 1건당 45초를 쓴다(읽기 10건이면 7분 남짓). 상태와 테스트 표면을 늘릴 값어치가 있는지
Manager 판단을 남긴다.

### ③ soak 일시정지 — soak은 파일을 모른다

`internal/runlock`(신규, os·time만 의존)이 마커를 소유하고, verify/console 배선이 run 동안
보유·1분 주기 갱신한다. **soak은 `Options.PauseWhile` 훅만 받는다** — `internal/soak`이
runlock을 직접 import해도 정적 가드는 통과하지만, "soak은 자기 기록 말고 아무것도 만지지
않는다"는 성질과 fake clock 테스트 가능성을 지키려고 seam으로 뒀다. 배선은 cmd/tossctl.

stale 기준은 mtime 5분(갱신 1분 = 4회 연속 실패해야 죽은 것으로 본다). 마커를 못 쓰면
verify는 한 줄 안내만 하고 그대로 진행한다 — 양방향 advisory다.

### 남은 위험 · 범위 밖

| 항목 | 상태 |
|---|---|
| 재측정도 프로세스당 1회 | `Console.spent` 그대로. 재측정 뒤 또 재측정하려면 콘솔 재시작 — 조건주문 존속 경계 보존이 우선 |
| 429 지속 시 run 소요 | 읽기당 최대 45초. 진행 로그에 재시도 사유·대기를 출력한다 |
| `cmd/tossctl` TestMain에 testenv 가드 없음 | 1.6에서 기록한 기존 상태 유지. 신규 테스트는 httptest·순수 stub만 쓴다 |
| 주문 접수 창 | 여전히 [미측정]. advisory는 평일 09:00–15:30 KST라는 가장 거친 선만 긋는다 |
| `docs/WORKFLOW.md` 미커밋 수정 | 작업 시작 시점에 이미 워킹트리에 있던 사용자 편집. 손대지 않았고 스테이징도 하지 않았다 |

## Manager 판정 (1.7 재측정·강건성, 2026-07-26)

독립 검증: 전체 스위트 -race 2589/48pkg 0 FAIL 재실행, RedoSet·retry·runlock·account-seq 코드 직접 판독. 편차 전건 승인:

- **mode=redo(신규 라우트 없음)**: 승인 — CSRF 게이트 문 1개 유지가 옳다.
- **redo 집합은 폼이 아닌 record에서만**(`TestTheRedoSetComesFromTheRecordAndNotFromTheForm`): 설계 핵심 그대로 — pass 단계 재실행 경로가 구조적으로 부재함을 확인.
- **`refused` 제외**: 승인 — 사람의 "아니오"를 자동 재시도하지 않는 방향이 맞다. `--confirm-each`가 웹에 오지 않는 한 도달 불가(레일 유지 확인).
- **plan 읽기·계정 조회까지 재시도 확대**: 승인 — M4에서 실제로 죽은 호출이고, 정책 상수 1곳(`ReadRetryExtraAttempts`/`ReadRetryBackoff`) 공유로 이중 정의 없음.
- **account-seq lazy 해석 제거(WithAccountSeq 선구성) + 표기·헤더 계좌 동일 엔트리 고정**: 승인 — "run당 1회 캐시" 지시보다 나은 해법이고, 다계좌 잠재 불일치("A라 적고 B를 측정")를 닫았다. seq=0 폴백으로 기존 경로 보존.
- **soak은 PauseWhile seam만**: 승인 — soak의 "자기 기록 외 무접촉" 성질 보존. lock 보유가 검증 run 동안만임을 확인(콘솔 대시보드 대기 중 soak 정상 동작).
- **run 전역 재시도 예산 없음(호출당 2회·최대 45s)**: 승인 — 운영자 참관 절차이고 진행 로그가 대기를 알린다. 실측 창이 나오면 1.3 retry matrix에서 수치로 대체.
- **시도 전건 Call 기록**: 승인 — 1.3의 입력이 시도 로그 그 자체라는 논거 타당.
- advisory 하드 차단 부재 + "미측정 달력을 레일화하지 않는" 정적 가드: 지시 취지 그대로.

docs/WORKFLOW.md 미커밋 수정은 사용자 편집으로 확인 — 스테이징하지 않음(사용자 소관).

## Manager 판정 (1.8 발주 — §0.7 비해당 사유, 2026-07-26)

- 콘솔·soak **재시작 버튼은 §0.7 운영 토글이 아니다**: 게이트·주문 능력·리스크 한도와 무관한 read-only 도구의 프로세스 재기동이며, 승인 등가성(새 nonce 타이핑)과 게이트 라우트 부재는 그대로다. 반대로 게이트 ON은 여전히 콘솔 범위 밖(2c 후 사람 승인).
- 웹 재시작의 핵심 위험은 "존속 측정의 프로세스 경계 훼손"이다 — re-exec은 pid가 유지되므로, 존속 판정 기준이 `process.instance_id`(기동마다 신규)임을 확인·고정하는 것을 1.8의 완료 조건으로 명시했다. 클라이언트 상태 소멸이 측정의 본질이고 instance_id가 그 증거다.
- 한글화에서 evidence record의 `title`은 영어 유지 — 이미 쌓인 레코드와의 비교성이 표시 언어보다 우선한다. 화면 라벨은 렌더 계층 매핑으로 해결(전 단계 drift 테스트).

## 2026-07-27 · task 1.8 (콘솔 운영 자동화·전면 한글화) 구현

### ① 존속 판정 기준 — 확인 결과: **이미 instance_id 기준**, pid 기준 없음

`steps.go stepConditionalPersist`는 `registrar == r.process.InstanceID`로 판정하고,
`registeringProcess()`는 레코드의 `Process.InstanceID`를 돌려준다. `Process.PID`를 읽는 곳은
`record.go NewProcess`(기록)와 `runner.go writeBanner`(배너 출력) 둘뿐이다. **교정할 것이 없었고,
대신 양방향으로 고정했다:**

- 런타임(`TestTheProcessBoundaryIsTheInstanceIDAndNotThePID`): ⓐ **pid 동일 + instance_id 신규**
  → `pass`(= re-exec은 유효한 프로세스 경계), ⓑ **instance_id 동일 + pid 상이** → `awaiting-restart`
  (= 자기 조건주문의 존속을 스스로 증명할 수 없다). 테스트는 기록의 pid 집합이 1개임까지 확인해
  하네스가 pid 고정을 그만두면 통과가 우연이 되지 않게 했다.
- 정적(`TestTheProcessBoundaryIsNeverJudgedByPID`): record.go·runner.go 밖에서 `process.PID`를
  읽으면 실패. runner.go에서도 `Fprint` 줄이 아니면 실패. `registrar == r.process.InstanceID`
  문자열이 사라져도 실패.

### ① 재시작 — 핸들러는 재실행하지 않는다, Serve가 한다

`POST /restart`(세션+CSRF)는 ⓐ 진행 중 실행 거부 → ⓑ 핸드오프 mint → ⓒ 인터스티셜 렌더 →
ⓓ 포트를 채널로 Serve에 넘김, 까지만 한다. **핸들러가 exec하면 안 되는 이유는 명확하다**: 방금
응답한 커넥션을 쥐고 있고 리스너가 열려 있어 후속 프로세스가 그 포트를 bind할 수 없다. Serve가
settle → Shutdown → `Relaunch(port)` 순으로 수행한다. 실제 루프백 소켓 위에서 후속 프로세스가
같은 포트를 bind할 수 있음을 `TestServeIsWhatExecutesTheRelaunchAndItReleasesThePortFirst`가 확인한다.

**진행 중 검증에서는 재시작을 거부한다**(게이트가 아니라 안전 사유): 미체결 주문을 남긴 채
프로세스를 버리는 유일한 경로이기 때문이다. 먼저 끝내거나 거부하라고 화면이 말한다.

### ① 핸드오프 토큰 생애주기 (구현된 그대로)

| 단계 | 동작 |
|---|---|
| mint | 이미 세션+CSRF를 통과한 `POST /restart` 핸들러에서만. `internal/handoff`가 20바이트 crypto/rand 토큰을 `console-handoff.json`(0600, 데이터 디렉터리 0700 안)에 쓰고 fsync. TTL 2분 |
| 전달 | 인터스티셜이 **같은 origin**의 `/?handoff=<tok>`으로 3초 뒤 meta refresh. 포트를 고정하므로 상대 URL로 충분하다 |
| consume | 새 프로세스의 `session0`이 `handoff` 쿼리를 보면 `Consume` 호출. **파일을 먼저 지우고 나서** 판정한다 — 일치/불일치/만료 전부 토큰을 소모한다 |
| 성공 | 세션 쿠키 발급 + 쿼리에서 `handoff`·`session` 제거 후 303. 주소창·히스토리·referrer에 남지 않는다 |
| 재사용 | 파일이 없으므로 `ErrNoHandoff` → 403 "1회용" 안내 |
| 오답 | `ErrMismatch` → 403. **그리고 진짜 토큰도 함께 소모된다** — 지시("재사용 거부")보다 엄격하다. 추측 창을 남기지 않는 방향이고, 비용은 이미 앉아 있는 콘솔에서 클릭 한 번이다 |
| 만료 | `ErrExpired` → 403, 파일은 이미 소모 |
| 미청구 | 2분 뒤 무효. `Discard()`도 있지만 기동 시 호출하지 않는다(후속 프로세스가 브라우저보다 먼저 뜨므로) |

**넘어가는 것은 세션뿐이다.** CSRF 토큰도, 승인도, `spent`도 넘어가지 않는다 —
`TestARestartResetsTheProcessCapAndNothingElse`가 ⓐ `spent=false`, ⓑ 이전 CSRF 거부,
ⓒ 시작하면 여전히 **새 nonce 폼**에서 정지·전송 0건을 고정한다. 새 세션 URL은 종전대로 터미널에도
출력된다(터미널 점유가 여전히 신뢰의 뿌리다).

### ② soak 재시작 — autostart 스크립트와 같은 메커니즘

`cmd/tossctl/soakproc.go`가 `pgrep -f "tossctl soak run"`(스크립트와 **동일 패턴**, 테스트로 고정)
→ SIGINT → 종료 대기(최대 30초) → `setsid`로 재기동, 출력은 기존 `soak.log`에 append.
**종료하지 않으면 새로 띄우지 않는다** — 기록 하나에 서베이 둘이 붙고 rate 예산을 나눠 쓰는 편이
사람이 필요한 서베이보다 나쁘다. 자기 pid는 절대 신호하지 않는다. 네 동작(find/signal/alive/spawn)
전부 패키지 변수 seam이라 테스트는 fork·signal하지 않는다.

### ② soak 자기 업그레이드 — 사이클 경계에서만, 실패 방향은 전부 "계속 돈다"

`soak.Options`에 `Binary`(지문)·`ReExec`(재실행) seam 추가. 사이클을 기록·fsync한 **뒤**에만
비교하므로 업그레이드가 측정을 잃게 할 수 없다. 지문을 못 읽으면·바뀌지 않았으면·재실행이 실패하면
전부 nil을 돌려주고 서베이는 계속한다(한 빌드 뒤처지는 것은 대시보드의 한 줄, 멈추는 것은 하루 손실).

**기록·streak 무영향을 실제로 보였다**(`TestTheRecordAndTheStreakSurviveAnUpgrade`):
가짜 시계로 3일치 사이클 → 사이클 사이에 재설치 → 경계에서 핸드오버 → **같은 파일**을 여는 후속
러너가 2일 추가 → 파일에서 다시 읽은 streak = 5일, 앞선 3줄은 바이트 그대로, 마지막 사이클의
빌드 지문은 후속 프로세스의 것.

### ② stale 판정 — 콘솔은 자기 자신, soak은 자기 기록 (선택과 사유)

| 대상 | 근거 | 왜 이것인가 |
|---|---|---|
| 콘솔 | 기동 시 1회 뜬 `binstamp.Self()` vs 렌더 시점의 같은 경로 | 이 프로세스는 자기가 무엇을 적재했는지 안다. 물어볼 곳이 없다 |
| soak | **사이클마다 기록에 찍는 자기 지문**(`Cycle.binary`, additive·`omitzero`) vs 설치된 바이너리 | 프로세스 테이블 스크래핑도, 한 시간 뒤면 의미 없는 pid 추적도 필요 없다. 서베이 자신의 진술이고 대시보드는 이미 그 파일을 읽고 있다 |

지문은 mtime+size다(해시 아님 — 사이클마다 40MB를 해싱할 이유가 없다). **관측 실패는 변경이
아니다**: `Stamp.Same`은 어느 쪽이든 unknown이면 "같다"를 돌려주므로, 근거 없는 경고가 화면에
뜨는 일이 없다(실패 방향은 놓친 알림이지 거짓 알림이 아니다). `binstamp.SelfPath`는 리눅스가
붙이는 `" (deleted)"`를 떼어내 rename 방식 재설치도 감지된다.

### ③ 한글화 — 번역한 것과 원문으로 남긴 것

**번역**: 콘솔 전 화면, runner 진행 출력(`r.out`), 배치·단계별 승인 프롬프트, `--list` 단계 목록,
리포트·`verify status` 텍스트, preflight 생략 사유, 단계 판정 사유, `Step.Proves`·`Step.Procedure`.

**원문 유지 — 각각의 사유:**

| 대상 | 사유 |
|---|---|
| step ID, verdict 값, observation key/value/**detail** | 기록의 조인 키이자 증거 삼중항 그 자체. 화면은 verdict를 `pass (통과)`처럼 **기록값을 앞세워** glossing한다 |
| 브로커 어휘 (`clientOrderId`·`sellableQuantity`·`order-hours-closed`·`HTTP 422`·endpoint 경로·`WATCHING`·`SINGLE`) | 번역하면 증거를 브로커 응답과 대조할 수 없다. `TestTheBrokerVocabularyIsNotTranslated`가 고정 |
| `Step.Title` | 레코드 `title` 필드. Manager 판정대로 기존 레코드와의 비교성 우선. 화면 라벨은 `verifylive.StepLabel` 매핑이고 `TestEveryCatalogueStepHasAKoreanLabel`이 전 단계 커버를, `TestTheEnglishTitleIsUntouched`가 Title에 한글이 섞이면 실패시킨다 |
| `PlannedMutation`의 `Ends`/`Note`/`Pricing`/`Quantity` | **`approval.plan_digest`의 해시 입력**. 번역하면 이미 디스크에 있는 레코드(M3의 `sha256:fac7f233…`)와 digest가 갈라지고 "같은 계획이 재시작을 넘어 재승인되었다"가 확인 불가능해진다 |
| JSON 다운로드(`/report.json`) | 지시대로 무변경 |

**digest를 지키면서 화면을 한글로 하는 방법**: `EndsKO`/`NoteKO`/`PricingKO`를 `json:"-"`
사이드카로 나란히 싣고 `Plan.WriteLines`가 그것을 렌더한다. digest 불변을
`TestTheDisplayLanguageCannotMoveThePlanDigest`가 고정한다(태그 한 글자를 지우면 다른 증상 없이
digest가 움직이므로 테스트가 필요하다). 번역이 빠진 줄은 `orFallback`이 원문을 보여준다 —
라이브 요청의 되돌리기 자리가 공백이 되는 것보다 낫다. 카탈로그 전수 커버는
`TestTheApprovalSummaryIsRenderedInKorean`이 강제한다.

**CLI도 한글이다**(`verify run`은 runner 출력을 공유한다 — 태스크가 허용). grep 가능성을 위해
영어를 남겨야 하는 곳은 위 표의 다섯 줄뿐이고, 모두 테스트로 고정했다.

### 편차 (전건, 사유 포함)

1. **재실행 argv에 `--port`를 고정한다.** 지시는 "argv 보존"이지만, `--port` 없이 띄운 콘솔은
   OS가 고른 포트에 브라우저가 앉아 있다. 그 포트를 유지하지 않으면 핸드오프 토큰이 갈 곳이 없고
   ①의 "브라우저 연속성"이 성립하지 않는다. 나머지 argv(서브커맨드·`--config-dir`·기타 플래그)는
   순서까지 보존한다. `argvWithPort` 순수 함수 + 4케이스 테스트.
2. **오답 핸드오프도 토큰을 소모한다.** 지시는 "재사용 거부"였다. 추측 창을 남기지 않는 방향이
   더 보수적이고, 비용은 이미 콘솔 앞에 앉은 사람의 클릭 한 번이다.
3. **진행 중 검증에서 콘솔 재시작 거부.** 지시에 없던 거부를 하나 추가했다. §0 취지 — 살아 있는
   주문을 남긴 채 프로세스를 버리는 경로를 만들지 않는다.
4. **observation `detail`은 번역하지 않았다.** "화면 전부 한글" 지시와 "evidence-record field
   values 원문 유지" 지시가 겹치는 지점이다. key/value/detail을 증거 삼중항으로 묶어 통째로
   원문에 두는 쪽을 택했다 — 절반만 번역하면 기록이 두 언어로 갈린다. 리포트 화면의 chrome
   (제목·열 이름·판정·미검증 목록 안내)은 한글이다. **Manager 판단을 남긴다.**
5. **`conditional-trigger`의 deferred 사유 중복 접두사 제거.** 카탈로그의 `Deferred`가 이미
   "별도 세션 — 시장 조건 필요:"로 시작하는데 `deferStep`이 같은 문구를 한 번 더 붙이고 있었다
   (1.7 이전부터의 중복). 한 번만 말한다.
6. **`soak.Cycle`에 `binary` 필드 추가**(additive·`omitzero`). 이 필드가 없던 빌드가 쓴 줄은
   그대로 파싱되고 unknown으로 읽힌다.
7. **정적 가드 2건 추가·1건 확장**: 콘솔 패키지는 `os/exec`·`syscall`을 import하지 않는다(전부
   seam); 라우트 표의 상태 변경 목록에 `/restart`·`/soak/restart` 추가(라우트 하한 7→9);
   "콘솔은 쓰지 않는다" 가드의 주석에 핸드오프 seam을 명시.
8. **비유닉스 플랫폼**: `reexecSelf`는 `errors.ErrUnsupported`를 담은 정직한 오류를 돌려주고
   버튼은 그대로 보인다(사라져서 어느 빌드인지 헷갈리게 하지 않는다). `setsid`도 no-op.
   이 change가 측정하는 것은 어차피 리눅스에서만 돈다.

### 레일 확인 (무변경)

- 게이트 라우트 없음(9개 라우트 전부 대시보드·검증·재시작·리포트). 대시보드는 여전히 "이 콘솔은
  게이트를 켜지 않는다"고 적는다.
- 승인 우회 없음 — 판정은 여전히 `verifylive.Batch.Verify` 하나뿐(`static_test.go`).
- 비대화 승인 경로 신설 없음 — 핸드오프는 **세션**만 준다.
- 배치 승인은 여전히 화면에 방금 표시된 nonce를 사람이 타이핑해야 한다.
- redo 집합은 여전히 기록에서만 계산한다(폼 가드 무변경).
- 재시작은 도구 재기동이다(§0.7 비해당 — Manager 판정 기록됨).

### 남은 위험 · 범위 밖

| 항목 | 상태 |
|---|---|
| observation `detail` 영어 | 위 편차 4. Manager 판단 대기 |
| `pgrep` 의존 | autostart 스크립트와 같은 의존이다. 없으면 재기동이 오류를 돌려주고 화면이 그대로 보여 준다 |
| 핸드오프 중 브라우저가 늦게 돌아옴 | 2분 창을 넘기면 터미널의 새 세션 링크를 쓰면 된다. 인터스티셜이 그렇게 적는다 |
| 재시작 인터스티셜의 3초 | 고정값. 느린 기계에서 이르면 브라우저 오류 페이지가 뜨지만 링크가 같은 페이지에 있다 |
| soak 재기동 후 첫 사이클까지의 공백 | 재기동은 새 프로세스이므로 interval 처음부터. streak는 일 단위라 영향 없다 |
| `docs/WORKFLOW.md` 미커밋 수정 | 작업 시작 시점에 이미 워킹트리에 있던 사용자 편집. 손대지 않았고 스테이징도 하지 않았다 |

## Manager 판정 (1.8 콘솔 자동화·한글화, 2026-07-27)

독립 검증: 전체 스위트 -race 2658/50pkg 0 FAIL 재실행, instance_id 판정(steps.go:630)·핸드오프 선삭제 소모(handoff.go Consume)·실행 중 재시작 거부(errRunInFlight)·digest 고정 테스트 직접 확인. 편차 8건 전건 승인:

- **--port re-exec argv 고정**: 승인 — 핸드오프 착지의 전제.
- **오추측이 실토큰 소모**: 승인 — 추측 창 제거, 지시보다 엄격한 방향.
- **검증 실행 중 재시작 거부**: 승인 — 실주문이 걸려 있는 프로세스를 버리지 않는다(§0 정합). 지시에 없던 거부의 추가는 보수 방향.
- **observation detail 영어 유지**: 승인 — 증거 삼중항(key·value·detail)을 한 언어로 유지하는 것이 기존 레코드 비교성에 옳다. 화면 크롬 한글화로 사용자 요구 충족. (Manager 판단 요청 건)
- **plan digest 보존 설계**(영문 필드+json:"-" 한글 사이드카+digest 고정 테스트): 승인 — M3 "재시작 간 동일 digest" 증거의 연속성이 표시 언어보다 우선. 이 판단이 1.8의 하중 결정.
- 중복 prefix 제거·Cycle.binary additive·정적 가드 확장(라우트 9·os/exec 금지)·비unix ErrUnsupported: 전건 승인.
- instance_id는 이미 올바랐음(양방향 테스트로 고정됨) — 발주 시 우려는 기우로 판명, 고정 테스트는 가치 있음.
