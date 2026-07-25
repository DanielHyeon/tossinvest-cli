# Tasks: add-tossos-foundation

> 실행 주체 — [M]=Manager(총괄 아키텍트), [T]=구현 에이전트(Teammate)

## 1. Fork 셋업 [T]

- [x] 1.1 upstream 전체 히스토리 클론 (483 커밋, 75 태그, non-shallow)
- [x] 1.2 `git remote rename origin upstream` — origin 미설정
- [x] 1.3 HEAD 커밋 고정: `57348a7ffb234c98d6d0c9ee1d6ae3c9a5af2867`
- [x] 1.4 작업 브랜치 `feat/p0-foundation` 생성
- [x] 1.5 upstream push URL 차단 (`git remote set-url --push upstream DISABLED`) 및 확인
- [x] 1.6 LICENSE(MIT)·원저작권 고지 보존 확인

## 2. 베이스라인 기록 [T]

- [x] 2.1 `go build ./...` — PASS
- [x] 2.2 `go vet ./...` — PASS
- [x] 2.3 `go test ./... -count=1 -cover` — 650 통과 / 0 실패
- [x] 2.4 upstream 알려진 갭 원문 추출 (architecture.md 4건 + TODOS.md 2건)
- [x] 2.5 `docs/baseline.md` 작성 (198줄, 전체 커버리지 51.9% 실측 포함)

## 3. SDD 스캐폴딩 [M]

- [x] 3.1 `openspec init --tools claude` (openspec/ 트리 + .claude/ 스킬)
- [x] 3.2 `openspec new change add-tossos-foundation` + proposal/design/specs 작성
- [x] 3.3 `docs/ROADMAP.md` 확정 (Phase 0~6, 정찰 결과 반영)
- [x] 3.4 `docs/WORKFLOW.md` 확정 (역할 분리·리뷰 게이트·불변 규칙)
- [x] 3.5 `openspec validate add-tossos-foundation --strict` 통과
- [x] 3.6 gstack 리뷰 실행 (autoplan: codex + CEO/Eng/DX 4보이스, 발견 66건) — 기록: review.md
- [x] 3.7 리뷰 결정 반영본 재검증 (`openspec validate --strict` 재통과 확인)

## 4. 개발 도구 [T]

- [x] 4.1 Makefile 타겟 추가: `vet`, `cover`, `validate` (기존 타겟·ldflags 보존)
- [x] 4.2 `make build/test/vet/validate` 동작 검증 (전부 PASS)
- [x] 4.3 .gitignore 보강: `.claude/settings.local.json`, `*.db`, `*.sqlite`, `*.sqlite3`, `.env`, `.env.*`
- [x] 4.4 `git status`로 시크릿·로컬 파일 미추적 확인
- [x] 4.5 .gitignore sidecar 보강 (`*.db-*`, `*.sqlite-*`) — 리뷰 반영
- [x] 4.6 `tools/gate.sh` + `make gate CHANGE=<id>` 타겟 작성·검증 — 리뷰 반영

## 5. StockOS 인벤토리 [M]

- [x] 5.1 `docs/stockos-inventory.md` 작성 (UI·불변조건·순수 로직 경로와 이식 판정)

## 6. 발견성·스코프 정리 [M] — 리뷰 반영

- [x] 6.1 저장소 루트 `CLAUDE.md` 작성 (WORKFLOW.md 우선 규칙 포인터)
- [x] 6.2 `AGENTS.md` 상단 스코프 헤더 추가 (런타임 운용 vs 개발, mutating 규칙의 대체 조건)
- [x] 6.3 `openspec/project.md` 작성 (openspec 스킬 경유 작업도 게이트를 타도록)

## 7. 완료 게이트 [M]

- [x] 7.1 diff 전체 리뷰 — 기존 Go 코드 무변경 확인
- [x] 7.2 `go test ./...` 독립 재실행 green 재확인
- [x] 7.3 feat/p0-foundation에 분리 커밋 (SDD 문서 / 개발 도구)
- 7.4 (게이트 명령 자체) `make gate CHANGE=add-tossos-foundation` 통과 후 완료 선언
- 7.5 (사용자 확인 후) `openspec archive add-tossos-foundation` — review.md 미결정 사항 U1~U6 답변 뒤 실행
