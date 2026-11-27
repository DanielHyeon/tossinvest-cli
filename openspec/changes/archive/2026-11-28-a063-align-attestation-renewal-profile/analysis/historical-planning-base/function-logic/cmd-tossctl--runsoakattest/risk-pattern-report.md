# Risk Pattern Report: `runSoakAttest`

- Source: `cmd/tossctl/soak.go`
- AST evidence was generated before implementation at `/tmp/a063-run-soak-attest-ast.json`.

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| attestation publication | `runSoakAttest` | review-required | `function-logic-map.md` |

The function writes the engine-interlock attestation. The implementation must refresh this report against the edited source before verification.
