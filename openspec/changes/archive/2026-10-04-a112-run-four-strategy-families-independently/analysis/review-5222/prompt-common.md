# a112 5.2.2.2 로트 적대 리뷰 — 공통 브리프

## 안전 규칙(먼저)

- 리뷰 트리는 `/tmp/claude-1000/a112-review-80ae96a5`(`git archive 80ae96a5` 전체, git 이력 없음). 실제 저장소 `/mnt/D/Axipient/workspace/TossOS` 는 **읽기 전용** —
  git 은 `git -C /mnt/D/Axipient/workspace/TossOS show|log|diff` 만. 상태를 바꾸는 git 명령 · 저장소 안 파일 생성/수정 금지.
- 실험은 자기 사본에서만. 모든 셸 블록 첫 줄:
  `set -euo pipefail; D=$(mktemp -d /tmp/claude-1000/a112-rev5222-XXXX); cp -a /tmp/claude-1000/a112-review-80ae96a5/. "$D"/; cd "$D"; case "$PWD" in /mnt/*) echo REFUSE; exit 1;; esac`.
  사본에서도 `git init` 금지(사본 안에 `.git` 이 없어야 한다 — 있으면 멈춰라). 변이는 한 번에 하나 · 원복 확인 · 무변이 대조군 GREEN 먼저. `GOFLAGS=-trimpath`,
  `GOCACHE=$D/.gocache` 권장. 끝나면 사본을 지운다. 리뷰 트리는 공유 — 고치지 마라.
- LIVE 주문 · 운영 토글 · 엔진 기동 · `mutating: true` 금지. 운영 원장(`~/.config/tossctl/`) · 자격 증명 열지 마라.
  **`~/.codex` 아래 어떤 파일(기억 · 설정 포함)도 읽거나 검색하지 마라. 어겼다면 출력 맨 위에 적어라.**
- `rtk` 가 `go test -v` 출력을 요약할 수 있다 — 셀 때는 `rtk proxy go test …` 또는 `go test -json`.

## 대상 — 커밋 `80ae96a5`(브랜치 feat/a112-four-family-runtime, 부모 7ab8cd12)

a112 5.2.2.2: 하류 권한(결과 · 위험 · 계좌 · 1차 레그 · worker 승격 · projection)을 소유자 범위 단위로 옮기고 1차 레그의 시장 단위 개수 관문을 제거.
읽을 것: `git -C … show 80ae96a5 --stat`, change `openspec/changes/a112-run-four-strategy-families-independently/` 의 `review.md` 끝 절
「2026-10-01 태스크 5.2.2.2 …」(Pre-Edit · 전수표 (a)~(e) · 편집 로트 · 잔여 R1~R5 · Manager 판정), `tasks.md` 5.2.2.2, 코드:
- `internal/app/engine/strategy_owner_scope_authority.go`(범위 키 · `strategyScopeRefusal`)
- `strategy_account_first_leg_authority.go` `collectStrategyFirstLegAuthority` · 계좌 `collectMarket`
- `strategy_risk_authority.go` 위험 `collectMarket` · `strategyRiskMarketFromScopes` · `forScope`
- `strategy_proposal_authority.go` `ResultAuthority` · 제안 `collectMarket`(digest)
- `strategy_first_leg_admission.go` `admit` · `strategy_dispatch_cycle.go` `dispatch` · `strategy_market_handoff_delivery.go` `deliverEachStrategyHandoff`
- `strategy_entry_supervisor.go` `buildProductionStrategyMarketWorker` · `strategy_runtime_projection.go`
- 시험: `a112_owner_scope_trading_test.go`(Done · 거울 다리 · 트립와이어 · 위조 종단 · 승격) · `a112_scope_refusal_census_test.go` · `a112_first_leg_owner_scope_seal_test.go` ·
  `a112_owner_scope_handoff_test.go` · `strategy_dispatch_handoff_guard_test.go`(census)
- 증거: `analysis/measurements/lot-5.2.2.2/`(RED · 변이 원장 · 스키마 핀 영수증 · J2 · 격리 검증 · check 델타 · 편집 전 번들), `analysis/harness/{a112_lot_mutate.py,render_5222_bundles.py}`
`analysis/review-5222/` 의 다른 파일은 열지 마라(이 브리프와 자기 프롬프트만 예외).

## 설계 배경(Manager 판정 — 다툴 필요 없음, 어긋난 구현만)

J1 범위별 목록 · J2 무캡(상한에 이름) · J3 범위별 거절(그 범위만 fail-closed, 봉투 폴백 금지, 기록 · 가시) · J4 타입 분류 거절(문구 아닌 타입, journal/gateway/central 오류를
범위 거절로 오분류 금지) · J5 관문 전수표. 판정 B: 생산 위험 적재기의 스키마 핀 27(실원장 v35 거절)은 별도 change — 시험은 원장 행을 stub 으로 복사하는 다리를 쓴다.
판정 ④: 같은 파도 두 범위 동시 발급 불가는 공유 버킷 CAS 보호의 설계 — 범위별 발급은 파도 순차. R2~R5 는 이름 붙인 이월(다툴 필요 없음).
활성화 없는 시장(오늘 생산 전부)의 판정은 편집 전과 같아야 한다(토글 OFF = upstream).

## 판정 기준 · 출력

- 모든 주장은 파일:줄 + 가능하면 **실행**(사본에서 시험 · 변이). 추측은 추측이라고.
- 등급: P0 = 안전 불변식 위반 · 생산 동작 변화 · 주문/손절 약화 · 위조가 1차 레그 발급까지 가는 경로 · 틀린 범위의 권한으로 발급 / P1 = 착지 상태로 둘 수 없는 결함
  (뚫린 경계 · 거짓 증거) / P2 / (T).
- 맨 위 **판정 한 줄 APPROVE / BLOCK**, 발견 표(`# | 등급 | 주장 | 증거 | 권고`), 필수 항목별 한 줄 판정, 마지막 줄 `Recommendation: …`. 만든 사본 경로를 적고 지워라.
