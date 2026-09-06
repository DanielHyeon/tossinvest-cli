# Function-logic coverage status

Comparison base remains immutable
`da80ce31b6a1ab5d443016768f970a82bab102db`; it was not reset or changed for
this work.

On 2026-09-06, direct `validate_target` checks passed for eleven scoped
bundles: five production functions (`newSoakAttestCmd`, `runSoakAttest`,
`Console.readAttestation`, `Summary.Evaluate`, and `BuildAttestation`) and six
test functions derived by the checker against that base. The six are exact
current requirements, not the earlier stale count of four:

- `TestSoakAndLiveEndpointsCoverTheEngineInterlock` (current)
- `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` (current)
- `TestBuildAttestationRefusesAnIncompleteSoak` (current)
- `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing` (base)
- `TestSoakAttestWritesAVerifiableAttestation` (base)
- `TestTheDashboardReportsAnUnstartedMachineWithoutFailing` (base)

The three `revision: base` ASTs were extracted retrospectively from the frozen
commit during this verification. They distinguish immutable-base structure from
current files but do not claim that pre-edit maps were captured before the
implementation. All scoped maps are post-hoc AST/test alignment evidence.

`python3 tools/logic-map/check_analysis.py --change
a063-align-attestation-renewal-profile` exited **1**. It reports exactly **327
missing-evidence rows**, all outside the a063 scoped bundles, because the
shared worktree differs from the immutable base in unrelated functions. The
global result is therefore still blocked; this verification neither creates
unrelated maps nor treats it as an a063 completion result.

The preceding branch-map-drift statements are obsolete: current
`runSoakAttest` maps B1-B20 and `Summary.Evaluate` maps B1-B2. No unmatched
scoped target was reported by the current direct validations.
