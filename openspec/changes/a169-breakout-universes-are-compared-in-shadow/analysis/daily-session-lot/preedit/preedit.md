# P9 Pre-Edit — 영속 세션 제한 로트

- 기준: `232d6b86c024dfd71cc8ffe6fd0fbba342b3e9ec`, `analysis/daily-session-plan.md`. 이 선언은 설계 리뷰와 Manager READY를 대신하지 않는다.
- 단일 writer 소유: 신규 `internal/shadowsession` production 파일 및 신규 package 테스트, 이 증거 디렉터리. 기존 production/test 함수 본문은 수정하지 않는다. 필요하게 되면 중단하고 해당 함수의 사전 FLM/BTM과 별도 선행 커밋을 준비한다.
- 예상 API: Root의 절대 정규 경로와 기대 dev/inode, Begin(ctx, root, day), Session Pause/Abort/Complete/Close. 상태 전이·canonical JSONL·128 KiB 제한·ext4/root/file 검증은 계획 §2–4를 따른다. 운영 caller, HTTP, engine, config, DB, 기존 journal의 변경 없음.
- 잠금: root/day별 검증된 동일 lock inode를 상태 접근 전에 nonblocking exclusive flock하고 Session 폐기까지 보유한다. 잠금 파일을 삭제하지 않는다. 파일/디렉터리 sync, identity 재검사, 실패 시 handle 폐기와 commit 후 오류 의미를 계약대로 구현한다.
- 안전 전제: Pause/Complete 전 caller drain 및 전송 중단, Abort 전 신규 작업 중단. 이 패키지는 실제 wire/drain, 현재 KST 거래일, storage rollback이나 물리적 내구성을 증명하지 않는다. 오류 Begin은 handle 없이 wire 허가 0; 이후 판정은 검증된 기록에 따른다.
- CodeGraph: `codegraph-status.json`, `checked-dir.json`, `checked-file.json`, `provision-observation.json`, `shadowsession.json`. 1.6.0 별도 worktree index, 최초 HEAD 기준 3,177 files, pending 0; `.sdd/index-state.json` 기록 또는 freshness gate PASS 아님. 새 패키지 query 결과 빈 배열이며 현재 production caller 0의 한정된 색인 근거다. 대상은 신규 심볼이라 definition/callers/callees/impact 및 기존 함수 FLM/BTM은 사전 해당 없음. 추가되는 각 함수의 Go AST/분기 지도는 구현 후 별도 보존한다.
- 기존 근거: `../../daily-session-research/provenance.md`의 checkedDir/checkedFile 및 probe.startSession AST와 source hash. `runlock`은 배타적이지 않고 `enginelock`은 수명 flock만 제공한다. accountbudget helper는 비공개·전용이며 export 또는 기존 경로 수정 없음. CodeGraphContext/GBrain 신규 색인 미실행, 후보 문맥은 source 및 연구 AST로 확인; 보조 그래프가 hard evidence를 대체하지 않는다.
- 시험 매핑: 최초 시작/정상 pause·resume/abort·complete 및 active kill 차단, 날짜·상태 검증, 동시 process lock, root/file identity·권한·symlink·hardlink·FS, malformed/canonical/checksum/sequence/size, create/write/short-write/file sync/dir sync/unlock/close 전후 오류와 재시작. 테스트 root는 명시적인 ext4 임시 경로, 시스템 TMPDIR 변경 없음. caller drain은 API 전제이며 이 테스트만으로 runtime 결속을 주장하지 않는다.
- RED 후 최소 GREEN, 신규 package 한정 normal/tagged race·필요 subprocess 반복·unsupported compile·선택 변이, 구현 후 scoped gstack 및 독립 리뷰 예정. 전체 테스트/gate, 실제 broker, 운영 변경·실측·배포 미실행. 기존 upstream production 변경 없음; 영향 package 시험과 caller 0을 확인하고 미실행 범위를 보고한다.
- UI 없음: `impeccable: N/A — backend/documentation-only`. 기존 함수 변경이 없어 upstream 상속 함수 회귀 영향 없음으로 예상하되 테스트 전 PASS 아님. 새 dependency 없음, 영속 파일만 신규이며 자동 복구·삭제 없음. 승인 없는 live 주문·토글 변경 없음.
- 작성 시 scoped `openspec` 1.4.1 strict와 PM `--check`, `git diff --check` 통과. 체크 이후 계약이 바뀌면 다시 수행한다. Manager READY 및 계약/Pre-Edit 선커밋 전 소스·테스트 편집 금지.
