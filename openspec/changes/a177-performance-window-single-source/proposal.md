# a177 — 성과 이력 조회 창 표기 단일화('최근 30일' 리터럴 vs '최근 0일' 기본값)

프로그램 맥락: `openspec/changes/a173-console-ux-common-cleanup/proposal.md` 「프로그램 개요」 S5.

## Why

2026-10-25 UX 전수 검사(HEAD `b27a1ea87`, 빈 설정 폴더):
- `/optimization?category=performance-history` 본문 notice: **최근 30일 · 전체 시장 · 전체 lane · complete lineage · 읽기 전용**.
- `/performance-history` 「서버 조회 기본값」 칩: **최근 0일 · 전체 시장 · 전체 lane · complete lineage only**.

`internal/console/templates_optimization.go:217` 의 "최근 30일" 은 템플릿 안의 글자다. 성과 이력 화면은 서버 조회 기본값을 렌더한다. 같은 사실을 두 화면이 다른 값으로 말한다. 사용자 결정(2026-10-25): 배치 문제가 아니라 값 문제이므로 별도 버그 change 로 **근본 원인부터**(규칙 3) 찾는다.

## 확정 (구현 1.1/1.2, 2026-10-25) — H1
리터럴 `templates_optimization.go:217` 은 데이터를 읽지 않는다. 칩은 `.View.Query.PeriodDays` 를 찍는데 `Unwired`/`LoadErr` 에서 `View` 가 영값이라 0일이 찍힌다(성공 경로는 `DefaultQuery` 30일). 30일이 의도된 기본값(H1), 0일은 미배선의 가면(둘째 결함). 일수는 바꾸지 않는다.

## 가설 (원문 보존)
- H1: 리터럴이 낡았고 서버 기본값(0일)이 정본 → 리터럴을 제거하고 성과 이력 화면이 쓰는 같은 값을 넘겨 렌더한다.
- H2: 서버 기본값 배선에 결함이 있어 0일이 잘못된 값(의도는 30일) → 기본값을 **바꾸는 것은 사람 결정** 뒤에만 한다. 이 change 는 표기 단일화까지만 하고 값 변경은 `issues.md` 에 사람 결정 항목으로 남긴다.
- 어느 쪽이든 두 화면은 **한 출처**에서 읽는다(spec 851 의 "요약을 재계산하지 않고 옮긴다" 원칙 준용).

## What Changes
- 조회 창·시장·lane·lineage 표기를 성과 이력 화면이 쓰는 값 객체에서 읽어 최적화 카테고리에 넘긴다. 리터럴 "최근 30일" 삭제.
- 시험: 두 화면을 같은 fixture 로 렌더해 같은 창 문구를 단언(RED→GREEN). 리터럴을 되돌리면 FAIL.

## Scope / Non-goals
- 범위: `internal/console/templates_optimization.go`·성과 이력 뷰 모델·시험.
- 비목표: 성과 DB 질의·prune·조회 기본값 일수 자체·주문·설정 저장. UI 배치(a175).

## Risk
Low. 표시 문자열의 출처 변경. High-risk 경로 아님.

## not-applicable
FLM/BTM·변이 원장·다중 적대 리뷰: 비례 원칙. 기존 Go 함수 내부를 편집하면 `review.md` 에 면제 문구.
