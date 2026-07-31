## 1. 선행조건과 RED

- [x] 1.1 a045/a046 완료를 선행 gate로 유지하고 첫 lane를 StockOS commit `d75113d3`, source-set digest `09260ac…`, KRX `parker_vwap_trend_v1` conservative constants로 proposal-freeze에서 고정한다.
- [x] 1.2 engine/gateway/Guardian/journal/console operating 함수의 CodeGraph impact와 Function Logic·Branch Test Map을 작성한다.
- [ ] 1.3 lane 순수성, ApprovedCandidate pass+candidate-life/set/evidence provenance, OFF exit 보존, Guardian/protection/gate refusal, duplicate identity와 legacy RiskIntent canonical compatibility RED 테스트를 추가한다.
- [ ] 1.4 activation manifest 모든 binding field의 mismatch/expiry/revocation과 decision→dispatch TOCTOU RED 테스트를 추가한다.
- [ ] 1.5 sorted source path/blob manifest가 frozen source-set digest를 재현하는지 검증하고, 불일치 시 not_configured/OFF인 RED 테스트를 추가한다.

## 2. Strategy domain과 orchestration

- [ ] 2.1 `internal/strategy`의 EntryLane, ApprovedCandidate와 EntryDecision 계약을 구현한다.
- [ ] 2.2 `krx_parker_vwap_conservative_v1`을 frozen gate order/constant와 Go golden fixture로 순수 이식하고 broker/journal dependency를 금지한다.
- [ ] 2.3 orchestrator에 RiskIntent→Guardian→durable attempt→official gateway 경로를 연결한다.
- [ ] 2.4 immutable activation manifest repository/validator와 submit-time digest 기록을 구현한다.
- [ ] 2.5 official 1m decimal을 KST 정규장 closed contiguous 5m bar로 exact 집계하고 incomplete/gap/장외/naive/stale input을 거부한다. authoritative symbol-state source가 없거나 stale이면 OFF를 유지한다.

## 3. 운영 lifecycle

- [ ] 3.1 `strategy-runtime` descriptor와 read-only 상태 card에 lane field의 label/help/default/desired/effective/unit/range/provenance/apply timing 및 a045/a046/a048/source blocker를 구현하고 not_configured/OFF를 고정한다.
- [ ] 3.2 전략 파라미터, lane state, auto-start와 LIVE approval의 별도 section descriptor를 구현하되 실제 write action은 a050에 맡긴다. arbitrary input·typed confirmation·`enable all` 부재를 DOM/정적 테스트로 고정한다.
- [ ] 3.3 lane desired/effective state와 audit, refusal reason, 기본 OFF, kill switch를 연결한다.
- [ ] 3.4 paper/shadow/canary 주문 경로가 없고 lane OFF에서 exit가 동일함을 검증한다.

## 4. 검증과 LIVE 인계

- [ ] 4.1 race/restart/full test·vet·validate와 security/adversarial review를 통과한다.
- [ ] 4.2 `make gate CHANGE=a047-add-strategy-engine`을 통과하고 activation blockers와 dormant 인계 절차를 기록한다. 이 change에서 LIVE 승인이나 운영 토글을 실행하지 않는다.
