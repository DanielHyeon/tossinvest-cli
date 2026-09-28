# Branch Test Map: `newEngineCmd`

The AST has no branches (`branches: null`, 114–127). The one a066 statement (the registration) is covered by
`TestEngineRiskRelaxationCommandsDeclareWhatTheyAre`, which walks `newRootCmd()` and fails when any of the three
commands is missing or mis-annotated (mutation C06 CAUGHT).
