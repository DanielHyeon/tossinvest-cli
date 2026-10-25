# a175 tasks

Story: `STORY-TOS-a175` · Low(읽기 전용 렌더 재배치). FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`(proposal); 기존 Go 함수 편집은 0 을 목표(gate ⑤). 착수 조건: a173 착지(a174 와 병행 가능). 작업 위치: `worktree/a175`.

## 0. 정의
- [x] 0.1 Story 등록 · change 정의 · spec 델타(ADDED 2) · proposal-freeze 경량 리뷰(`review.md`, 착수 시 base 기록)

## 1. 기준선·확인
- [ ] 1.1 `base-commit.txt` 갱신(Manager) · design §1 재확인 · `strategy-optimization`·`candidate-veto-thresholds`·`strategy-runtime` spec 의 표시 요구 목록을 `analysis/spec-display-requirements.md` 에 전수(값·사유·미측정·독립 projection)
- [ ] 1.2 a173 하네스 기준선(`measure/baseline-<utc>.json`) · `check_analysis.py` required 0 확인

## 2. RED
- [ ] 2.1 `TestOptimizationHasOneNavigation`: 개요 렌더에 `optimization-nav` 1개, 카드 그리드 0, KPI 블록은 개요에만(다른 카테고리 렌더에서 0), 보조 링크 3개는 메뉴 안에 1회
- [ ] 2.2 `TestRelayCategoriesAreOneLineAndOneLink`: 종목별 관리·전략 실행·성과 이력 렌더에 요약 `<p>` 1 + 링크 1, KPI 0
- [ ] 2.3 `TestExitPresetsCompareInOneTable`: `table.preset-compare` 1개, 열 3, `/optimization/exit-policy` 폼 3(열마다 1), 폼 action/필드/CSRF 집합 전후 동일
- [ ] 2.4 `TestCandidateFiltersRenderAsMatrix`: `table.veto-matrix` 1개(행 3 × 열 2), 공통 설명 각 1회, 사유 코드·미측정 표기 전수 유지
- [ ] 2.5 `TestStrategyRuntimeIsTwoColumnsWithFoldedContract`: `.runtime-grid` 안 KR·US 블록 2, `details.tech` 안에 Schema·Golden 레지스트리 항목 전부 존재, 안전 상태 단어(UNWIRED·NOT_CONFIGURED·UNKNOWN) 는 접히지 않음
- [ ] 2.6 `TestSignalsCommonCopyOnceAndShadowBandsFolded`: 공통 문장 1회, 그림자 밴드 `details` 2(기본 닫힘), veto 표·후보 수는 `details` 밖

## 3. GREEN (design §2 순서) 
- [ ] 3.1 내비 하나·KPI 개요 한정 · 3.2 중계 3개 · 3.3 프리셋 비교표(+`TestAmbientMotionContract` 제외 목록 확장 기록) · 3.4 후보 필터 행렬 · 3.5 runtime 두 열 + 기술 정보 · 3.6 신호
- [ ] 3.7 `gofmt`·`vet`·`go test ./internal/console/ -count=1`·`cmd/tossctl -run Console`. 고정 리터럴 수정 전수 기록 · `check_analysis.py` required N 기록

## 4. 측정·리뷰
- [ ] 4.1 하네스 재촬영·전후 표(높이 5개 화면 + 잘림 측정) · 4.2 인지 부하 표(최적화 6 + runtime + 신호) 시각 "높음" 0 · 4.3 `design-critic`·`fixing-accessibility`(표 머리·스코프·링크) → `analysis/review-after.md`
- [ ] 4.4 gstack `review` 경량 → `review.md`. 변이 1회: 행렬에서 US 열 제거 → 2.4 FAIL 확인 후 원복
- [ ] 4.5 Manager 검증(크롬: 최적화 6 카테고리·runtime·신호 × 데스크톱/모바일 × 라이트/다크)

## 5. 종결 (Opus)
- [ ] 5.1 rebase → FLM 번들(필요 시) → tracker → sdd-sync → gate → archive → push → retain

## 고정 리터럴 시험 수정 기록
(없으면 "없음")
