# Function Logic Map: `Console.handleQuarantineReleasePreview`

- Source: `internal/console/exit_quarantine.go` (176-209)
- AST evidence: `ast.json` — AST 분기 5 · return 4 · defer 0
  (source_sha256 `d61ec3a1...` — ast.json 이 정본, a109 §2.5 편집 후 생성)
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
| `quarantine_token` | 서명·용도·유효기간이 맞아야 한다 | `verifyQuarantineToken` | 403 (B2) |
| preview 결과 | capability 가 비면 안 된다 | 엔진 commander | 500 (B4) |
| `WaitSeconds` | 0 이하이면 3초 | 엔진 응답 | 기본값 대입 (B5) |

**불변식**: 이 경로는 **아무것도 변경하지 않는다**. 모든 거절 문구가 「아무것도 변경되지
않았다」로 끝나는 것이 그 계약의 표현이고, 공유 상수도 그 끝맺음을 유지한다.

**불변식 2 (사용자 지시)**: 타이핑 확인·추가 승인 마찰을 넣지 않는다. 이 편집은 문구만
바꾸고 새 확인 단계를 만들지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| **B1 (:178)** | commander 미배선 | 501 거절 + **정직화된 문구** | 조기 return | `TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage` |
| B2 (:184) | token 검증 실패 | 403 거절 | 조기 return | 기존 a079 테스트 |
| B3 (:192) | preview 호출 실패 | `writeQuarantineError` 로 위임 | 조기 return | 기존 a079 테스트 |
| B4 (:196) | capability 가 비었다 | 500 거절 | 조기 return | 기존 |
| B5 (:201) | `WaitSeconds <= 0` | 3초 기본값 | 없음 | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.exitQuarantines` | 엔진 command seam | 미배선은 501 (B1) | AST · :177 |
| `c.refuse` | 거절 화면 | 부작용은 응답뿐 | AST · :179·185·197 |
| `commander.PreviewQuarantineRelease` | 엔진에 preview 요청 | 실패는 `writeQuarantineError` 로 | AST · :188 |
| `c.render` | preview 화면 | — | AST · :204 |

## State mutations and fallbacks

- **상태 변경 없음** (preview 다). 부작용은 HTTP 응답뿐.
- fallback: 없음. 미배선은 거절이고, 그 거절 문구가 이번에 정직해졌다.

## Safety conclusion

- Safe edit boundary: **B1 의 본문 문자열**. 상태·분기·상태코드는 그대로다.
- High-risk impact: **no** — 문구만이다. 다만 이 문구가 운영자의 다음 행동을 정하므로,
  거짓 원인을 말하면 운영자를 무한 재시작에 넣는다(그것이 이 편집의 이유다).
- 금지: 새 확인 마찰 추가. 상태코드·제목 변경(기존 테스트가 제목을 핀한다).
