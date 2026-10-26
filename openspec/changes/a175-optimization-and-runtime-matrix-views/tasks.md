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
- [x] 3.1 내비 하나·KPI 개요 한정 · 3.2 중계 3개 · 3.3 프리셋 비교표(+`TestAmbientMotionContract` 제외 목록 확장 기록) · 3.4 후보 필터 행렬 · 3.5 runtime 두 열 + 기술 정보 · 3.6 신호
  - 구현자 증거: `internal/console/{optimization_matrix_view.go,templates_optimization.go,templates_signals.go,templates.go}`와 `a175_optimization_runtime_matrix_test.go` — registry 읽기 전용 행렬, 1·2·1 중계 링크, 999 sentinel 치환, runtime 독립 열, Help·shadow 설명만 접힘을 렌더함. lifecycle 미확인은 `OFF`가 아닌 `확인 불가`로, candidate preview 계약은 접힘 밖에 표시함. `TestOptimizationShowsExactlySixCategoriesAndThreeOwnerPolicies`·`TestOptimizationShowsExactlyThreePoliciesAndExternalBehavior`의 기존 `약 50%` 계약도 비교표 머리에 유지함.
- [x] 3.7 `gofmt`·`vet`·`go test ./internal/console/ -count=1`·`cmd/tossctl -run Console`. 고정 리터럴 수정 전수 기록 · `check_analysis.py` required N 기록
  - 구현자 증거: `analysis/{gofmt-output.txt,vet-output.txt,green-console-output.txt,tossctl-console-output.txt,check-analysis-output.txt}` — 모두 exit=0. logic-map required 5이며 대상 FLM 함수는 `TestAmbientMotionContract`, `TestCandidateFilterCardsContainExactMarketMetricMatrix`, `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`, `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`, `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`임.

## 4. 측정·리뷰
- [x] 4.1 하네스 재촬영·전후 표(높이 5개 화면 + 잘림 측정)
  - 구현자 증거: `analysis/measure/{after-review-20261026T1130Z.json,after-review-20261026T1130Z-outline.json,after-review-20261026T1130Z-run.txt}`, `analysis/measurement-comparison.md` — 96 GET-only metrics, preset 잘림 0, PID `1944443` 종료·38329 포트 해제.
- [x] 4.2 인지 부하 표(최적화 6 + runtime + 신호) 시각 "높음" 0
  - 구현자 증거: `analysis/cognitive-load.md` — 8개 대상 중 "높음" 0.
- [x] 4.3 `design-critic`·`fixing-accessibility`(표 머리·스코프·링크) → `analysis/review-after.md`
  - 구현자 증거: `analysis/{review-after.md,accessibility-static-check.txt,impeccable-detect.json}` — scope/region/summary 복원, 720px runtime 1열, Help 설명 전용 details 확인.
- [x] 4.4 gstack `review` 경량 → `review.md`. 변이 1회: 행렬에서 US 열 제거 → 2.4 FAIL 확인 후 원복
  - 구현자 증거: `analysis/{mutation-us-column-red-output.txt,mutation-us-column-green-output.txt,review-red-output.txt,review-green-output.txt,review-contract-green-output.txt}`, `review.md` — US 열 제거 시 2.4 FAIL, 원복 뒤 `candidate-filters-US` 2회·header 1회 확인; 독립 경량 리뷰의 4개 critical을 RED→GREEN으로 해소함.
- [x] 4.6 Manager 수정 라운드: 시장 key 합집합·미등록 cell·descriptor 전체 비교·동일 상태 신호 재측정
  - 구현자 증거: `analysis/{p2-red-output.txt,p2-green-output.txt,p2-signals-red-output.txt,p2-signals-green-output.txt,p2-isolated-measurement.md,p2-impeccable-detect.json,p2-final-{gofmt,vet,console,tossctl-console,check-analysis}-output.txt}`, `review.md` — KR/US 전용 key가 first-seen 합집합으로 남고 없는 시장은 실제 렌더에서 명시됨. 신호는 `discovery_unreadable` 동일 상태로 900/941px, 중계 본문·내비 분리값을 재측정했음. check_analysis required 5: `TestAmbientMotionContract`, `TestCandidateFilterCardsContainExactMarketMetricMatrix`, `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`, `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`, `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`.
  - 라운드 2 구현자 증거: `analysis/{p3-red-output.txt,p3-green-output.txt,p3-isolated-measurement.md,p3-impeccable-detect.json,p3-final-{gofmt,vet,console,tossctl-console,check-analysis}-output.txt}` — 1280px preset region이 `879/879`으로 overflow 없이 맞고, 관련 화면은 제목 있는 목록이며 링크 수·href를 유지함. check_analysis required 5는 모두 complete임.
