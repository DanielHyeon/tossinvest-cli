# proposal-freeze 적대 리뷰 r1 — codex (원문)

- 날짜 2026-11-23 · 보이스: codex-cli 0.165.0 `codex exec -s read-only -C ~/tossos-work/a197-impl` (model 로그 표기 `gpt-6-astra`), stdin </dev/null, rc 0 · 기준 HEAD 064f59fd9 + 작업트리 문서.
- 프롬프트: 아래 「프롬프트」 절. 질문 5 는 D6 의 결정 발행기 대조를 명시적으로 물었다(작성자가 probe 로 먼저 찾은 항목 — 그 결론은 알려 주지 않음).
- 아래는 `-o` 로 받은 최종 메시지 원문(무편집).

## 프롬프트

```text
TossOS(실제 돈을 다루는 Go 자동매매 엔진) change a197 의 proposal·design·spec delta·tasks 를 적대적 Eng 관점으로 리뷰하라. 파일을 수정하지 말 것(읽기 전용).

대상 문서: openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/proposal.md, openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/design.md, openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/specs/exit-policy/spec.md, openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/tasks.md.
이 change 는 실제 돈을 다루는 엔진이 **사람이 낸 매수 주문을 취소**하는 새 경로다(기본 꺼진 운영 토글). 보호 손절이 브로커에서 opposite-pending-order-exists 로 확정 거절되면, 엔진이 그 종목의 미체결 목록을 브로커에서 읽어 엔진에 귀속되지 않는 매수 주문을 취소한다.

현재 소스와 정본 spec 을 직접 확인하라: openspec/specs/exit-policy/spec.md 의 「보호 청산의 길 치우기는 엔진 귀속 주문만 다룬다」, openspec/specs/order-execution/spec.md, openspec/specs/engine-safety/spec.md. 코드: internal/app/engine/exitloop.go(ExitObserver.submit, record, clearTheSymbol, alertProposalRefused, isProtective), internal/app/engine/exitwiring.go(Context.ExitObserver), internal/execgw/(gateway.go Cancel/submit/prepareRequest/checkSymbolFree, guardian.go checkDecision/checkPreimage/checkIdempotencyKey, riskguardian.go IssueReduction, issue.go Issuer.IssueReduction, classify.go, refusal_code.go, indoubt.go ScanOrders, orders_source.go), internal/journal/(fills.go LiveOrdersForSymbol, daily_buy.go EngineBrokerOrderIDs, recovery.go PendingAttempts, decision.go, durability.go), internal/flatten/flatten.go, internal/config/engine.go, internal/app/engine/interlock.go recordGateSettings, internal/accountbudgetruntime/testdata/a166-wiring/shapes.json.

찾아서 P0/P1/P2 로 분류하라(각 항목에 file:line 근거):
1. 잘못된 주문을 취소하는 경로 — 다른 종목, 매도, 엔진 자신의 주문(원장에 번호가 아직 없거나 park 된 것 포함), 다른 포지션의 보호 주문, 식별 불완전 행.
2. 손절을 더 늦추는 경로 — 조회 시한, 루프 차단, 새 취소 attempt 가 IN_DOUBT·SymbolInFlight 로 남아 다음 손절을 막는 반복, rate limit.
3. 토글 꺼짐이 오늘 동작과 달라지는 곳.
4. 기존 정본 SHALL NOT 과의 모순 — ADDED 만으로 충분한가, MODIFIED 가 필요한가.
5. 계약이 기대는 기존 경로가 실제로 동작하는가(예: design D6 「기존 취소 경로(Issuer.IssueReduction → 게이트웨이 취소)」 가 생산 배선에서 성립하는가 — 결정 발행기의 Kind·Side·멱등 키와 게이트웨이 검증을 대조).
6. 감사 기록의 구멍(토글 변경 audit, 취소 사실, 알림).
7. 시험 전략(design §4)이 놓치는 변이.
결론은 한글로, 발견 목록 + 각 발견의 근거 좌표 + 제안(새 가드·정책은 「제안」으로만 표시).
```

## 최종 메시지

**판정: 현재 계약으로 freeze·구현 진행을 승인하기 어렵습니다. P0 확정 발견 없음, P1 7건·P2 4건입니다.** 핵심은 생산 취소 결정 발행 불일치, park된 엔진 주문의 외부 주문 오인, 새 취소의 `IN_DOUBT`가 손절을 막는 경로입니다.

