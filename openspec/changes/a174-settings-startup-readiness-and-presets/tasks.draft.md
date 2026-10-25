# a174 tasks

Story: `STORY-TOS-a174` · Low~Medium(설정 폼 마크업 재배치, 핸들러·POST 불변). FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`(proposal) — 단 gate ⑤ 는 기존 Go 함수 본문을 고치면 함수마다 번들을 요구하므로 **새 함수·새 파일로 구현**하고 기존 함수 편집은 0 을 목표로 한다. 착수 조건: a173 착지. 작업 위치: `worktree/a174`(브랜치 `feat/a174-settings-startup-readiness-and-presets`, base = a173 착지 커밋).

## 0. 정의
- [x] 0.1 Story 등록(FEAT-TOS-004) · change 정의 · spec 델타(`specs/operator-console/spec.md` ADDED 3) · proposal-freeze 경량 리뷰(`review.md`, 착수 시 Manager 가 base 를 적는다)

## 1. 기준선·확인
- [ ] 1.1 `base-commit.txt` 를 a173 착지 커밋으로 갱신(Manager), design §1 의 file:line 재확인 결과를 design.md 에 표로 추가
- [ ] 1.2 a173 하네스로 기준선 촬영(`measure/baseline-<utc>.json`: 22화면 × 1280/390 × light/dark, 수치만) · `python3 tools/logic-map/check_analysis.py --change a174-…` 로 required 0 확인

## 2. RED
- [ ] 2.1 `TestStartupReadinessAppearsOnThreeScreensWithoutControls`: 상시·당일·검증 콘솔 렌더에 `section.readiness` 1개, 그 안 `<form`·`<input`·`<button` 0, 행 = 게이트·자동 시작·한도 5·거래 정책 4·판정 불가 2, 각 행에 소유 탭 링크, 검증 콘솔에서는 `/engine/start` 폼보다 **앞**에 위치
- [ ] 2.2 `TestGuardianPresetsRenderAsOneMatrixPerCurrency`: 당일 렌더에 `table.preset-matrix` 2개(KRW·USD), 열 = 현재 + 프리셋, 프리셋 열마다 `/settings/limits/preset` 폼 1개, FX 설명 통화당 1회, 1280 에서 표 폭 ≤ 본문 폭(하네스로 측정 보조)
- [ ] 2.3 `TestTradingPolicyRowsAppearOnce`: `.toggle-row` 4개, 각 행 안에 체크박스·배지·설명이 함께, 현재 상태 표 0, 필드 이름 집합 == {place, sell, cancel, allow_live_order_actions}
- [ ] 2.4 `TestSettingsFormsAreByteIdenticalAfterLayout`: 폼 action/method/필드명/hidden 값 집합이 a173 착지 렌더와 동일(기준 스냅숏을 시험 fixture 로 고정)
- [ ] 2.5 a173 계약 시험(상자 ≤3·원인 1회·식별자 0·경계 배지) 가 설정 세 탭에서 계속 통과하는지 먼저 확인(RED 가 아니라 회귀 가드)

## 3. GREEN (design §2 순서)
- [ ] 3.1 `readiness.go`(새 파일) + `readiness` define + 세 자리 배치 + 기존 인터록 표·중복 거부 상자 정리
- [ ] 3.2 프리셋 비교표(뷰 모델 새 함수 + 템플릿 + CSS `.preset-matrix`, 모바일 규칙)
- [ ] 3.3 거래 정책 한 줄 형식
- [ ] 3.4 외부 편입 설명 중복 정리(상시·`/position-management`)
- [ ] 3.5 `gofmt -l`·`go vet ./internal/console/`·`go test ./internal/console/ -count=1`·`go test ./cmd/tossctl/ -run Console -count=1`. 고정 리터럴을 바꾼 시험은 아래에 전수 기록. `check_analysis.py` 결과(required N) 기록

## 4. 측정·리뷰
- [ ] 4.1 하네스 재촬영 → `measure/after-<utc>.json`, 전후 표(`/settings/daily` 높이·상자·폼·입력, 세 화면 체크리스트 유무)
- [ ] 4.2 `cognitive-load-assessment` 설정 3탭 + 검증 콘솔 표(`analysis/cognitive-load-after.md`), 결정·기억 "높음" 0
- [ ] 4.3 `design-critic` 재점수 + `fixing-accessibility`(체크박스 라벨·표 머리·링크 텍스트) 점검 → `analysis/review-after.md`
- [ ] 4.4 gstack `review`(경량) → `review.md` 「구현 리뷰」. 변이 1회: `readiness` 행 생성에서 거래 정책 4행 제거 → 2.1 FAIL 확인 후 원복(기록)
- [ ] 4.5 Manager(Fable) 검증 — 크롬으로 상시·당일·검증 콘솔 × 데스크톱/모바일 × 라이트/다크 직접 확인, 버튼 0

## 5. 종결 (Manager 승인 뒤, Opus)
- [ ] 5.1 rebase → FLM 번들(required 가 0 이 아니면) → tracker → sdd-sync → `make gate` → archive → push → retain

## 고정 리터럴 시험 수정 기록
(없으면 "없음")
