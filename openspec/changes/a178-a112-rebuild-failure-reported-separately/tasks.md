# a178 tasks

Story: `STORY-TOS-a178` · Low(측정 도구 진단 문구, 판정 불변). FLM/BTM 적용(기존 함수 `verifyPrescribedBinary` 편집 → 번들, 편집 전). 변이 원장·다중 적대 리뷰는 비례 원칙. 작업 위치: `/mnt/D/Axipient/workspace/TossOS.worktrees/a178-a112-rebuild-failure-reported-separately`(형제 디렉터리). 구현자: GPT-5.6 Terra.

## 0. 정의
- [x] 0.1 Story 등록(FEAT 는 a112 와 같음) · change 정의 · spec 델타(ADDED 1) · proposal-freeze 경량 리뷰(`review.md`)

## 1. 확인
- [x] 1.1 design §1 재확인 · `fakeBuildRunner` 가 명령 문자열을 바꿀 수 있는지 · HOLD 오류 원문이 receipt 에 쓰이는지 확인해 design 에 기록 · `check_analysis.py --change a178-…` 로 편집 전 required 0 확인

## 2. RED
- [x] 2.1 `TestVerifyPrescribedBinaryReportsRebuildFailureSeparately` · 2.2 `…ReportsUnexpectedCommandSeparately` · 2.3 기존 hash-mismatch 시험에 문구 단언 추가 — FAIL 원문을 `analysis/red-output.txt` 에

## 3. GREEN
- [x] 3.1 **편집 전에** `verifyPrescribedBinary` FLM 번들(TODO 0) → B6 분리(design §2) → 3.2 `docs/WORKFLOW.md` 함정 절에 재시도 관례 1줄 → 3.3 gofmt·vet·`go test ./tools/a112-mb-us-source/ -count=1`·check_analysis(required N 기록)

## 4. 리뷰
- [ ] 4.1 변이 1회: rebuild 갈래 `%w`→`%v` → 2.1 FAIL 확인 후 원복(심볼 수로 확인) · 4.2 gstack review 경량 → `review.md` · 4.3 Manager 검증(diff·시험 원문)

## 5. 종결 (Opus)
- [ ] 5.1 FLM 번들 · tracker · sdd-sync · `/tmp` 밖 격리 worktree 에서 워밍업(`codegraph status .`) 뒤 `make gate` · archive · push · retain

## 고정 리터럴 시험 수정 기록
(없으면 "없음")
