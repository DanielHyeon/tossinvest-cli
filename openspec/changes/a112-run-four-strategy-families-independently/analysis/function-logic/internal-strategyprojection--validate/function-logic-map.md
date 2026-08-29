# Function Logic Map: `Validate`

- Source: `internal/strategyprojection/model.go`
- Current source SHA-256: `0662dc5ab11eda0213bc4e887cdccbb71feb5115bfd5b4627dc71de81090d08f`
- Signature: `Validate(params=1, results=1)`
- Source range: `266:1`–`283:2`
- AST evidence: `ast.json`, regenerated from the post-edit worktree; AST 분기 5개.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 결정 54의 `ConfigDigest`/`BuildDigest` 노출.

## 이 lot 이 무엇을 바꿨는가

envelope 검사 바로 뒤에 분기 하나(B2)가 늘었다. runtime identity 가 반쪽이거나 비정규 digest 면
스냅샷 전체를 무효로 만든다.

## 왜 판정이 필요한가 — 이 값은 신뢰 경계를 건넌다

스냅샷은 `strategyprojectionrpc` 의 Unix socket 을 건너 콘솔·httpapi 프로세스로 온다. 받는 쪽은
`WithRuntimeIdentity` 를 거치지 않은 값을 볼 수 있다. 그리고 이 두 문자열은 운영자가 **그대로
서명 매니페스트에 옮겨 적는** 값이라, 검사 없이 화면에 올리면 형식이 깨진 숫자를 사람이 받아 적는다.

판정 규칙은 이 패키지의 다른 짝 필드와 같다 — `pairedIdentity` 로 「함께 있거나 함께 없거나」,
그리고 있으면 `validDigest`(소문자 64자리 hex).

## Inputs and invariants

- 입력은 AST signature 그대로다.
- 불변식: 판정은 **거절만** 한다. 값을 고치거나 채우지 않는다.
- 불변식: dormant/unavailable 스냅샷(둘 다 nil)은 유효하다 — 엔진이 만들지 않은 스냅샷에는
  엔진의 build 를 알 방법이 없고, 그 부재가 정직한 답이다.

## Branches and early returns

- Exact AST return nodes: `268:3`, `271:3`, `276:4`, `279:4`, `282:2`.

| Branch | AST kind | Source location | Edited by this lot | Disposition |
|---|---|---|---|---|
| B1 | if | 267:2 | 아니오 | envelope 3항(schema·시각·시장 수). 기존 그대로. |
| B2 | if | 270:2 | **예 — 이 lot 이 추가** | runtime identity 판정. `validateRuntimeIdentity` 로 위임한다. |
| B3 | range | 273:2 | 아니오 | KR·US 두 시장 루프. 기존 그대로. |
| B4 | if | 275:3 | 아니오 | 시장 부재/교차 판정. 기존 그대로. |
| B5 | if | 278:3 | 아니오 | 시장 레코드 판정 위임. 기존 그대로. |

## Calls and live bindings

| Callee expression | Source location |
|---|---|
| snapshot.GeneratedAt.IsZero | 267:48 |
| len | 267:81 |
| errors.New | 268:10 |
| validateRuntimeIdentity | 270:12 |
| fmt.Errorf | 271:10 |
| fmt.Errorf | 276:11 |
| validateMarketProjection | 278:13 |
| fmt.Errorf | 279:11 |

## State mutations and fallbacks

- 상태 변경 없음. fallback 없음 — 판정 실패는 오류이지 기본값이 아니다.

## Safety conclusion

- 읽기 전용 관측 계약의 판정이다. 주문·손절·사이징·Guardian·원장 경로에 닿지 않는다.
- 추가된 B2 는 **거절을 늘릴 뿐** 어떤 스냅샷도 새로 통과시키지 않는다. 기존 스냅샷은 두 필드가
  모두 nil 이므로 `pairedIdentity` 를 통과한다 (`TestDormantSnapshotHasNoRuntimeIdentity` 실측).
