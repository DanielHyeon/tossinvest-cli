# Function Logic Map: `Client.Read`

- Source: `internal/strategyprojectionrpc/transport.go`
- Current source SHA-256: `c2a7eaf8469508e309803d53120f0cfbb7f6ec35c24c42e27c5fbcb8bd55f617`
- Signature: `Client.Read(params=1, results=2)`
- Source range: `55:1`–`101:2`
- AST evidence: `ast.json`, regenerated from the post-edit worktree; AST 분기 8개.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 결정 56 (독립 리뷰 P1-2).

## 이 lot 이 무엇을 바꿨는가

`decoder.DisallowUnknownFields()` 한 줄을 지웠다. 분기 수는 그대로 8개이고, B6 의
**조건이 넓어진 것이 아니라 좁아졌다** — 모르는 필드는 더 이상 디코딩 실패가 아니다.

## 왜 — 승인된 spec 이 이것을 요구한다

a112 spec `runtime lineage와 health는 lane 단위로 결정적으로 관측된다`:
"older readers가 additive unknown fields를 무시할 수 있어야 한다 (SHALL)", 그리고
`legacy projection reader` 시나리오: "additive lane/coordinator fields를 무시해도
read가 실패하지 않는다".

이 client 가 그 시나리오의 legacy reader 자리다. 엄격 디코딩을 켜 두면 엔진이 필드를
하나 더 실어 보내는 순간 구버전 콘솔의 화면 전체가 죽는다 — 시나리오가 금지하는 실패다.
a112 자신이 `lanes[8]`·`coordinators[2]` 를 추가할 change 이므로 이 충돌은 언제든
터질 예정이었고, 이번 lot 의 envelope 필드가 먼저 건드렸을 뿐이다.

## 무엇이 남아서 지키는가 — 관용은 판정을 버린 것이 아니다

| 위협 | 지키는 것 | 위치 |
|---|---|---|
| 남의 프로세스가 응답 | 0700 control dir · 0600 socket · bearer 토큰 | Dial/`openVerifiedDescriptor` |
| 거대 응답 | `MaxProjectionBytes` | B4 (그대로) |
| 형식이 깨진 JSON | 디코딩 실패 | B6 (그대로) |
| JSON 값 두 개 | trailing 검사 | B7 (그대로) |
| 의미가 틀린 스냅샷 | `strategyprojection.Validate` | B8 (그대로) |

디스크 파일을 읽는 `readDescriptor` 의 `DisallowUnknownFields` 는 **유지한다** —
네트워크 상대가 아니라 파일이라 위협 모델이 다르다.

## Inputs and invariants

- 입력은 AST signature 그대로다.
- 불변식: 이 client 에는 mutation 메서드가 없다(`TestAuthenticatedUnixProjectionCrossNamespaceAndReadOnlyAuthority` 가 잰다).
- 불변식: 돌려주는 값은 `Clone` 이므로 호출자가 store 를 만질 수 없다.

## Branches and early returns

- Exact AST return nodes: `58:3`, `62:3`, `67:3`, `72:3`, `75:3`, `91:3`, `95:3`, `98:3`, `100:2`.

| Branch | AST kind | Source location | Edited by this lot | Disposition |
|---|---|---|---|---|
| B1 | if | 57:2 | 아니오 | nil client·토큰 길이. 그대로. |
| B2 | if | 61:2 | 아니오 | 요청 생성 실패. 그대로. |
| B3 | if | 66:2 | 아니오 | 전송 실패. 그대로. |
| B4 | if | 71:2 | 아니오 | **크기 상한** — 읽기 실패 또는 MaxProjectionBytes 초과. 그대로 남는다. |
| B5 | if | 74:2 | 아니오 | HTTP 상태 코드. 그대로. |
| B6 | if | 90:2 | **예 — 관용해졌다** | 디코딩 실패. `DisallowUnknownFields()` 제거로, 모르는 필드는 더 이상 이 가지를 타지 않는다. 형식이 깨진 JSON 은 여전히 탄다. |
| B7 | if | 94:2 | 아니오 | 값이 하나인가. 그대로 — 뒤에 붙은 두 번째 JSON 값은 여전히 거절. |
| B8 | if | 97:2 | 아니오 | `strategyprojection.Validate` — 의미 판정. 그대로. |

## Calls and live bindings

| Callee expression | Source location |
|---|---|
| len | 57:34 |
| errors.New | 58:18 |
| http.NewRequestWithContext | 60:18 |
| request.Header.Set | 64:2 |
| c.http.Do | 65:19 |
| response.Body.Close | 69:8 |
| io.ReadAll | 70:15 |
| io.LimitReader | 70:26 |
| len | 71:19 |
| errors.New | 72:18 |
| fmt.Errorf | 75:18 |
| json.NewDecoder | 89:13 |
| bytes.NewReader | 89:29 |
| decoder.Decode | 90:12 |
| fmt.Errorf | 91:18 |
| decoder.Decode | 94:12 |
| errors.Is | 94:40 |
| errors.New | 95:18 |
| strategyprojection.Validate | 97:12 |
| strategyprojection.Clone | 100:9 |

## State mutations and fallbacks

- 지역 변수만 쓴다. fallback 없음 — 어떤 실패도 오류이지 기본 스냅샷이 아니다.

## Safety conclusion

- 읽기 전용 관측 transport 다. 주문·손절·사이징·Guardian·원장 경로에 닿지 않는다.
- 이 완화는 **거절 집합을 줄인다.** 줄어든 부분이 정확히 "모르는 필드 하나 때문에
  화면 전체가 죽는 것"이며, 그것이 spec 이 금지한 실패다.
