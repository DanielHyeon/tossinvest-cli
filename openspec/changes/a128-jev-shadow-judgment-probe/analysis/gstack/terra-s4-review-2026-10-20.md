# a128 s4(J3·J4 질문 추가) 경량 리뷰 — Terra, 2026-10-20

## 방식

- gstack `/review` 스킬을 불렀으나 preamble(`gstack-skill-start`)이 이 격리 worktree 의 실행 가드에 막힘(원문:
  "this command runs gstack-skill-start with a value computed at runtime (the variable PPID) inside a construct too complex to verify
  … Refusing to run it"). 그래서 같은 스킬의 `review/checklist.md` 범주(CRITICAL: SQL·동시성·LLM 출력 신뢰 경계·셸 주입·
  Enum 완전성 / INFORMATIONAL: 필드 이름·LLM 프롬프트·완전성·시간창·형 변환)로 **구조화한 자기 리뷰**를 함. Review Army·적대
  다중 리뷰·Codex: not-applicable(비례 원칙 — 측정 도구, High-risk 경로 무접촉, 게이트 판정 변경 0).
- 범위: 이 worktree 의 미커밋 diff(base `bbe4ceaa6`) + 새 파일 `tools/a128-jev-shadow-probe/s4_test.go`·`testdata/pre-s4/`·
  `analysis/function-logic/tools-a128-jev-shadow-probe--rendercalibrationreport/`.

Scope Check: CLEAN
Intent: J3·J4 를 같은 요청에 더하고(J1·J2 불변), 보고서에 J3·J4 적중 규칙과 개장 후 경과 시간대 표를 더함.
Delivered: `state.go` 질문 2 개(rev a128-r2), `report.go` 적중 규칙 2 개 + 시간대 표 + state 파일 읽기 전용 로더, `main.go` 배선 1 줄,
`ledger.go` 필드 1 개·주석, 시험(새 9 · 기존 행 수 갱신), OpenSpec design·tasks.

## 결과: 고칠 것 1(이 change 밖) · 고친 것 1 · 정보 5

| # | 등급 | 확신 | 위치 | 내용 | 처리 |
|---|---|---|---|---|---|
| 1 | P1(착지 선행 조건) | 9/10 | `tools/a145-jev-offline-analysis/analysis.py:123` · `tools/a147-jev-mfi-offline/analysis.py:175` | 원장을 읽는 다른 도구 둘이 한 state 의 행이 J1·J2 두 개뿐이라고 가정. a145: `if set(pair) != {J1,J2}: raise ValueError('unpaired questions')` → s4 행이 하나라도 있으면 **분석 전체가 멈춤**. a147: `valid_pair = … and len(rows) == 2 and …` → s4 의 모든 후보가 `invalid_pair` 로 **조용히 강등**. | **수리(같은 날, Manager 지시)** — 두 도구가 J3·J4 행·라벨을 짝 판정 전에 빼고 수를 `ignored_s4_question_rows` 로 남김. 실측으로 a145 는 84 행 `invalid judgment` 에서 먼저 멈췄음. 영수증·변이 5/5 는 tasks 5.5 |
| 2 | P3 | 9/10 | `ledger.go` `InputTokens` 주석 | 「한 요청이 J1·J2 두 행을 만들므로 2배」 — s4 부터 4 행이라 낡음 | 수리(AUTO-FIX) — 질문 수 배수와 요청 단위로 묶는 방법을 적음 |
| 3 | 정보 | — | design.md 「질문 추가 (s4)」 | J3·J4 를 같은 요청에 더해도 J1·J2 답이 그대로라는 근거는 계약 문서 인용뿐(독립 평가·key 미전송). 실측 아님 | tasks 5.7 에 s4 전후 J1·J2 p 분포 대조를 넣음 |
| 4 | 정보 | — | 비용 | 질문 원문만큼 입력 토큰이 늘고, state 가 질문마다 다시 세어지는지는 계약 범위에서 미기재 | tasks 5.6 에 첫 사이클 `input_tokens` 중앙값 대조를 넣음 |
| 5 | 정보 | — | `report.go` 시간대 표 | 질문 × 로트 × 시장마다 표 하나 — 라이브 원장(J1·J2 만)에서 18 개 · 보고서 712 줄. J3·J4 가 들어오면 질문 수만큼 늘어남 | 수용(로트·시장 분리는 기존 구간표와 같은 규칙) |
| 6 | 정보 | — | `check_analysis.py` | 이 change 에 `base-commit.txt` 가 없어 FLM 판정 도구가 「missing base-commit.txt」로 멈춤 | AST 산출물은 만듦(B19~B22). base 동결은 Manager 결정으로 남김 |
| 7 | 정보 | — | 측정 경계 | J1·J2 는 s4 전후가 한 로트(의도). s4 이후만 보려면 공유 트리 반영 시각 또는 첫 J3 행 `at` 이 경계 | tasks 5.5 에서 반영 시각을 적게 함 |

