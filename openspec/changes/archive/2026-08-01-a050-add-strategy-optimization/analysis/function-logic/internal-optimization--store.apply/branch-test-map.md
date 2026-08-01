# Branch Test Map: `Store.Apply`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | blank token/actor | hardening tests | yes | yes |
| B2 | begin transaction failure | DB fault path | n/a | n/a |
| B3 | candidate absent | hardening tests | yes | yes |
| B4 | candidate query failure | DB fault path | n/a | n/a |
| B5 | immutable replay | lifecycle replay test | existing | existing |
| B6 | malformed not-before | `TestApplyRejectsCorruptCandidateTimes` | yes | yes |
| B7 | malformed/ordered expiry | `TestApplyRejectsCorruptCandidateTimes` | yes | yes |
| B8 | malformed created time | corruption test | yes | yes |
| B9 | forged schedule | `TestApplyRejectsCandidateScheduleThatBypassesRiskWait` | yes | yes |
| B10 | early candidate | lifecycle test | existing | existing |
| B11 | expired candidate | lifecycle test | existing | existing |
| B12 | missing confirmation | lifecycle test | existing | existing |
| B13 | current snapshot failure | DB fault path | n/a | n/a |
| B14 | version conflict | concurrent lifecycle test | existing | existing |
| B15 | invalid category/source/reason | metadata tamper test | yes | yes |
| B16 | malformed changes JSON | corruption test | yes | yes |
| B17 | malformed evidence JSON | corruption test | yes | yes |
| B18 | empty change set | corruption test | yes | yes |
| B19 | duplicate/unknown/inactive change | metadata tamper test | yes | yes |
| B20 | before/timing/safety mismatch | metadata tamper test | yes | yes |
| B21 | invalid option | metadata tamper test | yes | yes |
| B22 | derive risk flag | metadata tamper test | yes | yes |
| B23 | derive restart flag | metadata tamper test | yes | yes |
| B24 | metadata boolean range | metadata tamper test | yes | yes |
| B25 | effective entry mismatch | metadata tamper test | yes | yes |
| B26 | rollback desired removal | rollback lifecycle test | existing | existing |
| B27 | desired update | lifecycle test | existing | existing |
| B28 | immediate neutral effective removal | lifecycle test | existing | existing |
| B29 | immediate neutral effective update | lifecycle test | existing | existing |
| B30 | random audit ID failure | entropy fault path | n/a | n/a |
| B31 | risk clears manifest/entry | lifecycle test | existing | existing |
| B32 | non-immediate check | lifecycle test | existing | existing |
| B33 | effective version update | lifecycle test | existing | existing |
| B34 | control CAS DB failure | concurrent lifecycle test | existing | existing |
| B35 | busy/locked conflict mapping | concurrent lifecycle test | existing | existing |
| B36 | CAS no-row conflict | concurrent lifecycle test | existing | existing |
| B37 | snapshot insert failure | transaction fault path | n/a | n/a |
| B38 | audit insert failure | transaction fault path | n/a | n/a |
| B39 | application insert failure | transaction fault path | n/a | n/a |
| B40 | commit failure | transaction fault path | n/a | n/a |
| B41 | successful immutable result | lifecycle tests | existing | existing |
| B42 | computed settings digest | lifecycle tests | existing | existing |
| B43 | snapshot clone result | lifecycle tests | existing | existing |
| B44 | final result return | lifecycle tests | existing | existing |
| B45 | payload MAC verification | `TestApplyRejectsSelfConsistentCandidatePayloadTamperWithoutMAC` | yes | yes |
| B46 | raw payload/candidate ID/replay integer MAC binding | `TestApplyMACBindsCandidateIDRawPayloadAndReplayIntegers` | yes | yes |
| B47 | per-change audit row/digest insert failure aborts transaction | audit transaction fault-path review | audit digest absent | defensive branch reviewed |
| B48 | immutable application-row insert failure aborts transaction | transaction fault-path review | existing branch renumbered after hardening | defensive branch reviewed |
| B49 | commit failure returns no successful snapshot result | transaction fault-path review | existing branch renumbered after hardening | defensive branch reviewed |
