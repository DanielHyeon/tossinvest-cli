# retained gap/Last 상태 모순 delta

기준 22784a1b53bbb7390089f6dc2af1c4bcaca3899a. Manager와 독립 reviewer가 발견한 P2: canonical state에 같은 attempt가 잔존 gap과 Last.NoResponse로 동시에 들어가도 기존 seenAttempts가 event만 보유해 허용했다. 생산 수정은 gap loop에서 검증한 attempt를 기존 집합에 추가하는 한 줄이다. 파일 형식·API·정상 writer 동작은 바꾸지 않는다. Pre-Edit FLM/BTM과 pre-AST는 gap-last-pre-edit.md와 gap-last-validation-pre.json; post bundle은 function-logic/internal-accountbudget--validateobservationstate이다.

canonical JSON을 event.state에 넣고 Snapshot으로 거절하는 TestA166ObservationRejectsRetainedGapAsNoResponse는 baseline RED (`gap-last-red.log`)→GREEN. 한 줄 제거 overlay도 동일 assertion RED (`gap-last-mutant.log`, patch gzip SHA `gap-last-mutant.sha256`). 관련 정확 6개 선택 normal 5.158초·race 7.170초 PASS (`gap-last-normal.log`, `gap-last-race.log`). 해당 범위는 정상 no-response·gap ring·full bounded state·2process crash·disk fault 회귀를 포함한다.

수동 gstack: bounded ring, 정상 상태 생성 가능성, 오류의 Snapshot 전파 및 동시 writer 파일 잠금 순서를 검토했고 추가 P0/P1 발견 없음. 네트워크·실계좌·운영·전체 gate 미실행, 독립 delta 검증 대기. 증명 범위는 보존된 ring이며 이미 evict된 과거 attempt의 완전한 이력은 증명하지 않는다. UI/UX 비해당, 신규 의존성 없음.