기준은 HEAD `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`와 현재 작업트리 문서입니다. 아래 `design`, `delta`, `tasks`, `proposal`은 요청한 a197 경로를 뜻합니다. **제안은 승인된 정책이 아닙니다.**

1. **P1 — D6의 “기존 취소 경로”는 생산 exit 배선에서 성립하지 않습니다.**

   `Context.ExitObserver`는 `c.Guardian`을 발행기로 넣습니다. 생산 `RiskGuardian.IssueReduction`은 SELL만 받고 `KindPlace` 결정을 기록합니다. 이 결정에는 멱등 키가 붙습니다. 이를 CANCEL에 넘기면 **journal Prepare 단계에서 거절되어 취소 attempt조차 만들어지지 않습니다.** 키 문제만 우회해도 BUY 취소에는 SELL preimage가 불일치합니다.

   반면 flatten은 별도 `execgw.Issuer`에 `KindCancel`, 대상 주문의 실제 Side를 전달합니다. 같은 이름의 `IssueReduction`을 같은 계약으로 취급한 오류입니다.

   근거: `design:21`; [exitwiring.go:326](/home/daniel/tossos-work/a197-impl/internal/app/engine/exitwiring.go:326), [riskguardian.go:592](/home/daniel/tossos-work/a197-impl/internal/execgw/riskguardian.go:592)·`:614`, [decision.go:373](/home/daniel/tossos-work/a197-impl/internal/journal/decision.go:373), [durability.go:267](/home/daniel/tossos-work/a197-impl/internal/journal/durability.go:267), [guardian.go:363](/home/daniel/tossos-work/a197-impl/internal/execgw/guardian.go:363)·`:395`, [flatten.go:360](/home/daniel/tossos-work/a197-impl/internal/flatten/flatten.go:360)·`:659`.

   **제안:** 취소 결정 발행 계약과 생산 주입 지점을 먼저 확정하고, 실제 발행기→실제 Gateway→httptest 취소 전송까지 검증하십시오. 기존 `clearTheSymbol`도 같은 결함이 있으므로 이를 함께 고칠지 별도 범위로 분리할지 결정해야 합니다.

2. **P1 — “접수 확정 번호가 없음 = 외부 주문”은 park된 엔진 매수를 취소 대상으로 만듭니다.**

   D2와 delta는 부정 증거만으로 외부 귀속을 판정합니다. 번호 없는 `UNRESOLVED_IN_DOUBT` 엔진 매수가 브로커에 살아 있으면 이 정의에 들어갑니다. `PendingAttempts`에는 park 상태가 없고, Gateway도 위험 감소 요청에는 unresolved 진입 차단 검사를 적용하지 않습니다.

   구분이 필요합니다. 기존 `RECORDED/DISPATCH_STARTED/ACKED/IN_DOUBT`는 같은 종목의 선행 치움에서 차단됩니다. **문제는 그 보호가 park 상태까지 이어지지 않는다는 점**입니다. `EngineBrokerOrderIDs` 역시 번호가 없는 주문을 찾지 못합니다. `LiveOrdersForSymbol`은 CONFIRMED·소유 유일·미종결 조건이라 외부 귀속 증명용으로 더 부적합합니다.

   근거: `design:17`, `delta:29,43–45`; [recovery.go:30](/home/daniel/tossos-work/a197-impl/internal/journal/recovery.go:30), [gateway.go:814](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:814), [daily_buy.go:350](/home/daniel/tossos-work/a197-impl/internal/journal/daily_buy.go:350), [fills.go:1867](/home/daniel/tossos-work/a197-impl/internal/journal/fills.go:1867), [exitloop.go:1655](/home/daniel/tossos-work/a197-impl/internal/app/engine/exitloop.go:1655).

   **제안:** 외부·엔진 귀속·귀속 불명 세 상태를 구분하고, 번호 없는 park 및 소유 충돌 상태에서의 처리 정책을 명시하십시오.

