# Function Logic Map: `LoadProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `LoadProductionFamilyActivation(params=2, results=2)`
- Source range: `429:1`–`494:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 미선언 판정이 맨 앞 — 엔진 `familyGateFor` 의 `errors.Is(err, ErrProductionFamilyActivationUndeclared)` 판별 불변.
- 읽기 결함(공유 읽기 함수 `readProductionRouteFile` 의 오류)과 핀 불일치는 다른 메시지 · 다른 사슬. 그 읽기 함수는 OS 원인을 자기 sentinel 하나로 접는다 — OS 원인 노출은 이 로트 밖(잔여).

## Branches and early returns

- Exact AST return nodes: `435:3, 438:3, 441:3, 461:3, 468:3, 471:3, 476:3, 482:3, 486:3, 489:3, 492:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 434:2 | 핀이 비었으면 미선언 → `%w(Undeclared): manifest_digest pin is empty` — **맨 앞 순서 불변**(엔진 판별이 errors.Is 로 의존) |
| B2 | if | 437:2 | ctx 가 nil → `%w: context is nil` |
| B3 | if | 440:2 | ctx 취소 → ctx.Err() 그대로 |
| B4 | if | 449:2 | 설정 결속(복합 — 분기 하나) → `%w: config binding: <어긋난 필드 전부>` |
| B5 | if | 467:2 | 매니페스트 파일 읽기 결함 → `%w: manifest file …: %w(읽기 함수 오류)` — 불일치와 다른 종류 |
| B6 | if | 470:2 | **(새)** 핀이 파일 바이트와 다름 → `%w: manifest_digest: …` |
| B7 | if | 475:2 | 해석 거절 → 해석기가 붙인 이유 그대로(sentinel 포함) |
| B8 | if | 481:2 | 폐기 → `%w(Revoked): revoked=true` |
| B9 | if | 485:2 | 검증 거절 → 검증기 오류 그대로 |
| B10 | if | 488:2 | 끝의 ctx 취소 재확인 |

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

## State mutations and fallbacks

- 상태 변경 없음 — 파일 읽기 · 검증.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.
- 복합 결속은 분기 하나 그대로 두고 그 안에서 필드별 비교를 모은다(Q-B1=(c)); 항은 전부 순수 비교 · 부작용 없는 검사라 무조건 평가해도 판정이 같다(2026-10-04 단락 평가 전제 확인).
- B5/B6 분리로 분기 하나가 늘었지만 둘 다 거절(같은 sentinel)이라 수락 집합 불변.
