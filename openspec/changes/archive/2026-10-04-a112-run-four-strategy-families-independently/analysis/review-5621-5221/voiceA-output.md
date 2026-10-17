**판정: APPROVE** (P0·P1 없음, P2 둘과 (T) 셋은 착지 전에 문서만 고치면 됨)

사본 `/tmp/claude-1000/a112-rev-Kjc4`를 만들어 쓰고 지웠음(전용 GOCACHE 포함). 실제 저장소 `git status`는 리뷰 전과 같음(untracked 셋만 있음).

## 발견

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | P2 | 오늘-동등성 핀의 fixture는 생산 조립 모양이 아니다. 시험 주석의 「조립과 같은 모양」과 review.md·코드 주석의 「B2 = 유일한 방어」는 과장이다. (a) 위험·계좌 권한이 한 범위짜리 fixture에서 온 것이라, 범위가 둘인데도 Ready=true로 남는다. 생산에서는 그 상태가 나올 수 없다. (b) 항목이 [005930, 000660] 순서인데, 조정자는 소유자 범위를 사전순으로 놓으므로 생산이라면 000660이 앞이다. | 생산 모양으로 탐침(probe)하면 `ResultAuthority ready=false handoff refusal="HANDOFF_OVER_CAPACITY"`, 계좌 `collectMarket`(:154)은 `ready=false reason=PROPOSAL_NOT_READY`다. 즉 생산에서는 B2의 개수·위험·계좌 세 조건이 함께 거짓이다. 조정자 순서로 바꿔 B2 개수 조건을 지우면 `scope 0 err=…risk authority scope changed`, `scope 1 err=…proposal identity changed`, spy 호출 0이다. 이 순서에서는 :221·:228이 대신 막고, 핀은 errs[0] 문구 검사로만 빨개진다. | 주석과 review를 「의도적으로 만든 최악 조건(생산에서 도달 불가)」으로 고쳐 적을 것. 5.2.2.2의 기준선이 생산 모양이 아니라는 점도 적을 것. 조정자 순서 사례를 하나 더하는 것은 선택. |
| 2 | P2 (잠재, 활성화된 시장에만 해당) | 두 범위 활성화 시장의 관측 동작이 바뀌었는데 review에 적혀 있지 않다. 편집 전에는 `dispatchHandoff`가 OverCapacity를 내고 Deliver가 nil로 조용히 끝났다. 편집 뒤에는 매 poll마다 범위 1이 dispatch에 들어가 다음을 한 뒤 B2 오류를 돌려준다: `ObserveStrategyProtection`, `ObserveStrategyEntryGate`(둘 다 브로커 읽기), `AcquireStrategyDispatchOwner`(원장 쓰기, 프로세스당 한 번). 그 오류는 refreshOnly 갈래에서 `recordSwallowedCycleError`로 매 주기 세어진다. 주문 0은 유지된다. | `strategy_dispatch_cycle.go:96`·`:146`·`:150`·`:232`, `strategy_entry_supervisor.go:936`. 원래 핀을 돌리면 `scope 0/1 err=…paired production authority is incomplete`, spy 0이다. | review.md 「토글 OFF」 절에 활성화 시장의 부수 효과 차이를 적을 것. 오늘 생산은 핀 0이라 영향 없음. |
| 3 | (T) | 대조 실행은 fixture를 새로 만들므로 「같은 조립」은 사실이 아니다. 다만 귀속 결론은 맞다: 두 범위 실행에서 거짓인 B2 조건은 개수 하나뿐이다. | 탐침 결과: `count!=1=true riskNotReady=false fxNotReady=false acctNotReady=false schedNotReady=false activationNil=false` | 주석을 「같은 모양의 새 조립」으로 고칠 것. |
| 4 | (T) | 첫 오류에서 멈추는 방식은 안전 쪽이다. 다만 사전순으로 앞선 범위가 계속 실패하면 뒤 범위가 매 주기 굶는다(liveness 문제). | `deliverEachStrategyHandoff`의 `return err`, 이월 표 5.2.2.2 | 5.2.2.2 결정 항목에 「굶음」을 명시할 것. |
| 5 | (T) | 5.6.2.1의 B13은 오늘 생산에서 발동하지 않는다. 중앙 무결성 표지를 만드는 생산 호출자가 0이다. | codegraph 1.6.0 기준 `StrategyCentralIntegrityFailure` 호출자 5개가 전부 테스트이고, `ErrStrategyCentralIntegrity` 참조는 감독자 파일 안에만 있다. | 없음. 기록만 남길 것. |

