# Function Logic Map (편집 전): `LoadProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `LoadProductionFamilyActivation(params=2, results=2)`
- Source range: `429:1`–`494:2`
- AST evidence: `ast.json` — 편집 **전**(a112 8.5 응답 로트(HEAD 178cc196)).

## Inputs and invariants

- 편집 계획: 응답 로트 ④(P2-d): 읽기 결함 갈래(B5)의 안쪽 오류를 %w → %v 로 — 편집 전(f473d815) errors.Is 사슬 복원(ErrProductionRouteUnavailable 를 더는 만족하지 않음). 분기 불변.

## Branches and early returns

- Exact AST return nodes: `435:3`, `438:3`, `441:3`, `461:3`, `468:3`, `471:3`, `476:3`, `482:3`, `486:3`, `489:3`, `492:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 434:2 | `if strings.TrimSpace(config.ManifestDigest) == "" {` |
| B2 | if | 437:2 | `if ctx == nil {` |
| B3 | if | 440:2 | `if err := ctx.Err(); err != nil {` |
| B4 | if | 449:2 | `if fields := failedFields(` |
| B5 | if | 467:2 | `if err != nil {` |
| B6 | if | 470:2 | `if productionRouteDigest(data) != config.ManifestDigest {` |
| B7 | if | 475:2 | `if err != nil {` |
| B8 | if | 481:2 | `if manifest.Revoked {` |
| B9 | if | 485:2 | `if err != nil {` |
| B10 | if | 488:2 | `if err := ctx.Err(); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strings.TrimSpace` | 434:5 |
| `fmt.Errorf` | 435:30 |
| `fmt.Errorf` | 438:30 |
| `ctx.Err` | 440:12 |
| `filepath.Clean` | 443:21 |
| `strings.TrimSpace` | 443:36 |
| `strings.TrimSpace` | 444:26 |
| `productionRouteOwnerUID` | 445:20 |
| `ProductionFamilyActivationFileName` | 446:10 |
| `failedFields` | 449:15 |
| `filepath.IsAbs` | 452:29 |
| `config.ObservedAt.IsZero` | 453:29 |
| `productionRouteDigestValid` | 454:34 |
| `productionRouteIdentity` | 455:37 |
| `productionRouteDigestValid` | 456:40 |
| `productionRouteDigestValid` | 457:37 |
| `productionRouteIdentity` | 458:35 |
| `productionRouteIdentity` | 459:31 |
| `len` | 460:5 |
| `fmt.Errorf` | 461:30 |
| `strings.Join` | 461:109 |
| `readProductionRouteFile` | 463:15 |
| `filepath.Join` | 463:39 |
| `fmt.Errorf` | 468:30 |
| `productionRouteDigest` | 470:5 |
| `fmt.Errorf` | 471:30 |
| `decodeProductionFamilyActivation` | 474:19 |
| `fmt.Errorf` | 482:30 |
| `validateProductionFamilyActivation` | 484:16 |
| `ctx.Err` | 488:12 |
| `productionRouteTime` | 491:16 |

## Safety conclusion

- 판정 · 수락 집합 불변 — sentinel(ErrProductionFamilyActivationUnavailable) 그대로, 사슬에서 공유 읽기 함수의 sentinel 만 빠진다(소비자 0 — grep).
