# Branch Test Map: `Dial`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty descriptor path is refused | client contract | yes | yes |
| B2 | insecure descriptor open is refused | `TestOpenPrivateDescriptorRejectsInsecureFilesystemObjects` | yes | yes |
| B3 | malformed, unknown, trailing, or oversized JSON is refused | `TestDecodeDescriptorRejectsUnknownTrailingAndOversizedContent` | yes | yes |
| B4 | invalid PID/token/address fields are refused | RPC descriptor validation contract | yes | yes |
| B5 | non-loopback endpoint is refused | RPC endpoint validation contract | yes | yes |
| B6 | health error/false prevents client construction | endpoint integration contract | yes | yes |
| B7 | false health without transport error becomes an explicit error | endpoint health contract | yes | yes |
