# Branch Test Map: `ProtectionVerifier.verifyParsed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil/foreign verifier state | sealed verifier test | parsed state was externally authorizable | rejected |
| B2 | exact scope/evidence prevalidation fails | scope/evidence tables | unbounded work occurred after root read | rejected before final trust sampling |
| B3 | key is revoked after evidence validation | `TestProtectionVerifierRechecksRevocationAfterEvidenceValidation` | final root was read before hashing | callback changes root, final phase rejects |
| B4 | final policy cannot be re-read safely | policy artifact table | parse snapshot could suffice | rejected |
| B5 | parsed generation/digest conflicts with current policy | parse-then-rollback/reuse | stale parse could authorize | rejected |
| B6 | root missing, noncanonical, digest/generation mismatched | trust-root tables | partial validation | rejected |
| B7 | signer removed/replaced | key lifecycle table | copied key survived | rejected |
| B8 | current key does not verify exact envelope | signature tamper table | stale verification risk | rejected |
| B9 | matrix issue is future at final clock sample | time table | time sampled before evidence | rejected |
| B10 | matrix expired at final clock sample | time table | time sampled before evidence | rejected |
| B11 | key revoked/outside full matrix window | revoke-after-evidence + key-window tables | copied active status survived | rejected |
| B12 | verifier previously observed newer/different fully validated policy | direct verify rollback/reuse | cross-call rollback possible | rejected |
