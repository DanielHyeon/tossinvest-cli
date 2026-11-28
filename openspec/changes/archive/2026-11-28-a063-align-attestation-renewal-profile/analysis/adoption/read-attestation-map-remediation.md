# `Console.readAttestation` map remediation

- Exact source condition: `v.Path` is initialized from `c.opts.Attestation`, and `internal/console/data.go:268-270` returns the initialized view when `strings.TrimSpace(v.Path) == ""`; therefore it returns before either `v.readRenewalStatus(now)` call at lines 276 and 304.
- Changed claim: replaced “calls `readRenewalStatus` on every path” with the precise behavior: only the load-error and loaded-success flows after a nonblank trimmed path run the renewal diagnostic; the blank-path early return does not.
- Validation: the generic/TODO scan, `rg -n -i 'TODO|TBD|FIXME|XXX|generic' openspec/changes/a063-align-attestation-renewal-profile/analysis/function-logic/internal-console--console.readattestation/function-logic-map.md`, returned no matches; `git diff --check` returned success; `git diff --name-only c727ad12a42dcd15c494c1997e92816edaf17b6b -- ':!openspec/**'` returned no paths, confirming no non-metadata diff versus S `c727ad12a42dcd15c494c1997e92816edaf17b6b`.
- No tests or runtime commands were run or claimed.
