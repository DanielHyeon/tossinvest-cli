# a172 tasks

Story: `STORY-TOS-a172` · Low risk(CSS·장식 마크업만). FLM/BTM·변이·다중 적대 리뷰 `not-applicable`(proposal).

## 0. 정의
- [x] 0.1 Story 등록(FEAT-TOS-004) · change 정의 · base `fd6afe23b`

## 1. 1단계 — 날것의 담백한 구조
- [x] 1.1 징후 측정 → scratchpad a172/stage1-scan.md
- [x] 1.2 제거(eyebrow·균일 반경 중립화·평면 카드 배경)
- [x] 1.3 gofmt·vet·console/cmd 시험, 렌더 확인

## 2. 2단계 — 프리미엄 미니멀 테마(hallmark 대체)
- [x] 2.1 `:root` 토큰(`--space-*` clamp·`--radius-*` 역할별 3조합·`--surface-0/1/2`·`--line`·`--ink-*`·`--accent`·`--ok/--warn/--bad`)과 다크는 변수만 재정의
- [x] 2.2 비대칭 반경(card·control·soft, 알약 유지), 중첩 테두리를 면·약한 그림자·구분선 하나로 대체(테두리 선언 25 → 15)
- [x] 2.3 본문·제목 clamp 스케일, `.eyebrow/.section-kicker` 셀렉터 삭제, 기본 링크 색 `a { color: var(--accent) }`
- [x] 2.4 gofmt·vet·console/cmd 시험, 20장 렌더(`scratchpad/a172/stage2/`)

## 3. 3단계 — 디자인 시스템 무결성(타입 스케일·팔레트·대비·CSS 변수 재조정)
- [ ] 3.1 타입 스케일 · 3.2 팔레트(Tailwind 기본값 탈피)·대비 4.5:1 · 3.3 변수 정리

## 4. 4단계 — 스프링 물리 인터랙션 + GPU 가속 최적화(`ui-ux-pro-max` + `fixing-motion-performance`)
- [ ] 4.1 스프링 느낌 모션(`transform`/`opacity` 만, `prefers-reduced-motion` 존중) · 4.2 컴포지터 속성만·`will-change` 최소 · 4.3 주문·토글·승인·청산 버튼에는 지연 전환 금지(안전 불변식 4)

## 5. 검증·포트
- [ ] 5.1 전체 콘솔 시험 · 렌더 · 경량 리뷰 · 포트

## 고정 리터럴 시험 수정 기록
- 2단계 `display_primitives_test.go` `TestTheHeadingStepsAreDistinguishable` + 헬퍼(`declaredFontSize`·`declaredBodyFontSize`): 제목·본문 크기가 `clamp()` 라 `Nrem`/`Npx/` 리터럴 파서가 읽지 못함. 파서가 clamp 의 최소·최대를 읽도록 바꾸고, 제목 단계 순서(h1>h2>h3>본문)를 가장 좁은 뷰포트와 가장 넓은 뷰포트 양쪽에서 검사하며 본문 최소 15px 기준은 유지(본문 clamp 최소 0.9375rem = 15px). 약화 아님: 검사 지점이 1곳에서 2곳으로 늘어남.
- 2단계 `trading_views_test.go` `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA`: 다크 색이 개별 규칙(`.ok { color: #22c55e; }`)에서 `:root` 변수로 옮겨져 규칙 리터럴이 사라짐. 다크 블록의 변수 값을 읽어 `--ok/--warn/--bad/--bad-soft/--accent/--ink-1~3` 이 `--surface-0/1/2` 세 면 위에서 모두 4.5:1 이상인지 잼(기존: 두 색 × 한 면 `#1d1d22`). 약화 아님: 대상 색 2 → 8, 면 1 → 3.
