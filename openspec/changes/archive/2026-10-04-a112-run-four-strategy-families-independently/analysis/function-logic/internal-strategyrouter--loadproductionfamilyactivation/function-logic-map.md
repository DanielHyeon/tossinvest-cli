# Function Logic Map: `LoadProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `bac16c04b38d49c479a381fb326d7dd066e5525997322619d8ceb496e3ddb0d8`
- Signature: `LoadProductionFamilyActivation(params=2, results=2)`
- Source range: `429:1`–`495:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.5-R).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 미선언 판정이 맨 앞 — 엔진 `familyGateFor` 의 `errors.Is(err, ErrProductionFamilyActivationUndeclared)` 판별 불변.
- 읽기 결함(공유 읽기 함수 `readProductionRouteFile` 의 오류)과 핀 불일치는 다른 메시지 · 다른 사슬. 그 읽기 함수는 OS 원인을 자기 sentinel 하나로 접는다 — OS 원인 노출은 이 로트 밖(잔여).
- sentinel 배타성: 각 거절은 자기 sentinel 하나만 만족한다(`a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel` — 네 sentinel 전부를 모양마다 대조; 8.5 보이스 3 P1-2).
- **설정 결속의 market 항(`name == ""`)은 판정을 바꾸지 못한다 — 기록(8.5 보이스 2 P2-2, 편집 전부터)**. 가림 가드 둘: (1) 디렉터리 읽기 실패 — 항이 없으면 `filepath.Join(dir, "")` 가 디렉터리를 가리켜 `readProductionRouteFile` 이 정규 0400 파일이 아니라고 거절(B5); (2) 몸통 결속 `body.Market != config.Market || !validMarket(body.Market)`(validateProductionFamilyActivation B1). 이 블록을 닫아 두는 등식: `ProductionFamilyActivationFileName(m) == "" ⇔ !validMarket(m)` (둘 다 {KR, US} 위의 닫힌 대응 — production_family_activation.go `ProductionFamilyActivationFileName` · types.go `validMarket`). 이 항의 유일한 핀은 메시지 시험 `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(config binding: unknown market)이다.

## Branches and early returns

- Exact AST return nodes: `435:3, 438:3, 441:3, 461:3, 469:3, 472:3, 477:3, 483:3, 487:3, 490:3, 493:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 434:2 | 핀이 비었으면 미선언 → `%w(Undeclared): manifest_digest pin is empty` — **맨 앞 순서 불변**(엔진 판별이 errors.Is 로 의존) |
| B2 | if | 437:2 | ctx 가 nil → `%w: context is nil` |
| B3 | if | 440:2 | ctx 취소 → ctx.Err() 그대로 |
| B4 | if | 449:2 | 설정 결속(복합 — 분기 하나) → `%w: config binding: <어긋난 필드 전부>` |
| B5 | if | 468:2 | 매니페스트 파일 읽기 결함 → `%w: manifest file …: %v(읽기 함수 오류 문장)` — 불일치와 다른 종류, 사슬에 활성화 sentinel 하나 |
| B6 | if | 471:2 | **(새)** 핀이 파일 바이트와 다름 → `%w: manifest_digest: …` |
| B7 | if | 476:2 | 해석 거절 → 해석기가 붙인 이유 그대로(sentinel 포함) |
| B8 | if | 482:2 | 폐기 → `%w(Revoked): revoked=true` |
| B9 | if | 486:2 | 검증 거절 → 검증기 오류 그대로 |
| B10 | if | 489:2 | 끝의 ctx 취소 재확인 |

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
| `fmt.Errorf` | 469:30 |
| `productionRouteDigest` | 471:5 |
| `fmt.Errorf` | 472:30 |
| `decodeProductionFamilyActivation` | 475:19 |
| `fmt.Errorf` | 483:30 |
| `validateProductionFamilyActivation` | 485:16 |
| `ctx.Err` | 489:12 |
| `productionRouteTime` | 492:16 |

## State mutations and fallbacks

- 상태 변경 없음 — 파일 읽기 · 검증.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.
- 복합 결속은 분기 하나 그대로 두고 그 안에서 필드별 비교를 모은다(Q-B1=(c)); 항은 전부 순수 비교 · 부작용 없는 검사라 무조건 평가해도 판정이 같다(2026-10-04 단락 평가 전제 확인).
- B5/B6 분리로 분기 하나가 늘었지만 둘 다 거절(같은 sentinel)이라 수락 집합 불변.

0.5 응답 로트: 같은 파일의 주석 · 한 글자 편집(본문 구조 불변)으로 파일 SHA 만 바뀜
