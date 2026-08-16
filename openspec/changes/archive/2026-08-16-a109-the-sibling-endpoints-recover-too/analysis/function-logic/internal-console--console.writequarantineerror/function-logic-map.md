# Function Logic Map: `Console.writeQuarantineError`

- Source: `internal/console/exit_quarantine.go` (257-283)
- AST evidence: `ast.json` — AST 분기 9 · return 0 · defer 0
  (source_sha256 는 ast.json 이 정본, a109 §2-fix F5 편집 후 재생성)
- Risk scan: `risk-pattern-report.md`
- **a109 편집 대상: 미배선 거절의 본문 문자열 한 곳** — 원인 단정(「엔진 control plane이
  격리 해제를 제공하지 않는다」= 사실상 「빌드가 낡았다」)을 공유 상수
  `quarantineUnwiredDetail` 로 바꾼다. **문구만이고 분기·상태·화면은 그대로다**(D3a-2).
- **왜 세 함수인가**: 같은 문장이 세 벌 있었고 설계는 한 곳(`writeQuarantineError`)만
  인용했다(issues.md T2-2). 정정의 단위는 줄이 아니라 **값**이므로 셋을 함께 고친다 —
  운영자가 먼저 만나는 것은 대개 preview 경로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `err` | `exitquarantine` 의 sentinel 들 또는 그 밖 | 엔진 commander | sentinel 마다 다른 상태코드·제목·본문 |
| 상태코드 매핑 | 501/404/412/425/410/403/400 | 이 함수 | 매핑 자체는 a079 계약이고 이번에 바꾸지 않았다 |

**불변식**: 모든 가지가 「아무것도 변경되지 않았다」로 끝난다. 거절이 상태를 남기지
않는다는 사실을 운영자가 매번 읽을 수 있어야 한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:258) | 오류 종류 분기 | 없음 | 아래 가지들 | 기존 a079 표 테스트 |
| **B2 (:259)** | `ErrUnwired` | 501 거절 + **정직화된 문구** | — | `TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage` · 기존 a079(제목) |
| B3 (:262) | `ErrNotQuarantined` | 404 | — | 기존 a079 |
| B4 (:265) | `ErrVersionMismatch` | 412 | — | 기존 a079 |
| B5 (:268) | `ErrCapabilityTooEarly` | 425 | — | 기존 a079 |
| B6 (:271) | `ErrCapabilityExpired` | 410 | — | 기존 a079 |
| B7 (:274) | `ErrCapabilityInvalid` | 403 | — | 기존 a079 |
| B8 (:277) | `ErrConfirmationRequired` | 400 | — | 기존 a079 |
| B9 (:280) | 그 밖 | 400 + 원문 | — | 기존 a079 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.Is` | sentinel 판별 | 순수 | AST · :259–280 |
| `c.refuse` | 거절 화면 | 부작용은 응답뿐 | AST · 각 가지 |

## State mutations and fallbacks

- 상태 변경 없음. 이 함수는 오류를 화면으로 옮기기만 한다.

## Safety conclusion

- Safe edit boundary: **B2 의 본문 문자열**. 나머지 여덟 가지의 상태코드·제목·본문은
  a079 계약이므로 건드리지 않는다.
- High-risk impact: **no** (문구만).
- 금지: 상태코드·제목 변경(기존 a079 테스트가 제목을 핀한다).
