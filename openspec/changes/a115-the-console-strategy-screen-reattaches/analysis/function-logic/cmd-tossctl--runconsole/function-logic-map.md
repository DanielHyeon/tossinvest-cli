# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` — **구현 후 재생성**(revision current, :209–517, 분기 33).
  편집 전 base `8688f74f` 는 :210–525, 분기 38.
- Risk scan: `risk-pattern-report.md`

## 편집 전후 대조 (분기 번호는 위치다 — 옛/새 ast.json 을 difflib 으로 정렬, 손 재번호 아님)

정렬 결과(분기 kind + 소스 줄 텍스트 시퀀스, `difflib.SequenceMatcher`): 편집 전 **B33–B37**(:398–405, 전략
descriptor stat · Dial · Dial 실패 경고+nil 접힘 · client 고정 · 비부재 stat 경고+nil)은 **delete**, 편집 전
**B38**(`if soakBoot != nil {`, :494) → 편집 후 **B33**(:486). B1–B32 는 번호·종류·텍스트 그대로다(줄은
`strategyprojectionrpc` import 한 줄 삭제로 1 당겨짐). 삭제된 블록은 B32 안의 한 줄
`strategyRuntime = consoleStrategyRuntimeReaderFor(ctx, engineDir, cmd.ErrOrStderr())` 가 됐다. dial 은 이제
새 파일 `console_strategy_attach.go` 의 `resolveConsoleStrategyRuntime`(부팅 1회 + wrapper 재시도)에만 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 콘솔 수명(signal.NotifyContext) | B1 | wrapper 의 시도·펌프가 이 ctx 로 끝난다(`TestTheConsoleStrategyPumpReturnsWhenTheConsoleEnds`) |
| `engineDir` | `engineJournalDir(root)` 또는 "" | B14 | "" 면 전략 reader nil interface = 진짜 미배선(dormant) |
| 전략 descriptor | 있음/없음/stat 오류 | wrapper 해석 | 편집 후: 없음 → 부재(dormant, 조용히) · stat 오류 → 경고+부재(선언된 접힘) · dial 실패 → sentinel(도달 불가) · 성공 → live |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if ctx == nil {` (:211) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B2 | `if err != nil {` (:220) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B3 | `if err != nil {` (:225) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B4 | `if err != nil {` (:229) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B5 | `if err != nil {` (:233) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B6 | `if err != nil {` (:237) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B7 | `if err != nil {` (:241) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B8 | `if err != nil {` (:246) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B9 | `if journalPath != "" {` (:254) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B10 | `if err != nil {` (:256) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B11 | `} else {` (:259) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B12 | `if err != nil {` (:263) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B13 | `} else {` (:266) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B14 | `if dir, derr := engineJournalDir(root); derr == nil {` (:277) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B15 | `} else {` (:280) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B16 | `if os.Getenv("TOSSOS_CONTAINER") == "1" {` (:288) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B17 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` (:291) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B18 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` (:291) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B19 | `} else {` (:293) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B20 | `if cerr != nil {` (:295) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B21 | `} else {` (:303) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B22 | `if updater, uerr := localupdate.New(self); uerr != nil {` (:297) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B23 | `} else {` (:299) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B24 | `if updater != nil {` (:306) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B25 | `if uerr != nil {` (:310) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B26 | `} else {` (:312) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B27 | `if engineDir != "" {` (:319) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B28 | `if err != nil {` (:322) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B29 | `if engineBoot != nil {` (:336) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B30 | `if engineBootNote != "" {` (:343) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B31 | `if soakBoot != nil {` (:364) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B32 | `if engineDir != "" {` (:392) — engineDir 해석됨 | **a115**: 안에서 `consoleStrategyRuntimeReaderFor` 한 줄이 전략 reader 를 만든다(언제나 non-nil wrapper). a114 의 lifecycle 줄도 이 안 | — | `TestRunConsoleNeverDialsTheStrategyProjectionItself` · `TestTheConsoleStrategyScreenRecoversWhenTheEngineStartsLater` · `TestTheConsoleStrategyScreenShowsADeadDescriptorAsUnreachable` |
| B33 | `if soakBoot != nil {` (:486) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleStrategyRuntimeReaderFor` (새 파일) | 전략 reader — 부팅 1회 동기 해석 + 백그라운드 재부착(펌프) | 부재는 조용히, stat 오류·dial 실패는 경고 한 줄(「dormant로 뜬다」 없음), 재시도 경고는 discard | AST call (B32 안) |
| `strategyprojectionrpc.Dial` | **runConsole 에서 사라짐** | — | 편집 전 AST :399 · `TestRunConsoleNeverDialsTheStrategyProjectionItself` |
| `console.ListenAndServe` | 서버 | — | `Options.StrategyRuntime` 에 wrapper |

## State mutations and fallbacks

- 편집 전: dial 실패·stat 오류를 nil 로 접었다(fallback = dormant 오귀속, 부팅 1회라 영구). 편집 후: 부재=nil 자리 ·
  도달 불가=sentinel · live 를 wrapper 가 들고, 펌프(간격 절반마다 무조건 wake)와 화면의 presence 질문이 재부착을 깨운다.

## Safety conclusion

- Safe edit boundary: B33–B37 블록 → B32 안 한 줄 + import 한 줄 삭제. 나머지 분기·defer·반환 무변경.
- High-risk impact: no — 조회 전용 projection client, 주문·토글 경로 접촉 0.
