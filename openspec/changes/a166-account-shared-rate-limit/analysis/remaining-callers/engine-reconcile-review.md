# engine reconcile 부분 결과 — 전체 로트 미완

client 생성만 root factory와 실제 ctx에 연결하고 private fixture 옵션을 전달한다. 계좌 선택·journal 순서는 유지한다. source pre-edit AST/FLM/BTM은 해당 함수 bundle에 보관했다. CodeGraph worktree 인덱스 부재로 조회 불가.

정적 fixture 옵션 결속 RED→GREEN, local TLS 정상 OFF 및 required token/Accounts, required manifest 오류 wire 0, 선택 tagged race 통과. factory 우회 변이는 required 오류·wire 0 시험이 잡았다. 테스트는 각각 lo만 있는 별도 user/net namespace에서 실행했고 namespace ID·interface·route를 원문에 남겼다. fixture가 setns, 부모 namespace FD, ExtraFiles를 사용하지 않음을 소스 검색했다. 수동 gstack 범위 검토: 기존 계좌 선택, 오류 전파, 정리 및 private override 확인. 전체 gstack mixed-model workflow, 전체 tests/gate, 실계좌·운영은 미실행.

첫 required fixture의 manifest 키 오류는 실패 원문으로 남겼고, JSON 태그 수정 뒤 GREEN이다. 과거 외부 401 incident는 RED/PASS에 포함하지 않는다. verify-reconcile, credential 검증, console/HTTP API와 나머지 호출자는 미완이다.
