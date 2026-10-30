# offline budget tool 독립 설계 검토

판정: P1 두 항목 정정 후 구현 가능. 대상 `analysis/offline-budget-tool-plan.md` 초안 및 v2 후보 21b9bc85b. source 손읽기이며 AST·시험·코드 변경·운영 실행 없음.

- P1 경로/취소 보장 과장: `internal/accountbudget/budget_linux.go:69–73` Provision은 경로를 재개방하고 context와 dirfd를 받지 않는다. `:104–111` OpenContext도 경로를 다시 연다. 'root fd에 이후 전체 동작 결속' 및 '둘 모두 기존 API에 context 전달'은 사실이 아니다. 기존 신뢰 전제(동일 UID 악의적 경로 교체 없음)를 명시하고 최초 root fd identity와 재개방된 객체를 Provision 뒤·게시 직전 재검사하여 불일치 시 fail closed; Provision 전후 취소 검사만 약속한다. manifest parent의 openat/renameat2는 fd 결속 가능. 적대적 동일 UID까지 완전 방어가 필수라면 별도 dirfd API가 필요하다.
- P1 check clock 주장: 후보 `internal/accountbudget/observed_linux.go:14–20`, `observation_linux.go:99–115`는 boot source 존재와 state 읽기만 하고 현재 boot/time namespace snapshot을 채취하지 않는다. check 성공이 현재 clock 지원을 확인한다고 하려면 기존 SystemBootSource에서 명시 snapshot 검증을 추가하거나 그 주장을 제외해야 한다.
- writer schema: runtime v2 `runtime_linux.go:35–42,141–219`의 manifest 타입/decoder는 private이다. 소규모 내부 공개 직렬화·검증 경계 또는 같은 패키지의 좁은 provision API가 필요하며 writer가 중복 schema를 발명하지 않도록 기존 parser와 roundtrip fixture를 둔다. `ProvisionObservation` event ID는 별도 pin; 실제 multi-namespace 동일 inode는 사람이 확인한다.
- O_EXCL 및 renameat2 NOREPLACE, file sync→rename→parent fsync, 게시 후 fsync 실패 시 오류와 최종 파일 보존, 부분 root 삭제 금지는 적절. 기존 destination을 교체하는 fallback이나 자동 복구 금지. 신규 의존성 없이 x/sys v0.46.0 사용 가능(go.mod:14).
- check는 Factory.Open/Close로 network/client 없이 queue·event state/identity를 검사할 수 있으나 이 결과만으로 모든 참여자·무전송·운영 준비·권한을 증명하지 않는다. 전환/복구의 drain·30초·fencing은 사람이 담당한다.
