# proposal-freeze 적대 리뷰 r2 — codex (원문, 마지막 회차)

- 날짜 2026-11-25 · 보이스: codex-cli 0.165.0 `codex exec -s read-only -C ~/tossos-work/a197-impl -o …` (model 로그 표기 `gpt-6-astra`), stdin </dev/null, rc=0 · 기준 HEAD 6a3306f8c(재개 개정 커밋, 작업트리 깨끗).
- 작성자 계열(Claude Opus)과 다른 계열. 적대적 Eng 관점. 아래는 `-o` 로 받은 최종 메시지 원문(무편집).

## 프롬프트

```text
TossOS(실제 돈을 다루는 Go 자동매매 엔진) change a197 의 **개정본**(재개 개정 2026-11-25)을 적대적 Eng 관점으로 리뷰하라. 이것이 proposal-freeze 의 마지막 회차(r2)다. 파일을 수정하지 말 것(읽기 전용).

대상 문서(전부): openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/ 의 proposal.md, design.md(D1~D11, §2~§7), specs/exit-policy/spec.md(ADDED 7 · MODIFIED 4), specs/order-execution/spec.md(MODIFIED 1 — 「IN_DOUBT 해소」), tasks.md, issues.md, review.md 의 「Manager 결정 (2026-11-25)」 절. 근거 산출물: analysis/resume/decision-2-indoubt.md, analysis/resume/decision-2-addendum.md, analysis/probe/resume-probe-output.txt, analysis/probe/resume2-engine-probe-output.txt, analysis/probe/resume2-execgw-probe-output.txt, analysis/cited-ast/README.md, analysis/function-logic/*/function-logic-map.md.

change 요지: 기본 꺼진 토글이 켜진 엔진은 보호 손절이 브로커에서 opposite-pending-order-exists 로 확정 거절되면, 그 관측 주기의 **끝**(ObserveOnce 의 defer endExitOrderCycle 첫 문장)에서 그 종목의 미체결 목록을 읽어 엔진에 귀속되지 않는 매수 주문 하나를 취소한다. 주기당 1 건, 같은 주문 번호에는 결과와 무관하게 한 번만(표시 = 원장의 CANCEL attempt 행), 취소가 IN_DOUBT 면 그 자리에서 Attempt.ResolveUnresolved 로 park(계정 래치 Gate.Block 없음) — 그 대가로 그 종목 신규 진입이 운영자 해동까지 거절된다.

현재 소스와 정본을 직접 확인하라. 정본: openspec/specs/exit-policy/spec.md(:554 · :885 · :905 · :917), openspec/specs/order-execution/spec.md(「IN_DOUBT 해소」 :34-260, 특히 :44 · :46 · :125 · :126 · :131). 코드: internal/app/engine/exitloop.go(ObserveOnce · Run · record · submit · clearTheSymbol · alertProposalRefused · isProtective), internal/app/engine/exit_order_log.go(endExitOrderCycle), internal/app/engine/exitwiring.go(Context.ExitObserver), internal/app/engine/in_doubt_settle.go, internal/app/engine/a198_in_doubt_cancel_close.go, internal/app/engine/attempt_thaw_command.go, cmd/tossctl/engine_attempt_resolve.go, internal/execgw/(gateway.go Cancel · checkSymbolFree · attemptTargets · unsettledFor, a094_shared_judgements.go, indoubt.go Resolve · park · ScanOrders, amend_indoubt.go resolveCancel, cancel_refusal.go, cancel_refusal_latch.go, retry.go EntryGate), internal/journal/(resolution.go ResolveUnresolved · UnresolvedAttempts · OperatorResolve, durability.go Prepare · transition, recovery.go PendingAttempts, daily_buy.go EngineBrokerOrderIDs, exit_proposal_release.go, a198_cancel_target_close.go, cancel_refusal_latch.go, schema.go), internal/reconcile/(recovery.go Run, compare.go LocalStateFromJournal, mismatch.go), internal/app/engine/precheck.go, internal/official/client.go · token.go, internal/config/engine.go(mergeEngine), internal/app/engine/engine.go(NewContext, Context.Audit), internal/audit/audit.go.

특히 다음을 겨냥해 찾고 P0/P1/P2 로 분류하라(각 항목에 file:line 근거):
1. park 가 손절·익절·길 치우기·수동 flatten 을 **다른 문으로** 막는 경로(진입 게이트 · 대사 · 편입 · 발의 해제 판정 · 체결 감지 · 기동 복구 · 알림 래치 · strategy 경로 등). design D11 의 「막지 않는다」 주장이 거짓인 곳.
2. 종목 범위 park 의 구멍: 시장·종목 정규화 차이로 다른 종목을 막거나 같은 종목을 못 막는 곳, 계정 래치가 다른 경로로 걸리는 곳, park 가 재시작·해동 뒤 남기는 것.
3. 「한 번만」 표시(D7 — kind CANCEL ∧ source engine ∧ target_order_id 바이트 ∧ 같은 계좌·시장·종목·거래일 ∧ 상태 무관)의 구멍: 재시작, 경합, 거래일 경계(미국장 ET 자정 · KR 야간), 번호 재사용, 게이트웨이가 기록 전에 거절한 경우, RECORDED 에서 죽은 경우, flatten 이 같은 번호를 취소한 경우, 길 치우기 CANCEL 행과의 혼동.
4. 주기 끝 실행(D5)의 순서 문제: defer 순서, 이른 반환 경로, ctx 취소, ExitOrderLog nil 갈래(B1), 같은 주기의 다른 defer(flushStrategyProtectionAlerts)와의 순서, 다음 주기의 재발의와의 순서, 판정 경로에서 넣은 대기열 항목이 주기 끝에서 낡는 경우(사이에 그 포지션이 청산·변경됨).
5. 거부하게 되는 정상 입력: design §3 · spec 이 빠뜨린 것, 「종목 신규 진입 차단」의 실제 범위(같은 종목의 다른 포지션·편입·전략 레인 포함), 운영자 해동 절차의 오류(attempt-resolve 의 CONFIRMED/FAILED 번역, --broker-order-id 값).
6. D4 조회 빈도 선택(거절 주기마다 종목 필터 OPEN 1 GET — 공표 스냅숏 없음)의 근거와 비용이 맞는지, 더 작은 안이 코드에 있는지.
7. 토글 꺼짐이 오늘 동작과 달라지는 곳, audit 자리(Context.ExitObserver 의 c.Audit, 열린 항목 O-1).
8. spec delta 의 MODIFIED 가 정본 원문과 어긋나거나(복사 누락·변경) 다른 SHALL 과 모순되는 곳.
9. design §4 시험 · 변이 목록이 놓치는 것.
결론은 한글로: 판정 한 줄, 발견 목록(P0/P1/P2, 근거 좌표), 각 발견에 제안(새 가드·정책은 「제안」으로만 표시).
```

