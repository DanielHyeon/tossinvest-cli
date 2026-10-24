# a172 — 콘솔 anti-AI 리디자인

## Why

규칙 20(사용자 지시 2026-10-25)은 콘솔 UI 의 AI 티를 금지한다: 기계적 대칭(균일 반경·간격) 깨기, `clamp()` 유동 스케일, 스프링 느낌 모션, 저밀도 프리미엄 질감. 현재 콘솔 CSS(`internal/console/templates.go` 의 `{{define "style"}}`, 단일 스타일시트)는 이 규칙 이전에 쓰였다. 4단계로 나눠 반영한다.

## What Changes

- 1단계 날것의 담백한 구조: 징후 측정 → 장식 제거(eyebrow, 균일 반경, 평면 흰 카드 배경).
- 2단계 프리미엄 미니멀 테마(hallmark 대체): 토큰·비대칭 반경·여백 층·clamp 여백.
- 3단계 디자인 시스템 무결성: 타입 스케일·팔레트·대비·CSS 변수 재조정.
- 4단계 스프링 물리 인터랙션 + GPU 가속 최적화(`prefers-reduced-motion` 존중).
- 5단계 검증·포트.

## Scope / Non-goals

- 범위: `{{define "style"}}` CSS + `internal/console/*.go` 템플릿의 class·장식 요소 마크업.
- 비목표: Go 로직·핸들러·라우트·템플릿 데이터, 비콘솔 패키지, class 이름 변경(추가만 허용), 웹폰트·JS·의존성·CDN, 밀도(a058 12px/18px) 변경, 새 UI 기능.

## Risk

Low. 주문·승인 버튼은 장식 때문에 지연되지 않는다(전환은 시각 속성만, 클릭 경로 무관). 고정 리터럴 시험은 정당한 사유 시에만 고치고 tasks.md 에 전수 기록한다.

## Skills

설치된 `ui-skills-root`(baseline-ui·improve-ui 등), `ui-ux-pro-max`, `fixing-accessibility`, `fixing-motion-performance` 만 쓴다. 사용자가 지명한 `avoid-ai-design`·`hallmark` 는 미설치(설치는 사람 결정).

## not-applicable

FLM/BTM·변이 원장·다중 적대 리뷰: Go 함수 내부 로직을 편집하지 않고 High-risk 경로·게이트 판정을 바꾸지 않는다(docs/WORKFLOW.md 「비례 원칙」). 시험 + lint + 렌더 확인 + 경량 리뷰로 갈음한다.
