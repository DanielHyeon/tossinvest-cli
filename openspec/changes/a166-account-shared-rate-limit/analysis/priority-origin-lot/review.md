# priority/origin 구현·검증

기준 c4fb6a91f에서 시작해 rootcaller 착지 c412c0e55f2948e16e1f33f1e1e8fadaf2092c56를 ff-only 반영했다. Manager가 승인한 `analysis/priority-origin-plan.md` 및 `pre-edit.md`를 따른다. engine.go의 borrowed Budget 전달은 Manager가 추가 승인했다. 기존 함수의 편집 전 AST, 편집 후 AST/FLM/BTM과 위험 보고는 `analysis/function-logic`과 이 디렉터리에 보존한다. CodeGraph는 미초기화였으며 그 증거를 주장하지 않는다. UI: impeccable N/A — backend-only. search-first: 기존 context, exact-intent event, decision preimage, QueryFallback 및 기설치 라이브러리 재사용; 신규 의존성·schema 없음.

ExitObserver의 Query 전에 critical, 보호 취소/발주에 critical, 익절 취소/발주에 normal을 지정한다. flatten 명령은 계좌 bootstrap 전부터 critical이다. required replay는 기존 guard 뒤 private budgeted transport에서만 scope·decision hash/preimage·exact origin을 검증한다. 불명·충돌은 QueryFallback이며 기존 state/count 유지, DB 오류는 반환한다. BUY 또는 flatten과 모순되는 exit event도 거절한다. engine gateway는 borrowed budget의 private replay를 구성하고 실패를 전파하며 attestation nil을 유지한다. OFF HTTPReplay의 기존 전송·판정 흐름을 유지한다.

## 근거

- `journal-red/green.log`, `observe-red/green.log`, `proposal-red/green.log`, `full-profit-red.log`/`full-profit.log`, `replay-count-red/green.log`, `flatten-red/green.log`, `wiring-red/green.log`: 요청 분기의 RED/GREEN. full 익절 시험은 상위 critical context에서도 취소·발주 normal을 확인한다.
- `isolated-normal.log`, `isolated-race.log`, `isolated-tag.log`: Go 1.26.5, 4개 지정 패키지·지정 테스트의 normal/race/`tossos_testseams` PASS. 정확한 패키지·정규식은 `isolated-test.sh`. `isolated-go125.log`는 Go 1.25의 더 좁은 지정 집합 PASS다.
- `engine-integration-race.log`, `final-integration-tag-race.log`: required factory→실제 engine 조립, 같은 budget과 private replay 확인 및 실제 flatten replay 큐 critical·저장 body 원문을 local TLS에서 확인했다. 최초 컴파일 수정 이후 성공한 결과만 이들 로그에 표시한다. `401-critical.log` 및 최종 선택 격리 로그는 실제 official 401 뒤 토큰의 critical 상속을 확인한다. `ack-context-race.log`는 ACK 조회 fake reader가 priority와 부모 deadline을 보존함을 확인한다.
- `isolated-mutation-{replay-required,origin-account,origin-duplicate,refresh-context}.log`, `engine-budget-mutation.log`, `replay-priority-mutation.log`, `ack-context-mutation.log`, `full-profit-red.log`는 해당 보호 검사 삭제·완화 변이의 시험 실패 원문이다. 첫 계좌 변이는 `mutation-origin-account.log`에서 생존했으며 gateway 계좌 단독 불일치 반례 보강 후 `mutation-origin-account-after.log` 및 격리 최종 변이에서 검출했다. 변이 patch는 같은 디렉터리에 둔다.
- 모든 최종 네트워크 시험과 관련 변이는 별도 user+net namespace에서 loopback만 활성화하고 route가 없는 상태로 실행했다. 각 격리 로그의 namespace ID·주소·route 출력이 실행 영수증이다. 신규 TLS fixture는 httptest loopback listener 주소와 동일한 연결만 허용하는 DialContext 및 서버 인증서로 결속했다. fixture/선택 기존 테스트에서 setns, unshare, namespace FD 상속 사용을 검색했고 발견하지 못했다. 초기 비격리 시험은 local TLS 옵션을 명시했으며 변이가 endpoint·transport 옵션을 삭제하지 않았으나, 최종 네트워크 증거는 격리 실행이다.

수동 gstack review는 review SKILL과 checklist 두 pass를 현재 diff에 적용했다. SQL은 매개변수 바인딩과 exact ID 먼저 읽기·scope 모순·bounded 후보·오류 반환을 확인했다. context inheritance, OFF/required 분기, count/상태, 409/NOT_SENT 회귀, 닫기 소유권과 egress 경계를 검토했다. 추가 P0/P1 발견 없음. 설치·업그레이드·telemetry·별도 specialist는 수행하지 않았다. 개별 기존 모든 분기의 회귀·독립 리뷰를 대체하지 않는다.

exit event 조회와 replay count 증가는 별도 DB 연산이다. 그 사이 외부의 비정상 동일 intent event 추가를 원자적으로 막는다고 주장하지 않는다. 실제 운영 환경, 활성화, 실계좌, 전체 테스트·전체 gate 및 archive는 실행하지 않았다. a087 admission과 replay attestation은 nil 유지; 출처 미입증의 보수적 fallback은 보호 replay를 지연할 수 있다.
