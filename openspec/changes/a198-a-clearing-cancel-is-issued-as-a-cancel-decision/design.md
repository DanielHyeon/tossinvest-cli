# a198 Design

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | 길 치우기의 취소 결정 = 종류 `CANCEL`, 등급 위험 감소, 방향 = **취소할 주문의 방향**, 최대 수량 = 그 주문의 수량, 사유 문장은 종전 것(「cancelling the working … order … before liquidating position …」). 수동 flatten 의 취소 결정과 같은 필드 구성 | probe 대조군이 접수된 모양이다. 방향을 매도로 고정한 종전 결정은 매수 주문의 취소에 대해 내용도 틀렸다(「매도 감소」가 아니라 「그 주문의 철회」) |
| D2 | **Manager 결정 2026-11-23 — (가-B)**: 발행 자리는 관측기 안이다. `clearTheSymbol` 이 관측기가 이미 가진 값(`opts.Journal` · 관측기 시계 `o.clk` · `opts.AccountRef`)으로 `execgw.Issuer` 를 만들어 flatten 과 **같은 함수** `Issuer.IssueReduction`(종류 `CANCEL`)을 부른다. TTL·ID 생성기는 비워 둬 Issuer 기본값(`execgw.DefaultDecisionTTL` 60s · `execgw.randomID` 128비트)을 쓴다. 새 옵션 칸·새 인터페이스·`Context.ExitObserver` 편집·시험 하네스 변경 없음. 기존 `ReductionIssuer`(제출용)는 그대로 | 고칠 기존 함수가 가장 적고(analysis/edit-surface.md §3), 결정 기록이 flatten 과 같은 함수에서 나오며, a087 번들이 고정한 exitwiring.go 를 건드리지 않는다. 선행 확인(analysis/claim-check.md 「D2 선행 확인」): 운영에서 관측기 시계와 게이트웨이 시계는 같은 `clock.System()` 값 하나다. 취소 결정에 운영 모드·킬 스위치 검사가 걸리지 않음은 AST 로 확인(claim-check (a)) |
| D3 | **삭제 — Manager 결정 2026-11-23.** (가-B) 에서는 조립이 공급할 발행기가 없다(이미 필수인 Journal·AccountRef 에서 유도). 새 기동 실패 조건도 생기지 않는다 | r1 P2-5 의 「조립 실패가 엔진 런타임 전체를 막는다」는 이 삭제로 소멸 |
| D4 | **Manager 결정 2026-11-23 — 운영에서 보이는 기록.** `clearTheSymbol` 의 취소 결정 발행 실패(B9)와 취소 제출의 비확정 결과(B11 — 오류 또는 `CONFIRMED` 아님)마다 a129 의 기존 운영 로그 경로(`noteExitOrder` → 주기 끝 `endExitOrderCycle` → `writeExitOrder` → `opts.ExitOrderLog`, 운영 배선 cmd/tossctl/engine.go:680)에 한 줄을 넣는다. 줄 = 종류 취소 · 주문 방향 · 종목 · 시장 · 포지션 · 대상 주문 번호 · attempt · 결과 상태 · 사유 **코드**(토큰 거름 `exitOrderToken`). Detail · 원문 오류 · 계좌 · 수량 · 가격은 싣지 않는다(취소 줄에는 수량·가격 **키 자체**를 내보내지 않는다 — `args` 의 취소 분기). 같은 결과가 연속 주기에 이어지면 한 번(기존 `refusal` 래치). `warnEpisode` 와 `opts.Log` 연결 안은 기각(운영 무기록 / 계좌 노출). `res.cleared = false` 판정과 그 뒤 흐름은 한 글자도 바꾸지 않는다 | 실현 가능성(AST): 그 경로에서 고칠 기존 함수는 `exitOrderLine.args` **1** — 지금 `side "SELL"`·`order_type "LIMIT"` 을 고정해 내보내므로(exit_order_log.go:143-144, 분기 1) 취소 줄을 표현하려면 종류·방향을 필드에서 내보내야 한다(취소가 아닌 줄은 종전과 바이트 동일). `noteExitOrder`·`endExitOrderCycle`·`writeExitOrder` 는 무편집으로 그대로 쓴다. 줄 생성 helper·사유 코드 상수는 새 파일. 기준(기존 함수 2 개 초과면 멈춤) 이내 |
| D5 | 손절·익절 **제출** 경로(`submit` 의 `IssueReduction`)는 무편집 | 그 경로는 운영에서 동작한다(PLACE CONFIRMED 111 건) |
| D6 | 토글 없음 | 정본이 이미 요구하는 동작의 결함 수리다. 꺼 둘 「종전 동작」이 「손절 무기한 보류」다 |

