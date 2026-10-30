# verify-reconcile / OpenAPI 부분 구현 검토

기준 5762b161b. verify-reconcile의 최초 Accounts와 검증한 seq로 다시 만든 읽기 client는 같은 root factory를 빌린다. profile credential과 행·계좌 수·seq 거절, 좁은 읽기 wrapper는 유지한다. OFF 직접 호출은 기존 constructor를 사용한다. OpenAPI test/status, init, doctor 및 console onboarding 검증은 root factory를 통해 첫 token/Accounts를 호출한다. factory 오류는 성공으로 보이지 않으며 onboarding은 임시 토큰을 지우고 저장하지 않는다. 사용자 설정으로 노출하지 않는 private test option은 기존 official authority 제한을 유지한다.

사전 AST/FLM/BTM은 함수별 analysis/function-logic에 보관. 격리 worktree의 CodeGraph 인덱스 부재는 별도 증거 한계다. 정적 binding RED→GREEN, 정상 선택 및 tossos_testseams race는 receipts의 exact-normal/exact-tag-race를 참조한다. 최초 및 bound reader의 factory 우회, OpenAPI factory 우회, console root 누락 변이는 각 기대 시험에서 실패했다. local TLS 검증은 각각 독립 user/network namespace의 loopback만 활성화하고 route가 없음을 같은 영수증에 기록했다. 테스트는 고정 가짜 자격을 쓴다. 최초 reader는 다른 lease가 보유된 상황에서 wire 0으로 대기·취소하며 bound client는 동일 Budget 결속을 검사한다.

수동 gstack 범위 검토: 사용자 자격 노출, 계좌 선택, error-to-success 변환, token 격리, WTS 강등 및 private 옵션 authority를 확인했다. 혼합 모델 gstack·독립 검증은 미실행이다. console 전체 lifecycle, verifyBrokerFactory, 나머지 CLI/tool, 429 통합, 전체 tests/gate·실계좌·운영·archive는 이 후보 범위 밖이며 완료라고 주장하지 않는다. UI 재설계 없음(impeccable N/A — backend wiring); 기존 UI의 렌더링 검증은 미실행이다.
