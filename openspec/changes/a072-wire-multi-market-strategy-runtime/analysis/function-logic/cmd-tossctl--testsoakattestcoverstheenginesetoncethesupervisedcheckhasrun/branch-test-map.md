# Branch Test Map: TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | soak attestation command fails | this test | deployment regression converted to RED | yes |
| B2 | attestation cannot load | this test | deployment regression converted to RED | yes |
| B3 | combined evidence still misses a shared endpoint | this test | RED failed on exchange-rate GET | yes |
| B4 | supervised proof count drifts | this test | prior mutation guard retained | yes |
| B5 | inspect supervised proofs | this test | prior mutation guard retained | yes |
| B6 | supervised source is empty | this test | prior mutation guard retained | yes |
| B7 | inspect issued endpoint evidence | this test | prior authority guard retained | yes |
| B8 | WTS evidence claims official OAuth FX | this test | prior authority guard retained | yes |
