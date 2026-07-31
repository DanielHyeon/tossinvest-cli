# Branch Test Map: `checkProtectionFileInfo`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | required in follow-up RED | pending implementation |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | required in follow-up RED | pending implementation |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | required in follow-up RED | pending implementation |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | required in follow-up RED | pending implementation |
| B5 | hard-link count is unavailable or not exactly one | `TestProtectionMatrixRejectsUnsafeFileAndDirectParent/hardlink` | yes: linked file was accepted before guard | pass |
| B1+ | B1 non-regular/symlink → refuse; B2 mode !=0600 → refuse; B3 owner unavailable/mismatch → refuse; B4 otherwise accept. | mode, owner, symlink, directory, hardlink, direct-parent symlink/mode/owner, post-read identity. | yes: adversarial parser suite failed before guards | pass |
