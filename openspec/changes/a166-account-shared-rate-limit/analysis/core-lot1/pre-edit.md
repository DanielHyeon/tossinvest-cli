# 제품 첫 로트 Pre-Edit

- 기준: `da67f3e1e9282e181160467fb2a476345a08627a`; base-commit은 최초 생성 후 불변.
- 범위: 신규 `internal/accountbudget/**`, 본 분석과 작업표 1.x; 공식 client/engine/운영 설정 미수정.
- 현재 HEAD에서 해당 패키지 없음. 기존 제품 함수 본문 변경이 아니므로 기존 함수 FLM/BTM 해당 없음. 연구 프로토타입은 제품 소스가 아니며 사전 검증을 대체하지 않음.
- CodeGraph: 신규 worktree에서 `codegraph status` 결과 not initialized. 초기화하지 않았고 그래프 호출관계는 미검증. `officialfx/production_owner_unix.go`, `production.go`, `positionpolicyrpc/private_fs_unix.go`, `clock/clock.go` 원문 직접 확인; 수동 검색 증거이며 AST 열거를 주장하지 않음.
- 연구: queue `41c8c16a7`, transport `d0485309a`는 설계 참고만; Go stdlib, 기존 `golang.org/x/sys v0.46.0`, 내부 clock 재사용. 신규 의존성 없음.
- 불변식: 기존 production 호출자/토글 변경 없음; shared root/scope/lock failure는 fallback 없이 fail closed; wire → queue 역순 금지; context와 revoke/drain은 현재 시도에만 적용; actual live/브로커/운영 접근 금지.
- 계획 검증: 신규 패키지의 RED/GREEN, 정상/race·다중 프로세스·선별 변이·gstack. 전체 테스트/전체 gate 미실행. UI 없음: impeccable N/A — backend-only.
