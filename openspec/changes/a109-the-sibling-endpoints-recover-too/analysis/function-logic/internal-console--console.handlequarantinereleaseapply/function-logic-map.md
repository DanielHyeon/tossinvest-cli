# Function Logic Map: `Console.handleQuarantineReleaseApply`

- Source: `internal/console/exit_quarantine.go` (211-242)
- AST evidence: `ast.json` — AST 분기 3 · return 3 · defer 0
  (source_sha256 는 ast.json 이 정본, a109 §2.5 편집 후 생성)
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
| `c.exitQuarantines()` | 배선됐거나 아니거나 | 콘솔 조립 | 미배선이면 501 + `quarantineUnwiredDetail` (B1) |
| `capability` | 비면 안 된다 | preview 가 발급 | 403 (B2) |
| `confirm` | `"yes"` 여야 실제 적용 | 폼 | 엔진이 판정 (`ErrConfirmationRequired`) |
| 해제 결과 | 성공하면 리다이렉트 | 엔진 commander | 실패는 `writeQuarantineError` (B3) |

**불변식**: 해제는 **기준선을 지우지 않는다** — 진입·초기 손절·저장된 보호선은 그대로이고
다음 관측부터 판정이 재개된다. 그 사실을 알리는 notice 문구는 이번 편집에서 건드리지 않았다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| **B1 (:213)** | commander 미배선 | 501 거절 + **정직화된 문구** | 조기 return | `TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage` |
| B2 (:219) | capability 가 비었다 | 403 거절 | 조기 return | 기존 a079 테스트 |
| B3 (:227) | 해제 호출 실패 | `writeQuarantineError` 로 위임 | 조기 return | 기존 a079 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.exitQuarantines` | 엔진 command seam | 미배선은 501 (B1) | AST · :212 |
| `commander.ReleaseQuarantine` | **유일한 상태 변경** | 실패는 `writeQuarantineError` | AST · :224 |
| `http.Redirect` | 성공 후 화면 이동 | — | AST · :241 |

## State mutations and fallbacks

- 유일한 상태 변경은 `ReleaseQuarantine` 이고, 그것은 `exit_snapshot_quarantines` 한 표만
  쓴다(파일 주석의 a081 근거). 이번 편집은 그 경로에 닿지 않는다.
- fallback: 없음.

## Safety conclusion

- Safe edit boundary: **B1 의 본문 문자열**.
- High-risk impact: **no** (문구만). 단 이 화면이 없으면 격리된 포지션의 미판정이
  유지되므로, 문구가 원인을 잘못 말하면 그 상태가 길어진다.
- 금지: 새 확인 마찰 추가. 상태 변경 경로 수정.
