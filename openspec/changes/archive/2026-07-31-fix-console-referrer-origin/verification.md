# Verification

## Safe runtime reproduction before implementation

An actual Chrome form submission from
`https://127.0.0.1:37085/` replaced the rendered CSRF value with
`intentionally-invalid` before submitting `/restart`.

Baseline request evidence:

```text
Origin: null
Referer: <absent>
response: origin refusal (HTTP 403)
```

Browser-only `same-origin` policy override:

```text
Origin: https://127.0.0.1:37085
Referer: https://127.0.0.1:37085/
response: CSRF refusal (HTTP 403)
```

The invalid CSRF value ensures the restart handler did not execute.

## TDD and deployment

Proposal-freeze attempt 3: APPROVE, no unresolved P0/P1.

RED command:

```text
go test ./internal/console -run 'TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal|TestConsoleDocumentsUseSameOriginReferrerPolicy|TestExplicitOpaqueOriginCannotReachMutationHandler' -count=1
```

Result: FAIL as required. The exact remote/rendered headers and both document
meta policies were still `no-referrer`; the explicit opaque-origin safety
contract remained rejected.

GREEN results:

```text
focused policy/origin tests: PASS (0.017s)
go test ./internal/console -count=1: PASS (9.272s)
go test -race ./internal/console -count=1: PASS (33.127s)
```

Only the two response header literals and two HTML meta literals changed in
production code. Function Logic Map validation passes against the updated
source and the base-revision inherited test map.

Repository verification:

```text
make test: PASS
make vet: PASS
make validate: PASS (46 OpenSpec items)
make sdd-sync: CodeGraph and CodeGraphContext current; GBrain advisory busy
make sdd-check: PASS; CodeGraph matches worktree, PM hierarchy current,
                83 Python SDD tests and Go logic-map tests pass
```

Deployment:

```text
docker compose build tossos: PASS
image manifest: sha256:44005273ec5eb3863c0e9dce0de2661dc89d16afb23d4102921e3e21030861c6
docker compose up -d --no-deps --force-recreate tossos: PASS
container health: healthy, failing streak 0
published port: 127.0.0.1:37085 -> 37085/tcp
```

Post-deploy Chrome invalid-CSRF form probe:

```text
method: POST
url: https://127.0.0.1:37085/restart
Origin: https://127.0.0.1:37085
Referer: https://127.0.0.1:37085/
response: HTTP 403 CSRF refusal
origin refusal: false
```

The DOM form's process CSRF value was replaced with `intentionally-invalid`
before submission. Therefore the request proved the browser origin contract and
could not enter the restart handler.
