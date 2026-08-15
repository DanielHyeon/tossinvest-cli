# Branch Test Map: `runVerifyRun`

- Source: `cmd/tossctl/verify.go:272-400`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `cmd/tossctl/verify.go:278` — `if opts.list {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `cmd/tossctl/verify.go:284` — `if ctx == nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `cmd/tossctl/verify.go:293` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `cmd/tossctl/verify.go:302` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `cmd/tossctl/verify.go:306` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `cmd/tossctl/verify.go:309` — `if err := validateM0TriggerMode(opts, prior); err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `cmd/tossctl/verify.go:313` — `if opts.includeTrigger {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `cmd/tossctl/verify.go:315` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `cmd/tossctl/verify.go:323` — `if steps := verifylive.StepCount(prior); steps > 0 && !opts.resume && len(opts.redo) == 0 {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `cmd/tossctl/verify.go:331` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `if` at `cmd/tossctl/verify.go:336` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B12 | `if` at `cmd/tossctl/verify.go:342` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B13 | `if` at `cmd/tossctl/verify.go:347` — `if holding == "" {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B14 | `if` at `cmd/tossctl/verify.go:351` — `if market == verifylive.MarketUS && !verifylive.SameMarket(verifylive.MarketOf(symbol), market) {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B15 | `if` at `cmd/tossctl/verify.go:359` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B16 | `if` at `cmd/tossctl/verify.go:386` — `if err != nil {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B17 | `if` at `cmd/tossctl/verify.go:395` — `if runErr != nil && (errors.Is(runErr, context.Canceled) \|\| errors.Is(runErr, context.DeadlineExceeded)) {` | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
