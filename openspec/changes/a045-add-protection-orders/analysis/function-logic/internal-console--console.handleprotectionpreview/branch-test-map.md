# Branch Test Map: `Console.handleProtectionPreview`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unwired seam refuses | `TestExitProtectionDefaultIsHonestAndHasNoFreeFormInputs` | route absent | pass |
| B2 | empty opaque action token refuses | `TestExitProtectionPreviewRequiresOpaqueActionToken` | unchecked token | pass |
| B3 | commander preview failure refuses | exit-protection preview table | unchecked seam error | pass |
| B4 | non-weakening/capability-less preview refuses | exit-protection preview table | unsafe preview accepted | pass |
