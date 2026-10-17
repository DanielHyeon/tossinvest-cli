# codex — a112 6.2 봉인 로트 독립 리뷰(넓게)

아래 공통 브리프를 따른다(이 파일 아래에 붙어 있다). 읽기 전용 샌드박스 — 실행이 안 되면 코드 인용으로 판정하고 그렇다고 적어라.
작업 디렉터리는 `git archive 686b94e4` 트리(git 이력 없음). **`~/.codex` 아래 어떤 파일도 읽거나 검색하지 마라 — 어겼다면 출력 맨 위에 적어라.**

질문(넓게): 이 봉인(1차 레그 권한의 소유자 범위 재유도 + identity 대조)이 「조립이 중재하지 않은 봉인 제안은 1차 레그로 발급되지 않는다」를 **종결**하는가?
볼 것: (1) 위조가 1차 레그 발급(`collectStrategyFirstLegAuthority` 가 err=nil)까지 가는 경로가 남는가 — 다섯 축 바깥 포함, (2) 범위 선택 → identity 대조 순서가 자기 참조로
무너지는 조합, (3) A-lite 가 생산(활성화 없는 시장)을 바꾸는가 · 봉인과의 틈, (4) strategyflow census 가 기본 빌드의 봉인 주조를 닫는가, (5) 실행 증거 하네스의 한계,
(6) 토글 OFF = upstream(새로 통과하는 입력이 있으면 P0).

---
# a112 6.2 봉인 로트 적대 리뷰 — 공통 브리프

## 안전 규칙(먼저)

- 리뷰 트리는 `/tmp/claude-1000/a112-review-686b94e4`(`git archive 686b94e4` 전체). 실제 저장소 `/mnt/D/Axipient/workspace/TossOS` 는 **읽기 전용** —
  git 은 `git -C /mnt/D/Axipient/workspace/TossOS show|log|diff` 만. 상태를 바꾸는 git 명령 · 저장소 안 파일 생성/수정 금지.
- 실험은 자기 사본에서만: `set -euo pipefail; D=$(mktemp -d /tmp/claude-1000/a112-rev62-XXXX); cp -a /tmp/claude-1000/a112-review-686b94e4/. "$D"/; cd "$D"`.
  사본에서도 `git init` 금지. 변이는 한 번에 하나 · 원복 확인 · 무변이 대조군 GREEN 먼저. `GOFLAGS=-trimpath`. 끝나면 사본을 지운다. 리뷰 트리는 공유 — 고치지 마라.
- LIVE 주문 · 운영 토글 · 엔진 기동 · `mutating: true` 금지. 운영 원장(`~/.config/tossctl/`) · 자격 증명 열지 마라.
  **`~/.codex` 아래 어떤 파일(기억 · 설정 포함)도 읽거나 검색하지 마라. 어겼다면 출력 맨 위에 적어라.**
- `rtk` 가 `go test -v` 출력을 요약할 수 있다 — 셀 때는 `rtk proxy go test …` 또는 `go test -json`.

## 대상 — 커밋 `686b94e4`(브랜치 feat/a112-four-family-runtime)

a112 6.2 의 차단 선결조건(결정 (1)) · 결정 (5) 의 봉인 로트 몫. 읽을 것: `git -C … show 686b94e4`, change `openspec/changes/a112-run-four-strategy-families-independently/` 의
`review.md` 끝 절 「2026-10-01 6.2 봉인 로트」, `tasks.md` 6.2 · 6.2.0, `HANDOFF.md` 결정 (1) 과 §4, 코드:
- `internal/app/engine/strategy_account_first_leg_authority.go` `collectStrategyFirstLegAuthority`(B2 준비 · B3 소유자 범위 선택 실패 · B4 개수 관문 · B5 identity 가드)
- `internal/app/engine/strategy_first_leg_owner_scope.go` `authorityForOwnerScope`
- `internal/app/engine/strategy_dispatch_handoff.go` `dispatchHandoffs`(A-lite) · `strategy_proposal_set_digest.go`
- 시험: `a112_first_leg_owner_scope_seal_test.go` · `strategy_first_leg_backstop_shape_test.go` · `strategy_first_leg_identity_backstop_test.go` · `strategy_dispatch_handoff_guard_test.go`
- `internal/strategyflow/seal_census_test.go` · `testdata/exported_surface.golden` · `types.go`(봉인 함수)
- 증거: `analysis/measurements/lot-6.2-seal/`(RED · 변이 원장 · 편집 전 번들), `analysis/harness/{a112_lot_mutate.py, verify_named_tests.py, render_62_seal_bundle.py}`
`analysis/review-62-seal/` 의 다른 파일은 열지 마라(이 브리프와 자기 프롬프트만 예외).

## 설계 배경(Manager 판정 — 다툴 필요 없음, 어긋난 구현만)

봉인 = 마지막 권한(1차 레그)의 범위-단위 재유도(의미 봉인). 조정자는 엔진이 생성 · 공급 · 구동하므로 신뢰 주조자가 아니다(타입 토큰은 「어떤 중재가 돌았다」만 증명).
범위로 고르고 identity 로 대조한다(자기 참조 함정 회피). 시장 단위 개수 관문은 봉인이 아니라 상한이고 5.2.2.2 가 걷어 낸다. A-lite 는 2차 방어 · 가독 계약(봉인 아님).
이 로트는 판정을 더 엄격하게만 한다(생산 동작 변화 0 · 토글 OFF = upstream 이어야 한다).

## 판정 기준 · 출력

- 모든 주장은 파일:줄 + 가능하면 **실행**(사본에서 시험 · 변이). 추측은 추측이라고.
- 등급: P0 = 안전 불변식 위반 · 생산 동작 변화 · 주문/손절 약화 · 위조가 1차 레그 발급까지 가는 경로 / P1 = 착지 상태로 둘 수 없는 결함(뚫린 봉인 · 거짓 증거) / P2 / (T).
- 맨 위 **판정 한 줄 APPROVE / BLOCK**, 발견 표(`# | 등급 | 주장 | 증거 | 권고`), 필수 항목별 한 줄 판정, 마지막 줄 `Recommendation: …`. 만든 사본 경로를 적고 지워라.