3. **P1 — 취소가 모호하면 “다음 거절 주기에 다시 시도”할 수 없습니다.**

   외부 BUY 취소가 `IN_DOUBT`로 남으면 다음 주기의 `clearTheSymbol`이 미종결 attempt를 발견해 손절 발의를 억제합니다. 직접 Gateway까지 가더라도 `SymbolInFlight`입니다. 따라서 새 단계의 진입 조건인 **브로커의 opposite 거절 자체가 다시 발생하지 않습니다.** 사람이 실제 매수를 취소해도 로컬 취소 attempt가 남으면 동일합니다.

   정본은 세션 중 해소를 PLACE에 한정하고 CANCEL/AMEND는 기동에서 해소한다고 명시합니다. D7과 시험 10의 자동 재시도 기대는 이 계약과 맞지 않습니다.

   근거: `design:22,60`; [exitloop.go:1399](/home/daniel/tossos-work/a197-impl/internal/app/engine/exitloop.go:1399)·`:1659`, [gateway.go:802](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:802), [order-execution/spec.md:118](/home/daniel/tossos-work/a197-impl/openspec/specs/order-execution/spec.md:118).

   **제안:** 취소 결과별 다음 상태·해소 담당·운영자 조치를 계약에 넣으십시오. 세션 중 취소 해소를 추가하려면 별도의 정책 변경으로 결정해야 합니다.

4. **P1 — 3초 제한은 전체 단계의 지연 상한이 아닙니다. 뒤 포지션의 손절 판정까지 누락될 수 있습니다.**

   D4는 조회만 제한하고 D2는 외부 매수 전부를 순차 취소합니다. 취소 수·누적 전송 시간·예산 대기 상한은 없습니다. 공식 클라이언트 기본 HTTP timeout은 요청당 15초입니다. “한 주기에 한 번”이어도 취소가 여러 건이면 지연이 누적됩니다.

   `ObserveOnce`는 가격을 먼저 읽고 포지션을 직렬 판정합니다. 앞 포지션의 치움이 길어지면 뒤 포지션은 `quoteUsable`에서 만료되어 **그 주기 판정을 건너뜁니다.** 단순히 몇 초 늦게 판정하는 것보다 영향이 큽니다. 요청 수 계측만으로 합격 여부도 결정할 수 없습니다.

   근거: `design:17,19–20,56,63`; [exitloop.go:511](/home/daniel/tossos-work/a197-impl/internal/app/engine/exitloop.go:511)·`:521`·`:531`, [official/client.go:22](/home/daniel/tossos-work/a197-impl/internal/official/client.go:22), `tasks:13,27`.

   **제안:** 조회·취소·rate-budget 대기를 포함한 전체 지연 계약을 정하고, 여러 포지션·다수 주문·429·느린 취소에서 뒤 포지션의 손절 판정 시점을 측정하십시오. 실행 구조나 취소량 제한 변경은 Manager 결정 사항입니다.

5. **P1 — 기존 `ScanOrders`는 불완전한 목록을 성공으로 반환할 수 있습니다.**

   `HasNext=true`인데 `NextCursor=""`인 응답도 `!HasNext || next == ""`에 의해 성공 반환합니다. 앞 페이지의 외부 매수는 취소하고, 뒤쪽의 식별 불완전 행은 보지 못하는 구현이 가능합니다. D3의 “끝까지 읽지 못하면 취소 0”과 반대입니다.

   근거: `design:18,56`, `delta:29`; [indoubt.go:754](/home/daniel/tossos-work/a197-impl/internal/execgw/indoubt.go:754), [orders_source.go:58](/home/daniel/tossos-work/a197-impl/internal/execgw/orders_source.go:58).

   **제안:** 페이지 메타데이터가 모순되는 경우의 실패 계약을 명시하십시오. `HasNext=true + 빈 cursor`, cursor 반복, 후속 페이지 오류, 마지막 페이지의 미지 행을 각각 시험해야 합니다.

6. **P1 — 취소 대상 식별을 기존 Gateway가 재검증해 준다는 보장이 없습니다.**

   D2는 같은 계좌·시장·종목을 요구하지만 D3의 불완전 판정과 시험은 주로 번호·방향에 집중합니다. `OfficialOrders`는 종목 필터를 전달할 뿐 반환 행의 일치를 검증하지 않습니다. Gateway는 호출자가 준 `OrderRef`와 결정의 일치를 검사하며, 실제 취소 HTTP 요청은 **계좌 헤더와 주문 번호만** 사용합니다. 잘못 고른 ID에 현재 포지션의 종목·BUY를 붙이면 Gateway 검증으로 오선택을 잡을 수 없습니다.

   flatten 판독을 재사용하는 것도 그대로는 안전하지 않습니다. `parseTarget`은 ID를 trim하고 시장을 추론합니다. 정본은 주문 ID 원문 보존·바이트 비교를 요구합니다. 이것은 a197 구현 결함 확정이 아니라 **재사용 경로의 계약 공백**입니다.

   근거: `design:17–18,53–55`, `delta:29`; [orders_source.go:47](/home/daniel/tossos-work/a197-impl/internal/execgw/orders_source.go:47), [gateway.go:407](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:407), [orders_write.go:222](/home/daniel/tossos-work/a197-impl/internal/official/orders_write.go:222), [flatten.go:542](/home/daniel/tossos-work/a197-impl/internal/flatten/flatten.go:542), [order-execution/spec.md:389](/home/daniel/tossos-work/a197-impl/openspec/specs/order-execution/spec.md:389).

   **제안:** 응답 행→검증된 대상→취소 요청의 식별 보존 계약을 정하십시오. 다른 종목·시장, 같은 ID의 상충 행, 실제 SELL인데 BUY로 잘못 구성한 요청, 공백을 포함한 opaque ID를 시험에 넣으십시오. 정상적으로 식별된 SELL과 다른 포지션의 보호 매도는 현재 delta가 명시적으로 제외합니다.

