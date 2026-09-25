# Evidence reconciliation — a115

| 사실 | CodeGraph | AST/HEAD | 결론 |
|---|---|---|---|
| 전략 dial 자리 | `strategyprojectionrpc.Dial` 호출자에 `runConsole` 포함 | AST runConsole B33–B37(console.go:398–405), dial :399 | 일치 — 편집 블록 하나, dialErr·비부재 stat 오류 둘 다 nil 접힘(proposal 의 병 확인). 좌표는 freeze 리뷰 P1-5 가 정정 |
| httpapi 원형과의 판정 동치 | `resolveStrategyRuntimeReader` 호출자 1(`strategyRuntimeReaderFor`) | AST 4분기: 부재→nil · 비부재 stat 오류→경고+nil · dial 실패→sentinel · 성공→client | 콘솔 부팅 블록과 판정 구조 동일, 차이는 sentinel 유무와 경고 문구 — design D1 의 재사용 근거 |
| wrapper 의 콘솔 인터페이스 충족 | implements 목록에 `MultiMarketStrategyRuntimeReader` · `StrategyRuntimePresence` | httpapi_strategy_attach.go:60–314, `StrategyRuntimeConfigured` :113 | wrapper 를 콘솔 `Options.StrategyRuntime` 에 그대로 꽂을 수 있다 — 구조적 인터페이스라 import 불필요(design D2) |
| 소비자 nil 판정 | `StrategyRuntime` 소비 함수 2 | :48–49 · :291 — 둘 다 `== nil` 로 dormant 를 가른다 | wrapper 는 non-nil 이므로 두 자리 모두 부재 신호(`StrategyRuntimeConfigured`) 판정으로 교체해야 한다 — 안 하면 부재가 「읽지 못했다」로 회귀(design D2) |
| 영향 테스트 | 기본 0 / filter 863 | cmd/tossctl · internal/console 각각 한 패키지 | filter 결과 채택, 두 패키지 스위트 전체 + 기존 a108/a109/multimarket 회귀로 덮는다 |

불일치 해소 후 편집 차단 사유 없음. 분기 번호는 `analysis/function-logic/` AST 번들이 정본이다.
