# Branch Test Map: `LoadActivationRecord`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unknown field or trailing value | `TestActivationRecordIsStrictAndApprovalTimeIsBounded` | API absent | zero + error |
| B2 | field validation dispatch | same test plus binding matrix | API absent | complete record only |
| B3 | missing version | strict activation matrix | API absent | zero + error |
| B4 | wrong market | strict activation matrix | API absent | zero + error |
| B5 | wrong session | strict activation matrix | API absent | zero + error |
| B6 | malformed set digest | strict activation matrix | API absent | zero + error |
| B7 | malformed evidence digest | strict activation matrix | API absent | zero + error |
| B8 | missing approved_at | strict activation matrix | API absent | zero + error |
| B9 | missing approved_by | `TestActivationRecordIsStrictAndApprovalTimeIsBounded` | old approval embedded in set | zero + error |