- [x] 4.5 Manager 검증(Manager 가 체크 — 구현자는 체크하지 않음. 크롬: 최적화 6 카테고리·runtime·신호 × 데스크톱/모바일 × 라이트/다크)
  - Manager 판정(Opus 5.5, 2026-10-26): 라운드 1(`293a19f`)·2(`efcd925`) diff 를 읽었다. 크롬 2차 실측(HEAD `efcd925`, 격리 38351, 1280 다크, 클릭 0)에서 프리셋 표 region 879/879 로 넘침 0, 폼 3, 관련 화면 소제목+평문 링크 3, nav 667. 콘솔 PID 종료·포트 해제를 확인했다. 직접 보지 않은 것: 신호 화면의 시장 데이터 상태(격리 콘솔에서 discovery_unreadable — fixture 시험으로 대신함), 390 라이트, `/strategy-runtime` 모바일. 완료 기준 조정은 proposal 에 적었다. 판정: 승인, 5.1 로 넘김.
  - Manager 실측(2026-10-26): 격리 콘솔 `127.0.0.1:38351`, HEAD `293a19f`, 빈 config·임시 HOME·클릭 0·콘솔 오류 0. PID `2355508` 종료와 포트 해제를 확인함. exit-protection 1280px는 KPI 0·요약 한 줄·열 4(단계+3)·폼 3·`999` 0·`고정 목표 없음` 1. candidate-filters는 1280px 높이 2083(본문 1533)·행 3·공통 3·잘림 0, 390px dark는 페이지 390·표 342/342·가로 스크롤 없음·펼침 표식 `list-item`. position-management 1280px는 페이지 1261·본문 386·nav 710·KPI 0·링크 1. `/strategy-runtime` 1280px는 KR·US 두 열(left 57/641)·계약 접힘·미기동 줄을 확인함.

## 5. 종결 (Opus)
- [ ] 5.1 rebase → FLM 번들(필요 시) → tracker → sdd-sync → gate → archive → push → retain
  - 사람 승인: 사용자가 a175 종결(Manager 편집 커밋·5.1 체크·tracker·`/tmp` 밖 격리 gate `/mnt/D/Axipient/workspace/TossOS.worktrees/a175-close`·전체 `make gate`·archive·ff·push(로컬 `526650e50` 포함)·정리·retain)을 "진행"으로 승인(2026-10-27, a175 한정). 격리 gate 를 /tmp 밖에 두는 것은 `/tmp/.git` 원인(Codex workspace-write 샌드박스) 조사의 결정 1. 운영 이미지·배포는 범위 밖.

## 고정 리터럴 시험 수정 기록
- `TestCandidateFilterCardsContainExactMarketMetricMatrix`: 시장별 카드 3개라는 이전 구조 문자열 검사를 삭제하고, 실제 `.veto-matrix` 하나의 `scope=col`·시장 순서 합집합 `scope=row`·registry key 1회씩을 검사하도록 변경함.
- `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`: `<table>` 0개라는 이전 구조 검사를 삭제하고, 실제 행렬의 scope·시장 합집합 행 수·시장 deep link·키보드 scroll region·native 읽기 전용 표·`dl`을 검사하도록 변경함.
- `TestAmbientMotionContract`: 사라진 `.choice-tile` 대신 POST form을 실제로 품는 `.preset-compare`를 glow 금지 대상에 넣음. 임의 속성 추가나 존재 전용 marker는 사용하지 않음.
- `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`: lifecycle/evidence marker를 비개요 exit 화면에 가짜로 되살리지 않고, 실제 marker를 렌더하는 overview route에서 검사하도록 변경함.
