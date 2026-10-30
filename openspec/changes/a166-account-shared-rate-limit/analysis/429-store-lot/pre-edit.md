# 429 observation store leaf Pre-Edit

기준 c412c0e55f2948e16e1f33f1e1e8fadaf2092c56. 승인된 analysis/429-store-draft.md를 runtime 원본에서 복사했다. 기존 Budget Provision/Open/Lease.Release, transport, runtime, official·caller·schema 편집 없음. 신규 동일 package leaf가 Lease.mu를 보유한 동안 동일 Budget·wire·released를 검사해 Release 경합을 직렬화한다. 잠금 순서는 wire→event, snapshot은 event 해제 후 wire NB→event generation 재검증. 기존 함수 편집 없음: 기존 함수 FLM/BTM N/A. 신규 코드의 상태와 검증을 이 디렉터리에 기록한다.

search-first: accountbudget의 dirfd/NOFOLLOW·파일 안전검사·flock·dirty/sync/atomic rename, clock.RemainingBootHint를 재사용한다. event lock은 독립 identity이며 변경 가능한 event state/dirty는 inode pin하지 않는다. 새 의존성 없음. ai-memory 검색 결과 없음. CodeGraph 미초기화. impeccable N/A — backend-only. 실계좌·네트워크·운영·전체 gate 없음.

상태 후보: 버전/세대와 attempt/event/gap uint64, typed pending(kind + poll일 때 고정 32-byte key), 128 event와 gap 창, 256 poll slot, 최대 직렬화 256KiB. 최대 증가 overflow·불명 enum/key/형식/dirty·파일 불안전은 failclosed. no-response 정상 종료와 orphan gap 구분, poll orphan poison 지속. valid hint는 기존 미만료 잔여와 max, reboot는 최초 anchor를 같은 lock 아래 저장. snapshot cursor overrun/정상 pending/비차단 재검증 별도 표현. 저장 파일에는 account·header·URL·body 원문을 담지 않는다.

Manager 승인: Lease struct에 private observationStarted/observationSequence만 추가하며 기존 함수 본문은 변경하지 않는다. 동일 Lease에서 중복 Begin 또는 다른 handle 재시도와 타 Lease Finish를 거절한다. Finish는 명시 ctx를 받아 요청 취소와 별도 bounded settle ctx를 호출자가 선택할 수 있게 한다; transport 설계는 후속이다.
