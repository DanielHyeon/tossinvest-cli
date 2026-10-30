# a166 engine poll 독립 검토

대상 ffa2dadd49848afcba2863efc8e2473cb0d5f994, base cf0f0919e. 범위 내 P0/P1 발견 없음. 아래 정본문서 P2 표현 명료화 권고. 독립 source 손읽기, 작성자 AST 해시·영수증 대조 및 제한 시험; 생산 소스 수정 없음.

- `internal/execgw/engine_poll.go`는 observed Budget, 동일 client/seq/account 및 네 Query 종류를 검증하고 Query마다 collector 생성; `retry.go`는 반환 AuthFatal의 latch/escalation을 429·local 오류보다 먼저 수행한다. OFF nil binding은 기존 분기다. transport pre-Acquire collector latch는 token 포함 동일 논리 Query에만 적용하며 prewire body close/ErrNotSent를 보존한다. hint 동종 조회 skip은 freshness 성공이 아니다.
- 생산 Query 9곳은 tracer 4, exitloop 1, exitwiring 2, adoption 1, position policy 1로 source grep과 constructor값 흐름을 손으로 대조했다. 작성자 postedit AST 7개 source SHA가 후보 본문과 일치한다. 해당 AST는 전체 임의 Go caller proof가 아니며 실제 공식 client/selected account 값의 장부 결속은 기존 resolveAccountRef 및 caller 흐름에 의존한다. 전체 생산 송신 census는 별도 과제다.
- 실제 token200/401 Finish fault 시험은 observation syncDir의 여섯째 호출을 오류로 만들어 postcommit 경계를 주입하고 최종 가격/토큰 wire 수를 확인한다. guard 제거 변이 원문에서 각각 price 1, price 2/token 1이 실제 발생하므로 단순 dirty state가 후속 Begin을 막아 green이 된 것이 아니다. 반례는 test-only unsafe/fake TLS이며 실제 powerloss·운영 권한을 입증하지 않는다. 다른 reader guard 변이의 처음 false-green은 별도로 폐기됐고 최종 mutant의 구체 오류 조건은 RED다. 15개 압축 증거 SHA 목록 일치.
- 지정 작성자 regex Go1.25.0 normal 및 tagged race 모두 PASS. 원문 `/tmp/a166-engine-poll-independent-normal.log`, `/tmp/a166-engine-poll-independent-tag-race.log`: 각각 user/net namespace lo-only/no-route, local TLS exact dial. 패키지 normal accountbudget/execgw/engine 1.745/11.185/1.880초, tagged race 2.839/12.337/6.197초. 시험명·실행 로그 원문 참조.
- 정본 retry matrix diff는 required/OFF의 429·hint·보호 대기·최종 AuthFatal을 구분한다. P2: `retry-matrix.md:14-19`의 '**어떤 오류에도 자동 재시도하지 않는다**', ':15 공식 API 멱등성 키가 없다'는 단정과 바로 뒤 취소 409 예외·a165의 clientOrderId 관련 결정은 읽는 사람에게 모순을 남긴다. 기존 문구에서 유래했지만 이번 정합화가 상위 규칙을 만지므로 예외 포함한 사실관계로 별도 문구 정리 권고; 이 관찰은 engine poll 실행 결함 판정이 아니다.
- 전체 tests/gate, 모든 운영 참여자, 실제 broker/계좌/배포, a087/a169 활성화 미검증. gstack은 작성자 범위 지정 수동 검토이며 전체 자동 pipeline 아님. 최종 통합 SHA 및 Manager 승인 별도.
