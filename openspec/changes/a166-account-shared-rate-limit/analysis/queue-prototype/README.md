# 큐 프로토콜 연구

판정: **제한된 후보 검증 완료, 생산 통합 NOT READY**. 제품 경로에는 연결되지 않음. UI 영향 없음: impeccable N/A — 연구용 Go와 문서만 변경.

Manager가 정한 dirty sentinel 절차를 시험함. queue 잠금 아래 dirty 생성 및 file/dir sync 후 state 갱신 및 dir sync, dirty 삭제 후 dir sync 순서. dirty 존재 또는 확인 오류는 읽기를 거절함. injected write도 sentinel을 우회하지 않음. marker 준비 실패는 state 갱신 전 거절, state 갱신 후 dir sync 실패는 dirty 유지. 삭제 후 마지막 dir sync 실패는 state가 이미 동기화된 상태이고 marker 삭제의 내구성만 불확실함: 현재 marker 부재 시 진행 가능, 재부팅 후 marker 재등장 시 거절. 자동 복구 없음.

`GO111MODULE=off go test -race -count=1 .` 결과 `final-race.log` PASS. post-rename directory sync 오류 주입 뒤 별도 child aux/normal/critical 거절, 세 단계 directory sync 오류, FIFO(같은 프로세스), owner kill(다른 프로세스), waiter kill, 선취소 및 잠금 대기 취소, 취소 파일 정리와 dead 파일 정리, 선택 후 Close 잠금 역순 방지 포함. `dirty-mutation.log`: 임시 사본에서 marker read 검사를 제거하자 child aux 허용으로 RED. 기존 priority 변이 원본은 `/tmp/a166-queue-mutation.log`이며 현재 revision에서 재실행하지 않음. 기존 `directory-sync-red.log`는 sentinel 이전 RED이며 현재 해당 회귀 시험 PASS. `cleanup-red.log`, `dead-cleanup-red.log`, `inversion-red.log`, `cancel-red.log`는 각 보정 전 RED 원문.

동기화 실패는 callback 주입이며 실제 kernel fsync EIO, 디스크 전원 손실 및 재부팅은 미시험. marker sync, rename 및 unlink 각 실패의 모든 조합, 동시 Close/Wait, 모든 취소-선택 경합, 다중 프로세스 FIFO와 비선점은 미검증. Waiter는 한 호출자가 소유하며 Close/Wait 동시 호출은 이 연구 범위 밖임. Register의 queue 잠금은 context-aware가 아님. 생산 환경의 신뢰 가능한 디렉터리·stable inode·링크와 파일 권한 검증, 파일 형식 무결성/중복 ticket 검증, 실패 시 operator recovery, 파일 삭제의 내구성 및 프로세스 관리 필요. 현재 동작을 제품에 적용하면 안 됨. 실제 TLS/wire와 350ms 간격, 3 req/s 보장도 별도 연구임.

`Queue.read`, `Queue.write`, `Queue.Register`, `Waiter.Close`, `Waiter.tryContext`의 pre-edit AST는 `pre-*.json`에 보존. 첫 번째 연구 인수 시 기존 `Wait`/`Try` 변경은 사전 FLM/BTM 없이 진행했고 `wait-ast.json`, `try-context-ast.json`은 당시 post-edit 스냅샷임. 전체 FLM/BTM 번들, CodeGraph/CodeGraphContext freshness는 미충족. 연구용 신규 untracked 프로토타입이며 제품 함수 변경 없음. repo 검색에서 flock 및 임시 파일/rename/디렉터리 sync 기존 패턴 확인(`internal/officialfx/production_owner_unix.go`, `internal/officialfx/production.go`); 표준 라이브러리만 사용, 신규 의존성·외부 레지스트리 조사 없음.

gstack checklist 수동 리뷰: sentinel failure 경계·역순 잠금 RED를 보강했으나 위 생산 통합 위험은 미해결. 자동 설치·업그레이드·telemetry 없음. 독립 Manager 검토 미완. 전체 저장소 테스트·전체 게이트·실계좌·브로커·운영 변경·배포·활성화·merge·push·archive 미실행.
