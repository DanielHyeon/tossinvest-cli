# shadow tool owner 2차 후보

기준 1137b93ca. a128 JEV·a140 breakout의 기존 credential·flag 검증 뒤 공통 2env parser 및 Factory Open/NewClient를 사용하고, client가 필요한 run 수명이 끝난 뒤 Close 오류를 기존 오류와 결합한다. 기존 signal context를 Auxiliary로 전달한다. 외부 judge는 이 budget transport를 사용하지 않는다. OFF는 Factory.NewClient의 기존 official.New 경로다. 기능·운영 배포 완료가 아니다.

사전 AST/PreEdit 및 baseline required missing manifest RED를 보존한다. 처음 baseline a140 fixture가 금지된 evidence 파일 이름으로 잘못 실패해 경로를 고친 뒤 유효 baseline RED를 재실행했다. 두 RED는 기본 endpoint 요청을 시도할 수 있지만 user+net namespace에 loopback 외 route가 없어 연결되지 않았다. 실제 테스트는 fake credentials/local TLS/transport/loopback 주소 엄격 검증을 사용한다.

두 도구별 required 누락 refusal, OFF local TLS token 오류 및 Auxiliary context, 타 critical lease 보유 시 local 서버 wire 0, synthetic Close 오류 결합 선택시험이 Go1.25 normal/tagged-race PASS. a140 busy는 ErrUnavailable, JEV는 기존 루프 뒤 context deadline으로 돌아간다. 외부 재시도가 없음을 server 요청 수만으로 포괄 증명하지 않으며 실제 운영 exit code·전 성공 출력 및 filesystem Close 실패는 검증하지 않았다. required 우회, Auxiliary→Normal, Close 오류 삭제 변이 각각 두 도구에서 RED다. 최근 원문 및 SHA는 raw-sha256.txt. 실행 중 테스트가 시장·운영 데이터를 사용하지 않았다.

수동 gstack 범위: env 실패/우회, Close 소유 순서, context와 local egress, 기존 판단 및 external judge 분리 손읽기. 전체 자동 gstack·독립 검증, 나머지 네 도구, 전체 tests/gate, 실계좌·실제 도구·운영·활성화는 미실행. 이전 함수 pre AST와 새 실제 함수 post AST의 차이는 runCommand가 test-option seam wrapper로 나뉜 것이다. 필요 시 관리자에게 source 리뷰 요청한다.
