# Function Logic Map: `readDescriptor`

- Source: `internal/strategyprojectionrpc/transport.go` (160-184)
- AST evidence: `ast.json` — revision `current`. AST 분기 5
- Risk scan: `risk-pattern-report.md`

Fix 라운드가 이 함수를 **둘로 갈랐다**(design D1-2). 앞부분 — 경로 모양·디렉터리
0700·정확히 0600인 정규 파일·no-follow open·열기 전후 inode 동일성 — 은 새 함수
`openVerifiedDescriptor`로 나갔고, 여기 남은 것은 **내용**뿐이다. 분기는 10 → 5로 줄었다.

## 왜 갈랐는가

회수는 형식만 봐야 하고 소비는 내용까지 봐야 하기 때문이다.

- **형식**은 "이것이 우리가 만든 파일인가"를 묻는다. 답이 아니오면 회수도 소비도 거부다.
- **내용**은 D2 이후 **어떤 판정에도 쓰이지 않는다.** 생존 판정은 socket connect probe이고
  descriptor의 PID는 읽지도 않는다. 그래서 회수가 내용 파싱 실패를 거부로 읽으면 그
  거부는 순수한 손실이다 — 우리 자신이 반쯤 쓴 파일 때문에 부팅이 영구히 막힌다
  (A1 F2 실측 3/3).

가르지 않고 회수 쪽에서 "파싱 실패는 무시"로 처리할 수도 있었지만, 그러면 형식 검사와
내용 검사가 한 함수 안에서 서로 다른 엄격도를 갖게 되고 그 차이가 호출부에 없다.
두 함수로 가르면 **호출부가 자기 엄격도를 이름으로 고른다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 파일 형식 | `openVerifiedDescriptor`의 계약 전부 | 그 함수 | B1 — 오류 그대로 전파 |
| 본문 크기 | ≤ 4096 바이트 | `io.LimitReader` | B2 `unreadable or oversized` |
| JSON | 알 수 없는 필드 금지 | `DisallowUnknownFields` | B3 `descriptor JSON invalid` |
| 값의 개수 | 정확히 하나 | 두 번째 `Decode`가 `io.EOF` | B4 `must contain one value` |
| 필드 | 스키마 v1 · socket 이름 일치 · 토큰 32자 이상 · PID > 0 | 상수 비교 | B5 `descriptor fields invalid` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 형식 검사·열기 실패 | 없음 | 원본 오류(`errDescriptorChanged` 포함) | `TestDescriptorPublicationIsAtomic` |
| B2 | 읽기 실패 또는 4096 초과 | 없음 | `unreadable or oversized` | 없음 |
| B3 | JSON이 아니다 | 없음 | `descriptor JSON invalid` | `TestStartRecoversFromTruncatedDescriptorLeftover` (회수 쪽에서 이 실패가 **일어나지 않아야** 함을 잰다) |
| B4 | 값이 둘 이상이다 | 없음 | `must contain one value` | 없음 |
| B5 | 필드가 유효하지 않다 | 없음 | `descriptor fields invalid` | 없음 |

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `openVerifiedDescriptor` | 형식 검사 + no-follow open | 오류 그대로 전파. `errDescriptorChanged`는 rename 경합의 정상 답이다 | AST calls, B1 |
| `io.ReadAll(io.LimitReader(...))` | 크기 상한 | 초과는 거부 | AST calls, B2 |
| `json.Decoder` (`DisallowUnknownFields`) | 엄격 파싱 | 알 수 없는 필드 거부 | AST calls, B3·B4 |

## State mutations and fallbacks

- 없다. 조회 전용이고 디스크를 바꾸지 않는다. 파일 핸들은 `defer`로 닫는다.
- fallback 없음. 실패는 그대로 올라가고 `Dial`이 그것을 그대로 전파한다.

## Safety conclusion

- Safe edit boundary: 필드 검사(B5)를 느슨하게 하면 남이 놓은 descriptor의 토큰을
  우리가 쓰게 된다. 형식 검사는 `openVerifiedDescriptor`가 소유하므로 여기서 건드릴
  것이 아니다.
- High-risk impact: no (조회 전용) — 다만 회수가 이 함수를 다시 부르게 만드는 편집은
  A1 F2를 되살린다. 뮤테이션 M16이 그 회귀를 잰다.

## gstack Fix 라운드 (2026-08-14) — 이 함수는 편집하지 않았다

분기·return 불변(5·6). 파일 해시가 움직인 것은 같은 파일의
`openVerifiedDescriptor` 주석에서 잘못된 상호 참조(「아래 reclaimStaleControlDirectory」
— 그 함수는 다른 파일에 있다)를 고쳤기 때문이다.
