# a205 tasks

> 사용자 지시 2026-11-27(두 시장 모두 5s → 2s, 개발 · 시험 · 리뷰 · 커밋 · push 까지). 배포 · `make image` · 엔진 교체 · 운영 토글 · main 병합 · archive 는
> 사람 승인 별도 — 이 change 의 작업 브랜치에서 하지 않음. High-risk(손절 · 익절 관측 경로).

## 1. 계약 · 증거

- [x] 1.1 Story `STORY-TOS-a205` → FEAT-TOS-009, registry · PM generated 갱신, `--check` rc 0
- [x] 1.2 proposal · design(D1~D6) · exit-policy spec 델타 · `openspec validate a205-exit-observation-2s --strict` rc 0
- [x] 1.3 `base-commit.txt`(`e2921fc00`) 고정
- [x] 1.4 CodeGraph 1.6.0 hard evidence · CodeGraphContext not-applicable 사유 · evidence reconciliation(`analysis/code-context/`)
- [x] 1.5 분기 주장 AST 근거 7개(`analysis/ast-evidence/claims.md` C1~C4)
- [x] 1.6 편집할 기존 시험 함수 3개(r1 #1 뒤 1개 추가)의 FLM/BTM 번들(`analysis/function-logic/`), Pre-Edit 선언(`review.md`)

## 2. RED → GREEN

- [x] 2.1 사전 탐침: 상수만 2s 로 바꾼 엔진 패키지 시험 rc 0(406.7s) — 엔진 시험 함수 편집 불필요 확인 뒤 상수 되돌림
- [x] 2.2 RED: `internal/app/engine/a205_exit_observation_interval_test.go`(기본 2s · 옵션 없는 관측자 2s · 하한가 예산 읽기 1s · 사전조회 1s) ·
      콘솔 두 시험 기대값 2s(+ r1 #1: 개요 재로드 시험 횟수 10 → 16) — 실패 출력 `analysis/verify/red.txt`
- [x] 2.3 GREEN: `DefaultExitObservationInterval` · `lineRefreshInterval` · `engineRuntimeInterval` 2s, 주석(outage 「four intervals」 · a197 시한 근거) — 출력 `analysis/verify/green.txt`
- [x] 2.4 FLM 번들 갱신(GREEN 뒤 AST 재추출) · `check_analysis.py --change a205-exit-observation-2s` 결과 기록

## 3. 검증 · 리뷰

- [x] 3.1 대상 시험: `go test ./internal/app/engine/... ./internal/console/...`, `cmd/tossctl`, `make lint`, 엔진 · 콘솔 태그 시험(`make test-seams` 범위 중), `make test-race` 의 엔진 목록 — 원문 `analysis/verify/`
- [x] 3.2 gstack review(적대적 Eng 관점, 작성자와 다른 계열 — Codex) 최대 2회차, 지적 · 처분 `review.md`(새 가드 추가 금지)
- [x] 3.3 커밋 전 `git diff --cached --name-only` 전수 확인, 브랜치 push

## 4. 사람 승인 뒤 (이 change 범위 밖에서 사람이 시작)

- [ ] 4.1 `make image CHANGE=a205-exit-observation-2s` · 닫힌 창에서 엔진 교체(design 「배포」)
- [ ] 4.2 교체 뒤 `exit_states.last_observed_at` 간격이 ≈2.1s 인지 읽기 전용으로 확인, journal.db 성장 관찰(R3)
- [ ] 4.3 main 병합 · archive
