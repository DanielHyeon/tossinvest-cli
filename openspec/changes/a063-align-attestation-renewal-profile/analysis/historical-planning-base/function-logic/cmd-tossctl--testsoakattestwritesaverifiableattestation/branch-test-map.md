# Branch Test Map: `TestSoakAttestWritesAVerifiableAttestation`

Source: `cmd/tossctl/soak_test.go` (399-434); AST revision: `base`.

| Branch | Raw AST-position condition | Test consequence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `	if err != nil {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B2 | `	if err != nil {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B3 | `	if err := a.Verify(time.Now(), "123-45-678901", soak.RequiredEndpoints()); err != nil {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B4 | `	if a.SoakDays < 3 {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B5 | `	for _, e := range a.Endpoints {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B6 | `		if !strings.HasPrefix(e, "GET ") {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B7 | `	for _, remaining := range soak.LiveOnlyEndpoints() {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B8 | `		if !strings.Contains(out, remaining) {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B9 | `	if strings.Contains(out, "678901") {` | `TestSoakAttestWritesAVerifiableAttestation` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |

This is an AST structural inventory with raw source conditions. It does not turn a current result into historical pre-edit execution evidence.
