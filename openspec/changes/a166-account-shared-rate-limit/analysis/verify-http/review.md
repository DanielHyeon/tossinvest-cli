# verify factory·console broker·HTTP API 부분 후보

기준 bbe5066ec. verify의 최초 계좌 조회와 양수 seq 재생성 client는 동일 root factory를 사용하고, 호출별 새 client를 만든다. seq 0의 기존 lazy 경로는 유지했다. verify run/abort와 console verify starter의 현재 context가 전달된다. console 읽기 broker는 request와 signal lifetime을 함께 따르고 성공만 cache하며 실패 후 재시도한다. metadata의 최초 계좌 조회도 동일 factory다. HTTP API의 accountRef에 요청 context를 전달하며 서버 선오류와 signal 종료가 같은 5초 Shutdown 경로를 사용하고 실패 시 Close와 오류 결합을 한다. 기존 publisher cancel/join은 유지한다.

사전 분석과 각 함수 AST는 preedit/에 있다. 선택 normal/tagged race는 receipts의 freeze 영수증을 참조한다. local TLS는 가짜 자격, 명시 옵션 static guard, lo-only 별도 network namespace에서만 실행했다. 기존 계좌 선택·HTTP API 마스킹의 선택 회귀를 포함했다. 최초/seq-bound budget 우회, HTTP 요청 context 제거, Shutdown 오류 누락, console lifetime 제거 변이는 각각 RED다. Shutdown fixture는 helper에 인위적인 server 오류와 실제 handler timeout을 함께 전달해 두 원인 보존을 검사한다. 실제 runHTTPAPI server 선오류 전체 경로는 source 손읽기이며 통합 fixture로 증명했다고 주장하지 않는다.

수동 gstack 범위: OFF, 계좌 ref/seq 결속, 오류 전달, cache·취소·publisher 종료와 JSON/status/auth/mutation 무변경을 점검했다. CodeGraph 현재 worktree 인덱스 없음. 혼합 모델 gstack·독립 검증, 실제 UI rendering, candidate/soak·독립 tools 나머지, 전체 tests/gate·실계좌·운영·archive는 미실행. impeccable N/A — backend context/lifecycle wiring. base 이후 main의 console lifecycle은 별도 커밋이며 통합 시 겹치는 console.go를 확인해야 한다.