## 최종 메시지

**판정: proposal-freeze 보류 — P0 확인 없음, P1 3건·P2 5건. 기존 O-1도 미결이며, 현재 개정본의 재시작·한 번만·운영자 인계 계약은 닫히지 않았습니다.**

아래는 현재 소스로 확인한 **설계 반례와 명세 결함**입니다. 새 구현을 실행해 재현한 결과는 아닙니다.

1. **P1 — IN_DOUBT → park → 결과 알림 사이의 중단을 복구하지 못합니다.**

   [design.md:29](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md:29)는 원장 쓰기 실패를 잔여로 다루지만, 정상적인 두 기록 사이에서 프로세스가 종료되는 경우도 있습니다.

   - **IN_DOUBT 기록 후 park 전 종료:** 기동 복구가 해당 attempt를 일반 Resolver로 보냅니다([recovery.go:266](/home/daniel/tossos-work/a197-impl/internal/reconcile/recovery.go:266), `:300`). 조회로 해소하지 못하면 [indoubt.go:376](/home/daniel/tossos-work/a197-impl/internal/execgw/indoubt.go:376)의 park가 **계정 전체 `ReasonUnresolvedInDoubt` 래치**를 겁니다. D11의 종목 범위 결과를 보장하지 못합니다.
   - **park 기록 후 두 번째 critical 전 종료:** park는 [PendingAttempts:30](/home/daniel/tossos-work/a197-impl/internal/journal/recovery.go:30)에서 제외됩니다. D7은 재시도를 막고, 설계에는 누락된 결과 알림을 복구하는 경로가 없습니다. 첫 알림만으로는 이후 생성된 attempt의 해동 절차를 전달하지 못합니다.

   **제안:** 두 중단 지점의 기동 처리·누락 알림 복구 정책을 명시하십시오. 기존 일반 Resolver를 사용할 경우 계정 래치 예외와 해제 절차까지 공시해야 합니다. §4의 RECORDED 중단 시험만으로는 부족합니다.

