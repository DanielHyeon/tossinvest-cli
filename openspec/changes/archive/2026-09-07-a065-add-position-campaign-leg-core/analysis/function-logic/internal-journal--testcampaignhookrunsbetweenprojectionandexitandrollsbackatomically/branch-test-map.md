# Branch Test Map: `TestCampaignHookRunsBetweenProjectionAndExitAndRollsBackAtomically`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | fixture setup | self | yes | yes |
| B2 | binding | self | yes | yes |
| B3 | Project order | self | yes | yes |
| B4 | Campaign order | self | yes | yes |
| B5 | Exit order | self | yes | yes |
| B6 | forced failure | self | yes | yes |
| B7 | transaction rollback | self | yes | yes |
| B8 | exact call sequence | self | yes | yes |
