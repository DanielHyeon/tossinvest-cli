# 콘솔 디자인 시스템

정본은 `internal/console/templates.go` 의 `{{define "style"}}` 이다. 이 문서는 그 요약이며 어긋나면 코드가 맞다.
범위는 콘솔 화면(`internal/console/`)뿐이고 주문·전략 경로에는 쓰지 않는다. change `a172`, 규칙 20(`.claude/CLAUDE.md`).

## 원칙
- 색·여백·크기는 `:root` 변수에서만 읽고, 다크 테마는 `prefers-color-scheme: dark` 블록의 `:root` 변수 재정의로만 만든다.
- 웜 스톤 중성 + 액센트 하나(`--accent`) + 채도를 낮춘 의미색. 그라디언트·웹폰트·Tailwind/Material 기본 팔레트는 쓰지 않는다.
- 정보 섹션은 상단 구분선 하나만 두고, 떠 있는 면(`--surface-1` + `--lift`)은 먼저 읽어야 하는 상태 스트립·경고·승인 바에만 준다. 측면 3px 띠(side-tab)는 금지(시험 `TestNoSideTabAccentBorders`).
- 보유 표(a058: 12px/18px, 8px 12px)와 44px 터치 타깃은 고정이다. 그 밖의 여백은 `clamp()` 유동 값이다.

## 색 토큰 (라이트 / 다크)
| 토큰 | 라이트 | 다크 |
|---|---|---|
| `--surface-0` 페이지 | `#f5f3ee` | `#141311` |
| `--surface-1` 묶음 | `#fcfbf8` | `#201f1b` |
| `--surface-2` 안쪽 홈·입력칸 | `#ebe8e1` | `#11100e` |
| `--line` / `--line-soft` 장식 구분선 | `#d9d5ca` / `#e7e3da` | `#38362f` / `#2a2924` |
| `--line-strong` UI 경계(3:1) | `#857f73` | `#756f64` |
| `--ink-1` → `--ink-3` 본문 → 흐린 설명 | `#201e1a` `#47433b` `#625d54` | `#ece9e1` `#cbc7bb` `#a49f93` |
| `--accent` 링크·포커스 링 | `#2e5a86` | `#9fbbd8` |
| `--ok` / `-line` / `-bg` | `#2f6a40` `#5f8f6c` `#e6eee3` | `#93c7a0` `#5a9068` `#1b2a1f` |
| `--warn` / `-line` / `-bg` | `#7f540f` `#a9761e` `#f8eed9` | `#e0b765` `#a97b2d` `#2a2314` |
| `--bad` / `-line` / `-bg` | `#9a3228` `#b4503f` `#f6e6e1` | `#ea9a90` `#c25f52` `#2d1a17` |

대비 기준(시험 `TestTradingViewsThemeColorsMeetWCAGInBothThemes`가 두 테마를 모두 잰다): 텍스트 토큰은 세 면 위에서 4.5:1,
잉크는 의미색 바탕 위에서도 4.5:1, `--line-strong`·`--accent`·의미색 `-line` 은 3:1 이상이다.

## 타입 스케일 (모듈러, 좁은 화면 1.2 → 넓은 화면 1.25, 각 단계 clamp)
| 토큰 | 최소 → 최대 | 쓰임 | 줄 간격·자간 |
|---|---|---|---|
| `--text--1` | 0.78 → 0.85rem | 설명·레이블·알약 | 1.3~1.5, +0.02em |
| `--text-0` | 0.9375 → 1.0625rem | 본문(최소 15px) | 1.65, 0 |
| `--text-1` | 1.125 → 1.33rem | h3 | 1.35, -0.004em |
| `--text-2` | 1.35 → 1.66rem | h2 | 1.28, -0.012em |
| `--text-3` | 1.62 → 2.07rem | h1 | 1.2, -0.02em |

숫자 열은 `font-variant-numeric: tabular-nums` 를 유지한다. 시험이 리터럴을 고정한 규칙(`table` 0.9rem, `.safety-line` 1rem, 보유 표 px)은 스케일 밖이다.

## 여백·반경
- 여백 `--space-1` … `--space-5` 는 clamp 유동 값이다.
- 반경은 역할별 세 조합이고 네 모서리가 다르다: `--radius-card`(0.9 0.35 0.8 0.3rem, 묶음·카드·모달·모바일 표 행), `--radius-control`(0.35 0.7 0.35 0.6rem, 버튼·입력칸·탭), `--radius-soft`(0.25 0.5 0.3 0.45rem, 알림·코드·안쪽 칸). 알약(`.status-pill`)만 999px.

## 금지 목록 (규칙 20)
- 보라 그라디언트, Inter 등 웹폰트 도배, 기계적 대칭(같은 반경·여백 복사), 고정 px 만으로 짠 스케일.
- 주문·토글·승인·청산 버튼을 장식으로 흐리거나 전환으로 늦추는 것(안전 불변식 4). 버튼은 항상 채워진 「누르는 것」으로 보인다.
- 접근성(포커스 링 `--accent` 3px·offset 2px, 44px 타깃)과 모션 성능 규칙이 미학보다 앞선다.

## 모션
4단계에서 추가 예정(스프링 느낌, `transform`/`opacity` 만, `prefers-reduced-motion` 존중). 현재 전환·애니메이션은 0개다.