2. **P1 — D7은 ‘같은 주문에 한 번’이 아니라 ‘취소한 거래일마다 한 번’입니다.**

   [design.md:25](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md:25)의 표시는 **취소 intent의 거래일**을 사용합니다. 게이트웨이는 이를 대상 주문 생성일이 아니라 현재 시각으로 만듭니다([gateway.go:1020](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:1020)). 경계는 시장별 달력 날짜입니다([market.go:154](/home/daniel/tossos-work/a197-impl/internal/clock/market.go:154)).

   따라서 동일 OPEN 주문이 경계를 넘어 존속하면, 미국은 **ET 자정**, 한국은 **KST 자정** 뒤 다시 취소 대상입니다. CANCEL park도 다음 CANCEL을 막지 않습니다. 반면 D6 알림 key에는 거래일이 없어, 재취소하면서 새 시도 알림은 생략될 수 있습니다.

   [exit delta:67](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/specs/exit-policy/spec.md:67)의 일자 제한과 `:73–75`의 무조건적인 재시작 보장은 서로 맞지 않습니다. 실제 브로커에서 해당 주문이 경계를 넘어 존속하는지는 제공 증거로 확인되지 않았습니다.

   **제안:** 주문 생애당 한 번인지 거래일당 한 번인지 Manager 결정을 고정하고, 표시·알림의 식별 범위를 맞추십시오. 양 시장 자정, 자정을 넘긴 재시작, 기존 번호 재사용을 시험에 추가하십시오.

3. **P1 — 주기 끝에서 실행할 때, 취소 권한의 근거가 된 포지션을 다시 확인하지 않습니다.**

   [design.md:35–49](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md:35)는 과거 손절 거절을 대기열에 넣고, 실행 시 PLACE/AMEND park와 OPEN 주문만 확인합니다. 사이에 포지션이 청산되거나 세대·수량·보호 상태가 바뀌어도 제외하는 규칙이 없습니다.

   실제 순서도 `flushStrategyProtectionAlerts`가 먼저, `endExitOrderCycle`이 나중입니다([exitloop.go:482](/home/daniel/tossos-work/a197-impl/internal/app/engine/exitloop.go:482)). 게이트웨이 CANCEL은 현재 보호 포지션의 유효성을 확인하는 경로가 아닙니다([gateway.go:407](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:407)). 따라서 **이미 사라진 손절 필요성을 근거로 사람의 매수를 취소하는 실행**이 허용됩니다.

   **제안:** 실행 직전 포지션 존재·세대·보호 필요성을 재검증하는 가드, 또는 관측 당시 권한이 언제까지 유효한지에 관한 정책이 필요합니다. 대기열 적재 후 청산·세대 변경·ctx 취소 시험과 해당 가드 삭제 변이를 추가하십시오.

