# 도구 owner 1차 사전 분석

기준 5b379e08d7600682250170a4b658e2c307063cec. 기존 root parseBudgetEnv와 a112 l1c run 사전 AST pre-*.json. search-first: 현 root 2변수 판정과 accountbudgetruntime.Open/NewClient/Close를 확장, 새 의존성 없음. 단일변수 FromEnv의 의미는 그대로 둔다. impeccable N/A — CLI/backend, 화면 없음. 전체 gate/운영 금지.

분기 계획: absent/absent 및 off/absent OFF, required+nonempty manifest required, 나머지 failclosed. 기존 root 판정은 신규 leaf에 위임. l1c run은 입력 거절 후 timeout context로 factory Open, 동일 client로 두 GET, 요청 종료 후 Close. OFF는 기존 wire 사용, required failure 첫 wire 0, auxiliary 혼잡 즉시 실패, close error 누락 금지. 기존 출력/timeout/오류 보존. 사전 RED: runtime matrix 테스트는 새 leaf 부재로 컴파일 실패. 전체 도구 및 운영 배선은 아직 미완.
