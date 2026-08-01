# Branch Test Map: `validatePrivateDirectory`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty path is rejected | private directory contract | yes | yes |
| B2 | symlink traversal is rejected | `TestOpenPrivateDescriptorRejectsInsecureFilesystemObjects/symlink_traversal` | yes | yes |
| B3 | lstat failure is rejected | private directory contract | yes | yes |
| B4 | symlink or non-directory leaf is rejected | insecure filesystem tests | yes | yes |
| B5 | wrong/unknown owner is rejected | ownership contract | yes | yes |
| B6 | control leaf must be exact 0700 | `TestOpenPrivateDescriptorRejectsInsecureFilesystemObjects/wrong_control_directory_mode` | yes | yes |
| B7 | engine parent must not be group/other writable | `TestOpenPrivateDescriptorRejectsInsecureFilesystemObjects/group_writable_engine_directory` | yes | yes |
