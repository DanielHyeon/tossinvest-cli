# Issues

## 2026-07-27 — 병행 작업의 Function Logic Map이 현재 change gate를 차단

- 분류: blocking
- 성공한 검증:
  - `make sdd-sync` (`--no-gbrain` 재실행 포함)
  - `make sdd-check`
  - `openspec validate --all --strict --no-interactive`
  - 독립 리뷰 CLEAN
- 실패한 검증:
  - `make gate CHANGE=consolidate-agent-workflow-contract`
- 원인:
  - change의 `base-commit.txt` 대비 worktree에는 작업 시작 전부터 존재한
    `internal/console/console.go`, `internal/verifylive/*.go`의 기존 함수 수정이 포함된다.
  - gate는 worktree 전체의 수정된 기존 Go 함수를 현재 change의 Function Logic Map 대상으로
    계산하므로, 별도 change 소유의 함수들이 이 문서 통합 change를 차단한다.
- 결정:
  - 다른 change의 코드나 분석 산출물을 수정·복제하지 않는다.
  - 이 change의 diff만 적용한 clean worktree에서 gate를 재실행해 병행 변경과 검증 범위를 분리한다.

## 2026-07-27 — 병행 change 병합 후 base commit rebase

- 기존 base: `2f5675996e4bf891d0e90ac792c07a43c03ce584`
- 새 base: `3a2bc148199927bc1e95df4395f20db914ad2a61`
- 사유:
  - 기존 base 이후 병행 change의 Go 수정과 해당 Function Logic Map이 현재 HEAD에 병합되었다.
  - 이 change는 문서·Python SDD 검사만 수정하며 Go 제품 코드는 수정하지 않는다.
  - 오래된 base를 유지하면 이미 병합된 타 change의 Go 함수를 이 change의 수정으로 오인한다.
- 안전성:
  - rebase 전후 이 change의 대상 파일 diff를 보존한다.
  - 현재 HEAD 이후 수정된 기존 Go 함수가 생기면 gate는 계속 fail-closed한다.
