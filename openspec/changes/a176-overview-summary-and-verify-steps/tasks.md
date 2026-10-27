# a176 tasks

Story: `STORY-TOS-a176` · Low~Medium(`/verify` 는 승인 화면 — 폼 불변). FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`(proposal); 기존 Go 함수 편집 0 목표. 착수 조건: a173 착지. 작업 위치: `/mnt/D/Axipient/workspace/TossOS.worktrees/a176-overview-summary-and-verify-steps`(형제 디렉터리). 기존 Go 함수 편집이 생기면 함수마다 gate ⑤ 번들이 필요하다 — 새 파일·새 메서드 우선.

## 0. 정의
- [x] 0.1 Story 등록 · change 정의 · spec 델타(ADDED 2) · proposal-freeze 경량 리뷰(`review.md`, 착수 시 base 기록)

## 1. 기준선·확인
- [x] 1.1 (Manager 몫) `base-commit.txt` 갱신 · design §1 재확인 · 단계 데이터 형태 확정 — 증거: base `0e8ba1c7c`(a175 archive, origin/main), 단계 데이터는 `WriteSteps` 문자열(`pages.go:180-188`), design §1·§2.2 ★ 정정. 구현자 몫으로 남는 것: 개요·verify 렌더로 구조적 미측정 설명 횟수와 높이를 다시 센다(1.2 기준선과 함께)
- [x] 1.2 a173 하네스 기준선 · `check_analysis.py` required 0 확인
  - 증거(Terra, 2026-10-27): `analysis/measure/baseline-20261027T001347Z.json` — 빈 config-dir·임시 HOME·자격증명 0·127.0.0.1:38324 격리 콘솔에서 GET만 96화면 촬영. 대상 기준선: `/dashboard` 1280 라이트 4453px·다크 4697px, `/verify` 접힌 전 5123px(1280)·4859px(390), `/verify` form 1·input 3·button 1. PID 1254588 종료와 포트 해제를 확인함. `analysis/pre-edit-analysis.txt`의 `check_analysis.py`는 required 0을 출력함.

## 2. RED
- [x] 2.1 `TestOverviewSummaryAnswersSixFactsAboveTheFold`: 요약 영역에 여섯 사실(①~⑥) 라벨과 각 상세 링크, 시장별 분리(KR·US 각각), 미측정 칸마다 사유 코드, 요약 영역 하단 ≤ 900px(하네스 측정으로 보조 — Go 시험은 구조만)
  - 증거: `internal/console/a176_overview_verify_test.go`; RED 원문 `analysis/red-output.txt`, GREEN은 3.4 범위 시험.
- [x] 2.2 `TestStructuralUnmeasuredIsExplainedOnce`(a173 의 같은 이름 시험이 있으면 확장): 소진분 축 설명 1회
  - 증거: 같은 새 시험이 구조적 미측정 설명 문장을 정확히 1회로 고정함.
- [x] 2.3 `TestVerifyStepsAreFoldableAndListTextIsPreserved`: `details.step` 수 == `len(verifylive.Steps())`, 각 summary 가 그 단계 ID 로 시작, 머리말+조각을 이으면 `WriteSteps(false,false)` 출력과 바이트 동일, 경계를 하나 지운 가짜 문자열에서는 `StepItems()==nil` 이고 종전 `<pre>` 전체가 렌더됨(fail-safe)
  - 증거: 새 `verifyPage.StepItems`; 본문 `<pre>` 제거 변이 FAIL `analysis/mutation-output.txt`, 원복 후 PASS·심볼 수 1.
- [x] 2.4 `TestVerifyFormsAreByteIdentical`: `/verify/start`·`approve`·`abort` 폼 마크업·CSRF·확인 문자열 블록이 변경 전 렌더와 동일(기준 fixture 고정)
  - 증거: `internal/console/testdata/a176_verify_forms.golden.html`에 승인 프롬프트·세 폼을 고정하고 새 시험에서 바이트 비교함.

## 3. GREEN
- [x] 3.1 개요 요약 세 묶음(템플릿 + 새 템플릿 함수, 패널 함수 본문 편집 0) · 3.2 소진분 설명 1회 · 3.3 `/verify` 단계 `details` 구조화
  - 증거: `templates_overview.go` 세 묶음과 상단 배치, `verify_steps.go` 무손실 분할, `templates.go` fail-safe 렌더. 기존 패널 함수·verify 핸들러·verifylive 편집 0.
- [x] 3.4 `gofmt`·`vet`·`go test ./internal/console/ -count=1`·`cmd/tossctl -run 'Console|Verify'`. 고정 리터럴 수정 전수 기록 · `check_analysis.py` required N 기록
  - 증거: `gofmt -l internal/console` 빈 출력, `go vet ./internal/console/` PASS, `go test ./internal/console/ -count=1` PASS, `go test ./cmd/tossctl/ -run 'Console|Verify' -count=1` PASS. `check_analysis.py` required 0, 함수 목록 없음(새 leaf만 추가). 고정 리터럴 수정 없음; fixture로 기존 리터럴을 비교 고정함.

## 4. 측정·리뷰
- [x] 4.1 하네스 재촬영·전후 표 · 4.2 인지 부하 표(개요·verify) 읽기·시각 "높음" 0 · 4.3 `design-critic`·`fixing-accessibility`(details/summary 키보드·44px) → `analysis/review-after.md`
  - 증거: `analysis/measure/baseline-20261027T001347Z.json`, `analysis/measure/after-20261027T0030Z.json`, `analysis/cognitive-load-after.md`, `analysis/review-after.md`. 개요 데스크톱 요약 하단 624px, 가로 넘침 0, 단계 summary 44px(모바일 66px). `/verify` 접힌 전체 2047px로 1800px 목표는 미달하며 은닉·간격 압축은 하지 않음.
- [x] 4.4 gstack `review` 경량 → `review.md`. 변이 1회: 단계 `details` 에서 `<pre>` 제거 → 2.3 FAIL 확인 후 원복
  - 증거: `review.md` 구현 후 경량 리뷰. 자동 gstack CLI는 사용할 수 없어 설치된 checklist 수동 적용(발견 0)으로 대체함. 변이 원문 `analysis/mutation-output.txt`, 원복 후 심볼 수 1·시험 PASS.
- [x] 4.6 Manager 수정 라운드 1: 검증 보류 표식·소진분 근거·원인 배너 순서 복구, 정상/장애 실측
  - 증거: `TestOverviewSummaryKeepsHoldAndConsumptionEvidence` RED `analysis/review-round1-red-output.txt` → GREEN. 정상 fake-console(원인 배너 0) 요약 하단 644px(1280)·1092px(390), 실제 빈 config-dir 격리 콘솔 장애 상태(원인 배너 1) 1036px(1280)·1581px(390), 가로 넘침 0px — `analysis/measure/review-round1-normal.json`, `analysis/measure/review-round1-isolated-empty.json`. 장애 데스크톱은 900px을 136px 넘지만 여백·내용 압축은 하지 않음. 단계 `details`·`/verify` 폼은 변경하지 않음.
- [ ] 4.5 Manager 검증(Manager 가 체크 — 구현자는 체크하지 않음. 크롬: 개요·verify × 데스크톱/모바일 × 라이트/다크, 승인 버튼 0)

## 5. 종결 (Opus)
- [ ] 5.1 rebase → FLM 번들(필요 시) → tracker → sdd-sync → gate → archive → push → retain

## 고정 리터럴 시험 수정 기록
(없으면 "없음")
