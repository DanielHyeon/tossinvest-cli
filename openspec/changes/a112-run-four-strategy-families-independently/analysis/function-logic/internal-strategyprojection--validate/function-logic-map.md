# Function Logic Map: `Validate`

- Source: `internal/strategyprojection/model.go`
- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`
- Signature: `Validate(params=1, results=1)`
- Source range: `270:1`–`294:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 새 두 분기는 거절 집합을 늘린다. 기본(미관측) 스냅숏은 계속 유효하다(`TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators`).
- 새로 거절되는 정상 입력: 자식이 없는 옛 모양 envelope(7.3 전 엔진이 보낸 것) — 같은 이미지가 엔진 · 콘솔을 함께 교체하므로 혼합 창은 배포 절차 밖(review 잔여).

## Branches and early returns

- Exact AST return nodes: `272:3, 275:3, 280:4, 283:4, 288:3, 291:3, 293:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 271:2 | envelope 3항 |
| B2 | if | 274:2 | runtime identity |
| B3 | range | 277:2 | KR · US |
| B4 | if | 279:3 | 시장 부재/교차 |
| B5 | if | 282:3 | 시장 레코드 판정 |
| B6 | if | 287:2 | **(새)** 레인 자식: 개수 8 · 고정 순서 · 열쇠 · enum · 거절 코드 ⇔ REFUSED(판정 (A)) · 미관측 무사실 · 관측 사슬(물결 ⇔ 트리거, 투입만 시작, 연 사이클만 결과) |
| B7 | if | 290:2 | **(새)** 조정자 자식: 개수 2 · KR,US · 미관측 무사실 · 사유 · 중재 코드 · 정렬 유일 gated · null 아닌 목록 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `snapshot.GeneratedAt.IsZero` | 271:48 |
| `len` | 271:81 |
| `errors.New` | 272:10 |
| `validateRuntimeIdentity` | 274:12 |
| `fmt.Errorf` | 275:10 |
| `fmt.Errorf` | 280:11 |
| `validateMarketProjection` | 282:13 |
| `fmt.Errorf` | 283:11 |
| `validateLanes` | 287:12 |
| `fmt.Errorf` | 288:10 |
| `validateCoordinators` | 290:12 |
| `fmt.Errorf` | 291:10 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
