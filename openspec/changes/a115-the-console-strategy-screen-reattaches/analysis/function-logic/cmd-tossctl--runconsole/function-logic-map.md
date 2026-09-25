# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (편집 전 base `8688f74f`, :210–525, 분기 38)
- Risk scan: `risk-pattern-report.md`

a115 편집 지점은 B32 안의 **전략 projection dial 블록 B33–B37**(a114 이후 번호) 하나다. `strategyRuntime = consoleStrategyRuntimeReaderFor(ctx, engineDir, cmd.ErrOrStderr())` 한 줄이 된다. a114 가 바꾼 lifecycle 줄(B32 안)과 나머지 분기는 편집 밖이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engineDir` | `engineJournalDir(root)` 또는 "" | B14 | "" 면 전략 reader nil = 진짜 미배선 |
| 전략 descriptor | 있음/없음/stat 오류 | B33 | 편집 전: 없음·dial 실패·stat 오류 전부 nil 로 접힘 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if ctx == nil {` (:212) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B2 | `if err != nil {` (:221) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B3 | `if err != nil {` (:226) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B4 | `if err != nil {` (:230) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B5 | `if err != nil {` (:234) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B6 | `if err != nil {` (:238) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B7 | `if err != nil {` (:242) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B8 | `if err != nil {` (:247) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B9 | `if journalPath != "" {` (:255) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B10 | `if err != nil {` (:257) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B11 | `} else {` (:260) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B12 | `if err != nil {` (:264) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B13 | `} else {` (:267) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B14 | `if dir, derr := engineJournalDir(root); derr == nil {` (:278) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B15 | `} else {` (:281) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B16 | `if os.Getenv("TOSSOS_CONTAINER") == "1" {` (:289) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B17 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` (:292) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B18 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` (:292) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B19 | `} else {` (:294) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B20 | `if cerr != nil {` (:296) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B21 | `} else {` (:304) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B22 | `if updater, uerr := localupdate.New(self); uerr != nil {` (:298) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B23 | `} else {` (:300) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B24 | `if updater != nil {` (:307) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B25 | `if uerr != nil {` (:311) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B26 | `} else {` (:313) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B27 | `if engineDir != "" {` (:320) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B28 | `if err != nil {` (:323) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B29 | `if engineBoot != nil {` (:337) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B30 | `if engineBootNote != "" {` (:344) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B31 | `if soakBoot != nil {` (:365) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |
| B32 | `if engineDir != "" {` (:393) | a115 후 이 안에서 전략 reader 는 wrapper(언제나 non-nil) | — | `TestTheConsoleStrategyScreenRecoversWhenTheEngineStartsLater` |
| B33 | `if _, statErr := os.Stat(strategyDescriptor); statErr == nil {` (:398) — 전략 descriptor stat 성공 | **a115 가 지우는 블록** | 부팅 1회 dial → wrapper | `TestRunConsoleNeverDialsTheStrategyProjectionItself` |
| B34 | `} else if !errors.Is(statErr, os.ErrNotExist) {` (:405) — 비부재 stat 오류 else-if | **a115 가 지우는 블록** | 부팅 1회 dial → wrapper | `TestRunConsoleNeverDialsTheStrategyProjectionItself` |
| B35 | `if dialErr != nil {` (:400) — Dial 실패 → 경고 + **nil 접힘** | **a115 가 지우는 블록** | 부팅 1회 dial → wrapper | `TestRunConsoleNeverDialsTheStrategyProjectionItself` |
| B36 | `} else {` (:402) — Dial 성공 → client 고정 | **a115 가 지우는 블록** | 부팅 1회 dial → wrapper | `TestRunConsoleNeverDialsTheStrategyProjectionItself` |
| B37 | `} else if !errors.Is(statErr, os.ErrNotExist) {` (:405) — 비부재 stat 오류 경고 + nil | **a115 가 지우는 블록** | 부팅 1회 dial → wrapper | `TestRunConsoleNeverDialsTheStrategyProjectionItself` |
| B38 | `if soakBoot != nil {` (:494) | 편집 밖 | — | not-applicable: a115 편집 밖 부팅 배관 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyprojectionrpc.Dial` (편집 전 B33 안) | 부팅 1회 | 200ms probe · 재시도 없음 | AST call — a115 후 runConsole 에서 사라진다 |
| `console.ListenAndServe` | 서버 | — | Options.StrategyRuntime |

## State mutations and fallbacks

- 편집 전: dial 실패를 nil 로 접는다(fallback = dormant 오귀속, 영구). 편집 후: 부재=nil 자리·도달 불가=sentinel·live 를 wrapper 가 들고 재부착한다.

## Safety conclusion

- Safe edit boundary: B33–B37 → 한 줄. 나머지 무변경.
- High-risk impact: no — 조회 전용 projection client.
