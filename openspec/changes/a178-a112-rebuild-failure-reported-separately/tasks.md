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
- [x] 4.1 변이 1회: rebuild 갈래 `%w`→`%v` → 2.1 FAIL 확인 후 원복(심볼 수로 확인)
- [x] 4.2 gstack review 경량 → `review.md`
- [x] 4.3 Manager 검증(diff·시험 원문)
  - Manager 판정(Opus 5.5, 2026-10-27): `4b868a18d..65c5f183d` diff 를 읽었다. 범위는 `tools/a112-mb-us-source/`·a178 문서·`docs/WORKFLOW.md` 뿐이다. `identity.go` 는 B6 하나만 순서 있는 세 검사로 나눴고, 판정(셋 중 하나면 오류)·성공 경로·120초 시한·HOLD 조건은 그대로다. 시험 세 갈래(DeadlineExceeded 감싸기·명령 불일치·digest 문구)와 가짜 builder 확장(기본값 불변)도 맞다. WORKFLOW 줄은 리뷰 뒤 「완료 게이트 (자동화)」 절로 옮겼다(`65c5f183d`, +2줄). 시험 원문 `analysis/verification-output.txt` 를 확인했다. Manager 는 시험을 다시 돌리지 않았다. 판정: 승인, 5.1 로 넘김.

## 5. 종결 (Opus)
- [ ] 5.1 FLM 번들 · tracker · sdd-sync · `/tmp` 밖 격리 worktree 에서 워밍업(`codegraph status .`) 뒤 `make gate` · archive · push · retain
  - 사람 승인: 사용자가 a178·a176 종결(Manager 편집 커밋·5.1 체크·tracker·`/tmp` 밖 격리 worktree 워밍업 뒤 전체 `make gate`·archive·ff·push(로컬 `77d82b327`·`4b868a18d` 포함)·정리·retain, 재시도는 a112 e2e `prescribed rebuild failed … deadline exceeded` 또는 ⑥ CodeGraph 15초 시한 초과일 때만 change 당 1회)을 "진행"으로 승인(2026-10-27, 두 change 한정). 운영 이미지·배포는 범위 밖.

## 고정 리터럴 시험 수정 기록
(없으면 "없음")
