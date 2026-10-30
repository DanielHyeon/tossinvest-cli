# a166 runtime TLS 시작 증거 — 사전 분석

기준 `d2ce75b6229107a8bc5f8d3c9ef1c9ef93a806f1`. 신규 test-only `internal/accountbudget/export_test.go` 및 외부 `package accountbudget_test` fixture만 추가. 기존 제품 함수·runtime 소스·기존 시험 수정 없음: 기존 함수 FLM/BTM N/A. 사용자 승인 범위 내 이 분석 디렉터리만 별도 기록. UI 없음: impeccable N/A — test-only.

search-first: 기존 `internal/accountbudget/transport_test.go`의 guarded 첫 Write test seam 및 두 프로세스 fixture, `internal/accountbudgetruntime/runtime_test.go`의 manifest·factory·actual official.Client token/accounts fixture를 현재 기준 원문에서 확인. Go 표준 `httptest`, 독립 프로세스 및 기존 `x/sys/unix` monotonic clock 사용. 새 의존성 없음. CodeGraph 별도 worktree not initialized; 그 결과를 evidence로 사용하지 않음. ai-memory/Hindsight는 앞선 runtime 독립검토에서 호출했고 현재 SHA를 재검증했다.

외부 test가 `accountbudgetruntime`을 import하면서 test-only export helper로 package 내부 `Transport.firstWrite`를 설정한다. official.Client의 private budgetTransport는 생산 exported observer 없이 test-only reflection/unsafe로 읽는 방안이며, 대상이 정확히 factory가 생성한 transport·같은 Budget임을 concrete `UsesAccountBudget`로 검증한다. 취약한 구조 변경 또는 type 오류는 테스트 실패한다. 이 helper는 `_test.go`에만 존재한다.

실제 local TLS server와 두 독립 프로세스는 같은 외부 trusted manifest/root/inode를 연다. 각 process는 distinct fake credentials로 실제 official.Client accounts 요청을 수행하며 최초 accounts 401→token 재교환→accounts를 유도한다. 각 요청 첫 plaintext TLS Write의 core seam에서 같은 호스트 `CLOCK_MONOTONIC` 값을 기록한다. 외부 파일 기록 overhead와 스케줄링 때문에 계측은 정확한 서버 도착 시각이 아니며, seam은 Write 직전 계측이다. 최소 8개 관측, 이웃 350ms 및 임의 4개가 1초 안에 없음을 검사한다. 구조 결함 또는 missing hook/calls는 즉시 실패하도록 기대 총수·서버 경로 및 401 횟수도 검사한다.

RED: 누락 helper/observer, 기록 누락 및 주입한 간격 위반(또는 temp overlay의 350ms 제거). 이후 Go1.25/1.26 대상 테스트 normal/race, 수동 gstack checklist, 변이 원문. 전 프로세스 기동/운영 namespace·실제 브로커·전체 테스트/gate는 범위 밖.
