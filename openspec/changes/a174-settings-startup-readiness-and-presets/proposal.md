# a174 — 설정: 기동 준비 체크리스트 · Guardian 프리셋 비교표 · 거래 정책 한 줄 형식 (정의, 미착수)

프로그램 맥락과 사람 결정: `openspec/changes/a173-console-ux-common-cleanup/proposal.md` 「프로그램 개요」 S2. 착수 조건: a173 착지 뒤.

## Why
2026-10-25 실측: `/settings/daily` 5008px·폼 7·입력 24. 프리셋 5장이 각각 7줄 표 + 거의 같은 FX 박스 + 버튼. 거래 정책 토글 4개가 현재 상태 표와 체크박스로 두 번. 체크박스 설명이 다음 체크박스에 붙어 보임(안전 오독). `/settings/standing` 의 "지금 켜면 기동이 거부된다" 는 조건(한도 5개·거래 정책)이 다른 탭에 있어 운영자가 기억해야 한다. 인지 부하 표에서 결정·기억·시각 세 항목이 "높음".

## 스펙 제약 (사람 결정 1·2 의 근거)
`operator-console` 851~880: 탭은 가역성·빈도(상시/당일/전략/도구) SHALL, 기본 진입 당일 SHALL, 컨트롤은 한 화면에만 SHALL NOT(복제 금지), 다른 탭 값의 **읽기 전용 표시 + 링크는 요구됨**, 머리에 운영 상태 요약 SHALL, 저장 의미·CSRF·audit 무변경 SHALL NOT. 634·662: 검증 콘솔의 엔진 시작·정지 컨트롤 유지 SHALL NOT(줄이지 말 것).
→ 탭 축·기본 진입·버튼 위치는 바꾸지 않는다. 기동 준비 체크리스트는 **읽기 전용 표시 + 링크** 로 상시·당일 머리와 검증 콘솔 엔진 버튼 위에 둔다(컨트롤 복제 0).

## What Changes
1. 기동 준비 체크리스트 카드(읽기 전용): 자동화 게이트·엔진 자동 시작·Guardian 한도 다섯과 통화(절 1개)·거래 정책(절 1개, 꺼진 토글 이름 포함)·판정 불가 절 2개의 현재 상태 + 각 컨트롤이 있는 탭 링크. 합산 판정("기동 가능")은 만들지 않는다. 상시·당일 머리(spec 851 머리 요약 요구를 확장)와 검증 콘솔 엔진 버튼 위. 기존 "기동 인터록 사전 판정" 표의 계산 값을 옮기고 재계산하지 않는다.
2. Guardian 프리셋 비교표: 통화별 표 하나(행 = 한도 다섯 + 통화, 열 = 현재 | `config.GuardianTiers()` 의 그 통화 프리셋들), 열마다 기존 POST 폼(hidden `tier`) 1개, 통화 결과 설명은 통화당 1회. 고급(개별 값 입력) 폼은 접힌 채 유지.
3. 거래 정책: 줄마다 `체크박스 · 라벨 · 현재 배지 · 한 줄 설명` 1회. 현재 상태 표 제거(값은 배지로 이동).
4. 외부 편입 설정: 상시 탭이 편집, `/position-management` 는 읽기 전용 표 + 링크(spec 허용)로 설명 중복만 줄인다.

## Scope / Non-goals
- 범위: 새 파일 `readiness.go`(뷰 모델·seam 읽기)·`templates_settings.go`·`templates.go`(`dashboard` 엔진 절의 체크리스트 자리 + `style` 끝 a174 블록)·`templates_position_policy.go`(설명 문단)·`pages.go`(`dashboardPage` 필드 1 + `handleDashboard` 한 줄)·시험.
- 비목표: 탭 축·기본 진입·POST 경로·필드명·CSRF·audit·저장 뒤 돌아갈 탭 매핑·확인 단계(추가 금지, 사용자 지시)·엔진 버튼 위치·한도 값·프리셋 값.

## Risk
Low~Medium. 토글·한도 폼 마크업을 재배치한다. 폼 action/필드/CSRF 를 렌더 비교 시험으로 전후 동일 단언. 클릭 수 불변.

## 완료 기준
Story acceptance(`STORY-TOS-a174`). `/settings/daily` 1280 높이 ≤ 2500px, 떠 있는 박스 ≤ 3, 결정·기억 항목 "높음" 0.

## not-applicable
변이 원장·다중 적대 리뷰: 비례 원칙(렌더·CSS·시험). 설정 저장 핸들러·`GatePreflight` 본문은 편집하지 않는다.

FLM/BTM 은 면제가 **아니다**: 기존 함수 `pages.go:Console.handleDashboard` 의 리터럴에 한 줄을 더하므로 gate ⑤ 가 번들 1개를 요구한다(`not-applicable` 표지는 required 0 일 때만 통한다). 번들은 편집 **전에** 만든다(tasks 3.1). 그 밖의 기존 함수 편집은 0 이 목표이고, 더 필요하면 구현자가 멈추고 Manager 에게 함수 이름을 보고한다. `GatePreflight` 의 절 구조(design §1)는 `tools/logic-map` AST 로 증거를 남긴다(tasks 1.1, 편집 0).
