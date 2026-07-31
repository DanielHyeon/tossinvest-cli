# Function Logic Map: `toConsoleOpenAPICredentialCheck`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The local state and fixed message are converted byte-for-byte into the one
console package type. There are no credentials in either value.

## Branches and early returns

No branches or side effects.

## Calls and live bindings

The only call is the enum type conversion
`console.OpenAPICredentialState(result.State)`, which cannot fail.

## State mutations and fallbacks

None.

## Safety conclusion

- Safe edit boundary: pure adapter in the sole CLI file allowed to import the
  console package.
- High-risk impact: no; it cannot access credentials or start a process.
