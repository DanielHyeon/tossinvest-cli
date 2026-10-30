# a087 probe import whitelist 정적시험 delta

기준 6e0968222; 생산 main.go가 이미 도입한 accountbudgetruntime 하나만 whitelist에 추가했다. 같은 TestImportsStayInsideTheAllowlist가 수정 전 RED(미허용 runtime import), 수정 뒤 Go1.25 normal 및 tossos_testseams tagged race GREEN이다. 별도 user/net namespace lo-only/no-route에서 정적 시험 하나만 실행했으며 network/실주문 없음. 사전 AST와 FLM/BTM pre-edit.md, 원문 gzip과 SHA 보존. 수동 범위 리뷰: 다른 허용 이름 및 검사 분기를 바꾸지 않았다. production 코드 변경 없음, 전체 tests/gate 미실행. 구조검사 확장은 별도 계획·승인 대상이다.
