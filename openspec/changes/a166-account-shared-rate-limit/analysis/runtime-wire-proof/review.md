# runtime factory 실제 TLS 시작점 — test-only 검증

기준 `d2ce75b6229107a8bc5f8d3c9ef1c9ef93a806f1`. 생산 Go 함수와 runtime 소스는 수정하지 않았다. 변경은 Linux `_test.go` 두 파일과 이 분석뿐이다. 계측 가능한 fixture 결과이며 전체 a166 승인, 모든 배포 참여자의 namespace 일치, 실계좌·브로커 도착률 보장은 아니다.

## 시험 경계

외부 `package accountbudget_test`가 같은 파일 ID를 고정한 신뢰된 외부 manifest를 사용해 별도 프로세스 둘에서 `accountbudgetruntime.Open` → `Factory.NewClient` → 실제 `official.Client.Accounts`를 실행한다. 각 프로세스는 distinct fake credentials로 localhost TLS fixture에 token→accounts 401→token 재발급→accounts 성공을 보낸다. 각 자식의 event 정확히 4개, 서버의 프로세스별 token 2/accounts 2 및 미예상 요청 0을 확인한다. 두 프로세스 합계 8개의 core guarded 첫 plaintext Write 직전 seam을 관찰하며 `CLOCK_MONOTONIC` 시각으로 인접 350ms 이상과 모든 연속 4개가 1초 밖에 있음을 확인한다.

factory가 만든 official client의 `budgetTransport`는 비공개라 production API에 observer를 추가하지 않고, test-only `reflect`/`unsafe`로 해당 필드를 읽는다. 필드명·정확한 `*accountbudget.Transport` 타입, nonnil 및 `UsesAccountBudget` 결속을 사전에 확인하고, 요청 전 `_test.go`의 helper로 관측 함수만 설정한다. client/transport/budget/authority는 교체하지 않는다. `go list`에서 helper는 TestGoFiles이고 production GoFiles에 없다. unsafe는 private 구조가 바뀌면 시험을 실패시키며 production 빌드에 포함되지 않는다.

관측 callback의 clock 읽기와 파일 기록은 실제 TLS Write 호출 직전 test seam이며 rate 의사결정에는 사용하지 않는다. 기록 자체의 파일 I/O와 스케줄링은 실행을 늦출 수 있다. 계측 값은 instrumentation 아래의 local 첫 Write 경계 표본이지 서버 도착 시각, 네트워크 분산, 비계측 운영 실측이 아니다. 두 child가 같은 kernel CLOCK_MONOTONIC namespace라는 fixture 조건을 사용한다.

## 실행 및 반증

- `seam-red.log`: test-only helper 추가 전 컴파일 RED, `seam-green.log`: 추가 후 GREEN.
- `go125-final-normal.log`, `go125-final-race.log`, `go126-final-normal.log`, `go126-final-race.log`: Go 1.25.0/1.26.5 각각 대상 패키지에서 `TestA166RuntimeWireSeamAvailable|TestA166RuntimeFactoryTwoProcessesTLSStarts`만 normal/race 실행. 각 원문에 8개 sample과 인접 차이, token/accounts 횟수가 기록된다.
- `mutation-final-interval.log`: Go overlay에서 350ms를 0으로 바꾼 변이는 인접 22.7657ms로 RED. `mutation-final-seam.log`: Go overlay에서 firstWrite 전달을 생략한 변이는 관측 파일 누락으로 RED. 변이는 /tmp 파일에만 적용했고 공유 source는 변경하지 않았다. 모든 네트워크 상대는 local TLS fixture다.
- 기존 core `TestTLSFirstWriteSpacingAcrossProcesses`와 runtime `TestTwoRealClientsShareGroup`은 각각 transport 및 factory fixture였고 이 신규 시험은 factory와 actual official.Client를 연결한 직접 경계다. 기존 결과를 이 시험으로 소급하지 않는다.

## 검토와 제한

사전 근거·연구 재사용·scope는 `pre-edit.md`. 기존 제품 함수 변경이 없어 기존 함수 FLM/BTM N/A이며, 작성한 시험 자체의 변경 전 AST는 `pre-TestA166RuntimeFactoryTwoProcessesTLSStarts.json`. CodeGraph는 미초기화, graph freshness PASS 아님. 로컬 gstack checklist 수동 적용: trust/authority 분리, concurrent hook 설치 시점, input·count·failure 경로와 mutation 확인. specialist·telemetry·설치·업그레이드 없음. Manager 독립 검증 전이다.

실제 운영 host/container에 같은 group이 배포됐는지, 첫 요청 외 수천 요청이나 시스템 시간 변동, 타 OS, 실제 fsync EIO/power loss, 비협조 참가자·관리자, 브로커 도착률·계좌 호출은 검증하지 않았다. 전체 test/gate, 운영, live, push, merge는 수행하지 않았다.
