# a172 — 콘솔 anti-AI 리디자인

## Why

규칙 20(사용자 지시 2026-10-25)은 콘솔 UI 의 AI 티를 금지한다: 기계적 대칭(균일 반경·간격) 깨기, `clamp()` 유동 스케일, 스프링 느낌 모션, 저밀도 프리미엄 질감. 현재 콘솔 CSS(`internal/console/templates.go` 의 `{{define "style"}}`, 단일 스타일시트)는 이 규칙 이전에 쓰였다. 4단계로 나눠 반영한다.

## What Changes

- 1단계 날것의 담백한 구조: 징후 측정 → 장식 제거(eyebrow, 균일 반경, 평면 흰 카드 배경).
- 2단계 프리미엄 미니멀 테마(hallmark 대체): 토큰·비대칭 반경·여백 층·clamp 여백.
- 3단계 디자인 시스템 무결성: 타입 스케일·팔레트·대비·CSS 변수 재조정.
- 4단계 스프링 물리 인터랙션 + GPU 가속 최적화(`prefers-reduced-motion` 존중).
- 5단계 검증·포트.
- 6단계 CSS-only "셰이더풍" 인터랙션(사용자 지시 2026-10-25, 사람 결정 선택지 2): 헤더·상태 스트립 뒤 저채도 radial-gradient 덩어리가 `translate`/`rotate`/`opacity` 로만 느리게 움직이고(엔진·데이터 톤이 색을 정함), 카드 hover 에 `opacity` 로 켜지는 radial 조명. JS 0·WebGL 0·새 라우트 0·CSP 불변.

## Scope / Non-goals

- 범위: `{{define "style"}}` CSS + `internal/console/*.go` 템플릿의 class·장식 요소 마크업.
- 비목표: Go 로직·핸들러·라우트·템플릿 데이터, 비콘솔 패키지, class 이름 변경(추가만 허용), 웹폰트·JS·의존성·CDN, 밀도(a058 12px/18px) 변경, 새 UI 기능.

## 6단계 범위 정정 (사람 결정 2026-10-25)
6.1 정찰 뒤 JS asset + `/static/` 라우트 + `script-src 'self'` 안을 한때 문서에 적었으나(`bbe176f5e`) 사람이 되돌렸다. 근거(스펙 원문):
- operator-console 「보호선 화면은 스크립트 없이 엔진 주기를 따라간다」: "client-side script를 도입하거나 배포 CSP를 완화해서는 안 된다(SHALL NOT)", "응답 CSP는 `default-src 'none'`을 유지하고 `script-src`를 추가해서는 안 된다(SHALL NOT)".
- operator-console 「CSP 안전한 포지션 관리 조작」(spec 174) 과 반응형 표시 요구(spec 684): "이 반응형 표시는 JavaScript나 외부 asset을 요구해서는 안 된다(SHALL NOT)".
- console-request-origin(spec 83): "The change SHALL NOT alter the existing deny-by-default CSP".
따라서 6단계의 변경은 `{{define "style"}}` CSS 와 시험뿐이다. Go 코드·CSP 상수·라우트·JS 는 변경하지 않는다.

## Risk

Low. 주문·승인 버튼은 장식 때문에 지연되지 않는다(전환은 시각 속성만, 클릭 경로 무관). 고정 리터럴 시험은 정당한 사유 시에만 고치고 tasks.md 에 전수 기록한다.

## Skills

설치된 `ui-skills-root`(baseline-ui·improve-ui 등), `ui-ux-pro-max`, `fixing-accessibility`, `fixing-motion-performance` 만 쓴다. 사용자가 지명한 `avoid-ai-design`·`hallmark` 는 미설치(설치는 사람 결정).

## not-applicable

FLM/BTM·변이 원장·다중 적대 리뷰: Go 함수 내부 로직을 편집하지 않고 High-risk 경로·게이트 판정을 바꾸지 않는다(docs/WORKFLOW.md 「비례 원칙」). 시험 + lint + 렌더 확인 + 경량 리뷰로 갈음한다.

6단계 위험: Low. 장식은 `pointer-events: none` 의사요소이고 주문 경로와 무관하다. FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`: Go 함수 내부 분기를 편집하지 않고 CSS 와 시험만 바꾼다(비례 원칙). 모션 계약 시험 변이 1회는 한다.