7. **P1 — 취소 접수와 critical 기록 사이의 crash가 취소 사실을 영구히 알리지 못하게 합니다.**

   설계 순서는 `취소 접수 → critical 기록`입니다. Gateway의 취소 종결 기록과 outbox 기록은 별도 작업입니다. 그 사이 crash 또는 outbox 기록 실패가 발생하면, 재시작 후 OPEN 목록에서 이미 취소된 주문이 사라져 알림을 다시 만들 계기가 없습니다. “접수된 주문마다 한 번 기록”을 현재 순서만으로 보장하지 못합니다. 모호한 취소가 기동 해소에서 CONFIRMED가 되는 경우도 빠져 있습니다.

   근거: `design:21,34,60`, `delta:49–53`; [gateway.go:733](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:733)·`:753`, [record_only.go:86](/home/daniel/tossos-work/a197-impl/internal/obs/record_only.go:86).

   **제안:** 영속된 취소 사실에서 미기록 알림을 복구할 책임과 식별자를 정하십시오. 접수 직후 crash, critical 기록 실패, 기동 해소 후 알림을 시험해야 합니다.

8. **P2 — ADDED만으로는 기존 SHALL NOT의 적용 경계가 명확해지지 않습니다.**

   기존 정본은 “보호 청산이 충돌 주문을 치울 때” 엔진 귀속만 허용하고, 손절 경로의 새 브로커 조회를 금지합니다. a197은 같은 보호 청산을 위해 외부 주문을 조회·취소하면서 “별도 단계”라는 이름으로 기존 요구를 유지합니다.

   기존 문서의 “외부 주문 취소는 이 요구가 다루지 않는다”는 주석은 별도 요구 추가를 뒷받침하지만, **본문의 넓은 SHALL과 새 예외의 관계를 완전히 해소하지는 않습니다.** ADDED만 충분하다고 단정하기 어렵습니다.

   근거: `proposal:49`, `delta:1,5`; [exit-policy/spec.md:554](/home/daniel/tossos-work/a197-impl/openspec/specs/exit-policy/spec.md:554)·`:557`·`:561`·`:573`·`:624`.

   **제안:** MODIFIED로 기존 요구를 제출 전 원장 기반 치움에 한정하고, 거절 후 외부 BUY 취소 예외를 교차 참조하십시오.

9. **P2 — OFF 동일성 및 “즉시 롤백” 주장이 실제 설계와 다릅니다.**

   proposal은 OFF면 “한 글자도 다르지 않다”고 하지만 D9와 delta는 OFF에서도 알림 문구를 바꿉니다. 시험 2도 변경된 문구를 기대합니다. 의도된 개선이지만 동일성의 명시적 예외입니다.

   또한 기존 토글 선례는 설정 편집 후 재기동이며, `Context.ExitObserver`가 기동 설정을 복사합니다. 같은 관례라면 파일에서 false로 바꾸는 것만으로 즉시 꺼지지 않습니다. OFF 전환은 이미 만든 `IN_DOUBT` 취소도 제거하지 않습니다.

   근거: `proposal:33–34,57`, `design:23–24,42,52`, `delta:21,55–57`; [config/engine.go:61](/home/daniel/tossos-work/a197-impl/internal/config/engine.go:61), [exitwiring.go:357](/home/daniel/tossos-work/a197-impl/internal/app/engine/exitwiring.go:357).

   **제안:** OFF 보존 범위에서 알림 변경을 명시적으로 제외하고, 토글 적용 시점과 미종결 취소가 남은 롤백 절차를 구분하십시오.

