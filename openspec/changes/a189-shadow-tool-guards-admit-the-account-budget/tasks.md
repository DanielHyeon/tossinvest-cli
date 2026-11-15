## 1. 계약과 증거 (1단계 — `.go` 무편집)

- [x] 1.1 memory recall(ai-memory·Hindsight)과 번호 재확인, worktree·브랜치 생성. 기록은 `review.md`.
- [x] 1.2 Story `STORY-TOS-a189`(FEAT-TOS-002) 선등록, change 문서 작성, `capture_change_base.py` 로 base 고정, PM tracker 재생성.
- [x] 1.3 CodeGraph(1.6.0) 호출자·`affected --filter '*_test.go'` 원문을 `analysis/code-context/` 에 남김.
- [x] 1.4 D2 열거(`analysis/enumeration/`), D5 영향 원문, 편집 전 AST 5+1개(`analysis/pre-edit-ast/`)를 만듦.
- [x] 1.5 `openspec validate --strict`·PM `--check` 원문을 `analysis/verify-1/` 에 남김.
- [ ] 1.6 Manager 결정: D4 선택지(A/B/A-wide)와 A 일 때 스펙 처리(A-1 해석 / A-2 개정). A-2 이면 `analysis/option-a-mb0-delta-draft.md` 를 specs/ 로 옮기고 strict validate 를 다시 돎.

## 2. 구현 (RED → GREEN)

- [ ] 2.1 RED: 편집 전 다섯 시험이 base 와 같은 문구로 실패함을 이 worktree 에서 원문으로 남김(5개 패키지 시험 지정 실행).
- [ ] 2.2 GREEN: proposal 「What Changes」의 이름만 더함. 다른 함수 본문은 고치지 않음. 기존 문구·금지 목록 무편집.
- [ ] 2.3 편집 뒤 FLM/BTM 번들(함수 5개, 선택지 B 면 그에 맞게)을 만들고 `check_analysis.py --change a189-shadow-tool-guards-admit-the-account-budget` 를 통과시킴.

## 3. 검증

- [ ] 3.1 사본 변이(무변이 대조군 GREEN 먼저, 사본에 pid, 한 번에 한 판, 사본 대상 단언 — design D7 목록). 각 변이가 어느 시험의 어느 줄에 잡히는지 원문을 `analysis/mutation/` 에 남김.
- [ ] 3.2 영향 패키지 시험: `internal/official`, `internal/accountbudgetruntime`, `tools/a128-jev-shadow-probe`, `tools/a140-breakout-signal-shadow`, `tools/a112-mb-us-source`(측정 소스 신원 집합에 가드 파일이 있음 — design D4). `go vet`·gofmt. 전체 gate·전체 시험은 사용자가 요청할 때만.
- [ ] 3.3 gstack `review` 1회(다중 적대 리뷰는 not-applicable — 비례 원칙), 처리 결과를 `review.md` 에 기록.
- [ ] 3.4 Manager 독립 재검증.

## 4. 종결

- [ ] 4.1 PM check, strict validate, archive 와 Story 경로 동기화. 검증된 교훈만 retain.
