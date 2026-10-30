# poll boot clock leaf Pre-Edit

기준 `ccc39f6f137236740efa924f96fa2a8f2d953bd9`. Manager 승인 계약과 design의 해당 복원 문단만 source worktree에서 복사했다. 기존 `clock.Clock`, `System`, lease 함수는 변경하지 않고 신규 `internal/clock` Linux leaf와 non-Linux unsupported stub 및 시험을 추가한다. UI: impeccable N/A — backend-only. 기존 제품 함수 본문 변경이 없으므로 기존 함수 FLM/BTM N/A; 신규 함수에 대한 결과·시험·변이를 기록한다.

search-first: `internal/clock/clock.go`의 주입 Clock/Sleep과 프로세스 내 LeaseAnchor/Elapsed, `cmd/tossctl/engineready.go`의 boot ID 프로세스 식별 선례, `internal/accountbudget`의 x/sys/unix와 `internal/clock/clock_test.go`의 lease 단조 핀·`internal/app/engine/a090_monotonic_pin_test.go`의 AST 기반 lease 호출 핀을 원문 확인했다. 기존 `/proc` 선례는 bounded/namespace 엄격 검증이 없어 이 leaf를 대체하지 못한다. Go stdlib 및 기설치 x/sys만 사용한다. CodeGraph 상태 미확인, graph hard evidence를 주장하지 않는다.

RED: strict boot UUID·offset 파서, namespace ID 및 boot 앞뒤 변화·자식 namespace 불일치·syscall 오류, 순수 remaining의 같은 boot/namespace·역행·다른 boot anchor 필요·재시작 재사용·duration 0/30s·overflow. Linux 로컬 proc read smoke는 ID를 출력하지 않는다. non-Linux unsupported stub. 새 전역 cooldown·store·anchor 영속화·HTTP-date·운영 설정·namespace 조작은 범위 밖. 기존 `Clock` 인터페이스와 System의 동작, 350ms admission은 그대로다.