## 범주별 확인(증거)

- **J1·J2 불변**: `TestJ1J2AreByteIdenticalAndPinned` 가 digest 를 `sha256:a362ac64…49d10`·`sha256:ea5fc246…16737` 로 고정 — 공유 원장의
  `question_digest` 값 분포를 `grep -o | sort | uniq -c` 로 센 결과 이 두 값이 각 7,710 행(그 밖 값 0). 변이 「J1 원문 한 글자」 CAUGHT.
- **선별은 J1·J2 만**: `collect.go` 선별 입력은 `result.P[screenUpQuestion]`·`result.P[screenRedFlagQuestion]` 둘이고 `screenInput` 필드는 2 개
  (`TestScreeningIgnoresJ3J4` 가 필드 수 단언). J3·J4 를 J1 과 반대·같은 방향으로 준 두 사이클의 선별 기록이 같고 손계산 결선과 일치.
  변이 「J1 대신 J3」·「J2 대신 J4」 둘 다 CAUGHT(처음 앞의 것이 살아남아 시험을 고침 — tasks 5.3).
- **Enum 완전성(새 질문 id 의 소비자)**: 저장소 전체 `j1_up_within_horizon`/`j2_hold_off_red_flag` 검색 — a128 밖 소비자는 a145·a147 둘(1번).
  a128 안: `hitRules`(추가) · 시험 mock 기본 답(추가 — 없으면 응답 JSON 이 깨져 판단 gap) · `screen.go` 상수(의도적으로 J1·J2 만) ·
  `renderScreenComparison`(J1 만, 의도) · `assertPublicJSON` 금지 열쇠 정규식(새 id 둘 다 비통과 — 사이클 시험이 실제 전송으로 확인).
- **LLM 프롬프트**: J3·J4 원문은 J1 과 같은 틀(`exactly horizon_minutes minutes after observed_at`)이고 참 기준이 라벨 정의와 같은 말
  (≥ 1.01 배 / ≤ 0.99 배). 비 ASCII 곱셈 기호 대신 `times` 로 씀(영어 원문 유지). 요청 본문은 원문 리터럴과 바이트 대조
  (`TestS4QuestionsRideInTheSameRequest`).
- **경계·형 변환**: 라벨러가 수익률을 소수 8 자리로 반올림(`writeLabels` `round(…, 8)`) — `TestLabelerStoresAnExactOnePercentMoveAsTheBoundary`
  가 실제 `writeLabels` 로 3.00 → 3.03(나눗셈 원값 0.00999…)이 0.01 로 저장되어 J3 적중임을 확인. 경계 시험은 상수 접기를 피해 변수로 계산.
- **파일 읽기 안전**: `readStateMinutes` 는 state_digest 가 16진 64 자일 때만 경로를 만들고(원장 값으로 states/ 밖을 못 가리킴), 내용 sha256 이
  digest 와 같을 때만 값을 씀. 쓰기·이동 없음. 실패는 「알 수 없음」 줄로 셈 — 변이 「해시 검사 삭제」 CAUGHT.
- **동시성·SQL·셸**: 해당 없음(보고서는 단일 고루틴, DB·셸 호출 없음).
- **기존 출력 불변**: base 코드로 만든 골든(`testdata/pre-s4/report-base.golden.md`)이 새 출력의 접두로 바이트 그대로; 라이브 원장에서 base(overlay)/new
  두 출력 대조로 23,235 바이트 접두 일치, 덧붙은 제목은 전부 「개장 후 경과 시간대」.
- **주석**: 새 주석은 한글 명사형·「~함」 종결, 의도(왜) 위주.

## 남은 위험

- 1번(a145·a147)은 수리됨. 아카이브된 a145·a147·a153 기록 속 소스 sha256 은 수리 전 값으로 남음(시점 영수증 — 고치지 않음).
- J3·J4 의 기저율이 낮으면 구간 n≥30 이 늦게 참(5.6 은 시장당 ≥5 세션).
- J1·J2 독립성은 미실측(3번).
