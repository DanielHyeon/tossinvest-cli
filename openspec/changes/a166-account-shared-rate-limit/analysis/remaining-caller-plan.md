# 나머지 CLI 호출자와 콘솔 수명 — Manager 설계안

기준 root 후보 `a8b7afd2966ad0f079be7df9c0c611e5db5bd18c`. `/tmp/a166-remaining-callers.md`, `/tmp/a166-console-lifetime.md`의 원문 위치와 해당 생산 함수를 Manager가 손읽기했다. 기존 `caller-lifetime-plan.md`의 borrowed factory 계약을 확장한다. 독립 검토 전 구현 승인 문서가 아니다.

## 생성과 계좌 선택

CLI root 소유 factory를 사용하는 작은 공식 client 생성 helper로 반복 배선을 모은다. 실제 호출 context, credentials, token path와 기존 official options를 받고, root owner가 없는 기존 시험/OFF 호환 경로만 기존 생성자를 쓴다. 실제 newRootCmd는 owner를 항상 가진다. owner가 required인 경우 오류를 raw client·nil 성공으로 바꾸지 않는다. 외부 transport/base override의 authority 제한은 유지한다. 실제 환경이나 설정 파일을 수정하지 않는다.

verify는 최초 Accounts와 선택된 seq로 재생성한 client 모두 동일 factory를 빌린다. seq=0의 기존 lazy resolution 정책도 보존한다. verify-reconcile의 profile-only credential, 정확히 한 계좌, 잘못된 행·seq 거부, 좁은 읽기 wrapper는 유지한다. engine reconcile도 최초 token/Accounts 전부터 guard를 받는다. 기존 verify의 15s/30s walk와 engine poll 정책을 이 로트에서 섞지 않는다.

verifyBrokerFactory와 context 없는 builder는 필요한 내부 시그니처를 context-aware로 바꾸고 모든 생산 호출자와 해당 fixture를 함께 갱신한다. 새 Background 대기를 추가하지 않는다. console read는 현재 요청 context와 console 수명을 모두 존중하고, verify 시작은 해당 run context를 사용한다. 계좌를 임의로 첫 행/단일 계좌로 통일하지 않는다.

MCP는 기존 official eligibility 분기에서만 factory를 열고 실패 시 WTS-only 성공으로 숨기지 않는다. soak·candidate panel은 기존 선택과 수명을 유지하며 factory만 빌린다. openapi test/login/init/doctor 및 console 설정의 credential 검증도 동일 root factory로 최초 token/Accounts를 보낸다. 기존 probeResult 분류는 보존하되 factory 실패를 검증 성공으로 표시하거나 저장 성공의 근거로 쓰지 않는다. 임시 token의 격리·삭제 절차와 secret redaction을 유지한다. 독립 tools는 별도 로트이며 여기서 전체 coverage 완료를 주장하지 않는다.

## 콘솔 수명과 종료

consoleBroker는 read client 성공만 캐시하고 실패 후 명시 재시도를 허용한다. verify starter는 매번 새 client와 계좌 확인을 유지한다. 공유 대상은 factory다. runConsole의 signal child context를 builder에 명시 전달한다. request context가 더 짧으면 그것도 지키며, rootOptions의 공유 context 필드를 concurrent handler마다 덮어쓰지 않는다. Background로 gate 또는 계좌 조회를 기다리는 생산 경로를 제거한다.

기존 verify worker의 cancel/settle 최대 2분과 HTTP Shutdown 최대 5초는 늘리지 않는다. 서버가 먼저 오류를 반환하는 경로도 동일 종료 처리를 거치며, 원래 서버 오류와 종료 오류를 보존한다. 종료 대기 만료는 drain 성공이 아니다. HTTP shutdown 실패 시 연결을 닫아 request context 취소를 전달하고 오류를 반환한다. 별도 background 작업이 종료하지 않으면 그 사실과 owner Close 실패를 보존하며, 정상 종료·정리 완료라고 보고하지 않는다. 무기한 join이나 강제 goroutine 종료는 도입하지 않는다.

