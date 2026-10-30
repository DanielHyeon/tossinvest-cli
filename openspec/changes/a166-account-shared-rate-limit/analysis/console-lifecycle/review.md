# Console lifecycle 부분 후보

기준 bbe5066ec. handler quiesce의 5초 Shutdown 후 worker의 실제 done을 최대 2분 확인한다. server 오류도 같은 정리 경로를 거치고 오류를 결합한다. Shutdown 실패 시 Close하고 재실행하지 않는다. worker finishedFlag가 실제 done보다 먼저 바뀌는 틈에 이전 run을 교체하지 않는다. 재실행 closure는 root owner Close 성공 뒤에만 exec한다. exec 실패 후 root의 재차 Close를 보존한다.

실제 worker done·미완 상태와 종료 시한, server 실패, owner Close 실패 및 fake exec 실패를 선택 normal/tagged race에서 확인했다. worker 완료 누락 및 owner Close 생략 변이는 RED였다. 영수증은 receipts.md; local HTTP는 lo만 활성화한 별도 network namespace에서 실행했다. 수동 gstack 범위는 worker·handler·owner 수명과 오류 보존을 검토했다. 정지 전 handler가 오래 유지되면 5초 Shutdown 실패가 먼저 나타나고 재실행하지 않는다. 서버 소켓 Close만으로 handler 종료를 주장하지 않는다.

CodeGraph worktree 인덱스 없음. 혼합 모델 gstack·독립 검증, HTTP API 및 verify factory의 context 배선, 나머지 CLI, 전체 tests/gate·실계좌·운영·archive는 미완이다. impeccable N/A — backend lifecycle; 렌더 UI 검증 없음. pre-edit 중 기존 runConsole bundle을 임시 덮었으나 제 생성분을 이 폴더에 보관하고 기존 HEAD 원문은 복원했다.
