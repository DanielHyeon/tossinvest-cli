## Why

TossOS의 production strategy registry에는 continuation, reversal, weekly-value의 KR/US 조합 6개가 존재하지만, 실제 engine runtime은 시장별 proposal이 정확히 1개일 때만 하나의 market worker에서 처리한다. 따라서 현재 구조는 전략군별 cadence, queue, fault latch를 독립적으로 운영하지 못하며, 돌파 전략을 descriptor 두 개만 추가하면 production manifest와 proposal assembly가 fail-closed로 거부한다.

weekly-value를 포함한 네 전략군을 실제로 독립 평가하려면 4개 family × 2개 market의 8개 lane runtime을 분리하면서도 account/symbol owner, Guardian, risk reservation, dispatch lease, official ExecutionGateway, fill/protection/reconciliation/exit 권위는 하나로 유지해야 한다. 이 경계를 명시하지 않으면 여러 evaluator가 같은 종목을 중복 매수하거나 account-wide 한도를 복제할 수 있다.

## What Changes

- continuation, reversal, weekly-value, breakout-retest를 canonical 4-family matrix로 정의하고 KR/US별 8개 lane descriptor를 모두 기본 `OFF/UNOBSERVED`로 등록한다.
- 각 lane instance가 자체 cadence, bounded work queue, deadline, health, failure counter와 entry-only latch를 갖는 independent evaluation worker가 되도록 production strategy runtime을 재구성한다.
- 시장별 coordinator가 같은 `(account, market, symbol, position_generation)`의 sealed proposals를 수집하고 versioned common score로 arbitration한 뒤 최대 1개만 기존 single-writer dispatch 경로에 전달하도록 한다. active owner가 있으면 기존 owner가 항상 우선하며 동점·비교 불가능 점수는 fail-closed refusal이다.
- KR/US breakout-retest v1 pure lane과 closed-bar state machine, typed evidence, conservative risk sizing, invalidation/timeout 계약을 추가한다. 첫 breakout touch 매수와 평균단가 낮추기는 금지한다.
- scheduler의 low-priority admission scope에 lane family를 추가하되 실제 endpoint/reset-generation quota와 safety reserve는 복제하지 않는다. 한 lane의 stale evidence, timeout, panic 또는 repeated failure는 해당 lane entry만 막고 peer lanes와 safety loops를 막지 않는다.
- official read-only market bar/quote evidence를 append-only snapshot으로 수집·재생하는 breakout evidence contract를 추가하되 evaluator에는 broker mutator, journal writer, toggle writer를 주입하지 않는다.
- 새로운 scale-in 실행 권한은 만들지 않는다. 이 change의 breakout production path는 first-leg proposal까지만 허용하며 기존 a066 campaign/risk owner와 a100 broker-resident protection 선행 gate가 완료되지 않으면 effective entry는 계속 OFF다.
- registry, tagged lane input, adapters, production manifest, proposal assembly, worker supervision, observability와 paired KR/US regression을 함께 변경하는 apply 순서와 test matrix를 제공한다.

## Capabilities

### New Capabilities

- `four-family-strategy-runtime`: 4개 전략군의 8개 market-specific evaluator를 독립 감독하고 sealed proposal만 하나의 shared arbiter/dispatch authority로 전달하는 production runtime 계약.
- `breakout-retest-strategy-lane`: closed-bar breakout, retest, reclaim과 invalidation을 결정적으로 평가하는 KR/US pure lane 및 evidence/state-machine 계약.

### Modified Capabilities

- `strategy-engine`: canonical descriptor/typed input/adapter/production proposal matrix를 6개에서 8개 lane으로 확장하고, lane evaluator가 mutation 권위를 갖지 않는 계약을 유지한다.
- `market-aware-scheduler`: market/horizon admission subscope를 market/horizon/family로 세분화하면서 physical endpoint quota와 safety-class reserve는 계속 단일 권위로 공유한다.

## Impact

- 주요 코드 영향: `internal/strategyflow`, `internal/strategyrouter`, `internal/strategyproposal`, `internal/strategyevidence`, 신규 `internal/breakoutlane`, `internal/app/engine/strategy_entry_supervisor.go`와 관련 configuration/console projection/tests.
- 재사용 경계: `strategyDispatchCycle.dispatch`, Guardian, journal owner/reservation, official ExecutionGateway, fill detection, protection, reconciliation과 exit lifecycle은 shared authority로 유지한다.
- 데이터 영향: 기존 evidence SQLite TEXT kind/schema는 additive kind를 수용하므로 destructive migration은 필요하지 않다. breakout setup/state/correction은 canonical payload/revision과 snapshot digest로 append-only 저장한다.
- 운영 영향: 모든 새 descriptor와 worker는 `OFF/UNOBSERVED`; 배포, 테스트 또는 change 적용만으로 automation, autostart, lane desired state나 LIVE approval을 변경하지 않는다.
- 선행 의존성: a064 evidence replay, a066 q_final/owner/bucket 및 exit-bypass 잔여 gate, a070 router/scheduler contract, a072 shared dispatch runtime과 a100 broker-resident protection readiness가 해당 범위에서 완료·검증되어야 한다. 또한 current main은 lossless official 1-minute authority가 KR-only이므로, 기존 권위를 넓히지 않는 M-B0 measurement seam과 별도 bounded M-B1 receipt가 US raw candle/quote source를 실제 계약으로 증명하고 Manager가 PASS를 기록해야 한다. M-B PASS는 L1 구현 착수만 허용하며 closed-bar finality나 production source acceptance를 대신하지 않는다. 미충족 시 production 8-lane wiring/배포는 OFF이고 exposure-raising dispatch는 0건이다.