성공한 syscall.Exec는 root 실행 helper로 돌아오지 않으므로 relaunch에는 진행 중 verify worker의 실제 완료, 필요한 HTTP handler의 종료, root owner.Close 성공이 모두 필요하다. settle은 timeout과 완료를 구분해 반환하며 2분 한도 만료는 오류·exec 0이다. HTTP Shutdown의 5초 만료도 연결 Close 여부와 무관하게 오류·exec 0이다. 연결 Close 자체를 handler 종료 증거로 쓰지 않는다. listener 종료와 이 종료 확인 뒤 relaunch closure에서 owner.Close를 수행하고 실패면 exec를 호출하지 않는다. 잠시 lease가 없는 미완료 worker에서도 Close가 성공할 수 있으므로 Close만으로 worker 완료를 추정하지 않는다. 성공 시 새 프로세스 이미지는 같은 manifest를 새로 검증한다. Exec 자체가 실패하면 기존 오류를 반환하고 root helper의 재차 Close는 안전해야 한다. 포트·argv·환경·재실행 승인 정책은 변경하지 않는다. 두 번째 엔진 신호의 hard exit130은 이 로트 범위 밖의 기존 강제 종료 예외다.

이 계약은 모든 goroutine이 강제로 종료된다는 보장이 아니다. cached client의 새 Acquire는 성공한 Budget.Close 뒤 거부되며, Close가 활성 owner 때문에 실패한 경우는 실패 종료로 다룬다. 실제 SIGKILL·hard exit 검증은 별도 제품 증거로 남는다.

## 수용과 작업 경계

후속 원문 확인: `httpapi.go`의 openHTTPAPIResources도 consoleBroker를 사용하고 `httpAPIReader`의 accountRef는 context 없는 함수다. 이 두 파일의 연결부를 같은 로트에 포함한다. accountRef에 실제 요청 context를 전달하고 signal child context도 수명 경계로 연결한다. runHTTPAPI의 서버 선오류·종료 신호를 공통 Shutdown 정리로 모으며 기존 5초 한도, publisher cancel/join, 원래 오류를 보존한다. Shutdown 실패는 연결 Close와 오류 반환으로 처리하고 handler 완료로 주장하지 않는다. JSON schema·응답 status·권한·mutation 경로는 변경하지 않는다. 이 확장은 새 API 기능이 아니라 기존 broker builder의 필수 caller 배선이다.

FLM/BTM·AST 사전 기록 후 local TLS fake credentials로 token/Accounts 첫 전송의 guard, 계좌 재생성 동일 budget, required 오류 wire0, OFF golden, console 동시 lazy 생성·실패 재시도·run별 새 client, 취소 중 bootstrap·server 오류·shutdown 만료·Close 실패 시 relaunch 0회·exec 실패의 중복 Close를 시험한다. closure에 실제 production seam을 연결한 fixture로 검증하며 브로커·실계좌·운영 명령은 금지한다. 정확 선택 normal/tag/race·변이·수동 gstack·독립 검증 뒤 별도 SHA를 판정한다. 전체 tests/gate 승인으로 해석하지 않는다.

search-first: 기존 root owner, console lifecycle, verify 생성자와 stdlib context/http shutdown/errors를 확장한다. 신규 의존성·공유 client singleton·계좌 DB 없음. grill-me/grilling: 종료를 무기한 기다릴지 재질문하지 않고 기존 2분/5초 한도와 실패 표시를 유지한다. impeccable: N/A — backend wiring/lifecycle-only; 기존 UI 경로·상태 표현을 재설계하지 않는다. 오류가 기존 화면의 실패 경로에 올바르게 드러나는지는 fixture로 확인하며 rendered UI를 검증했다고 주장하지 않는다.

독립 설계 검토는 owner.Close 성공만으로 relaunch를 허용하면 lease 사이의 미완료 verify가 끊길 수 있음을 지적했다. Manager가 실제 worker 완료·HTTP Shutdown 성공을 별도 조건으로 추가했다. 만료한 worker가 잠시 lease 0인 반례와 connection Close 후 handler 잔존 반례를 fixture에 포함한다. 독립 delta 재검토에서 P1 해소·조건부 PASS를 확인했다. Manager는 이 계약의 구현을 승인한다. source/AST·fixture·변이·독립 구현 검증은 별도 수용 조건이다.
