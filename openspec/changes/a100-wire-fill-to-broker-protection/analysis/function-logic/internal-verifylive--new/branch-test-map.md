# Branch Test Map: `New`

- Source: `internal/verifylive/runner.go:320-425`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/runner.go:321` — `if o.Broker == nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/runner.go:324` — `if o.Recorder == nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/runner.go:327` — `if o.Confirm == nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/runner.go:330` — `if o.ConfirmBatch == nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/runner.go:334` — `if strings.TrimSpace(o.AccountRef) == "" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/runner.go:338` — `if err != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/runner.go:341` — `if o.IncludeTrigger {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/verifylive/runner.go:342` — `if !o.ConfirmEach \|\| !o.Resume \|\| o.IncludeTTLEdge \|\| o.Receipt == nil \|\| !o.Receipt.usable() \|\|` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/runner.go:346` — `if len(Outstanding(o.Prior)) != 0 {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `internal/verifylive/runner.go:349` — `if checkpoint, ok, checkpointErr := M0Unsettled(o.Prior); checkpointErr != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `else` at `internal/verifylive/runner.go:351` — `} else if ok && checkpoint.Kind != "pending-create" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B12 | `if` at `internal/verifylive/runner.go:351` — `} else if ok && checkpoint.Kind != "pending-create" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B13 | `else` at `internal/verifylive/runner.go:353` — `} else if !ok {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B14 | `if` at `internal/verifylive/runner.go:353` — `} else if !ok {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B15 | `if` at `internal/verifylive/runner.go:358` — `if err := M0ExactPrerequisites(o.Prior); err != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B16 | `if` at `internal/verifylive/runner.go:362` — `if _, ok := official.M0ReadSourceFor(o.Broker); !ok {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B17 | `if` at `internal/verifylive/runner.go:392` — `if r.approvalChannel == "" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B18 | `if` at `internal/verifylive/runner.go:395` — `if r.out == nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B19 | `if` at `internal/verifylive/runner.go:398` — `if r.now == nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B20 | `if` at `internal/verifylive/runner.go:401` — `if r.sleep == nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B21 | `if` at `internal/verifylive/runner.go:404` — `if r.maxSellQuantity <= 0 {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B22 | `if` at `internal/verifylive/runner.go:407` — `if r.ttlWait <= 0 {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B23 | `if` at `internal/verifylive/runner.go:410` — `if r.triggerWindow <= 0 {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B24 | `if` at `internal/verifylive/runner.go:413` — `if r.process.InstanceID == "" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B25 | `range` at `internal/verifylive/runner.go:416` — `for _, id := range o.Redo {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B26 | `if` at `internal/verifylive/runner.go:420` — `if r.m0Receipt != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
