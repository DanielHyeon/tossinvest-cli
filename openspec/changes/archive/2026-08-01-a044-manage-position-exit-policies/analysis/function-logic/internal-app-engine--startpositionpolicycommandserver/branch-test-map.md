# Branch Test Map: `StartPositionPolicyCommandServer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil command service is refused | server constructor contract | yes | yes |
| B2 | empty engine directory is refused | server constructor contract | yes | yes |
| B3 | insecure engine directory fails before bind | `TestPositionPolicyControlRejectsInsecureControlFilesystem` | yes | yes |
| B4 | control directory creation error fails closed | filesystem constructor contract | yes | yes |
| B5 | newly created control directory is tracked for cleanup | endpoint lifecycle test | yes | yes |
| B6 | existing non-directory/symlink is not treated as creatable | `TestPositionPolicyControlRejectsInsecureControlFilesystem` | yes | yes |
| B7 | control directory validation failure aborts startup | `TestPositionPolicyControlRejectsInsecureControlFilesystem` | yes | yes |
| B8 | newly created invalid directory is cleaned | constructor cleanup contract | yes | yes |
| B9 | later startup failure cleans only the newly created directory | constructor cleanup contract | yes | yes |
| B10 | loopback bind failure aborts startup | transport constructor contract | yes | yes |
| B11 | entropy failure aborts startup | crypto/rand contract | yes | yes |
| B12 | health is GET-only | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` | yes | yes |
| B13 | position list is GET-only | endpoint transport contract | yes | yes |
| B14 | list command failure is typed | endpoint transport contract | yes | yes |
| B15 | private descriptor failure closes the listener | descriptor writer contract | yes | yes |
