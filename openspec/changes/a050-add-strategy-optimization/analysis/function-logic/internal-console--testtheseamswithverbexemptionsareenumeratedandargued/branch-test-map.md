# Branch Test Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | iterate registered capabilities | self | a050 seam unenumerated | yes |
| B2 | capability has no exemptions | self | existing contract | yes |
| B3 | exempt capability is not enumerated | self | a050 seam unenumerated | yes |
| B4 | iterate exemptions | self | existing contract | yes |
| B5 | exemption name absent from allow-map | self | existing contract | yes |
| B6 | exemption rationale blank | self | existing contract | yes |
| B7 | iterate expected exemption names | self | existing contract | yes |
| B8 | expected exemption absent from capability | self | existing contract | yes |
| B9 | iterate expected capability fields | self | existing contract | yes |
| B10 | expected field has no exemptions | self | existing contract | yes |
| B11 | iterate protected mutation verbs | self | existing contract | yes |
| B12 | iterate mutation verb registry | self | existing contract | yes |
| B13 | exact mutation verb found | self | existing contract | yes |
| B14 | required mutation verb missing | self | existing contract | yes |
| B15 | iterate account verb registry | self | existing contract | yes |
| B16 | exact account verb found | self | existing contract | yes |
| B17 | order account verb missing | self | existing contract | yes |