10. **P2 — “기존 audit 경로를 탄다”는 구현 계획이 아직 닫히지 않았습니다.**

    audit는 설정 구조체를 자동 순회하지 않습니다. `recordGateSettings`의 고정 목록에 항목을 추가하고 호출부에서 값을 전달해야 합니다. 선례대로라면 `NewContext`가 바뀌며, 이 함수는 a166 census 고정 대상입니다. `Context.ExitObserver` 자체는 해당 census 대상이 아닙니다.

    최신 tasks 1.3·1.4는 이 문제를 이미 **Manager 결정 필요**로 기록했습니다. 따라서 신규 미발견 결함이라기보다, freeze 전에 반드시 해소해야 할 알려진 blocker입니다.

    근거: `design:23`, `tasks:11–12,25`; [interlock.go:739](/home/daniel/tossos-work/a197-impl/internal/app/engine/interlock.go:739)·`:759`·`:805`, [engine.go:486](/home/daniel/tossos-work/a197-impl/internal/app/engine/engine.go:486), [shapes.json:42](/home/daniel/tossos-work/a197-impl/internal/accountbudgetruntime/testdata/a166-wiring/shapes.json:42).

    **제안:** audit 배선·census 변경 범위를 결정하고 ON/OFF 기록, 재기동, audit 기록 실패 시 활성화 방지를 수용 기준으로 넣으십시오.

11. **P2 — 알림 key는 “주문마다 한 번”을 정확히 표현하지 못합니다.**

    D6의 `타입·포지션·주문 번호`는 같은 주문을 다른 포지션이 처리하면 다른 key를 만듭니다. 반대로 같은 포지션에서 거래일이 달라 재사용된 주문 번호는 같은 key입니다. `remindAfter=0`은 정착한 행을 재무장하지 않으므로 후자의 새 취소 알림이 억제될 수 있습니다.

    근거: `design:21,61`, `delta:53`; [record_only.go:85](/home/daniel/tossos-work/a197-impl/internal/obs/record_only.go:85), [outbox.go:289](/home/daniel/tossos-work/a197-impl/internal/journal/outbox.go:289), [exit-policy/spec.md:686](/home/daniel/tossos-work/a197-impl/openspec/specs/exit-policy/spec.md:686).

    **제안:** “한 번”의 단위를 주문 인스턴스 또는 취소 에피소드로 확정하십시오. 동일 주문의 포지션 간 반복과 주문 번호 재사용을 별도 시험해야 합니다.

시험 전략의 가장 큰 약점은 **가짜 Gateway가 생산 계약 실패를 가린다는 점**입니다(`design:49`). 위 발견과 연결해 다음 변이가 추가로 필요합니다.

| 누락된 변이·실패 주입 | 필요한 관측 |
|---|---|
| 취소 결정의 Kind를 PLACE로, Side를 SELL로 변경 | 실제 Gateway 거절 및 브로커 호출 0 |
| 귀속 검사에서 park·번호 없는 attempt 제외 | 엔진 주문이 외부 취소 대상으로 들어가지 않음 |
| 종목·시장 검사를 삭제하거나 ID를 trim | 다른 주문 ID로 취소 전송되지 않음 |
| `HasNext=true`·빈 cursor를 완료로 처리 | 취소 0 |
| 첫 취소를 IN_DOUBT로 남김 | 다음 주기의 실제 차단·해소·알림 상태 |
| 취소 여러 건 지연·429·예산 대기 | 뒤 포지션 판정 시점과 가격 만료 여부 |
| 취소 접수 후 crash·outbox 실패 | 재기동 후 누락 알림 복구 |
| audit 연결 삭제·생산 토글 주입 누락 | 설정값과 실행·audit의 일치 |
| 거래일 다른 동일 주문 번호 | 새 취소 알림이 과거 key에 묻히지 않음 |

파일 수정·테스트·probe·운영 호출은 하지 않았습니다. 기존 AST 28개의 원본 소스 해시 일치를 확인하고 관련 분기와 현재 코드를 대조했습니다. 이는 새 AST 생성이나 실행 검증이 아닙니다. CodeGraph는 기존 산출물을 참조했으며 재질의하지 않았습니다. ai-memory 조회는 로컬 연결 권한 오류로 실패했습니다. 리뷰 중 다른 작업자가 tasks 1.1–1.6을 갱신한 것을 확인해 최종 내용도 반영했습니다. `impeccable: N/A — backend/documentation-only`; 새 설계 확정·구현 없이 계약 리뷰만 수행했습니다.