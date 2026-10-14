# a136 CodeGraphContext supporting context

- CodeGraphContext 질의는 이 세션에서 돌리지 않았다(advisory, GBrain MCP 는 병행 세션이 점유 가능). `not-applicable` 사유: 판정 훅 자리는
  CodeGraph hard evidence(`reserveRows` 호출자 5)와 HEAD 읽기로 닫혔고, 보조 문맥이 바꿀 수 있는 결론이 없다.
- 보조로 확인한 문맥(HEAD grep): `ReservationKindDailyLoss` 생산 사용처 0, `LimitsDigest` 생산 사용처 3(전부 Guardian 자기 유도), 콘솔 한도 편집기
  (`internal/console/settings_limits.go` · `config/limits.go` 의 tier · ceiling).
