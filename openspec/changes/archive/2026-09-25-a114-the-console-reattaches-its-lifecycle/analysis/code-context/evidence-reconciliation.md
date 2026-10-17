# Evidence reconciliation — a114

| 사실 | CodeGraph | AST/HEAD | 결론 |
|---|---|---|---|
| lifecycle dial 자리 | commander 생성자 호출자 = runConsole | AST runConsole B33–B37(:395–409) | 일치 — 편집 블록 하나 |
| impact 52 | 인터페이스 메서드 이름(List/Preview/Apply)로 엔진 서버·ops 까지 | 실제 타입 흐름은 runConsole → console.Options.PositionPolicies 뿐 | CodeGraph 과대 추정(이름 해소). 편집은 cmd/tossctl 새 파일 + runConsole 한 블록 |
| 영향 테스트 | 기본 0 / filter 860 | cmd/tossctl 패키지 전체가 한 패키지 | filter 결과 채택, 전체 cmd/tossctl 스위트 + internal/console 회귀로 덮는다 |
| 격리 해제 발견 | quarantineClient 호출자 3 | AST B1 :28 타입 단언 | wrapper 가 세 메서드를 구현해야 회귀 없음 — design D1 |

불일치 해소 후 편집 차단 사유 없음.