## 2. 이 수리가 처음으로 실행 가능하게 만드는 경로 — 달라지는 운영 동작

오늘은 취소가 원장에서 거절되어 그 뒤의 코드가 운영에서 한 번도 돌지 않았다. 이 change 뒤 처음 도는 것:

1. 엔진 귀속 **매수** 미체결이 있는 종목의 손절: 취소 접수 → 그 주기에 치움 완료 → 손절 제출.
2. 엔진 귀속 **매도** 미체결(무장된 발의의 주문 — 익절 지정가나 앞선 손절)이 있는 종목의 손절: 취소 접수 → 「종결 증거 대기」(정본 — 매도의 취소 접수는 치움의 증거가 아니다) → 체결 감지가 그 주문의 종결을 기록한 뒤 손절 제출.
3. 취소가 모호(IN_DOUBT)로 남는 경우: 같은 종목의 미종결 attempt 가드가 손절을 보류한다(정본 — 미종결 attempt 가 있으면 치우기는 완료가 아니다). `CANCEL` 은 세션 중 해소되지 않으므로 재시작 복구까지 보류가 이어질 수 있다 — **기존 정본 규칙의 귀결이고 이 change 가 만든 것이 아니지만, 이 change 가 그 경로를 처음 열므로 공시한다.** 기존 지연 경보와 연속 실패 트리거가 그것을 알린다(tasks 3.x 에서 그 경보가 실제로 나가는지 시험).

4. 취소 1건 = 사전 단건 조회 1 + 취소 요청 1, 둘 다 관측 루프 안에서 동기(Manager 결정 2026-11-23 — 정본의 「브로커 조회를 더하지 않는다」는 길 치우기가 **대상을 찾으려고** 브로커를 읽지 않는다는 뜻이고, 사전 단건 조회는 주문 경로의 기존 계약이다; retry matrix 행 37). 사전 조회 시한 2s(`precheckTimeout`, internal/app/engine/precheck.go:58), 취소 요청은 HTTP 클라이언트 시한 15s(`defaultTimeout`, internal/official/client.go:22·137). 대상 주문 수 × (2s + 15s) 는 **HTTP 시간 예산**일 뿐 최악 지연 상한이 아니다 — 토큰 획득(internal/official/client.go:336 `tm.token(ctx)` → token.go:63 context 로 끊기지 않는 `mu.Lock()`, 다른 루프의 갱신이 잠금을 쥐면 2s 시한이 지나도 대기)·401 재전송(최대 2회)·a166 계좌 예산 admission 대기·원장 처리는 계산 밖이다. **전체 최악 지연은 미확정·미측정**(r2 P1-2 정정).
5. 매수 취소 접수 뒤 같은 주기의 손절이 브로커에서 반대 미체결(`opposite-pending-order-exists`)로 거절되면(브로커가 취소를 늦게 반영하는 경우) 그 매수는 종결 기록 전까지 원장의 미체결 목록에 남아 다음 주기가 같은 매수를 다시 취소하려 하고, 그 재취소가 거절되면 그 매수의 종결 기록이 원장에 들어올 때까지 손절이 보류된다(체결 감지 주기에 묶인 일시 상태로 추정 — 미측정). 고치지 않는다(Manager 결정 2026-11-23, §6).

2·3 의 기존 규칙들이 실제 게이트웨이 위에서 문서대로 도는지는 이 change 의 결합 시험이 처음 잰다. 어긋남이 나오면 고치지 말고 Manager 에게 보고한다(범위 판단).

