## 1. 선행조건과 RED

- [ ] 1.1 a045/a046 완료를 확인하고 첫 StockOS lane source·시장·상수 provenance를 proposal-freeze에서 고정한다.
- [ ] 1.2 engine/gateway/Guardian/journal/console operating 함수의 CodeGraph impact와 Function Logic·Branch Test Map을 작성한다.
- [ ] 1.3 lane 순수성, OFF exit 보존, Guardian/protection/gate refusal, duplicate identity와 provenance RED 테스트를 추가한다.
- [ ] 1.4 activation manifest 모든 binding field의 mismatch/expiry/revocation과 decision→dispatch TOCTOU RED 테스트를 추가한다.

## 2. Strategy domain과 orchestration

- [ ] 2.1 `internal/strategy`의 EntryLane, ApprovedCandidate와 EntryDecision 계약을 구현한다.
- [ ] 2.2 최소 한 lane를 순수 판단으로 이식하고 broker/journal dependency를 금지한다.
- [ ] 2.3 orchestrator에 RiskIntent→Guardian→durable attempt→official gateway 경로를 연결한다.
- [ ] 2.4 immutable activation manifest repository/validator와 submit-time digest 기록을 구현한다.

## 3. 운영 lifecycle

- [ ] 3.1 `strategy-runtime` descriptor에 lane field의 label/help/default/desired/effective/unit/range/provenance/apply timing을 구현하고 proposal-freeze 전 not_configured/read-only를 고정한다.
- [ ] 3.2 전략 파라미터, lane state, auto-start와 LIVE approval을 별도 section/action으로 구현하고 lane·auto-start 기본 OFF와 `enable all` 부재를 테스트한다.
- [ ] 3.3 lane desired/effective state와 audit, refusal reason, 기본 OFF, kill switch를 연결한다.
- [ ] 3.4 paper/shadow/canary 주문 경로가 없고 lane OFF에서 exit가 동일함을 검증한다.

## 4. 검증과 LIVE 인계

- [ ] 4.1 race/restart/full test·vet·validate와 security/adversarial review를 통과한다.
- [ ] 4.2 `make gate CHANGE=a047-add-strategy-engine` 후 운영자의 단일 LIVE 승인 절차를 기록한다.
