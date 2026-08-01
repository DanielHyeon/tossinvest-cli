# Branch Test Map: `TestTheReadOnlyHandleHasNoWriteMethods`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for` at line 137: `for i := 0; i < typ.NumMethod(); i++ {`; invariant: missing/corrupt/alternate path is explicit | `TestTheReadOnlyHandleHasNoWriteMethods` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 139: `if !allowed[name] {`; invariant: missing/corrupt/alternate path is explicit | `TestTheReadOnlyHandleHasNoWriteMethods` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
