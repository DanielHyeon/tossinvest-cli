# Branch Test Map: `writePositionPolicyDescriptor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | descriptor marshaling failure is returned | JSON contract | yes | yes |
| B2 | control directory revalidation failure blocks staging | `TestPositionPolicyControlRejectsInsecureControlFilesystem` | yes | yes |
| B3 | temp creation error blocks publication | private descriptor contract | yes | yes |
| B4 | open temp must validate as exact private file | private descriptor contract | yes | yes |
| B5 | temp stat error blocks publication | private descriptor contract | yes | yes |
| B6 | chmod/write/short-write/sync/close failure blocks publication | `TestWritePositionPolicyDescriptorPreservesChmodAndWriteErrors` | yes | yes |
| B7 | closed temp remains exact 0600 owner-only single-link file | endpoint integration test | yes | yes |
| B8 | closed temp inode still matches opened inode | inode contract | yes | yes |
| B9 | control directory is revalidated before rename | insecure filesystem tests | yes | yes |
| B10 | rename error blocks publication | atomic publication contract | yes | yes |
| B11 | successful return retains published descriptor | endpoint lifecycle test | yes | yes |
| B12 | post-publish failure removes only matching staged inode | inode-scoped rollback contract | yes | yes |
| B13 | published descriptor metadata is valid | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` | yes | yes |
| B14 | published inode equals staged inode | endpoint integration test | yes | yes |
| B15 | control directory opens for durability sync | endpoint integration test | yes | yes |
| B16 | directory sync error fails publication | durability contract | yes | yes |
| B17 | directory close error fails publication | durability contract | yes | yes |
