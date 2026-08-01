# Branch Test Map: `loadRemoteCertificate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | wrong-host/invalid certificate is refused before serving | `TestRemoteAccessRejectsCertificateForAnotherHost` and shared certificate coverage | shared extraction not present | GREEN via full console suite |
| B2 | valid current ServerAuth certificate remains accepted | remote console construction/route tests | existing behavior | GREEN via full console suite |