**B2 조건별 변이 표** (사본에서 :217–218의 조건을 하나씩 `false`로 바꾸고 핀과 탐침을 실행)

| 변이 | 핀 | 탐침 |
|---|---|---|
| 개수 | **FAIL** (`placed 1 orders`) | scope0 err=nil, scope1=identity changed(:221), spy=1 |
| 위험 | PASS | 두 범위 모두 B2, spy=0 |
| FX | PASS | 같음 |
| 계좌 | PASS | 같음 |
| 일정 | PASS | 같음 |
| 활성화 | PASS | 같음 |

핀을 빨갛게 하는 것은 개수 변이 하나다(원장 F14와 일치). 핀 fixture 순서에서는 B2 개수 조건이 범위 1의 유일한 방어이고, 범위 2는 :221 identity 검사가 막는다. 조정자 순서에서는 :221·:228이 대신 막는다(1번 발견). errs[1]도 B2 문구다.

**추가 실행** (사본, `GOFLAGS=-trimpath`): 무태그·태그 `./internal/strategyhandoff` ok, 태그 `./internal/app/engine -run 'Handoff|OwnerScope|Admit|Central|FaultScope|RefreshOnly|Dispatch'` ok, `./internal/execgw -run 'Reason|Latch'` ok.

## 필수 항목 판정

1. **오늘-동등성 핀 공격**: 귀속은 fixture 안에서 참이다(측정했고, 개수 변이만 FAIL). 다만 fixture는 생산 모양이 아니고, 「유일한 방어」는 fixture 순서에서만 성립한다. P2(1번)와 (T)(3번).
2. **토글 OFF = upstream**: 동등하다. 활성화가 없으면 `dispatchHandoffs`는 `[]{dispatchHandoff()}`를 돌려주고, 헬퍼는 `Deliver(body)` 값을 감싸지 않고 그대로 돌려준다(`err != fault` 동일성 시험이 있음). 순서 `lanes.evaluate` → deliver도 그대로다. 시장이 어긋난 활성화는 들어갈 수 없다: `familyGateFor(market)`가 시장별 env를 읽고, 검증에서 `body.Market != config.Market`이면 거절한다(`production_family_activation.go:526`). 미선언은 digest env가 비었다는 것 하나로만 정한다(:432). 배포 yaml·compose에서 `FAMILY_ACTIVATION` grep 결과는 0건이다. secrets 디렉터리는 열지 않았으므로 그쪽의 핀 0은 HANDOFF의 측정을 인용한 것이다.
3. **활성 시장의 새 동작이 안전한가**: 안전 쪽이다. 같은 범위 중복은 시장 전체를 거절하고, 닫힘은 handoff 하나다. 첫 오류에서 멈추고, 부분 전달 뒤 범위 1은 다음 주기에 CAS≠FLAT이라 `return nil`로 건너뛴다. 잔여는 P2(2번 부수 효과)와 (T)(4번 굶음)다.
4. **5.6.2.1 배선**: 문제없다.
   - 진입만 닫힌다. `Gateway.checkEntry`는 `!plan.raisesExposure`이면 nil이므로 매도(buy가 아닌 쪽)·취소는 게이트를 보지 않는다(`gateway.go:851–855`). Guardian chain도 `SideSell`이면 reduction 경로라 latch 단계를 건너뛴다(`chain.go`의 Evaluate). 따라서 손절·청산 차단은 없다.
   - 루프는 살아 있다(`continue`). 게이트 없는 조립은 `signalCentral`로 올리고, 생산 생성자는 `c.Entry == nil`이면 거절한다. Recovery가 이미 같은 게이트를 요구하므로 기동 동작 변화는 없다.
   - 재시작 뒤 같은 고장이 다시 나면 첫 poll에서 다시 잠긴다. 그 사이에 전략 dispatch가 나갈 수 없는데, 고장이 같은 사이클의 dispatch 앞에서 검출되기 때문이다. 편집 전에는 영구히 삼켰으므로 엄격히 개선이다.
   - 오늘 생산에서는 발동하지 않는다(5번).

Recommendation: 5.6.2.1·5.2.2.1을 착지 상태로 두는 것을 APPROVE because 안전 불변식 위반과 오늘 생산 동작 변화가 없고 핀의 귀속도 측정으로 참임을 확인했다. 다만 1·2번(핀 fixture가 생산 모양이 아니라는 점, 활성화 시장의 부수 효과 차이)은 review.md와 시험 주석에 정정해 적어야 한다.
