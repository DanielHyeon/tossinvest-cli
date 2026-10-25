package console

// templates.go is the markup. It is one string on purpose: the console has three
// pages and a refusal, and a template directory with four files in it would be
// four more things to keep in step with an embed directive.
//
// The style sheet is inline and deliberately tiny. Anything that had to be
// fetched — a font, a framework, an icon set — would be a request this page
// cannot make (there is nothing to make it to) and a dependency on somebody
// else's server for a screen that authorises live orders.
//
// The dark semantic status colours follow the StockOS lane-console palette.
// They are intentionally separate from the light colours: both must keep normal
// text at WCAG AA contrast against their respective section backgrounds.

const pageTemplates = `
{{define "style"}}
/* 토큰: 색·여백·반경은 여기서만 정의하고 아래 규칙은 변수만 읽음. 다크는 맨 아래 블록에서 변수만 재정의함. */
:root {
  color-scheme: light dark;
  /* 여백 스케일: 뷰포트에 따라 유동(clamp). 같은 값을 반복하지 않고 단계마다 비율이 다름 */
  --space-1: clamp(0.25rem, 0.2rem + 0.15vw, 0.35rem);
  --space-2: clamp(0.5rem, 0.42rem + 0.25vw, 0.7rem);
  --space-3: clamp(0.8rem, 0.65rem + 0.55vw, 1.2rem);
  --space-4: clamp(1.1rem, 0.85rem + 0.9vw, 1.75rem);
  --space-5: clamp(1.6rem, 1.2rem + 1.4vw, 2.6rem);
  /* 타입 스케일: 좁은 화면 1.2 → 넓은 화면 1.25 비율의 유동 모듈러 스케일(본문 0 단계 최소 15px) */
  --text--1: clamp(0.78rem, 0.74rem + 0.15vw, 0.85rem);
  --text-0: clamp(0.9375rem, 0.88rem + 0.25vw, 1.0625rem);
  --text-1: clamp(1.125rem, 1.04rem + 0.4vw, 1.33rem);
  --text-2: clamp(1.35rem, 1.2rem + 0.7vw, 1.66rem);
  --text-3: clamp(1.62rem, 1.35rem + 1.3vw, 2.07rem);
  /* 반경: 역할별 세 조합, 네 모서리를 다르게 둠(기계적 대칭 회피). 알약은 상태 표시 전용 */
  --radius-card: 0.9rem 0.35rem 0.8rem 0.3rem;
  --radius-control: 0.35rem 0.7rem 0.35rem 0.6rem;
  --radius-soft: 0.25rem 0.5rem 0.3rem 0.45rem;
  /* 면: 0 = 페이지, 1 = 묶음, 2 = 묶음 안쪽 홈(입력칸 포함) */
  --surface-0: #f5f3ee; --surface-1: #fcfbf8; --surface-2: #ebe8e1;
  /* 선: line·line-soft 는 장식 구분선, line-strong 은 3:1 이상이 필요한 UI 경계(입력칸·꺼짐 상태) */
  --line: #d9d5ca; --line-soft: #e7e3da; --line-strong: #857f73;
  /* 잉크: 1 → 3 으로 갈수록 옅어짐(본문 → 보조 → 흐린 설명), 세 면 위에서 4.5:1 이상 */
  --ink-1: #201e1a; --ink-2: #47433b; --ink-3: #625d54;
  --accent: #2e5a86;
  /* 의미색: 각각 글자(fg) · 선(line, 3:1 이상) · 바탕(bg) 세 역할 */
  --ok: #2f6a40; --ok-line: #5f8f6c; --ok-bg: #e6eee3;
  --warn: #7f540f; --warn-line: #a9761e; --warn-bg: #f8eed9;
  --bad: #9a3228; --bad-line: #b4503f; --bad-bg: #f6e6e1;
  /* 앰비언트 톤 틴트(change a172 6단계): 글자색(--ok 등)은 어두워서 섞으면 바탕이 어두워지고 톤끼리 비슷해짐.
     그래서 바탕과 밝기가 비슷하고 색상만 벌어진 전용 틴트를 두고 테마별 투명도로 섞음.
     틴트 4색 쌍마다 ΔE76 12 이상, 글자 토큰 대비 4.5 이상을 TestAmbientToneLegibility 가 고정 */
  --tint-accent: #5b85d0; --tint-ok: #52b96e; --tint-warn: #e9b207; --tint-bad: #d8614b; --tone-alpha: 19%;
  --lift: 0 1px 2px rgba(32, 30, 26, 0.07);
  /* 모션: 시간 토큰 3단과 스프링 이징 2종. 기본값은 linear() 미지원 브라우저용 짧은 ease-out 이고
     지원하면 아래 @supports 블록이 감쇠 스프링 근사로 덮음. 움직임 크기도 토큰이라 reduced-motion 에서 0 으로 리셋됨 */
  --dur-1: 120ms; --dur-2: 220ms; --dur-3: 300ms;
  --ease-spring: ease-out; --ease-settle: ease-out; --ease-pop: ease-out;
  --nudge: -2px; --press: 0.97; --pop-from: 0.96;
  /* 떠 있는 층: 상세 모달(pop) · 하단 승인 바(bar) */
  --shadow-pop: 0 18px 54px rgba(20, 18, 14, 0.3), 0 0 0 1px var(--line);
  --scrim: rgba(20, 18, 14, 0.32);
  --shadow-bar: 0 -5px 20px rgba(20, 18, 14, 0.14);
}
* { box-sizing: border-box; }
body {
  margin: 0; padding: 0 var(--space-3) 4rem;
  font: var(--text-0)/1.65 ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Noto Sans KR", sans-serif;
  background: var(--surface-0); color: var(--ink-1);
  /* This console prints filesystem paths, record ids and hashes in ordinary
     prose, not only inside <code>. At 375px an unbroken path is wider than the
     screen and pushes the whole document sideways — measured, on the
     verification console (change a054, task 4.7). anywhere rather than
     break-word because it also shrinks the intrinsic minimum, which is what
     the grid columns below are sized against. */
  overflow-wrap: anywhere;
}
a { color: var(--accent); }
main, header > div { max-width: 72rem; margin: 0 auto; min-width: 0; }
header { border-bottom: 1px solid var(--line); margin-bottom: var(--space-4); }
header > div { display: flex; gap: var(--space-3); align-items: baseline; padding: var(--space-3) 0; }
header strong { font-size: var(--text-0); letter-spacing: 0.02em; }
nav { display: flex; flex-wrap: wrap; gap: 0.15rem var(--space-3); min-width: 0; }
/* min-height 44px: the navigation is the most-tapped thing on the screen and a
   text link with no padding is a 20px target (a054 §4 touch floor). */
nav a { color: var(--ink-2); text-decoration: none; display: flex; flex-direction: column;
  justify-content: center; min-height: 44px; padding: 0.2rem 0; min-width: 0; }
nav a small { color: var(--ink-3); font-size: var(--text--1); line-height: 1.3; font-weight: 400; }
nav a.on { font-weight: 600; text-decoration: underline; }
.header-note { margin-left: auto; }
h1 { font-size: var(--text-3); font-weight: 700; line-height: 1.2; letter-spacing: -0.02em; margin: var(--space-5) 0 var(--space-1); text-wrap: balance; }
h2 { font-size: var(--text-2); font-weight: 650; line-height: 1.28; letter-spacing: -0.012em; margin: var(--space-4) 0 var(--space-1); text-wrap: balance; }
h3 { font-size: var(--text-1); line-height: 1.35; letter-spacing: -0.004em; margin: var(--space-3) 0 var(--space-1); }
/* 묶음은 테두리 대신 면(surface-1)과 아주 약한 그림자로 구분함 */
section { padding: var(--space-3) 0; margin: var(--space-5) 0; border-top: 1px solid var(--line); }
/* 떠 있는 면은 먼저 읽어야 하는 것(상태 스트립·경고·승인)에만 줌 */
.sticky-save, .order-summary { padding: var(--space-3) var(--space-4); border-top: 0; background: var(--surface-1);
  border-radius: var(--radius-card); box-shadow: var(--lift); }
pre { overflow-x: auto; padding: var(--space-2) var(--space-3); background: var(--surface-2); border-radius: var(--radius-soft);
      font: var(--text--1)/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
table { border-collapse: collapse; width: 100%; font-size: 0.9rem; }
th, td { text-align: left; padding: 0.3rem 0.6rem 0.3rem 0; border-bottom: 1px solid var(--line-soft); vertical-align: top; }
code { overflow-wrap: anywhere; }
dl { display: grid; grid-template-columns: 8rem 1fr; gap: 0.25rem var(--space-3); margin: var(--space-2) 0; }
dt { color: var(--ink-3); } dd { margin: 0; }
.notice { border: 1px solid var(--warn-line); background: var(--warn-bg); box-shadow: var(--lift); padding: var(--space-2) var(--space-3); border-radius: var(--radius-soft); }
.danger { border: 1px solid var(--bad-line); background: var(--bad-bg); box-shadow: var(--lift); padding: var(--space-2) var(--space-3); border-radius: var(--radius-soft); }
.ok { color: var(--ok); } .bad { color: var(--bad); } .muted { color: var(--ink-3); }
input[type=text] { font: var(--text-0)/1.4 ui-monospace, Menlo, Consolas, monospace; padding: 0.45rem 0.6rem; min-height: 44px;
                   width: 22rem; max-width: 100%; border: 1px solid var(--line-strong); border-radius: var(--radius-control);
                   background: var(--surface-2); color: var(--ink-1); }
button { font: inherit; padding: 0.45rem 1rem; border-radius: var(--radius-control); border: 1px solid var(--ink-3);
         background: var(--ink-1); color: var(--surface-0); cursor: pointer; }
button.secondary { background: transparent; color: var(--ink-1); }
/* 상태 규칙은 즉시 반영됨(전환 없음). 누르는 것은 항상 채워진 모양을 유지하고 명도 단계만 바꿈 */
button:hover:not(:disabled), .button-link:hover { background: var(--ink-2); }
button:active:not(:disabled), .button-link:active { background: var(--ink-3); }
button.secondary:hover:not(:disabled), .secondary-link:hover { background: var(--surface-2); }
button.secondary:active:not(:disabled), .secondary-link:active { background: var(--line); }
button:disabled { cursor: not-allowed; opacity: 0.65; }
.button-link { min-height: 44px; display: inline-flex; align-items: center; justify-content: center;
  padding: 0.45rem 1rem; border: 1px solid var(--ink-3); border-radius: var(--radius-control); background: var(--ink-1);
  color: var(--surface-0); text-decoration: none; }
.secondary-link { background: transparent; color: var(--ink-1); }
form { display: inline; }
.page-intro { max-width: 52rem; }
.data-table { table-layout: fixed; }
.data-table caption { text-align: left; font-weight: 700; font-size: var(--text-0); padding: 0 0 0.65rem; }
.data-table th, .data-table td { padding: 0.75rem 0.65rem; overflow-wrap: anywhere; }
.data-table tbody tr:last-child > * { border-bottom: 0; }
.positions-table { font-size: 12px; line-height: 18px; }
.positions-table thead { font-size: 10px; line-height: 15px; }
.positions-table th, .positions-table td { padding: 8px 12px; }
.positions-table .number-column { text-align: right; font-variant-numeric: tabular-nums; }
.positions-table thead th:nth-child(5) { text-align: right; }
.position-col-instrument { width: 31%; }
.position-col-quantity { width: 6%; }
.position-col-average, .position-col-current { width: 11%; }
.position-col-lines { width: 18%; }
.position-col-value { width: 12%; }
.position-col-pnl { width: 11%; }
.position-col-detail { width: 7%; }
.holding-name { display: block; margin-bottom: 0.12rem; font-weight: 700; }
.holding-name code { font-size: inherit; font-weight: 500; }
.holding-verdicts { display: flex; flex-wrap: wrap; gap: 3px 5px; align-items: center; margin-top: 4px; }
.holding-verdicts .submetric { display: inline; margin-top: 0; }
.number-cell { text-align: right; font-variant-numeric: tabular-nums; white-space: nowrap; }
.line-item { display: flex; justify-content: space-between; align-items: baseline;
  gap: 5px; margin-top: 2px; font-size: 12px; line-height: 15px; font-variant-numeric: tabular-nums; }
.line-item > span { color: var(--ink-3); font-size: 10px; line-height: 12.5px; }
.line-item > strong { font-size: 11px; line-height: 13.75px; text-align: right; overflow-wrap: normal; }
/* One status primitive. .state-badge said the same thing under a second name
   and was used twice; this one was used eight times and is asserted by an
   existing test, so the merge keeps the common name (change a054). */
.status-pill { display: inline-block; border: 1px solid currentColor; border-radius: 999px;
  padding: 0.08rem 0.55rem; font-size: var(--text--1); }
.positions-table .status-pill { padding: 1px 6px; font-size: 11px; line-height: 15px; }
.positions-table .submetric { font-size: 11px; line-height: 15px; }
.metric { display: block; font-size: var(--text-0); font-weight: 700; }
.submetric { display: block; margin-top: 0.15rem; color: var(--ink-3); font-size: var(--text--1); }
.exit-line-stack { text-align: right; font-variant-numeric: tabular-nums; }
.exit-line-status { display: block; margin-top: 0.2rem; }
.safety-line { display: block; margin-top: 0.2rem; font-size: 1rem; line-height: 1.5; }
.actions { display: flex; flex-wrap: wrap; gap: 0.4rem; align-items: flex-start; }
.actions form { display: block; }
.actions button { padding: 0.32rem 0.65rem; font-size: var(--text--1); }
/* 펼침 요약은 어디서든 44px 터치 목표 */
summary { min-height: 44px; align-content: center; }
.row-details, .explain { margin-top: 0.45rem; }
.row-details summary, .explain summary { color: var(--ink-3); cursor: pointer; font-weight: 600; }
.row-details summary { min-height: 44px; display: flex; align-items: center; }
.detail-grid { display: grid; grid-template-columns: 1fr; gap: 0.25rem 0.9rem; margin-top: 0.5rem; }
.detail-grid span { min-width: 0; }
.detail-cell { min-height: 66px; text-align: right; white-space: nowrap; }
.detail-cell .row-details { min-height: 44px; margin: 0; }
.detail-cell .row-details summary { justify-content: flex-end; }
.detail-cell .row-details[open] { position: fixed; left: 50%; top: 50%;
  z-index: 21; width: min(42rem, calc(100% - 2rem)); max-height: min(70vh, 42rem);
  margin: 0; padding: 0.65rem 1.25rem 1.25rem; overflow: auto; transform: translate(-50%, -50%);
  box-sizing: border-box; border-radius: var(--radius-card);
  background: var(--surface-1); color: var(--ink-1); box-shadow: var(--shadow-pop);
  text-align: left; white-space: normal; }
.detail-cell .row-details[open] summary { position: sticky; z-index: 22; top: 0; min-height: 44px;
  justify-content: flex-end; padding: 0 0.4rem; border-radius: var(--radius-soft); background: var(--surface-1); }
.detail-cell .row-details[open] summary::after { content: "닫기"; margin-inline-start: 0.5em; }
.detail-cell .row-details[open] .detail-grid { align-content: start; margin-top: 0.25rem; }
.filter-bar { display: flex; flex-wrap: wrap; gap: 0.5rem 1.1rem; align-items: center; }
.filter-group { display: flex; flex-wrap: wrap; gap: 0.35rem 0.7rem; align-items: center; }
.filter-group .muted { margin-right: 0.15rem; }
.optimization-title { display: flex; gap: var(--space-3); align-items: flex-start; justify-content: space-between; }
.optimization-title > div { min-width: 0; }
.optimization-title, .optimization-shell, .optimization-content, .optimization-content > section,
.optimization-content article, .optimization-content fieldset { min-width: 0; max-width: 100%; }
.optimization-title code, .optimization-shell code, .preview-review code { overflow-wrap: anywhere; word-break: break-word; }
.status-strip { display: grid; grid-template-columns: repeat(2, minmax(7rem, 1fr)); min-width: min(100%, 28rem);
  background: var(--surface-1); border-radius: var(--radius-card); overflow: hidden; box-shadow: var(--lift); }
.status-strip > div { padding: 0.55rem 0.75rem; border-bottom: 1px solid var(--line-soft); }
.status-strip dt { font-size: var(--text--1); } .status-strip dd { font-size: var(--text-0); }
.status-strip dd small { display: block; margin-top: 0.1rem; color: var(--ink-3); font-size: var(--text--1); }
.status-strip[data-lifecycle-state=error] { box-shadow: inset 0 0 0 1px var(--warn-line); }
.status-strip[data-lifecycle-state=unavailable] { box-shadow: inset 0 0 0 1px var(--line-strong); }
.status-strip [data-evidence-state=stale], .status-strip [data-evidence-state=insufficient] { border-color: var(--warn-line); }
.status-strip [data-evidence-state=unavailable] { border-color: var(--line-strong); }
/* .console-status is the shell's strip: the same four facts, same place, every
   screen. auto-fit rather than a column count, so it reflows at any width
   without a second breakpoint to keep in step with the first. */
.console-status, .console-summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr));
  gap: 0; max-width: 72rem; margin: 0 auto var(--space-3); background: var(--surface-1);
  border-radius: var(--radius-card); overflow: hidden; box-shadow: var(--lift); }
.console-status > div, .console-summary > div { min-width: 0; padding: 0.5rem 0.75rem;
  border-left: 1px solid var(--line-soft); }
.console-status > div:first-child, .console-summary > div:first-child { border-left: 0; }
.console-status dt, .console-summary dt { color: var(--ink-3); font-size: var(--text--1); }
.console-status dd, .console-summary dd { font-size: var(--text-0); overflow-wrap: anywhere; }
.console-status dd small, .console-summary dd small { display: block; margin-top: 0.1rem;
  color: var(--ink-3); font-size: var(--text--1); }
/* The overview summary sits above the detail sections it points at, so it is
   a little louder than the shell strip and each market gets its own line. */
.console-summary { margin-bottom: var(--space-4); }
.console-summary dd { font-size: var(--text-0); font-weight: 600; }
.console-summary .summary-line { display: block; font-weight: 400; }
.console-summary .summary-line + .summary-line { margin-top: 0.15rem; }
/* Tone is never the only carrier: every toned cell also renders the word, so a
   reader who cannot see the colour reads the same verdict. */
.console-status [data-data-tone=ok] { color: var(--ok); }
.console-status [data-data-tone=warn], .console-status [data-engine-state=stopped] { color: var(--warn); }
.console-status [data-data-tone=stale] { color: var(--bad); }
.console-approval { max-width: 72rem; margin: 0 auto 0.9rem; }
/* --- the settings tabs and the card standard (change a055) ------------------ */
.settings-tabs { display: flex; flex-wrap: wrap; gap: 0.4rem; margin: var(--space-3) 0 var(--space-3); }
.settings-tabs a { min-height: 44px; display: inline-flex; align-items: center; padding: 0.4rem 1rem;
  border-radius: var(--radius-control); background: var(--surface-2); color: var(--ink-2); text-decoration: none; }
.settings-tabs a.on { box-shadow: inset 0 -2px var(--accent); font-weight: 700; color: var(--ink-1); }
.settings-state { margin-bottom: var(--space-4); }
.settings-state [data-state-tone=ok] { color: var(--ok); }
.settings-state [data-state-tone=warn] { color: var(--warn); }
/* 현재값 and 적용 후 are the same grid so the eye compares them row for row. */
.card-current, .card-preview { grid-template-columns: minmax(8rem, 14rem) 1fr; }
.card-preview { background: var(--surface-2); border-radius: var(--radius-soft); padding: var(--space-2) var(--space-3); margin: 0.7rem 0; }
.card-preview dt { color: var(--ink-2); font-weight: 600; }
.tier-card { background: var(--surface-2); border-radius: var(--radius-soft); padding: var(--space-2) var(--space-3); margin: var(--space-2) 0; }
.tier-card form { display: block; }
[data-limit-axis=강화] { color: var(--ok); }
[data-limit-axis=완화] { color: var(--warn); }
[data-limit-axis="통화 변경"], [data-limit-axis="최초 설정"] { color: var(--ink-2); }
.entry-list { grid-template-columns: minmax(8rem, 12rem) 1fr; gap: var(--space-3) var(--space-3); }
.entry-list dt a { min-height: 44px; display: inline-flex; align-items: center; font-weight: 700; }
/* A URL-driven disclosure. It gets no native toggle on purpose — see explain.go. */
.explain-link { margin: 0.5rem 0; }
.explain-link > a { font-weight: 600; color: var(--ink-2); }
.explain-link > div { margin-top: 0.4rem; }
.hypothetical-results th, .hypothetical-results td { white-space: nowrap; }
.hypothetical-fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 13rem), 1fr)); gap: 0.8rem; }
.hypothetical-fields label { display: block; font-weight: 600; margin-bottom: 0.25rem; }
.hypothetical-fields input, .hypothetical-fields select { box-sizing: border-box; width: 100%; min-height: 44px; padding: 0.4rem; }
.hypothetical-fields .danger { display: block; margin-top: 0.25rem; }
.hypothetical-fields :focus-visible { outline: 3px solid var(--accent); outline-offset: 2px; }
.optimization-shell { display: grid; grid-template-columns: minmax(13rem, 16rem) minmax(0, 1fr); gap: var(--space-3); align-items: start; }
.optimization-nav { position: sticky; top: 0.75rem; display: grid; gap: 0.35rem; }
.optimization-nav a { min-height: 56px; display: flex; flex-direction: column; justify-content: center; padding: 0.55rem 0.75rem;
  border-radius: var(--radius-control); background: var(--surface-1); color: var(--ink-1); text-decoration: none; box-shadow: var(--lift); }
.optimization-nav a.on { box-shadow: inset 0 -2px var(--accent); }
.optimization-nav small { color: var(--ink-3); line-height: 1.35; }
.optimization-content { min-width: 0; }
.optimization-steps { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--space-2); margin: var(--space-3) 0;
  padding: 0; list-style: none; counter-reset: none; }
.optimization-steps li { min-width: 0; padding: var(--space-2) var(--space-3); background: var(--surface-2); border-radius: var(--radius-soft); }
.optimization-steps strong, .optimization-steps span { display: block; }
.optimization-steps span { margin-top: 0.2rem; color: var(--ink-3); font-size: var(--text--1); }
.category-summary { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-3) var(--space-4); }
.category-summary article { border-top: 1px solid var(--line-soft); padding-top: 0.4rem; }
.category-summary h3 { margin-bottom: 0.1rem; font-size: var(--text-0); font-weight: 600; }
fieldset.setting-row { border: 0; border-radius: var(--radius-card); background: var(--surface-2); margin: var(--space-3) 0; padding: var(--space-3) var(--space-4); }
fieldset.setting-row legend { font-weight: 700; padding: 0 0.35rem; }
fieldset.setting-error { border: 1px solid var(--bad-line); }
.configuration-error { overflow-wrap: anywhere; }
.setting-values { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.setting-values > div { min-width: 0; border-top: 1px solid var(--line-soft); padding: 0.35rem 0; }
.choice-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--space-3); }
.choice-tile { background: var(--surface-1); border-radius: var(--radius-soft); padding: var(--space-3); min-width: 0; }
.choice-tile[data-selected=true] { background: var(--ok-bg); box-shadow: inset 0 0 0 1px var(--ok-line); }
.choice-tile header { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.4rem; }
.choice-tile h3 { margin: 0; } .choice-action { display: block; margin-top: 0.7rem; }
.optimization-nav a, .optimization-content button, .optimization-content .button-link,
.sticky-save button, .sticky-save .button-link, .confirm-check { min-height: 44px; }
.choice-action button { width: 100%; }

.table-scroll { max-width: 100%; overflow-x: auto; overscroll-behavior-inline: contain; }
.table-scroll:focus-visible { outline: 3px solid var(--accent); outline-offset: 2px; }
.table-scroll table { min-width: 36rem; }
/* 짧은 열이 글자 단위로 쪼개지지 않게 하는 최소 폭(긴 열은 그대로 줄바꿈) */
.table-scroll th, .table-scroll td { min-width: 4.5rem; }
.history-panel summary { min-height: 44px; display: flex; align-items: center; cursor: pointer; }
.preview-review { min-width: 0; max-width: 100%; }
.sticky-save { position: sticky; z-index: 4; bottom: 0.5rem; max-width: 100%; box-sizing: border-box;
  box-shadow: var(--shadow-bar), inset 0 2px var(--warn-line); }
.confirm-check { display: flex; gap: 0.55rem; align-items: center; margin: 0.7rem 0; }
.confirm-check input { width: 1.25rem; height: 1.25rem; }
/* 설정 토글 체크박스: 라벨 전체가 44px 터치 목표(마크업·문구·제출 동작은 그대로) */
label:not(.confirm-check):has(> input[type=checkbox]) { display: block; min-height: 44px; padding-block: 0.55rem; }
label:not(.confirm-check) > input[type=checkbox] { width: 1.25rem; height: 1.25rem; vertical-align: middle; margin-inline-end: var(--space-2); }
.approval-actions { align-items: stretch; }
.approval-actions > * { flex: 1 1 12rem; }
a:focus-visible, button:focus-visible, summary:focus-visible, input:focus-visible {
  outline: 3px solid var(--accent); outline-offset: 2px;
}
@media (max-width: 720px) {
  body { padding-inline: 0.75rem; }
  header > div { flex-wrap: wrap; align-items: flex-start; gap: 0.5rem; }
  header > div > .header-note { margin-left: 0; width: 100%; }
  nav { display: flex; flex-wrap: wrap; gap: 0.25rem 0.75rem; min-width: 0; }
  dl { grid-template-columns: minmax(7rem, auto) minmax(0, 1fr); }
  .card-current, .card-preview, .entry-list { grid-template-columns: minmax(0, 1fr); }
  button, summary, nav a, .filter-group a { min-height: 44px; display: inline-flex; align-items: center; }
  .detail-grid { grid-template-columns: 1fr; }
  .data-table { display: block; }
  .data-table caption { display: block; width: 100%; }
  .data-table thead { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px;
    overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
  .data-table tbody, .data-table tr { display: block; width: 100%; }
  .data-table tr > * { width: 100% !important; min-width: 0; }
  .data-table tr { background: var(--surface-2); border-radius: var(--radius-card); margin: 0.65rem 0; padding: 0.3rem 0.7rem; }
  .data-table td, .data-table th[scope="row"] { display: grid; grid-template-columns: minmax(7rem, 40%) minmax(0, 1fr);
    gap: 0.75rem; width: 100%; padding: 0.55rem 0; border-bottom: 1px solid var(--line-soft); }
  .data-table td::before, .data-table th[scope="row"]::before { content: attr(data-label); color: var(--ink-3);
    font-weight: 500; grid-column: 1; }
  .data-table td > *, .data-table th[scope="row"] > * { grid-column: 2; min-width: 0; }
  .data-table tr > :last-child { border-bottom: 0; }
  .number-cell { text-align: left; white-space: normal; }
  .line-item { justify-content: flex-start; font-size: 0.9rem; font-variant-numeric: tabular-nums; }
  .line-item > strong { text-align: left; }
  .detail-cell { text-align: left; white-space: normal; }
  .detail-cell .row-details summary { justify-content: flex-start; }
  .detail-cell .row-details[open] { width: calc(100% - 1.5rem); max-height: calc(100% - 3rem); }
  .actions { min-width: 0; }
  .optimization-title { display: block; }
  .status-strip { width: 100%; }
  .optimization-shell { display: block; }
  .optimization-nav { position: static; display: flex; flex-wrap: wrap; margin: 0.8rem 0; }
  .optimization-nav a { flex: 1 1 calc(50% - 0.35rem); min-width: 0; }
  .category-summary, .setting-values, .choice-grid, .optimization-steps { grid-template-columns: 1fr; }
  .sticky-save { bottom: 0; }
}
@media (max-width: 420px) {
  .status-strip { grid-template-columns: 1fr 1fr; min-width: 0; }
  .optimization-nav a { flex-basis: 100%; }
  .optimization-content section { padding-inline: 0.75rem; }
  .approval-actions > * { flex-basis: 100%; width: 100%; }
}
/* --- 모션(change a172 4단계) ------------------------------------------------
   원칙: 전환 대상은 transform 계열(translate·scale)·opacity 와 hover 색(background-color·color)뿐이고,
   hover 색은 --dur-1 이하. 레이아웃 속성·box-shadow·all·blur 는 쓰지 않음.
   주문·승인·청산·토글 등 상태를 바꾸는 POST 폼의 버튼과 승인 바는 전환 0(안전 불변식 4). */
@supports (animation-timing-function: linear(0, 1)) {
  :root {
    --ease-spring: linear(0, 0.5 7%, 1.143 14%, 1.287 21%, 1.097 29%, 0.934 36%, 0.923 43%, 0.986 50%, 1.024 57%, 1.019 64%, 1 71%, 0.992 79%, 0.996 86%, 1.001 93%, 1);
    --ease-pop: linear(0, 0.547 8%, 1.053 17%, 1.114 25%, 1.028 33%, 0.986 42%, 0.989 50%, 0.999 58%, 1.002 67%, 1.001 75%, 1 83%, 1 92%, 1);
    --ease-settle: linear(0, 0.488 10%, 0.933 20%, 1.045 30%, 1.028 40%, 1.005 50%, 0.998 60%, 0.999 70%, 1 80%, 1 90%, 1);
  }
}
nav a, .settings-tabs a, .optimization-nav a { transition: translate var(--dur-2) var(--ease-spring),
  background-color var(--dur-1) var(--ease-settle), color var(--dur-1) var(--ease-settle); }
nav a:hover, nav a:focus-visible, .settings-tabs a:hover, .settings-tabs a:focus-visible,
.optimization-nav a:hover, .optimization-nav a:focus-visible { translate: 0 var(--nudge); }
nav a:hover { color: var(--ink-1); }
.settings-tabs a:hover { background-color: var(--line-soft); }
.optimization-nav a:hover { background-color: var(--surface-2); }
/* 일반 버튼·링크 버튼: 누르면 작아졌다가 스프링으로 복귀. POST 폼·승인 영역 안의 버튼은 제외 */
button:not(:where(form:not([method=get]) *, form button, form input[type=submit], [formmethod], [formaction], .sticky-save *, .approval-actions *, .confirm-check *, .console-approval *)),
.button-link:not(:where(form:not([method=get]) *, form button, form input[type=submit], [formmethod], [formaction], .sticky-save *, .approval-actions *, .confirm-check *, .console-approval *)) {
  transition: scale var(--dur-2) var(--ease-spring), background-color var(--dur-1) var(--ease-settle), color var(--dur-1) var(--ease-settle); }
button:active:not(:disabled):not(:where(form:not([method=get]) *, form button, form input[type=submit], [formmethod], [formaction], .sticky-save *, .approval-actions *, .confirm-check *, .console-approval *)),
.button-link:active:not(:where(form:not([method=get]) *, form button, form input[type=submit], [formmethod], [formaction], .sticky-save *, .approval-actions *, .confirm-check *, .console-approval *)) {
  scale: var(--press); transition-duration: var(--dur-1); }
/* 제외 영역은 어떤 규칙이 나중에 들어와도 전환 0 으로 못 박음 */
form:not([method=get]) button, form:not([method=get]) .button-link, form button, form input[type=submit],
[formmethod], [formaction], .sticky-save, .sticky-save *, .approval-actions *,
.confirm-check, .confirm-check *, .console-approval * { transition: none; }
/* 상세 모달 열림: opacity + 개별 scale 속성. 위치용 transform 은 건드리지 않고(transform 전환 없음),
   transform-origin 을 레이아웃 박스 좌상단(= 화면 중앙)에 두어 중앙 기준으로 커짐. will-change 는 열린 동안 이 한 곳만 */
@keyframes pop-in { from { opacity: 0; scale: var(--pop-from); } }
/* 스크림: 열린 모달이 있는 셀의 ::before 가 화면 전체를 어둡게 덮음. pointer-events 를 꺼서 뒤의 어떤 버튼도 막지 않고,
   opacity 만 애니메이션함(흐림 효과는 쓰지 않음). 모달 박스 안쪽은 overflow 에 잘려서 셀 쪽에 둠 */
@keyframes scrim-in { from { opacity: 0; } }
.detail-cell:has(.row-details[open])::before { content: ""; position: fixed; inset: 0; z-index: 20; background: var(--scrim);
  pointer-events: none; animation: scrim-in var(--dur-3) var(--ease-settle) both; }
.detail-cell .row-details[open] { transform-origin: 0 0; animation: pop-in var(--dur-3) var(--ease-pop) both;
  will-change: opacity, scale; }
@media (prefers-reduced-motion: reduce) {
  :root { --dur-1: 0.01ms; --dur-2: 0.01ms; --dur-3: 0.01ms; --nudge: 0px; --press: 1; --pop-from: 1; }
}
/* --- 앰비언트 톤 배경과 hover 조명(change a172 6단계, CSS 만) ------------------------
   헤더·상태 스트립 뒤에 정적 저채도 radial-gradient 덩어리 둘을 두고, 움직이는 것은 덩어리의
   translate·rotate·opacity 뿐임(20~40초 루프). 그라디언트·배경 위치는 애니메이션하지 않음(매 프레임 paint 금지).
   색은 --tone(전용 틴트 토큰, 투명도 --tone-alpha 는 테마별)이고 --tone 은 상태 스트립이 이미 단어로도 말하는 엔진·데이터 톤을 따름: 정보 없는 장식이 아님.
   혼합 모드·필터·흐림 효과는 쓰지 않음. 덩어리는 pointer-events: none 이고 콘텐츠 아래(z-index: -1). */
body { --tone: var(--tint-accent); }
body:where(:has(.console-status [data-engine-state=running]):has(.console-status [data-data-tone=ok])) { --tone: var(--tint-ok); }
body:where(:has(.console-status :is([data-data-tone=warn], [data-engine-state=stopped]))) { --tone: var(--tint-warn); }
body:where(:has(.console-status [data-data-tone=stale])) { --tone: var(--tint-bad); }
body > header, .console-status { position: relative; isolation: isolate; }
/* 덩어리가 헤더 밖 본문 위로 번지면 h1 대비가 떨어지므로 헤더를 잘라 냄(.console-status 는 이미 overflow: hidden).
   overflow-clip-margin 은 nav 포커스 링(offset 2px + 3px)이 잘리지 않게 하는 여유 */
body > header { overflow: clip; overflow-clip-margin: 6px; }
body > header::before, body > header::after, .console-status::before, .console-status::after {
  content: ""; position: absolute; z-index: -1; pointer-events: none; border-radius: 50%;
  width: min(42%, 30rem); height: min(170%, 22rem); top: max(-35%, -7rem);
  background: radial-gradient(closest-side, color-mix(in srgb, var(--tone) var(--tone-alpha), transparent), transparent); }
body > header::before, .console-status::before { left: 8%; animation: amb-a 34s linear infinite alternate; }
body > header::after, .console-status::after { right: 8%; animation: amb-b 41s linear -9s infinite alternate; }
@keyframes amb-a { from { translate: -10% -6%; rotate: -6deg; opacity: 0.75; } to { translate: 10% 6%; rotate: 8deg; opacity: 1; } }
@keyframes amb-b { from { translate: 8% 5%; rotate: 5deg; opacity: 1; } to { translate: -9% -5%; rotate: -7deg; opacity: 0.7; } }
/* 자동 새로고침 화면(상태 스트립이 data-reload=on 을 렌더, 대시보드·검증 2초)은 정지 프레임: 재시작해 튀지 않게 */
body:has([data-reload=on]) :is(body > header, .console-status)::before,
body:has([data-reload=on]) :is(body > header, .console-status)::after { animation: none; }
@media (prefers-reduced-motion: reduce) {
  body > header::before, body > header::after, .console-status::before, .console-status::after { animation: none; }
}
@media (prefers-reduced-transparency: reduce) {
  body > header::before, body > header::after, .console-status::before, .console-status::after,
  :is(.optimization-nav a, .settings-tabs a)::after { background: none; }
}
/* hover 조명: ::after 하이라이트를 opacity 로만 페이드(포인터 추적 없음). 대상은 읽기 전용 이동 링크(최적화 nav·설정 탭)뿐임.
   POST 폼을 품은 카드(.tier-card·.choice-tile)에는 걸지 않고, 제외 선택자는 승인 영역에 나중에 링크가 생겨도 막는 방어 */
@media (hover: hover) {
  .optimization-nav a, .settings-tabs a { position: relative; isolation: isolate; }
  :is(.optimization-nav a, .settings-tabs a):not(:where(form:not([method=get]) *, .sticky-save *, .approval-actions *, .confirm-check *, .console-approval *))::after {
    content: ""; position: absolute; inset: 0; z-index: -1; pointer-events: none; border-radius: inherit; opacity: 0;
    background: radial-gradient(circle at 28% 0%, color-mix(in srgb, var(--accent) 12%, transparent), transparent 70%);
    transition: opacity var(--dur-2) var(--ease-settle); }
  :is(.optimization-nav a, .settings-tabs a):not(:where(form:not([method=get]) *, .sticky-save *, .approval-actions *, .confirm-check *, .console-approval *)):is(:hover, :focus-visible)::after { opacity: 1; }
}
/* 다크: 변수만 재정의함. 개별 규칙의 색 덮어쓰기는 두지 않음 */
@media (prefers-color-scheme: dark) {
  :root {
    --surface-0: #141311; --surface-1: #201f1b; --surface-2: #11100e;
    --line: #38362f; --line-soft: #2a2924; --line-strong: #756f64;
    --ink-1: #ece9e1; --ink-2: #cbc7bb; --ink-3: #a49f93;
    --accent: #9fbbd8;
    --ok: #93c7a0; --ok-line: #5a9068; --ok-bg: #1b2a1f;
    --warn: #e0b765; --warn-line: #a97b2d; --warn-bg: #2a2314;
    --bad: #ea9a90; --bad-line: #c25f52; --bad-bg: #2d1a17;
    --tint-accent: #7895dc; --tint-ok: #5bbe83; --tint-warn: #e4a722; --tint-bad: #d8565b; --tone-alpha: 15%;
    --lift: 0 1px 2px rgba(0, 0, 0, 0.4), 0 0 0 1px var(--line-soft);
    --shadow-pop: 0 18px 54px rgba(0, 0, 0, 0.6), 0 0 0 1px var(--line);
    --scrim: rgba(0, 0, 0, 0.55);
    --shadow-bar: 0 -5px 20px rgba(0, 0, 0, 0.5);
  }
}
{{end}}

{{define "head"}}<!doctype html>
<html lang="ko"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="referrer" content="same-origin">
{{if .Refresh}}<meta http-equiv="refresh" content="{{.RefreshSeconds}}">{{end}}
<title>tossctl console</title>
<style>{{template "style"}}</style>
</head><body>
<header><div>
  <strong>tossctl console</strong>
  {{/*
    Six items, in the order the engine actually works: 발굴 → 발주 → 보유 → 종결,
    then the settings that govern all four (change a055 §7). An order that can be
    learned is an order nobody has to remember.

    Twelve items were not a menu, they were a list of every screen that had ever
    been built, and two of the labels named a SECTION of a screen rather than the
    screen — "외부 종목 자동관리" pointed at /settings#adoption, so the Guardian
    limits, the operating toggles and the system update on the same screen were
    reachable from the navigation only by accident.

    Each item carries the question its screen answers. The labels alone do not
    separate 검증 콘솔 from 검증, 거래 이력 from 성과 이력, or 포지션 from
    포지션 정책 — the six screens that left the bar keep their routes and gain
    explicit entry points under 설정, which is the opposite of being hidden: none
    of them had a navigation entry of its own before.
  */}}
  <nav aria-label="주요 화면">
    <a href="/dashboard" {{if eq .Nav "overview"}}class="on" aria-current="page"{{end}}>개요<small>지금 무엇이 참인가</small></a>
    <a href="/signals" {{if eq .Nav "signals"}}class="on" aria-current="page"{{end}}>신호<small>후보와 차단 사유</small></a>
    <a href="/orders" {{if eq .Nav "orders"}}class="on" aria-current="page"{{end}}>주문<small>대기와 종결</small></a>
    <a href="/positions" {{if eq .Nav "positions"}}class="on" aria-current="page"{{end}}>포지션<small>보유와 보호</small></a>
    <a href="/history" {{if eq .Nav "history"}}class="on" aria-current="page"{{end}}>이력<small>왕복과 성과</small></a>
    <a href="/settings/daily" {{if eq .Nav "settings"}}class="on" aria-current="page"{{end}}>설정<small>상시 · 당일 · 전략 · 도구</small></a>
  </nav>
  <span class="muted header-note">127.0.0.1 전용 · 승인 외 읽기 전용</span>
</div>
{{template "statusstrip" .}}
</header>
<main>
{{end}}

{{/*
  "statusstrip" is the shell's four facts, in the same place on every screen.

  It is given the whole page rather than .Status because the reload cell reads
  .Refresh and .RefreshSeconds directly. Copying the period into the strip would
  make it a second statement of one fact, and the two would disagree the first
  time a screen changed its period — which is exactly the failure the strip
  exists to stop the rest of the console from having.
*/}}
{{define "statusstrip"}}
<dl class="console-status" aria-label="콘솔 상태">
  <div><dt>엔진</dt>
    <dd data-engine-state="{{.Status.EngineState}}">{{.Status.EngineText}}{{if .Status.EngineNote}}<small>{{.Status.EngineNote}}</small>{{end}}</dd></div>
  {{/*
    The open/closed decision is made here and not in Go. That is the whole point:
    the market-hours reading is advisory, the console never consults it before
    doing anything, and static_test.go keeps it that way by refusing any Go file
    that touches .Outside. A template is markup, so this is a render.
  */}}
  <div><dt>시장 세션</dt>
    <dd data-session-state="{{if .Status.Session.Outside}}closed{{else}}open{{end}}">{{.Status.Session.Label}}<small>advisory — 요일과 시각만 읽는다. 휴장일은 모르며 이 값으로 막는 것은 없다</small></dd></div>
  <div><dt>데이터 시각</dt>
    <dd data-data-mode="{{.Status.DataMode}}" data-data-tone="{{.Status.DataTone}}">{{.Status.DataText}}{{if eq .Status.DataTone "warn"}} <span class="status-pill">주의</span>{{else if eq .Status.DataTone "stale"}} <span class="status-pill">경고</span>{{end}}{{if .Status.DataNote}}<small>{{.Status.DataNote}}</small>{{end}}</dd></div>
  <div><dt>자동 재로드</dt>
    <dd data-reload="{{if .Refresh}}on{{else}}off{{end}}">{{if .Refresh}}{{.RefreshSeconds}}초마다{{else}}걸려 있지 않음<small>이 화면은 스스로 갱신하지 않는다</small>{{end}}</dd></div>
</dl>
{{if .Status.PendingApproval}}
<p class="notice console-approval" data-pending-approval="true"><strong>승인을 기다리는 검증 run이 있다.</strong>
승인 창은 짧고 지나가면 그 시장 창을 다시 기다려야 한다 —
<a href="{{.Status.PendingHref}}">검증 콘솔에서 승인</a>. 승인은 그 화면에서 한다.</p>
{{end}}
{{end}}

{{define "foot"}}
</main></body></html>
{{end}}

{{/*
  "hours" is the market-hours advisory (task 1.7 ②). It warns and it never
  blocks: the only thing measured is that a live order sent on a closed day comes
  back 422 order-hours-closed, and the window in which the broker DOES accept one
  is unmeasured. A tool that refused to start outside a hard-wired calendar would
  be asserting the half nobody has observed.
*/}}
{{define "hours"}}
{{if .Outside}}
<p class="notice"><strong>지금은 KR 정규장(평일 09:00–15:30 KST) 밖이다.</strong>
{{.Detail}}
<br>그래도 <strong>시작은 막지 않는다</strong> — 주문 접수 창의 실제 경계는 미측정이고, 휴장 달력을
도구에 박아 넣으면 관측하지 않은 것을 단언하게 된다.</p>
{{else}}
<p class="muted">{{.Detail}}</p>
{{end}}
{{end}}

{{define "refuse"}}<!doctype html>
<html lang="ko"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="referrer" content="same-origin"><title>거부됨 — tossctl console</title>
<style>{{template "style"}}</style></head><body><main>
<h1>{{.Title}}</h1>
<p class="danger">{{.Detail}}</p>
</main></body></html>
{{end}}

{{/*
  "restart" is the interstitial the browser holds while this process is replaced.
  It refreshes to a different URL than itself — the same origin, carrying the
  one-time handoff token — so it writes its own meta tag instead of the shared
  two-second one.
*/}}
{{define "restart"}}<!doctype html>
<html lang="ko"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="referrer" content="same-origin">
<meta http-equiv="refresh" content="{{.RefreshContent}}">
<title>재시작 중 — tossctl console</title>
<style>{{template "style"}}</style></head><body><main>
<h1>콘솔을 다시 시작하는 중</h1>
<p class="notice">같은 바이너리를 같은 포트로 다시 실행한다. <strong>{{.Delay}}초 뒤</strong> 이 페이지가
자동으로 새 프로세스로 돌아간다 — 돌아오지 않으면 아래 링크를 누르거나 터미널에 출력된 새 세션 링크를 열어라.</p>
<p><a href="{{.Target}}">지금 새 프로세스로 이동</a></p>
{{if .Note}}<p class="notice">{{.Note}}</p>{{end}}
<p class="muted">재시작이 하는 일은 하나뿐이다: <strong>새 프로세스 인스턴스</strong>를 만든다.
조건주문 존속 판정은 증거 기록의 <code>process.instance_id</code>(기동마다 새로 발급)를 보므로,
이것이 그 측정에 필요한 프로세스 경계다. 게이트·설정·계좌는 아무것도 바뀌지 않고, 배치 승인은
새 계획으로 처음부터 다시 받는다.</p>
</main></body></html>
{{end}}

{{/*
  "staleconsole" / "stalesoak" are the two halves of the stale-binary warning.
  They are advisory and they are quiet when there is nothing to say: a banner
  that is always on screen is a banner nobody reads.
*/}}
{{define "stalebinary"}}
{{if .ConsoleStale}}
<p class="notice"><strong>이 콘솔은 설치된 바이너리보다 오래되었다.</strong>
실행 중 {{.ConsoleAt}} · 설치됨 {{.InstalledAt}} — [콘솔 재시작]을 누르면 새 바이너리로 올라온다.</p>
{{end}}
{{if .SoakStale}}
<p class="notice"><strong>실행 중인 soak은 설치된 바이너리보다 오래되었다.</strong>
마지막 사이클이 보고한 빌드 {{.SoakBuildAt}} · 설치됨 {{.InstalledAt}} —
soak은 다음 사이클 경계에서 스스로 새 바이너리로 재실행한다. 기다릴 수 없으면 [soak 재시작].</p>
{{end}}
{{end}}

{{define "dashboard"}}
{{template "head" .}}
<h1>검증 콘솔</h1>
<p class="muted">이 화면은 로컬 파일만 읽는다. 네트워크 호출도, 설정 변경도 하지 않는다.
계좌 보유·exit 라인은 <a href="/positions">포지션</a>, 완결 거래는 <a href="/history">거래 이력</a>에 있다.</p>
{{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}
{{template "stalebinary" .Snap.Binary}}

<section>
  <h2>도구 프로세스</h2>
  <p class="muted">아래 두 버튼은 <strong>도구를 다시 띄우는 것</strong>이다 — 게이트·주문 능력·위험 한도와는
  무관하고, 승인 절차(새 계획을 보고 다시 승인)도 그대로다.</p>
  {{if .CanRestart}}
  <form method="post" action="/restart">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <button type="submit" class="secondary">콘솔 재시작</button>
  </form>
  {{else}}
  <p class="muted">콘솔 재시작 배선이 없다 — 종료(Ctrl-C) 후 직접 다시 시작하라.</p>
  {{end}}
  {{if .CanRestartSoak}}
  <form method="post" action="/soak/restart">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <button type="submit" class="secondary">soak 재시작</button>
  </form>
  {{else}}
  <p class="muted">soak 재기동 배선이 없다.</p>
  {{end}}
  <p class="muted">콘솔 재시작은 <strong>새 프로세스 인스턴스</strong>를 만든다 — 프로세스당 1회 검증 상한이
  초기화되고, 조건주문 존속 측정에 필요한 프로세스 경계가 생긴다. soak 재시작은 조회 전용 서베이를
  SIGINT 후 다시 띄우는 것이고 자격증명은 건드리지 않는다.</p>
</section>

<section>
  <h2>엔진 런타임</h2>
  {{with .Snap.Engine}}
  {{if .Wired}}
  <dl>
    <dt>상태</dt>
    <dd>{{if .Running}}<span class="ok">실행 중</span>{{if .PID}} (pid {{.PID}}){{end}}{{else}}<span class="muted">정지</span>{{end}}</dd>
    <dt>기동 시각</dt><dd>{{.StartedAtText}}</dd>
    <dt>마지막 갱신</dt><dd>{{.RefreshedAtText}}</dd>
    <dt>실행 중 빌드</dt><dd>{{.BuildAt}}</dd>
    <dt>활성 마커</dt><dd><code>{{.Marker}}</code></dd>
  </dl>
  {{if .Stale}}
  <p class="notice"><strong>실행 중인 엔진은 설치된 바이너리보다 오래되었다.</strong>
  엔진이 보고한 빌드 {{.BuildAt}} — 새 바이너리로 올리려면 [엔진 정지] 뒤 [엔진 시작].</p>
  {{end}}
  {{else}}
  <p class="muted">엔진 상태 배선이 없다 (활성 마커 경로 미지정).</p>
  {{end}}

  {{if .CanStart}}
  <form method="post" action="/engine/start">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit" class="secondary">엔진 시작</button>
  </form>
  {{end}}
  {{if .CanStop}}
  <form method="post" action="/engine/stop">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit" class="secondary">엔진 정지</button>
  </form>
  {{end}}
  {{if not (or .CanStart .CanStop)}}
  <p class="muted">엔진 기동/정지 배선이 없다 — 터미널에서 <code>tossctl engine run</code>으로 직접 띄워라.</p>
  {{end}}

  {{if .Note}}
  <p class="notice">마지막 응답 ({{.NoteAtText}}):</p>
  <pre>{{.Note}}</pre>
  {{end}}

  <p class="muted">이 두 버튼은 <strong>프로세스를 켜고 끄는 것</strong>이다. 엔진이 주문 능력을 갖는지는
  §0.7로 승인된 게이트 설정과 엔진 자신의 기동 인터록이 결정한다 — 콘솔은 게이트를 켜지 않고,
  인터록을 우회하지도 못한다. 인터록이 미충족이면 엔진 프로세스가 거부하고, 그 사유가 위에 그대로 뜬다.
  정지는 SIGTERM(루프 완주·journal 정합 close)이고, 활성 마커는 자문 신호다 —
  배타는 엔진이 쥔 journal 디렉터리 flock이 담당하며 크래시한 엔진은 최대 {{.StaleAfter}} 동안
  실행 중으로 보인다.</p>
  {{end}}
</section>

<section>
  <h2>soak (조회 전용 서베이)</h2>
  {{with .Snap.Soak}}
  {{if .Error}}<p class="danger">{{.Error}}</p>{{end}}
  {{if .Present}}
  <dl>
    <dt>계좌</dt><dd>{{.AccountRef}}</dd>
    <dt>사이클</dt><dd>{{.Cycles}}</dd>
    <dt>연속일</dt><dd>{{.StreakDays}} / {{.MinDays}}일</dd>
    <dt>기간</dt><dd>{{.FirstAt.Format "2006-01-02 15:04Z"}} → {{.LastAt.Format "2006-01-02 15:04Z"}}</dd>
    <dt>attest 가능</dt><dd>{{if .Ready}}<span class="ok">예</span>{{else}}<span class="bad">아니오</span>{{end}}</dd>
  </dl>
  {{if .Reasons}}<ul>{{range .Reasons}}<li>{{.}}</li>{{end}}</ul>{{end}}
  <div class="table-scroll" role="region" aria-label="soak (조회 전용 서베이)" tabindex="0"><table>
    <tr><th>날짜</th><th>사이클</th><th>인증 성공</th><th>인증 실패</th><th>판정</th></tr>
    {{range .Days}}
    <tr><td>{{.Date}}</td><td>{{.Cycles}}</td><td>{{.CredentialOK}}</td><td>{{.AuthFailures}}</td>
        <td>{{if .Pass}}<span class="ok">pass</span>{{else}}<span class="bad">fail</span>{{end}}</td></tr>
    {{end}}
  </table></div>
  {{else}}
  <p class="muted">아직 기록이 없다. <code>tossctl soak run</code>으로 시작하라. ({{.Record}})</p>
  {{end}}
  {{end}}
</section>

<section>
  <h2>capability attestation</h2>
  {{with .Snap.Attestation}}
  {{if .Present}}
  <dl>
    <dt>계좌</dt><dd>{{.AccountRef}}</dd>
    <dt>soak 일수</dt><dd>{{.SoakDays}}</dd>
    <dt>발급 / 만료</dt><dd>{{.IssuedAt.Format "2006-01-02"}} → {{.ExpiresAt.Format "2006-01-02"}}</dd>
    <dt>상태</dt><dd>{{if .Usable}}<span class="ok">유효</span>{{else}}<span class="bad">사용 불가</span>{{end}}</dd>
    <dt>검증된 endpoint</dt><dd>{{range .Endpoints}}{{.}}<br>{{end}}</dd>
  </dl>
  {{else}}
  <p class="muted">attestation 파일이 없다 ({{.Path}}). <code>tossctl soak attest</code>가 조건을 만족할 때 쓴다.</p>
  {{end}}
  {{if .Reasons}}
  <p class="notice">게이트가 열리지 않는 이유:</p>
  <ul>{{range .Reasons}}<li>{{.}}</li>{{end}}</ul>
  {{end}}
  {{if .ExpiryHorizon}}<p class="notice" role="status">attestation 만료까지 {{.ExpiryHorizon}} 남았다.{{if .ExpiryWarning}} 72시간 이내다.{{end}}</p>{{end}}
  <p class="notice" role="status">renewal 진단: {{.RenewalState}}{{if not .RenewalAttemptedAt.IsZero}} · 마지막 시도 {{.RenewalAttemptedAt.Format "2006-01-02 15:04Z"}} ({{.RenewalAge}} 전){{end}}</p>
  {{if .RenewalReasons}}<ul>{{range .RenewalReasons}}<li>{{.}}</li>{{end}}</ul>{{end}}
  <p class="muted">이 콘솔은 게이트를 켜지 않는다. 자동매매 토글은 사람이 직접 승인한다.</p>
  {{end}}
</section>

<section>
  <h2>실계좌 검증 진행</h2>
  {{with .Snap.Verify}}
  {{if .Error}}<p class="danger">{{.Error}}</p>{{end}}
  <dl>
    <dt>기록</dt><dd>{{.Record}}</dd>
    <dt>계좌</dt><dd>{{if .AccountRef}}{{.AccountRef}}{{else}}(없음){{end}}</dd>
    <dt>완료 / 전체</dt><dd>{{.Done}} / {{.Total}} 단계</dd>
  </dl>
  {{if .Steps}}
  <div class="table-scroll" role="region" aria-label="실계좌 검증 진행" tabindex="0"><table>
    <tr><th>단계</th><th>이름</th><th>판정</th><th>사유</th></tr>
    {{range .Steps}}<tr><td><code>{{.Step}}</code></td><td>{{stepLabel .Step}}</td><td>{{verdict .Verdict}}</td><td>{{.Reason}}</td></tr>{{end}}
  </table></div>
  {{end}}
  {{if .AwaitingRestart}}
  <p class="notice"><code>{{.AwaitingRestart}}</code> ({{stepLabel .AwaitingRestart}}) 단계는 <strong>새 프로세스</strong>를 기다린다 — 위의 [콘솔 재시작] 뒤 이어하기.</p>
  {{end}}
  {{if .Outstanding}}
  <p class="danger">계좌에 아직 살아 있는 객체:</p>
  <ul>{{range .Outstanding}}<li>{{.Kind}} {{.ID}} ({{.Symbol}}){{if .Deliberate}} — 존속 측정을 위해 의도적으로 남긴 것{{end}}</li>{{end}}</ul>
  {{end}}
  {{end}}
  <p><a href="/verify">검증 화면으로</a></p>
</section>
{{template "foot" .}}
{{end}}

{{define "verify"}}
{{template "head" .}}
<h1>검증 <span class="muted">실계좌 · {{.Market}} 시장</span></h1>
<p class="muted">검증 능력은 계좌와 <strong>시장</strong>의 속성이다 — 아래 모든 것은 {{.Market}} 기록만 읽는다.
한 시장의 판정은 다른 시장의 단계를 완료로 만들지 않는다.
{{if .Snap.IsUS}}<a href="/verify?market=KR">KR 검증으로</a>{{else}}<a href="/verify?market=US">US 검증으로</a>{{end}}
· <a href="/report?market={{.Market}}">이 시장 리포트</a></p>
{{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}

{{if .Spent}}
<section>
  <p class="notice"><strong>이 프로세스는 이미 검증을 수행했다.</strong> 조건주문 존속은 등록한 프로세스가
  죽은 뒤에만 관측할 수 있으므로 이 경계는 우회하지 않는다. 다른 시장을 이어서 측정할 때도 마찬가지다 —
  재시작한 뒤 그 시장 화면에서 시작하라.</p>
  {{if .CanRestart}}
  <p>아래 버튼이 콘솔을 같은 포트로 다시 띄운다 — 돌아오면 [이어하기]를 누르면 된다.</p>
  <form method="post" action="/restart">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <button type="submit">콘솔 재시작</button>
  </form>
  {{else}}
  <p>콘솔을 종료(Ctrl-C)하고 다시 시작한 뒤 이어하기를 눌러라.</p>
  {{end}}
</section>
{{end}}

{{if .Run}}{{with .Run}}
<section>
  <h2>실행 {{.ID}} <span class="muted">{{.Market}} 시장 · 시작 {{.StartedAt.Format "15:04:05Z"}}</span></h2>
  {{if .Remeasuring}}
  <p class="notice"><strong>재측정 {{len .Redo}}단계</strong> — {{.RedoList}}. 판정이 pass인 단계는
  건드리지 않는다. 아래 목록을 승인해야만 요청이 나간다.</p>
  {{end}}

  {{if .Awaiting}}
  <p class="danger"><strong>여기서부터 실제 계좌에 요청이 나간다.</strong> 아래 목록이 이 실행이 보낼 수 있는
  전부다. 목록에 없는 요청은 전송되지 않으며, 보내야 하는 상황이 되면 실행이 멈춘다.</p>
  <pre>{{$.Prompt}}</pre>
  <form method="post" action="/verify/approve">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit">위 목록을 승인하고 실행</button>
  </form>
  <form method="post" action="/verify/abort">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit" class="secondary">거부</button>
  </form>
  <p class="muted">승인은 이 화면의 클릭 한 번이다 — 세션 토큰과 CSRF 토큰을 갖춘 이 폼의 제출만 승인이 된다.
  승인 창(5분)이 지나면 아무것도 전송되지 않고 이 실행은 중단된다. 단계별 확인(--confirm-each)은 웹에서는
  제공하지 않는다 — 콘솔은 배치 승인 전용이다.</p>
  {{end}}

  {{if .Approved}}<p class="ok">승인됨 {{.ApprovedAt.Format "15:04:05Z"}} — 목록에 있는 요청만 나간다.</p>{{end}}
  {{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}

  {{if .Lines}}
  <h2>진행</h2>
  <pre>{{range .Lines}}{{.}}
{{end}}</pre>
  {{end}}

  {{if .Done}}
  <h2>결과</h2>
  {{if .Err}}<p class="danger">{{.Err}}</p>{{end}}
  {{if .Summary.Outcomes}}
  <div class="table-scroll" role="region" aria-label="결과" tabindex="0"><table>
    <tr><th>단계</th><th>이름</th><th>판정</th><th>사유</th></tr>
    {{range .Summary.Outcomes}}<tr><td><code>{{.Step}}</code></td><td>{{stepLabel .Step}}</td><td>{{verdict .Verdict}}</td><td>{{.Reason}}</td></tr>{{end}}
  </table></div>
  {{end}}
  {{if .Summary.Halt}}<p class="notice">중단: {{.Summary.Halt}}</p>{{end}}
  {{if .Summary.Outstanding}}
  <p class="danger">계좌에 아직 살아 있는 객체:</p>
  <ul>{{range .Summary.Outstanding}}<li>{{.Kind}} {{.ID}} ({{.Symbol}}){{if .Deliberate}} — 의도적으로 남긴 것, 이어하기가 취소한다{{else}} — <strong>취소되지 않았다. 직접 처리하라.</strong>{{end}}</li>{{end}}</ul>
  {{end}}
  {{if .AwaitingRestart}}
  <p class="notice"><strong>여기서 프로세스 경계다.</strong> 조건주문이 이 프로세스보다 오래 사는지가
  다음 측정이므로, 같은 프로세스에서 이어갈 수는 없다.</p>
  {{if $.CanRestart}}
  <form method="post" action="/restart">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit">콘솔 재시작 후 이어하기</button>
  </form>
  {{else}}
  <p>콘솔을 종료(Ctrl-C)하고 다시 시작한 뒤 이어하기를 눌러라.</p>
  {{end}}
  {{end}}
  <p><a href="/report">리포트 보기</a></p>
  {{end}}
</section>
{{end}}{{end}}

{{if .ShowStart}}
{{template "hours" .Snap.Session}}

<section>
  {{if .Resuming}}
  <p class="notice">이 기록에는 이미 검증이 있다. 시작하면 <strong>이어서</strong> 진행한다 —
  판정이 끝난 단계는 다시 주문을 내지 않는다.</p>
  {{end}}
  {{if .Snap.Verify.AwaitingRestart}}
  <p class="notice"><code>{{.Snap.Verify.AwaitingRestart}}</code> ({{stepLabel .Snap.Verify.AwaitingRestart}}) 단계가 새 프로세스를 기다리고 있었다. 이 콘솔이 그 새 프로세스다.</p>
  {{end}}
  {{if .Snap.Verify.Cleanup}}
  <p class="notice"><strong>이전 실행이 남긴 객체 {{.Snap.Verify.CleanupCount}}건</strong>을 이 실행이 먼저 취소한다 —
  승인 목록의 첫 줄에 표시된다. 그것이 노출 상한을 채우고 있는 동안에는 아래 단계들이 아무것도 보낼 수 없다.</p>
  {{end}}
  <p>시작하면 이 실행이 보낼 수 있는 <strong>모든 라이브 요청의 목록</strong>이 표시되고,
  그 목록을 승인하는 버튼을 누르기 전에는 아무것도 전송되지 않는다.</p>
  {{$nothingToResume := and .Resuming (not .Snap.Verify.Pending) (not .Snap.Verify.Cleanup)}}
  {{if $nothingToResume}}
  <p class="notice"><strong>이어할 단계가 없다</strong> — 모든 단계에 판정이 있어 이어하기는 아무것도
  측정하지 않는다(2026-07-27에 이 버튼이 두 번 눌려 장중 창이 측정 0건으로 끝났다).
  {{if .Snap.Verify.Redo}}다시 측정하려면 아래 <strong>재측정</strong>을 사용하라.{{end}}</p>
  {{end}}
  <form method="post" action="/verify/start">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <input type="hidden" name="market" value="{{.Market}}">
    <input type="hidden" name="mode" value="resume">
    <button type="submit" {{if or .Spent $nothingToResume}}disabled{{end}}>{{if .Resuming}}이어하기{{else}}검증 시작{{end}}</button>
  </form>
</section>

{{with .Snap.Verify}}
{{if .Redo}}
<section>
  <h2>재측정</h2>
  <p>재측정 대상은 <strong>{{.RedoCount}}개</strong>다 — 마지막 판정이 <code>fail</code>·<code>skipped</code>인
  단계, 그리고 <strong>통과했지만 그 통과가 만든 조건주문이 사라져 아래 단계들이 측정할 수 없게 된 단계</strong>다.
  이어하기는 이 단계들을 건너뛴다(판정이 이미 terminal이므로). 재측정은 이 단계들만 다시 시도한다.</p>
  <div class="table-scroll" role="region" aria-label="재측정" tabindex="0"><table>
    <tr><th>단계</th><th>이름</th><th>마지막 판정</th><th>사유</th></tr>
    {{$redo := .Redo}}
    {{range .Steps}}{{if inRedo $redo .Step}}<tr><td><code>{{.Step}}</code></td><td>{{stepLabel .Step}}</td><td>{{verdict .Verdict}}</td><td>{{.Reason}}</td></tr>{{end}}{{end}}
  </table></div>
  <p class="muted">대상: {{.RedoList}}</p>
  <p class="notice"><code>deferred</code>·<code>refused</code> 단계는 대상이 아니다. <code>pass</code>도
  대상이 아니다 — 이미 측정된 속성을 위해 실주문을 다시 내지 않는다. 단 하나의 예외는
  <strong>그 통과가 확립한 성질이 더 이상 참이 아닌 경우</strong>다: 조건주문 등록이 남긴 조건주문이
  사라지면 존속·정정·취소는 영원히 <code>skipped</code>가 되므로, 그 전제를 다시 세우는 것은
  아는 것을 다시 재는 것이 아니다. 대상 목록은 폼이 아니라 <strong>증거 기록</strong>에서 계산한다.
  재측정도 계획을 처음부터 다시 만들고 <strong>그 계획을 다시 승인</strong>해야 요청이 나간다.
  설계상 생략되는 단계(보유 0, <code>--include-ttl-edge</code> 옵트인)는 preflight가 다시 걸러 무해하다.</p>
  <form method="post" action="/verify/start">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <input type="hidden" name="market" value="{{$.Market}}">
    <input type="hidden" name="mode" value="redo">
    <button type="submit" {{if $.Spent}}disabled{{end}}>재측정 {{.RedoCount}}단계</button>
  </form>
</section>
{{end}}
{{end}}

<section>
  <h2>단계 목록</h2>
  <p class="muted"><code>tossctl verify run --list</code>와 같은 내용이다. 계좌를 건드리지 않는다.</p>
  <pre>{{.Steps}}</pre>
</section>
{{end}}
{{template "foot" .}}
{{end}}

{{define "report"}}
{{template "head" .}}
<h1>리포트</h1>
{{if .Error}}<p class="danger">{{.Error}}</p>{{end}}
<p><a href="/report.json">JSON 내려받기</a> — 측정된 속성과 미검증 목록 전체.</p>
<section><pre>{{.Text}}</pre></section>
{{template "foot" .}}
{{end}}
`
