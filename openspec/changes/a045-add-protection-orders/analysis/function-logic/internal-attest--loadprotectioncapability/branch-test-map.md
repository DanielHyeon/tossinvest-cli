# Branch Test Map: `LoadProtectionCapability`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | required in follow-up RED | pending implementation |
| B1+ | B1 current UID unavailable → file error; B2 filesystem checks/open/read/decode fail → typed refusal; B3 matrix/window/scope mismatch → typed refusal; B4 all checks pass → capability returned. | parse-only cannot yield verified capability; missing/fake/mismatched evidence bytes; valid bound evidence; unsafe parent/path tests. | required in follow-up RED | pending implementation |
