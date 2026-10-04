# Function Logic Map (편집 전): `LoadProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`
- Signature: `LoadProductionFamilyActivation(params=2, results=2)`
- Source range: `427:1`–`478:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD f473d815).

## Inputs and invariants

- 편집 계획: a112 8.8.4 로트 B 항목 1(Q-B1=(c)): 맨 sentinel 반환마다 필드명 %w 래핑(errors.Is 보존). 복합 결속은 분기 그대로 필드-diff 목록으로 한 번 반환. :455 읽기 실패는 underlying err 를 %w 사슬로 보존하고 digest 불일치만 필드명(두 종류 분리).

## Branches and early returns

- Exact AST return nodes: `433:3`, `436:3`, `439:3`, `451:3`, `456:3`, `460:3`, `466:3`, `470:3`, `473:3`, `476:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 432:2 | `if strings.TrimSpace(config.ManifestDigest) == "" {` |
| B2 | if | 435:2 | `if ctx == nil {` |
| B3 | if | 438:2 | `if err := ctx.Err(); err != nil {` |
| B4 | if | 445:2 | `if !ownerOK // name == "" // !filepath.IsAbs(config.ConfigDir) // config.ObservedAt.IsZero() //` |
| B5 | if | 455:2 | `if err != nil // productionRouteDigest(data) != config.ManifestDigest {` |
| B6 | if | 459:2 | `if err != nil {` |
| B7 | if | 465:2 | `if manifest.Revoked {` |
| B8 | if | 469:2 | `if err != nil {` |
| B9 | if | 472:2 | `if err := ctx.Err(); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strings.TrimSpace` | 432:5 |
| `ctx.Err` | 438:12 |
| `filepath.Clean` | 441:21 |
| `strings.TrimSpace` | 441:36 |
| `strings.TrimSpace` | 442:26 |
| `productionRouteOwnerUID` | 443:20 |
| `ProductionFamilyActivationFileName` | 444:10 |
| `filepath.IsAbs` | 445:32 |
| `config.ObservedAt.IsZero` | 445:68 |
| `productionRouteDigestValid` | 446:4 |
| `productionRouteIdentity` | 447:4 |
| `productionRouteDigestValid` | 448:4 |
| `productionRouteDigestValid` | 449:4 |
| `productionRouteIdentity` | 450:4 |
| `productionRouteIdentity` | 450:56 |
| `readProductionRouteFile` | 453:15 |
| `filepath.Join` | 453:39 |
| `productionRouteDigest` | 455:19 |
| `decodeProductionFamilyActivation` | 458:19 |
| `validateProductionFamilyActivation` | 468:16 |
| `ctx.Err` | 472:12 |
| `productionRouteTime` | 475:16 |

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is). 메시지에 필드명이 더해질 뿐. 미선언(Undeclared)이 맨 앞인 순서 불변.