4. **P2 — 5xx 뒤 ‘다음 손절 접수’를 명세가 보장합니다. park가 보장할 수 있는 결과가 아닙니다.**

   [exit delta:85–87](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/specs/exit-policy/spec.md:85)는 취소가 5xx면 다음 관측 주기의 손절이 **접수된다**고 단정합니다. 그러나 취소가 실제로 실행되지 않았다면 외부 매수는 남고, 브로커는 다시 같은 이유로 거절할 수 있습니다.

   실제 [gateway.go:814](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:814)가 보장하는 것은 park를 이유로 감소 mutation을 막지 않는다는 것뿐입니다. 이는 [design.md:62](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md:62)의 반복 거절 설명과도 충돌합니다.

   **제안:** 무조건 보장을 ‘park 때문에 제출이 차단되지 않는다’로 정정하십시오. ‘취소 실행 후 응답 유실’과 ‘취소 미실행 후 5xx’를 나눠 시험해야 합니다. §4 시험 12도 같은 수정이 필요합니다.

5. **P2 — 운영자 절차의 ‘매수가 아직 살아 있다 → FAILED_CONFIRMED’는 너무 넓습니다.**

   [design.md:103](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md:103)는 살아 있는 주문을 취소 실패로 번역합니다. 현재 해소기는 다르게 판단합니다.

   - 취소·정정 처리 중은 확정 실패가 아닙니다([amend_indoubt.go:116](/home/daniel/tossos-work/a197-impl/internal/execgw/amend_indoubt.go:116)).
   - OPEN도 안정 관찰 조건 뒤에야 실패로 닫습니다(`:127`).
   - REPLACED는 park합니다(`:145`).
   - 부분 체결 후 취소는 CONFIRMED입니다(`:81`).

   현 절차대로 앱의 ‘남아 있음’만 보고 해동하면, 아직 처리 중인 취소의 불확실성을 제거했다고 기록할 수 있습니다. **CONFIRMED에 취소 대상 매수 번호를 넣는 설명 자체는 현재 코드와 맞습니다.**

   **제안:** 처리 중·미지·대체됨은 즉시 해동하지 않는 상태별 확인표를 두십시오. 부분 체결 후 취소와 완전 체결도 구분하고, 실제 운영자 명령을 통한 두 종결 분기를 시험하십시오.

6. **P2 — ‘다른 종목 무관’과 ‘해동하면 풀림’에는 알림 실패 예외가 빠졌습니다.**

   D6은 실제 critical 입구를 사용합니다. 이 입구는 기록 실패 시 **계정 전체 `ReasonAlertUndelivered` 래치**를 걸고 운영 모드 승격도 시도합니다([record_only.go:137](/home/daniel/tossos-work/a197-impl/internal/obs/record_only.go:137), `:148`, `:157`). 따라서 D11에서 직접 `Gate.Block`을 호출하지 않는 것만으로 계정 범위 영향을 배제할 수 없습니다.

   이는 기존 안전 장치가 새 알림 경로에서 작동하는 경우입니다. attempt 해동은 해당 알림 래치나 승격된 모드를 해제하는 절차가 아닙니다([attempt_thaw_command.go:78](/home/daniel/tossos-work/a197-impl/internal/app/engine/attempt_thaw_command.go:78)).

   **제안:** 정상 알림 경로에서의 종목 범위 보장과 알림 실패 예외를 분리하고, 운영 절차에 별도 복구를 명시하십시오. 실제 Notifier의 기록·전달 실패 시험이 필요합니다. 기존 안전 래치를 제거하자는 제안은 아닙니다.

7. **P2 — 정상 외부 주문을 영구 제외할 수 있는 귀속 판독 범위가 §3에 빠졌습니다.**

   [EngineBrokerOrderIDs:350](/home/daniel/tossos-work/a197-impl/internal/journal/daily_buy.go:350)의 SQL에는 시장·종목·거래일 제한이 없습니다. 같은 계좌의 과거 PLACE/AMEND 번호와 일치하면 모두 제외됩니다.

   번호가 다른 거래일·시장·종목에서 재사용되는 경우, 오늘의 정상 외부 매수가 과거 엔진 주문으로 간주되어 취소되지 않습니다. D7은 거래일로 식별 범위를 제한하면서 D2는 무기한 계좌 전체 번호를 사용하는 비대칭입니다. **브로커 번호의 영구 전역 유일성은 제공 근거에서 입증되지 않았습니다.**

   **제안:** 현재 보수적 제외를 유지한다면 §3의 정상 입력 거부 사례로 명시하고 충돌 fixture를 추가하십시오. 귀속 범위 축소는 별도 정책 제안으로 판단해야 합니다.

