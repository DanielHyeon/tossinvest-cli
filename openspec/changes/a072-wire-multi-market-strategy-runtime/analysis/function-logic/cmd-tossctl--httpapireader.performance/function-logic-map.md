# Function Logic Map: `httpAPIReader.Performance`

- Source: `cmd/tossctl/httpapi_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The reader requires an immutable performance capability. Attribution is account-bound and queried separately for exact `KR` and `US` market keys with `link_missing` retained.

## Branches and early returns

B1–B8 cover nil/error fallback, dashboard failure, missing account seam, account resolution, two market reads, absent attribution generation, corrupt read failure and successful append.

## Calls and live bindings

`Dashboard` uses the fixed server query; `AttributionRows` enforces account, market and row bounds in `performance.ReadOnly`.

## State mutations and fallbacks

Only a returned value slice is extended. A never-produced generation is empty; integrity failures are returned and no market value is copied to its peer.

## Safety conclusion

This is an immutable read adapter and exposes no derived-store writer or trading capability.
