# Branch Test Map: `Console.handleSoakRestart`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | restart seam absent | `TestASoakRestartWithNoWiringSaysSo` | existing | pending |
| B2 | restart seam returns error | `TestAFailedSoakRestartIsReportedAndNotSwallowed` | existing | pending |
| B3 | restart seam returns blank success | `TestBlankSoakRestartResultUsesFixedNotice` | pending | pending |
| B4 | ready preflight continues | `TestReadyCredentialsRestartSoakInOneClick` | failed | passed |
| B5 | missing preflight redirects | `TestMissingCredentialsOpenSetupWithoutRestart` | failed | passed |
| B6 | rejected/unavailable returns without restart | onboarding state tests | failed | passed |
| new ready | saved credential preflight succeeds | `TestReadyCredentialsRestartSoakInOneClick` | pending | pending |
| new missing | no credentials | `TestMissingCredentialsOpenSetupWithoutRestart` | pending | pending |
| new rejected | file credentials rejected | `TestRejectedFileCredentialsOpenSetupWithoutRestart` | pending | pending |
| new transient | official probe unavailable | `TestTransientPreflightStopsWithoutSetupOrRestart` | pending | pending |
| new env | rejected environment credentials | `TestRejectedEnvironmentCredentialsRequireContainerUpdate` | pending | pending |
| new pending | incomplete saved generation reopens setup without spawn | `TestPendingGenerationRestartReopensSetupWithoutRestart` | yes | pending |
| B7 | all other unavailable states fail closed on dashboard | transient preflight tests | existing | pending |
