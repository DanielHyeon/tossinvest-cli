# a175 — 최적화·전략 화면: 내비 중복 제거 · 중계 페이지 축소 · 행렬화 (정의, 미착수)

프로그램 맥락: `openspec/changes/a173-console-ux-common-cleanup/proposal.md` 「프로그램 개요」 S3. 착수 조건: a173 착지 뒤(a174 와 병행 가능).

## Why
2026-10-25 실측:
- `/optimization`: 내비 셋(상단 nav·보조 링크 줄·왼쪽 카테고리), 개요 카드 6개 = 왼쪽 메뉴 중복, KPI 6칸 히어로가 모든 카테고리에서 다시 그려짐. 하위 라벨에 `a044`·`a046`·`a047`·`a049`.
- `?category=exit-protection`: 프리셋 카드 3장의 표가 오른쪽에서 잘려 비교 불가. 계산기 입력 10개.
- `?category=candidate-filters`: 같은 11줄 블록 6개(3932px), 값 거의 동일(미승인·`not_measured`).
- 중계 카테고리 3개(종목별 관리·전략 실행·성과 이력): 링크 하나를 위해 1186px.
- `/strategy-runtime`: KR·US 15줄씩 세로 적층, `not_observed` 59회, 개발자 계약(Shared projection·Golden field registry) 노출.
- `/signals`: KR·US 동일 구조 2회, "기록만 하고 판정하지 않는다" 6회, veto 아닌 그림자 밴드 펼침.

## What Changes
1. 개요 카드 그리드 삭제(왼쪽 메뉴 하나), KPI 6칸은 개요 카테고리에만.
2. 중계 카테고리는 요약 1줄 + 대상 화면 링크(spec 851 진입점 규칙 준용: 재계산 0, 대상 화면 값 옮김).
3. 익절 프리셋 → `단계 T1~T4 × 프리셋` 비교표 1개, 잘림 0, 선택 버튼은 그대로 프리셋당 1개.
4. 후보 필터 → 행=사유 코드 × 열=KR/US 행렬, 공통 행 1회.
5. `/strategy-runtime` KR/US 두 열, 계약·레지스트리는 접힌 기술 정보.
6. `/signals` 공통 설명 1회, 그림자 밴드 기본 접힘(수치는 접지 않음).

## Scope / Non-goals
- 범위: `templates_optimization.go`·`templates_signals.go`·strategy-runtime 템플릿·CSS·시험.
- 비목표: 카테고리 deep link(`?category=`, a050 계약)·POST(`/optimization/exit-policy`·`hypothetical`)·preview→apply 2단계·값·핸들러. 손절 비교 계산기의 입력 수(기능).

## 스펙 제약
`strategy-optimization`·`candidate-veto-thresholds`·`strategy-runtime` spec 을 구현 전에 읽고 표시 요구(값·사유·미측정 표기)를 전수 확인한다. 접기는 설명에만, 상태값·차단 사유는 접지 않는다.

## Risk
Low. 읽기 전용 렌더 재배치. 프리셋 선택 폼의 action/CSRF 전후 동일 단언.

## 완료 기준
Story acceptance(`STORY-TOS-a175`): 중계(종목별 관리·전략·실행) ≤ 700px — 성과·이력은 감사 표를 유지하므로 측정값만 기록, 후보 필터 ≤ 1200px, 프리셋 표 잘림 0(1280·390), 시각 항목 "높음" 0.


완료 기준 조정(Manager 4.5, 2026-10-26):
- 중계 700px 은 **본문(`.optimization-content`) 높이** 기준으로 바꾼다. 페이지 높이는 왼쪽 카테고리 메뉴(667px)가 정하므로 본문을 아무리 줄여도 페이지는 메뉴보다 낮아지지 않는다. 실측 본문 386px 로 충족.
- 후보 필터 1200px → **2100px**. 3950→2083px(−47%). 남은 높이는 접을 수 없는 값·사유·미측정·CAS 표시라 여백을 눌러 맞추지 않았다(a174 와 같은 판단).
## not-applicable
변이 원장·다중 적대 리뷰: 비례 원칙(읽기 전용 렌더). FLM/BTM 은 면제를 미리 주장하지 않는다: 기존 Go 함수 본문(시험 함수 포함, 예: hover 금지 목록 확장)을 고치면 gate ⑤ 가 함수마다 번들을 요구한다(docs/WORKFLOW.md 비례 원칙, 2026-10-26 문구). 새 함수·새 파일을 우선하고, 생긴 required 목록은 tasks 3.7 에 적는다.

Manager 착수 전 정정(2026-10-26): design §1 ★ 표시 — 프리셋은 `.choice-tile`(`.tier-card` 아님), 중계 카테고리의 링크 수 1·2·1 과 성과·이력 감사 표 유지, runner 999 sentinel 표시 모순, 후보 필터는 KR·US 가 같은 값일 때만 공통 행으로 합침.
