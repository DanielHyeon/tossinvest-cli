# Branch Test Map: TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | isolated soak attestation command fails | this test | deployment regression converted to RED | yes |
| B2 | generated attestation cannot load | this test | deployment regression converted to RED | yes |
| B3 | soak alone unexpectedly covers mutation endpoints | this test | deployment regression converted to RED | yes |
| B4 | enumerate remaining mutation gaps | this test | deployment regression converted to RED | yes |
| B5 | any shared GET remains missing | this test | RED failed on global exchange-rate GET | yes |
