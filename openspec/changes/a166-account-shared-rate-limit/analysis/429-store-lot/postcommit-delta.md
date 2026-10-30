# 마지막 directory fsync 오류 delta

기준 2907237a0982573c96d596680ec7a2b0c79c0213. Manager가 승인한 429-store-draft.md의 postcommit 오류 한 문단만 추가했다. 생산 source 변경 없음; 기존 함수 변경 없음, FLM/BTM N/A — 새 test-only fixture. impeccable N/A — backend test/documentation.

TestA166ObservationPostcommitBeginFailureBecomesGap은 state rename과 directory fsync 뒤 dirty unlink의 마지막 directory fsync만 오류로 주입한다. Begin은 오류와 sequence 0을 반환하고 fake caller의 전송 분기 0, 현재 pending은 유지된다. 실제 다음 Lease에서 gap이 생기고 이전 poll slot은 poison된다. 실제 transport/wire는 이 leaf 시험에 없으므로 wire 측정을 주장하지 않는다.

TestA166ObservationPostcommitFinishFailureNoDuplicate는 같은 마지막 오류 이후 원래 429 event가 이미 commit되고 pending이 없어지는 상태를 읽는다. 다음 Lease의 정상 NoResponse 완료 뒤에도 이벤트는 하나이고 gap 0이다. 전원 손실 실측은 아니다.

두 개 지정 normal 1.614초, race 2.627초 PASS (`postcommit-normal.log`, `postcommit-race.log`); 마지막 sync 오류를 삼키는 source overlay는 두 assertion에 의해 RED (`postcommit-mutation.log`). 변이 patch 원문 gzip과 해제 SHA는 `postcommit-mutation.patch.gz`와 `postcommit-mutation.sha256`. 수동 gstack: 상태 commit 지점·오류 보존·다음 owner의 sequence/gap을 범위 한정 재검토했으며 추가 발견 없음. 네트워크·실계좌·운영·전체 test/gate는 미실행이다.
