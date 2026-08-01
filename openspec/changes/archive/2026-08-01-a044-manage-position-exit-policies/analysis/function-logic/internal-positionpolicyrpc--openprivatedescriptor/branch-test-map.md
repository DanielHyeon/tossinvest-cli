# Branch Test Map: `openPrivateDescriptor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | wrong endpoint filename is rejected | secure path contract | yes | yes |
| B2 | insecure directory is rejected | `TestOpenPrivateDescriptorRejectsInsecureFilesystemObjects` | yes | yes |
| B3 | descriptor lstat failure is rejected | secure path contract | yes | yes |
| B4 | symlink/mode/owner/hardlink descriptor is rejected | `TestOpenPrivateDescriptorRejectsInsecureFilesystemObjects` | yes | yes |
| B5 | no-follow open error returns no handle | secure open contract | yes | yes |
| B6 | fstat error closes and returns no handle | secure open contract | yes | yes |
| B7 | deterministic inode replacement during open is rejected | `TestOpenPrivateDescriptorRejectsInodeReplacementDuringOpen` | yes | yes |
| B8 | opened inode is revalidated after identity checks | hardlink/mode contract | yes | yes |
