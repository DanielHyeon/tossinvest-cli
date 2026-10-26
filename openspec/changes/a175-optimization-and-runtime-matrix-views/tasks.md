# a175 tasks

Story: `STORY-TOS-a175` · Low(읽기 전용 렌더 재배치). FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`(proposal); 기존 Go 함수 편집은 0 을 목표(gate ⑤). 착수 조건: a173 착지(a174 와 병행 가능). 작업 위치: `/mnt/D/Axipient/workspace/TossOS.worktrees/a175-optimization-and-runtime-matrix-views`(형제 디렉터리, 사람 결정 2026-10-26). 기존 Go 함수 편집이 생기면 함수마다 gate ⑤ 번들이 필요하다(docs/WORKFLOW.md 비례 원칙) — 새 함수·새 파일 우선, 기존 시험 함수를 고치면 그것도 번들 대상.

## 0. 정의
- [x] 0.1 Story 등록 · change 정의 · spec 델타(ADDED 2) · proposal-freeze 경량 리뷰(`review.md`, 착수 시 base 기록)

## 1. 기준선·확인
- [x] 1.1 (Manager 몫) `base-commit.txt` 갱신 · design §1 재확인 — 증거: base `b7d34acad`(a174 archive, origin/main), design §1 표를 HEAD 에서 다시 잼(★ 표시 6곳 정정). 구현자 몫으로 남는 것:  `strategy-optimization`·`candidate-veto-thresholds`·`strategy-runtime` spec 의 표시 요구 목록을 `analysis/spec-display-requirements.md` 에 전수(값·사유·미측정·독립 projection)
  - 구현자 증거: `analysis/spec-display-requirements.md` — 세 spec의 값·사유·미측정·독립 projection과 접근성 경계를 원문 기준으로 목록화함.
- [x] 1.2 a173 하네스 기준선(`measure/baseline-<utc>.json`) · `check_analysis.py` required 0 확인
  - 구현자 증거: `analysis/measure/baseline-20261026T1019Z.json`(96 metrics), `baseline-20261026T1019Z-outline.json`(구조 count) — 빈 HOME/config·자격증명 0·127.0.0.1:38325 격리 콘솔에서 GET만 수행하고 PID `1467793` 종료·포트 해제를 확인함. `analysis/pre-edit-analysis.txt` — base→HEAD required 0.

## 2. RED
- [x] 2.1 `TestOptimizationHasOneNavigation`: 개요 렌더에 `optimization-nav` 1개, 카드 그리드 0, KPI 블록은 개요에만(다른 카테고리 렌더에서 0), 보조 링크 3개는 메뉴 안에 1회
  - 구현자 증거: `analysis/red-output.txt` — 메뉴 안 보조 링크 3개가 모두 0으로 실패함.
- [x] 2.2 `TestRelayCategoriesAreOneLineAndKeepTheirLinks`: 종목별 관리·전략·실행·성과·이력 렌더에 요약 `<p>` 1, 링크 수는 base 와 같음(1·2·1), 머리 KPI 0, 성과·이력의 감사 표는 Audit fixture 가 있을 때 존재
  - 구현자 증거: `analysis/red-output.txt` — 기존 relay 요약 문단이 3·2·3으로 실패함.
- [x] 2.3 `TestExitPresetsCompareInOneTable`: `table.preset-compare` 1개, 열 수 = `len(.Policies)`(레지스트리에서 셈), `/optimization/exit-policy` 프리셋 폼 = 열 수(열마다 1, hidden 5개·버튼 문구·selected 의 disabled 바이트 동일), 칸 값 = 각 정책 `Ladder.Rungs`, runner sentinel 칸에 `999` 문자열 0·"고정 목표 없음" 1, `ExitPolicyWritable=false` fixture 에서 폼 0·읽기 전용 문구 열마다 1
  - 구현자 증거: `analysis/red-output.txt` — preset comparison table 0개로 실패함.
- [x] 2.4 `TestCandidateFiltersRenderAsMatrix`: `table.veto-matrix` 1개(행 = 키 3 × 열 = 시장 2, 레지스트리에서 셈), 같은 값 행만 1회로 합치고 KR≠US fixture 에서는 시장별로 남음, 사유 코드·미측정 표기 전수 유지
  - 구현자 증거: `analysis/red-output.txt` — candidate filter matrix 0개로 실패함.
- [x] 2.5 `TestStrategyRuntimeIsTwoColumnsWithFoldedContract`: `.runtime-grid` 안 KR·US 블록 2, `details.tech` 안에 Schema·Golden 레지스트리 항목 전부 존재, 안전 상태 단어(UNWIRED·NOT_CONFIGURED·UNKNOWN) 는 접히지 않음, KR 만 실패한 fixture 에서 US 블록 값이 그대로 렌더됨(독립 projection)
  - 구현자 증거: `analysis/red-output.txt` — runtime grid 0개로 실패함.
- [x] 2.6 `TestSignalsCommonCopyOnceAndShadowBandsFolded`: 공통 문장 1회, 그림자 밴드 `details` 2(기본 닫힘), veto 표·후보 수는 `details` 밖
  - 구현자 증거: `analysis/red-output.txt` — 공통 문장·shadow details가 각각 0개로 실패함.

## 3. GREEN (design §2 순서) 
- [ ] 3.1 내비 하나·KPI 개요 한정 · 3.2 중계 3개 · 3.3 프리셋 비교표(+`TestAmbientMotionContract` 제외 목록 확장 기록) · 3.4 후보 필터 행렬 · 3.5 runtime 두 열 + 기술 정보 · 3.6 신호
- [ ] 3.7 `gofmt`·`vet`·`go test ./internal/console/ -count=1`·`cmd/tossctl -run Console`. 고정 리터럴 수정 전수 기록 · `check_analysis.py` required N 기록

## 4. 측정·리뷰
- [ ] 4.1 하네스 재촬영·전후 표(높이 5개 화면 + 잘림 측정) · 4.2 인지 부하 표(최적화 6 + runtime + 신호) 시각 "높음" 0 · 4.3 `design-critic`·`fixing-accessibility`(표 머리·스코프·링크) → `analysis/review-after.md`
- [ ] 4.4 gstack `review` 경량 → `review.md`. 변이 1회: 행렬에서 US 열 제거 → 2.4 FAIL 확인 후 원복
- [ ] 4.5 Manager 검증(Manager 가 체크 — 구현자는 체크하지 않음. 크롬: 최적화 6 카테고리·runtime·신호 × 데스크톱/모바일 × 라이트/다크)

## 5. 종결 (Opus)
- [ ] 5.1 rebase → FLM 번들(필요 시) → tracker → sdd-sync → gate → archive → push → retain

## 고정 리터럴 시험 수정 기록
(없으면 "없음")
