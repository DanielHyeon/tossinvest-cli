# a112 0.5 전체 diff 프리랜딩 리뷰 — 공통 브리프 (2026-10-05)

너는 이 변경을 **만들지 않은** 신선한 컨텍스트의 리뷰어다. 저자의 결론(주석 · review.md · tasks.md)을 믿지 말고 코드와 실행으로 반증하라.
판정 형식: **P0**(실주문 · 손절 · 노출 상승 경로를 깨거나 열 수 있음, 안전 불변식 위반) / **P1**(판정 · 수락 집합이 의도와 다름, 또는 증거가 주장을 못 받침) /
**P2**(진단 · 문서 · 시험 판별력 · 유지보수). 각 지적에 **file:line 과 그 줄의 원문 인용**, 그리고 **재현 명령 또는 변이(overlay/사본)와 그 결과**를 붙여라 —
재현 없는 지적은 P2 이하로만 적는다.

## 안전 규칙(어기면 리뷰 무효 — 어겼다면 출력 맨 위에 적어라)

- `~/.codex` 아래 어떤 파일도 읽거나 검색하지 마라. 리뷰 디렉터리의 `_codex*/` 도 열지 마라.
- 읽기 전용. 저장소 파일을 편집하지 마라 — 변이 실험은 `go test -overlay` 또는 스크래치 사본(`git -C /mnt/D/Axipient/workspace/TossOS archive <commit> | tar -x -C <사본>`)
  에서만.
- 모든 셸은 `set -euo pipefail` 로 시작하고, git 은 `git -C <절대경로>` 로만. 사본에서 git 을 쓰기 전에 `test "$(git -C <사본> rev-parse --show-toplevel)" = "<사본>"`
  을 단언하라(전례: 리뷰 픽스처의 mkdir 실패 뒤 cwd 가 실제 저장소라 git init/commit 이 돈 사고).
- 네트워크 · 브로커 · LIVE · 토글 · 엔진 lock · `mutating: true` 명령 금지. 시험은 `go test`(태그 `tossos_testseams` 허용)만.
- 시작과 끝에 `git -C /mnt/D/Axipient/workspace/TossOS rev-parse HEAD` 와 `status --short` 를 재고 같음을 보고하라.

## 좌표 · 입력

- **좌표 고정: base `1e25b3a31b7109cf7688768000914ce04360a5a9` → 착지 `4cbcfb36365c5910a681883841edb9e7fad9cc04`.** 작업 트리를 「후」로 쓰지 마라(그 뒤 미커밋
  증거 파일이 있다). 실험은 `4cbcfb36` 의 격리 사본에서.
- diff(줄이 잘리지 않은 파일 — 이 디렉터리): `diff-production-go.patch`(생산 Go 27 파일, +1785/−92) · `diff-tests-go.patch` · `diff-docs-make.patch`(OpenAPI ·
  docs/operations.md · ROADMAP · Makefile) · `commits.txt`(창의 커밋 35).
- 창의 주요 로트(근거는 `openspec/changes/a112-run-four-strategy-families-independently/review.md` 의 해당 절 — 주장으로만 읽고 검증하라):
  2.3 breakout 1.2 반사실(`internal/breakoutlane`) · 8.8.4 로트 B(활성화 거절 필드명, `strategyrouter/production_family_activation.go`) · 8.5 응답 로트(관문 판정 재계산 ·
  schedule 재검증 · 레인 step seam) · 8.2 가드(`internal/testenv`) · 3.8/4.5 시험 · **7.3.1 SHADOW 구현**(`internal/strategyshadow` · `strategyworker/shadow.go` ·
  엔진 `strategy_shadow_batch.go` · `strategy_lane_shadow*.go` · 투영 · `strategyprojection/lanes.go` · 도구 `tools/a112-family-shadow`) — 설계 브리프
  `analysis/shadow-2026-10/design-brief.md` v3.3(freeze 종결판).
- 안전 불변식(TossOS `.claude/CLAUDE.md`): 사람 승인 없는 LIVE 주문 0 · 토글 OFF = upstream 동작 · 손절/비상 청산 즉시성 불변 · 주문/손절/사이징/Guardian/원장/
  대사/인증/체결 경로는 High-risk. SHADOW 는 「읽기 전용 반사실」 — desired/effective/활성화 · dispatch · 원장에 닿으면 P0.

## 출력

맨 위에 전체 판정(SHIP / SHIP-WITH-FIXES / HOLD) 한 줄. 그다음 표: `| # | 등급 | file:line | 원문 인용 | 문제 | 재현/변이와 결과 | 고치는 법 |`. 그 뒤 「확인했으나 문제 없음」
목록(무엇을 어떻게 쟀는지 한 줄씩). 끝에 저장소 무변경(시작/끝 HEAD · status). SubagentHandback 으로 보고하라.
