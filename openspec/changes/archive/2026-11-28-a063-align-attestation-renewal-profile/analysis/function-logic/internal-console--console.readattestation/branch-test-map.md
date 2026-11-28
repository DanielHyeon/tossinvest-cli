# Branch Test Map: `Console.readAttestation`

| B-id | Scenario | Focused test and bounded claim |
|---|---|---|
| B1 | Empty configured path. | unmeasured by focused test. |
| B2 | Load failure. | unmeasured by focused test. |
| B3 | Non-missing load failure adds unreadable reason. | unmeasured by focused test. |
| B4 | Content-invalidity switch. | Named test supplies valid content; invalid cases unmeasured. |
| B5 | Blank account case. | unmeasured by focused test. |
| B6 | No-endpoints case. | unmeasured by focused test. |
| B7 | Expired attestation. | unmeasured by focused test. |
| B8 | 48-hour future expiry produces near-expiry dashboard text. | `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability`; horizon field itself unmeasured. |
| B9 | Missing endpoints are appended. | Named test has none; nonempty range unmeasured. |
