# Manager 검증 — 미배선 공유 관측 store

대상 `2907237a0982573c96d596680ec7a2b0c79c0213`, postcommit delta `22784a1b53bbb7390089f6dc2af1c4bcaca3899a`, 모순 검증 delta `04fdd134c65499c493f245841f95a0c2b875d797`. 이 leaf의 병합·푸시를 승인한다. 실제 wire 사전 기록·전체 429 관측·runtime v2·poll 통합·a166 전체 완료가 아니다.

Manager가 신규 상태/파일/Begin/Finish/PollRemaining/Snapshot 코드와 기존 Lease struct의 두 private 필드를 손으로 읽었다. 활성 동일 Budget Lease와 Release의 직렬화, reader dirfd pin, wire→event 순서, dirty/원자 교체, typed orphan/gap/poison, bounded ring/slot/cursor, invalid hint와 429 사건 구분, BOOTTIME anchor를 대조했다. 정상 NoResponse와 crash gap을 구분한다. 마지막 directory fsync 실패의 postcommit 경계를 설계에 정정했으며 state commit 이후와 부분 저장 실패를 혼동하지 않는다.

독립 Terra가 Go 1.25 `TestA166Observation*` normal 17.017초/tagged race 20.558초를 통과했다. postcommit 두 시험은 normal 1.643초/race 2.637초, 마지막 retained gap/Last 반례는 normal 0.022초/race 1.034초 통과했다. Manager는 `/tmp/a166-store-independent-*`, `/tmp/a166-store-postcommit-independent-*`, `/tmp/a166-store-gap-independent-*` 해당 normal/race 원문을 확인했다. 독립 리뷰는 source AST/hash와 변이 13개 및 delta 변이 원문·gzip SHA를 대조했다.

retained gap과 Last.NoResponse가 같은 attempt인 canonical state를 허용하는 P2를 발견해 사전 FLM/AST, baseline RED와 한 줄 수정·변이로 닫았다. 범위는 보존된 ring이며 과거 evicted 이력 전체의 무결성을 증명하지 않는다. postcommit 시험의 전송 분기 0은 fake caller이고 실제 TLS wire 증거가 아니다. fsync/write/rename은 fault injection이며 전원 손실 실측이 아니다. trusted filesystem·정상 fsync·비변조 관리자·BootSource 전제를 유지한다.

신규 store는 production에서 아직 사용하지 않는다. 기존 Budget 함수 본문, Transport, official, runtime, Retrier는 이 로트에서 바꾸지 않았다. CodeGraph 미초기화, 전체 tests/gate·네트워크·실계좌·운영·배포·활성화는 미실행이다. impeccable N/A — backend-only, 기존 stdlib/x/sys/clock 재사용, 신규 의존성 없음. 독립 검증 후 PM/OpenSpec/diff 선택 검사와 타 작업 보존으로 게시한다. archive는 하지 않는다.
