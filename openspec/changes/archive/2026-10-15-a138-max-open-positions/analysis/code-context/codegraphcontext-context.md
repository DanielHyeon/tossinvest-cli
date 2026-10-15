# a138 CodeGraphContext supporting context

- `not-applicable`: CodeGraphContext 질의는 돌리지 않았다(advisory). 판정 훅 자리는 CodeGraph hard evidence(`reserveRows` 호출자 — 모든 예약 tx)와
  HEAD 읽기 · AST(`analysis/ast-evidence/`)로 닫혔고 보조 문맥이 바꿀 결론이 없다. a136 이 같은 자리를 같은 방법으로 닫은 선례.
- 보조로 확인한 문맥(HEAD 읽기): 포지션 투영의 세 종류(`positions.entry_decision_id` · `adoption_id` · 둘 다 NULL = 외부, core_domain.go · adoption.go),
  외부 보유를 접어 넣는 생산자 `reconcile/external.go`(30초 대사 루프), 관리 술어 단일 철자 `position.ExitEligible`(a137 이 `positionSelect` +
  `Position.ExitEligible` 로만 가름 — 정적 시험이 지킴), 결정 preimage 의 `market`(소문자) · `symbol`(대문자) 정규화(`canonicalVenue`, decision.go:556).
