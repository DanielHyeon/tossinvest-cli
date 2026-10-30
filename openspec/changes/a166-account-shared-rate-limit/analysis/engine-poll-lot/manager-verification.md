# 엔진 poll 로트의 Manager 검증

2026-10-31, 후보 ffa2dadd4, base cf0f0919e. **개발 로트 승인**. a166 전체 완료·운영 전환·a087 활성화·a169 측정 승인이 아니다.

Manager는 required binding/key/remaining의 신규 구현과 기존 7개 함수 diff, 사전 함수 지도 및 분기 지도, 원문 race와 token200/GET401 postcommit-fault 변이 결과를 대조했다. closed/unobserved Budget 거절, 길이 구분 여섯 필드 key, 계좌·선택 seq·동일 client 확인, Query별 collector와 reader 경계는 승인 설계와 일치한다. Context.ReconcileDriver는 caller Retrier를 c.Retrier로 덮은 뒤 reader를 확인하며 OFF 주입을 보존한다. 원래 최종 AuthFatal의 gate/escalation을 관측 오류보다 먼저 처리하고 required429/기록 오류에서 재시도·freshness 성공을 중단한다.

독립 Terra는 동일 후보의 정확 선택 regex를 Go1.25 normal/tagged race, accountbudget·execgw·engine 세 패키지에서 별도 lo-only/no-route로 재실행해 PASS했다. postedit AST 7개 source SHA 및 gzip 15개 SHA 일치, Query 생산 자리 9곳 값 흐름을 손으로 읽었다. AST 분기 추출과 손읽기 값 흐름은 다른 증거다. 전체 생산 제어 흐름을 자동 검증했다고 하지 않는다. 원문과 상세 제한은 독립 보고서를 보존한다.

token200 Finish의 마지막 parent fsync만 실패시킨 변이에서 latch 제거 시 GET1, GET401 경우 token1+GET2가 실제 발생했다. dirty-state Begin 실패에 기대는 false-green이 아니다. 최초 HTTP 결과를 유지하면서 후속 요청만 중단한다. 별도 reader 변이의 처음 false-green은 제외하고 고유 guard 오류로 검출한 결과를 사용한다.

정본 retry matrix는 이 로트와 함께 required/OFF·token 내부401·최종 AuthFatal을 구분하도록 고쳤다. 독립 P2에 따라 '멱등성 키 없음/어떤 오류에도 mutation 무재시도'의 과도한 문구를 기본0·맹목재전송 금지·별도 guarded replay와 a165 취소 예외로 명료화했다. 기본 transient 수치와 staleness 임계는 바꾸지 않았다. TestDefaultsMatchThePublishedMatrix 선택 PASS는 수치 일치 증거이며 문서 모든 문장의 자동 증명은 아니다.

보호 Critical price429 뒤 adoption·재편입의 실제 함수가 같은 hint를 공유하는 fixture는 확인했지만 ExitObserver 전체 주기 자체는 미실행이다. test-only reflection/unsafe fault·selected seq 변경을 운영 권한 증거로 쓰지 않는다. 실제 broker·운영 환경·전체 tests/gate·최신 CodeGraph 색인은 미검증. gstack은 작성자의 수동 범위 검토와 독립 검증이며 자동 전체 pipeline이 아니다. impeccable N/A — backend 및 문서만 영향, 신규 의존성 없음.

최신 main 통합에서 source 충돌·새 동작이 없으면 필요한 정확 영향 선택만 확인하고 Terra가 정상 ff 게시한다. a166 최종 구조검사와 소비자 API/통합 증거·사람 tasks 4.2는 남는다. archive와 운영 토글은 수행하지 않는다.
