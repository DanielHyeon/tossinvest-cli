# Function Logic Map: TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse (base 이름, 현재 트리에 없음)

- Source: `internal/strategyprojectionrpc/transport_unix_test.go` — **base revision**, `aeeb209e`
- Base source SHA-256: `b765db14f7c1c5b4e1e7fe58b4f808d5128e5dad6355290239f27eb71c7602e8`
- Signature: `TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse(params=1, results=0)`
- Source range at base: `162:1`–`182:2`
- AST evidence: `ast.json` (`revision: base`); AST 분기 2개.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 결정 55/56.

## 이 lot 이 무엇을 했는가 — 이 함수는 **사라졌다**

이름이 `TestUnixClientRejectsOversizedResponse` 로 바뀌었고, 표의 두 팔 중
`"unknown field"` 가 제거됐다. 남은 팔은 크기 상한 하나다.

## 왜 — 승인된 spec 이 정반대를 요구한다

a112 의 spec `runtime lineage와 health는 lane 단위로 결정적으로 관측된다` 는
"older readers가 additive unknown fields를 무시할 수 있어야 한다 (SHALL)" 를 쓰고,
`legacy projection reader` 시나리오는 "additive lane/coordinator fields를 무시해도
read가 실패하지 않는다" 를 요구한다.

권위 순서는 안전 불변식 → **승인된 OpenSpec** → 현재 HEAD·실행 테스트다. 이 테스트가
고정하던 동작은 승인된 spec 이 금지하는 동작이므로, 테스트가 아니라 동작이 남을 수 없다.

## 엄격함이 실제로 무엇을 지켰는가 — 아무것도

이 응답은 0700 디렉터리 안 0600 소켓 뒤에서 bearer 토큰으로 인증된 **엔진 자신**이
보낸다. 내용은 바로 다음 줄의 `strategyprojection.Validate` 가 의미 단위로 다시
판정하고, 크기는 `MaxProjectionBytes` 가 막는다. 모르는 필드 거절이 유일하게 만든
결과는 **엔진이 필드를 하나 더 실을 때 구버전 콘솔 화면 전체가 죽는 것**이었다.

디스크 파일을 읽는 `readDescriptor` 의 엄격함은 그대로 둔다 — 위협 모델이 다르다.

## Inputs and invariants

- 입력은 base signature 그대로다(테이블 주도 하위 테스트 두 개).
- 불변식: 크기 상한은 그대로 강제된다.

## Branches and early returns

| Branch | AST kind | Source location (base) | 이 lot 의 처분 |
|---|---|---|---|
| B1 | range | 166:2 | 표가 두 행에서 한 행으로 줄었다. 루프 자체는 남는다. |
| B2 | if | 177:4 | 그대로. "거절되지 않으면 실패"의 단언. |

## Calls and live bindings

base 리비전의 호출은 테스트 하네스뿐이다 — `strategyprojection.DormantSnapshot`,
`json.Marshal`, `httptest.NewServer`, `Client.Read`. production 심볼 중 이 함수가 묶고
있던 유일한 계약이 `Client.Read` 의 엄격 디코딩이었고, 그것이 이 lot 이 바꾼 대상이다.

## State mutations and fallbacks

없음. 테스트 서버를 띄우고 응답을 읽을 뿐이며 디스크·저널·브로커에 쓰지 않는다.

## Safety conclusion

- 테스트 전용 함수다. production 동작을 만들지 않는다.
- 잃은 커버리지는 없다: 새 계약을 `a112_additive_fields_test.go` 가 **양쪽으로** 잰다 —
  모르는 필드는 무시하고(`TestClientIgnoresAdditiveFieldsFromANewerEngine`),
  의미가 틀린 스냅샷은 여전히 거절한다(`TestClientStillRejectsASemanticallyInvalidSnapshot`).
