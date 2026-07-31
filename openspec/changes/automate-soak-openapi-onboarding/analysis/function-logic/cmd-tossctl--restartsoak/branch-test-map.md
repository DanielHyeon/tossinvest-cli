# Branch Test Map: `restartSoak`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | executable resolution failure blocks restart | existing binstamp seam tests | existing | passed |
| B2 | process lookup failure blocks spawn | `TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted` | existing | passed |
| B3 | located processes are considered | `TestRestartingTheSoakInterruptsItThenStartsItAgain` | existing | passed |
| B4 | console PID is never signalled | `TestTheRestartNeverSignalsThisProcess` | existing | passed |
| B5 | signal failure blocks spawn | existing process seam tests | existing | passed |
| B6 | signalled process is awaited | `TestRestartingTheSoakInterruptsItThenStartsItAgain` | existing | passed |
| B7 | non-exiting survey blocks spawn | `TestASurveyThatWillNotStopBlocksTheRestart` | existing | passed |
| B8 | token fence runs after old exit before spawn | `TestRestartingTheSoakInterruptsItThenStartsItAgain` | yes | passed |
| B9 | token fence failure blocks spawn | `TestTokenGenerationFenceFailureBlocksSoakSpawn` | yes | passed |
| B10 | detached spawn error is returned | existing spawn seam tests | existing | passed |
| B11 | stopped count controls truthful notice | restart process tests | existing | passed |
| B12 | absent old soak starts one | `TestRestartingWithNothingRunningJustStartsOne` | existing | passed |
| B13 | one old soak reports its PID | `TestRestartingTheSoakInterruptsItThenStartsItAgain` | existing | passed |
| B14 | multiple old soaks produce one child | existing multi-process seam | existing | passed |
