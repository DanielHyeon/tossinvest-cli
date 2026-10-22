## ADDED Requirements

### Requirement: Inspection evidence has only named offline consumers and no runtime consumer

Among non-test Go files under `internal`, `cmd` and `tools`, the campaign-state evidence package SHALL be imported only by the three
lane `offline_evidence.go` files and by the offline follow-on assessment file `internal/offlinecontinuation/assessment.go`, and a
lane's `CaptureOfflineEvidence` SHALL be referenced through an imported lane package only from that assessment file. The assessment is
inspection-only and returns no execution authority. No such file SHALL import the offline follow-on assessment package, so the capture
path cannot reach engine, CLI, API or UI code through it. A static guard test MUST enforce these rules over that file set; unqualified
calls inside a lane package and Go files outside those three trees are outside the guard, as they were before this change.

#### Scenario: the named offline assessment is accepted

- **WHEN** the guard reads `internal/offlinecontinuation/assessment.go` importing the evidence package and calling a lane `CaptureOfflineEvidence`
- **THEN** the guard accepts it, and the same source placed at any other path is refused

#### Scenario: a runtime file reaching the assessment is refused

- **WHEN** a non-test file such as one under `internal/app/engine` imports the offline follow-on assessment package, written plainly or with an escaped import path
- **THEN** the guard refuses it as an unexpected inspection consumer
