# Branch Test Map: `TestAShadowBandCannotBeReadAsAVeto`

- Source: `internal/candidate/band_test.go`
- Function: `internal/candidate/band_test.go:TestAShadowBandCannotBeReadAsAVeto`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | NEW - range over the value type and the pointer type | mutation 0.5 - `func (b *ShadowBand) Dangerous() bool` is RED | yes | yes |
| B2 | range over the forbidden names | same | yes | yes |
| B3 | the type has such a method | same | yes | yes |