## 3. 거절·침묵하는 정상 입력

없음 — 새 거절을 더하지 않는다. 종전에 조용히 실패하던 취소가 나가고, 실패하면 운영 로그에 기록된다. D3 삭제로 새 기동 실패 조건도 없다.

## 4. 시험 전략 (RED 먼저)

새 시험 파일 `internal/app/engine/a198_clearing_cancel_test.go`. **가짜 제출자를 쓰지 않는다** — 실제 `execgw.Gateway` + 실 원장(임시 경로) + 가짜 브로커 + 운영과 같은 Guardian(`RiskGuardian`, 제출용). 기존 실제-게이트웨이 하네스(`buildE2EStack`, exit_e2e_test.go:267-277)를 재사용할 수 있다. 실패 기록 단언은 `ExitOrderLog` 에 넣은 싱크의 줄로 한다. 픽스처는 합성 값만.

1. **RED(결함 재현)**: 엔진 귀속 매수 미체결 1 + 보호 손절 발의 → 현재 HEAD 에서 브로커 취소 0·손절 미제출임을 단언 실패로 남긴다(`analysis/tdd/red.txt` — 원장 거절 문구 포함).
2. 엔진 귀속 매수 미체결 → 브로커 취소 1, 원장의 attempt 가 `kind=CANCEL`·`CONFIRMED`·대상 주문 번호를 갖고, 그 결정 행에 멱등 키가 없으며 preimage 의 방향이 주문 방향(매수) — 결정의 종류는 원장에 없다(키 유무로만 남음), 같은 주기에 손절 PLACE 제출·접수.
3. 엔진 귀속 매도 미체결(무장된 발의의 주문) → 브로커 취소 1(결정 방향 매도), 그 주기 손절 미제출(종결 증거 대기), 종결 체결 스냅숏 기록 뒤 주기에 손절 제출. 접수 확정된 취소를 다시 내지 않음(브로커 취소 여전히 1).
4. 취소 제출이 브로커에서 거절 → 손절 미제출, 운영 로그 줄 1(종류 취소·결과 상태·사유 코드·대상 주문 번호; Detail·계좌·수량·가격 없음), 다음 주기 재시도, 연속 주기의 같은 거절은 줄 1 로 유지.
5. 취소 결정 발행 실패(원장 쓰기 오류 주입) → 손절 미제출, 운영 로그 줄 1(사유 코드).
6. 취소가 5xx 로 IN_DOUBT → 다음 주기 손절 미제출(미종결 가드), 기존 지연 경보가 한계 뒤에 나감(가짜 시계)과 연속 실패 경보가 각각 나감. 경보는 관측 주기가 다시 `noteDelay` 에 닿아야 나가므로(첫 호출은 시계만 시작 — exitloop.go:1925) **전체 관측 주기를 반복해** 알림 기록으로 단언한다(`noteDelay` 직접 호출 금지). 변형: 취소 attempt 가 **ACKED 로 고착**(접수 뒤 최종 쓰기 전 중단을 원장으로 재현) → 손절 미제출, 연속 실패 계수 제외, **지연 경보(30 초)가 실제로 나감**을 단언 — 안 나가면 고치지 말고 Manager 에게 보고(범위 판단이 바뀐다).
7. 조립: 운영 조립(`Context.ExitObserver`)으로 만든 관측기가 실제 게이트웨이로 2 를 구동한다(운영 조립이 지도자인 실행).
8. 손절·익절 제출의 결정은 종전과 같은 발행(`IssueReduction`, 종류 `PLACE`)임을 원장으로 단언(회귀).
9. 기존 a094 길 치우기 시험·a129 줄 시험 전부 무편집 통과((가-B)는 새 인터페이스가 없어 하네스 변경 0). 기존 a129 줄(취소 아님)의 필드는 종전과 같다.

