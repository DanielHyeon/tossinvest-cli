# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` — **구현 후 재생성**(:210–525, 분기 38). 편집 전 base `634cf3c5` 는 :211–539, 분기 43·반환 21·defer 3.
- Risk scan: `risk-pattern-report.md`

## 편집 전후 대조 (분기 번호는 위치다)

편집 전 B33–B37(:395–409, lifecycle dial 블록 — descriptor stat · Dial 실패 경고 · commander 고정 ·
비부재 stat 경고)이 **삭제**되고 B32 안의 한 줄(`positionPolicyCommander =
consolePositionPolicyCommanderFor(ctx, engineDir, cmd.ErrOrStderr())`)이 됐다. 그 뒤 분기는 번호가
다섯씩 당겨진다: 편집 전 B38–B42(전략 projection dial, a115 몫) → 편집 후 **B33–B37**, 편집 전 B43(:508
soakBoot) → 편집 후 **B38**. B1–B32 는 번호·종류 그대로다(줄은 import 한 줄 삭제로 1 당겨짐).
`positionpolicyrpc` import 가 console.go 에서 사라졌다(dial 은 새 파일의 wrapper resolve 에만 있다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 콘솔 수명(signal.NotifyContext) | B1 | wrapper 의 시도·펌프가 이 ctx 로 끝난다 |
| `engineDir` | `engineJournalDir(root)` 또는 "" | B14 | "" 면 commander nil = 진짜 미배선 |

불변식: 콘솔은 엔진 journal 을 직접 쓰지 않고 인증된 loopback client 만 받는다. a114 는 client 를 **언제**
얻는가만 바꾼다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if ctx == nil {` (:212) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B2 | `if err != nil {` (:221) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B3 | `if err != nil {` (:226) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B4 | `if err != nil {` (:230) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B5 | `if err != nil {` (:234) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B6 | `if err != nil {` (:238) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B7 | `if err != nil {` (:242) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B8 | `if err != nil {` (:247) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B9 | `if journalPath != "" {` (:255) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B10 | `if err != nil {` (:257) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B11 | `} else {` (:260) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B12 | `if err != nil {` (:264) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B13 | `} else {` (:267) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B14 | `if dir, derr := engineJournalDir(root); derr == nil {` (:278) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B15 | `} else {` (:281) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B16 | `if os.Getenv("TOSSOS_CONTAINER") == "1" {` (:289) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B17 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` (:292) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B18 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` (:292) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B19 | `} else {` (:294) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B20 | `if cerr != nil {` (:296) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B21 | `} else {` (:304) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B22 | `if updater, uerr := localupdate.New(self); uerr != nil {` (:298) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B23 | `} else {` (:300) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B24 | `if updater != nil {` (:307) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B25 | `if uerr != nil {` (:311) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B26 | `} else {` (:313) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B27 | `if engineDir != "" {` (:320) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B28 | `if err != nil {` (:323) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B29 | `if engineBoot != nil {` (:337) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B30 | `if engineBootNote != "" {` (:344) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B31 | `if soakBoot != nil {` (:365) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B32 | `if engineDir != "" {` (:393) — engineDir 해석됨 | **a114**: 안에서 `consolePositionPolicyCommanderFor` 한 줄이 commander 를 만든다(언제나 non-nil) | — | `TestTheConsoleAttachesWhenTheEngineStartsLater` · `TestRunConsoleNeverDialsTheLifecycleItself` |
| B33 | `if _, statErr := os.Stat(strategyDescriptor); statErr == nil {` (:398) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B34 | `} else if !errors.Is(statErr, os.ErrNotExist) {` (:405) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B35 | `if dialErr != nil {` (:400) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B36 | `} else {` (:402) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B37 | `} else if !errors.Is(statErr, os.ErrNotExist) {` (:405) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |
| B38 | `if soakBoot != nil {` (:494) | 편집 밖 | — | not-applicable: a114 편집 밖 부팅 배관 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consolePositionPolicyCommanderFor` (새 파일) | lifecycle commander — 부팅 1회 동기 해석 + 백그라운드 재부착 | 부재는 조용히, 그 밖의 해석 실패는 경고 한 줄 | AST call (B32 안) |
| `positionpolicyrpc.Dial` | **runConsole 에서 사라짐** | — | 편집 전 AST :397 |

## State mutations and fallbacks

- 편집 전: 부재·dial 실패는 commander nil 로 굳었다(영구 미배선). 편집 후: engineDir 가 있으면 non-nil, 부착 전
  호출은 연결 없는 detached 오류, 펌프·요청이 백그라운드 시도를 깨운다.

## Safety conclusion

- Safe edit boundary: 블록 하나 → 한 줄 + import 한 줄 삭제. 나머지 분기·defer·반환 무변경.
- High-risk impact: no — 명령은 재전송하지 않는다(`TestTheLifecycleWrapperNeverResendsACommand`).