8. **P2 — D4의 21.05초·51.05초는 ‘유계 부분 최악’으로도 근거가 부족합니다.**

   [design.md:22](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md:22)의 계산은 토큰 **잠금 대기**만 별도로 제외합니다. 실제로는 요청 전 토큰 교환과 401 후 교환이 별도 HTTP 요청입니다([client.go:336](/home/daniel/tossos-work/a197-impl/internal/official/client.go:336), `:361`; [token.go:132](/home/daniel/tossos-work/a197-impl/internal/official/token.go:132)). 350ms도 계좌 예산의 최소 간격이지 대기시간 상한이 아닙니다. 단계 전체 ctx의 구체적인 제한값도 미정입니다.

   401 재전송 최대 두 번은 현재 코드에서 확인했습니다. 그러나 그 사실만으로 51.05초가 되지는 않습니다. 약 7%도 5초마다 OPEN 한 번이라는 조회 몫의 계산이며 전체 추가 요청 비용은 아닙니다.

   **제안:** 수치를 조건부 예산 추정으로 정정하고 인증·계좌 예산 대기·단계 ctx를 포함해 계산하십시오. §4 시험 16에 인증 갱신과 예산 경합을 포함해야 합니다. 확인한 경로에서는 재사용 가능한 공표 스냅숏이 없어, 새 조회 선택 자체를 결함으로 보지는 않습니다.

반대로, 다음은 이번 검토에서 결함으로 확인하지 않았습니다.

- **직접 park가 정상 완료된 CANCEL:** Pending 제외, 감소 mutation 조기 통과, 별도 취소 intent, PLACE/AMEND 중심 체결·대사 판독을 따라 확인했습니다. 이것만으로 손절·익절·길 치우기·flatten을 막는 추가 경로는 찾지 못했습니다.
- **차단 범위:** 같은 시장·종목의 다른 포지션과 전략 레인도 공통 게이트의 영향을 받습니다. 정확한 범위는 ‘신규 포지션’보다 넓은 **노출 증가 요청**이며, 수량·가격을 높이는 AMEND도 포함합니다([gateway.go:441](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:441)). 편입까지 같은 방식으로 차단된다고 볼 근거는 없습니다.
- **정규화·경합:** 현재 CANCEL 생산 경로의 시장·종목 정규화에서 확정적인 오차단 반례는 찾지 못했습니다. 단일 exit 호출자의 D7 판독·발행 경합도 확인하지 못했습니다. flatten의 별도 source 제외와 RECORDED 표시 유지 자체는 명시한 정책과 일치합니다.
- **D5·정본 복사:** B1 앞 호출은 nil 로거 문제를 처리합니다. MODIFIED 5개 Requirement 대조에서 의도한 추가 외 정본 원문 복사 누락은 발견하지 못했습니다. 다만 위 2·4번의 개정본 내부 계약 충돌은 남습니다.
- **OFF·audit:** D9 문장 변경은 명시된 예외입니다. 그러나 audit 실패 정책은 [O-1:139](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md:139)에 그대로 남아 있어 OFF 보존·관측기 생성 실패까지 완료 판정할 수 없습니다. 기존 미결 항목이며 새 발견 수에는 넣지 않았습니다.

검토 기준은 HEAD `6a3306f8c1999df9a77fd33420217089ec7f41f3`입니다. 지정 문서·근거 산출물을 읽고 현재 코드의 관련 함수·분기를 대조했습니다. AST/FLM의 원본 SHA 66개는 현재 파일과 일치했지만, 이는 **새 실행 증거가 아닙니다**. 기존 probe는 과거 가짜 브로커 실행 증거로만 사용했습니다. 테스트·probe·전체 게이트·그래프 재생성·파일 수정은 하지 않았고 작업 트리는 깨끗합니다. ai-memory 조회는 로컬 서비스 접근 제한으로 실패했습니다. `impeccable: N/A — backend/documentation-only`.