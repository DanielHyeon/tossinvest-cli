# a172 1단계 스캔 (대상: templates.go {{define "style"}} 306줄, base fd6afe23b)
| 항목 | 측정 | 처리 |
|---|---|---|
| gradient | 0 | 없음(지울 것 없음) |
| border-radius | 6px x9, 4px x7, 8px x1, 999px x2 (균일 4/6px) | 0 으로 중립화 18곳(6px 9·4px 7·8px 1·settings-tabs 999px 1; 2단계에서 토큰), status-pill 999px 1곳 유지 |
| font-family | 3 스택, 전부 시스템(ui-sans-serif/system-ui, ui-monospace). 웹폰트 0, Inter 0 | 유지 |
| 평면 흰 카드 배경 | 라이트 #fff 5곳(section, console-status/summary, optimization-nav a, settings-tabs a); 다크 #1d1d22 6곳 | 라이트 4 + 다크 4 제거(경계선이 묶음 표현). 상세 모달 배경은 불투명 필수라 유지 |
| eyebrow/section-kicker | CSS 2곳 + 마크업 4곳(templates_optimization.go) | "Optimization lifecycle"(순수 장식 영문) 삭제, 나머지 3곳은 정보라 class 만 muted 로(텍스트 보존). CSS 셀렉터는 무해하게 유지 |
| Tailwind 기본 팔레트 | #2878d0 x7, #fbbf24 x3, #4ade80 x3, #22c55e x3, #f43f5e x2, #60a5fa x2, #fca5a5 x1, #b42318 x1 | 1단계 미변경(색은 2단계 토큰·3단계 대비 몫) |
| transition/animation | 0 | 없음(모션은 3단계에서 신설) |
| box-shadow | 5(정보성 inset 3, 모달·sticky-save 2) | 유지 |
| clamp() | 0 | 2단계 |
| 인라인 style= | 2(margin-left:auto, width:8rem) | 유지 |
| class 이름 | 94 | 변경 0 |
