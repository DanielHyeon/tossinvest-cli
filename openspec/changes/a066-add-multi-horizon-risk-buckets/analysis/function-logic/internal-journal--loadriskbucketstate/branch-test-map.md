# Branch Test Map: `loadRiskBucketState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | active owner reads still delegate with released rows excluded | full journal risk-bucket replay suite | existing | GREEN |
| sibling | released receipt recomputes the same digest, then rejects late-fill drift | `TestRiskBucketLateFillCannotBindReopenedOwner`, `TestRiskBucketOwnerReleasedReplayRequiresExactSealedReceiptAndEvent` | released replay returned early without digest validation | GREEN |

Wave 2A (2026-09-25): AST re-extracted at HEAD `648df8ef` because the file changed around the function; the
function body text is identical to the 2026-08-04 revision and the old/new branch alignment is identical, so
the rows above are unchanged. Package suite `go test -count=1 ./internal/journal/...` PASS at `d72bc401` (495.5s)
and with `-coverprofile` at `648df8ef` (641.1s).