변이(전부 CAUGHT, 변이마다 닿음 확인, 사본 하네스·무변이 대조군 GREEN·시작 sha 단언): 취소에 종전 `IssueReduction` 결정을 싣기(= 결함 복원 — 2·3·7 이 죽어야 함) · 결정 방향을 매도로 고정(2 가 죽어야 함) · 결정에 멱등 키가 붙는 발행으로 되돌리기 · 실패 기록 삭제(4·5) · 취소 줄이 `side SELL` 로 나가게 되돌리기(4) · 매도 취소 접수를 치움으로 판정(3 이 죽어야 함 — 기존 규칙이 실제 경로에서 지켜지는지).

선택 시험: `./internal/app/engine -run 'A198|A094|ExitClear|ExitObserver'`, `./internal/execgw -run 'Cancel|RiskGuardian|Issue'`, `./internal/journal -run 'Prepare|Decision'`. `-race` 는 `-run A198` 만. 전체 스위트는 종결 게이트에서만.

## 5. 운영 (사람 승인)

배포 뒤 첫 길 치우기 취소가 일어나면 원장의 `kind=CANCEL` attempt 와 그 뒤 손절 접수 시각을 `review.md` 에 시각·상태만 기록한다. 발생을 인위로 만들지 않는다(LIVE 주문 금지).

## 6. 잔여·후속

- §2-3 · ACKED(Manager 결정 2026-11-23 — 범위 밖, 이름 붙인 잔여): 취소가 ACKED 또는 IN_DOUBT 로 남으면 그 종목의 손절은 보류되고, ACKED 는 연속 실패 계수에서도 빠진다(정확히는: 같은 종목의 미종결 attempt 가 치울 대상 **전부**의 진행 중 취소뿐일 때 — `onlyEngineCancelsInFlight`; 대상이 종결 기록으로 빠지거나 다른 대상이 남으면 다시 센다). 지연 경보(30 초)는 나간다. 지금 빌드에는 ACKED 로 남은 취소를 닫는 운영자 경로가 없다 — 후속 change 『CANCEL attempt 의 복구 계약(IN_DOUBT·ACKED)』의 대상이고, a197 재개의 선행이기도 하다. (근거: 기동 복구 internal/reconcile/acked_boot.go:51 「CANCEL/AMEND … stays ACKED」, 운영자 해소 internal/journal/resolution.go:143 은 UNRESOLVED_IN_DOUBT 에서만, 세션 해소는 PLACE 만.)
- 전일 주문의 거래일 결속(범위 밖 — **알려진 정본 미충족 잔여**: 정본의 「확정된 취소를 다시 내지 않는다」와 delta 의 「한 번 나간다」를 이 경우 충족하지 않는다): `ConfirmedCancelOf` 는 취소 intent 의 거래일 = 주문 거래일을 요구한다 — 전일 주문을 오늘 취소하면 대기가 성립하지 않아 확정된 취소가 주기마다 다시 나간다. 엔진 주문은 당일 주문이라 거래일을 넘겨 원장에 「살아 있는」 주문은 브로커에서 이미 만료됐을 가능성이 높고, 그 정리는 체결 감지(a191)의 영역이다. 고치지 않는다.
- 매수 취소 접수 뒤 재취소(§2-5): 고치지 않는다.
- a087 과의 순서: a198 은 기다리지 않는다. `clearTheSymbol`(exitloop.go)과 `exitOrderLine.args`(exit_order_log.go)가 a087 의 gate ⑤ 창에 들어가는 것은 a087 종결 쪽의 기존 비용(착지 기록 또는 번들)이다. 참고: a087 번들의 `source_sha256` 은 파일 해시다 — exitloop.go 쪽은 이미 HEAD 와 다르고(과거 커밋 5c0aeb2b8 과 일치), exit_order_log.go 쪽(`exitOrderLine.args`·`exitOrderLineFor`)은 지금 HEAD 와 같아 D4 편집으로 달라진다. exitwiring.go 는 (가-B) 로 무편집.
- 과거 운영에서 이 결함이 실제 손절을 막았는지는 미확인이다.
- 「가짜가 생산 계약을 가린다」: 주문 경로의 다른 결합(정정 등)에도 같은 공백이 있는지는 보지 않았다.
