# 1.6 · 1.7 — 계좌 일치 비교의 자리(D1a)와 진입 래치의 자리(D4)

HEAD `e4165fa3d`. AST: `cited-ast/internal-attest--attestation.verify.ast.json` · `internal-attest--sameaccountmasked.ast.json` ·
`internal-execgw--gateway.checkentry.ast.json` · `internal-execgw--entrygate.checkaccountentrylocked.ast.json` · `internal-app-engine--restorealertentrylatch.ast.json` ·
`internal-execgw--allreasoncodes.ast.json`.

## 1.6 계좌 일치 비교 (D1a)

- `Attestation.Verify` 의 계좌 비교는 B6@228 `!sameAccount(a.AccountRef, accountRef)` → `ErrAccountMismatch`. `sameAccount`(attest.go:296)는 **비공개**, 규칙은
  `accountDigits(a) != "" && accountDigits(a) == accountDigits(b)`(숫자만 남겨 동일 — 하이픈 무시, 가림 `****` 은 숫자를 잃으므로 다름).
- 공개 `SameAccountMasked(full, other)`(148-153, AST B1@149 빈 값 거절)는 `sameAccount(full, other) ‖ Mask(full) == TrimSpace(other)` — **가림 꼴도 같다고 보는
  더 관대한 규칙**이고, 소비자는 soak 의 감독 증거 결속(검증 기록은 계좌를 가린 채 씀). `Verify` 의 AST calls 에 `SameAccountMasked` 없음 → **Verify 와 같은 비교가 아님.**
- `Expired(now)`(261)·`MissingEndpoints`(242)는 공개지만 계좌와 무관.
- **결론: Verify 와 같은 비교의 공개 진입점은 없다.**

**구현 자리 제안(구현하지 않음):** `internal/attest/` 의 **새 파일**(예: `account_match.go`)에 새 메서드 하나 —
`func (a Attestation) NamesAccount(accountRef string) bool { return sameAccount(a.AccountRef, accountRef) }`(이름은 2단계 확정).
`Verify` 본문 무편집(gate ⑤ required 0 유지 — attest 패키지에서 기존 함수 편집 0). 「Verify 와 같은 비교」는 두 함수가 같은 `sameAccount` 를 부른다는
구조로 성립하고, 시험은 (1) 하이픈 다른 같은 계좌 → true, (2) 가림 꼴 → false(= Verify 와 같음), (3) 빈 AccountRef → false(Verify 는 그 전에 `ErrIncomplete`) 로 Verify 와 같은 답을 고정.
공용 판정(인터록 · 감시)은 `Verify` 오류가 `ErrExpired` 일 때만 이 메서드를 부른다 — endpoint 누락이면 Verify 가 이미 B6@228 을 통과했으므로.

## 1.7 EntryGate 래치 (D4)

### API 와 사유
- `EntryGate.Block(reason, detail)`(retry.go:558 — 없을 때만 삽입 · 처음 설명 유지 · revision++), `Clear(reason)`(569 — 해제 세대 +1), `Blocks()`(662).
- 사유 상수는 `internal/execgw/reason.go` 의 `ReasonCode` 문자열. 래치에 쓰이는 사유의 가장 가까운 선례: `ReasonStrategyCentralIntegrity`(273 — 「해제 경로는 **재시작뿐**」,
  운영 해제 명령 없음). 새 사유 가칭 `ReasonCapabilityAttestationStale = "capability_attestation_stale"`.

### 닫힌 목록(새 사유를 더하면 함께 바꿔야 하는 것)
| 자리 | 종류 | 새 사유를 안 넣으면 |
|---|---|---|
| `execgw.AllReasonCodes()`(failclosed.go:254-311, 분기 0) | 함수 본문 — **기존 함수 편집**(FLM 번들) | `TestReasonCodeEnumIsStable` 이 golden 과 비교 — 새 사유를 낼 수 있는데 열거에 없음 |
| `internal/execgw/testdata/reason_codes.golden` | 데이터 | 위 시험 실패 |
| `a098_sender_down_reason_test.go` `reasonCodesRegisteredAfterA098`(시험 파일 패키지 변수) | 길이 단언(`29+1+len(var)`) | `TestTheSenderDownReasonIsRegisteredInTheEnumeration` 실패 — 변수에 이름으로 더하는 것이 그 파일의 규칙(주석 30-32) |
| `execgw.latchOrder`(retry.go:677-713, 패키지 변수) | 보고 순서 | 차단은 하지만 `checkAccountEntryLocked` B3@638 맵 순회로 **사유 이름이 비결정**(a098 시험 주석 93-116 이 경고) |
| `a112_strategy_central_integrity_reason_test.go` | 중앙 무결성이 모드보다 앞 · 송신자 정지보다 뒤 | 새 사유를 중앙 무결성 뒤 · 모드 앞에 붙이면 기존 쌍의 상대 순서 무변 → 통과 예상 |
| 정본 | engine-safety 1342 「자기 사유 코드를 가져야 한다」(배달 실행자) · risk-management:7 의 괄호(「401/403·SLO 위반·RECONCILE·recovery 미완료」)는 **예시 목록**(알림 미전달 · 모드 · 중앙 무결성도 없음) | 닫힌 목록 아님 → order-execution · exit-policy · risk-management MODIFIED 불요(intent D 의 확인 결과) |
| 콘솔 · 상태 출력 | `alertops.entryBlockReasons`(alertops.go:187 — `Blocks()` 의 키를 정렬해 그대로) · position_policy_command 는 RECONCILE 블록만 | 일반 표시 — 깨질 열거 없음, 새 사유는 `engine alert-ack` 류 출력에 이름으로 나타남 |

### 기동 때 래치를 세우는 선례
- `restoreAlertEntryLatch`(engine gateway.go:158-173, AST B1@160 읽기 실패 → 오류, B2@163 0건 → nil, 그 밖 `gate.Block(ReasonAlertUndelivered, 건수)`) — `buildGateway` 안 287
  (**인터록보다 먼저**), 원장에서 유도, 「only ever latches」. 운영 모드 복원 실패 래치(`bindOperatingModeProjection` → mode_projection_wiring.go:32)도 같은 자리.
- 둘 다 `buildGateway`(**a166 census 키**) 안이고, 우리 판정(낡음)은 인터록 단계에서야 알 수 있어 그 자리에 둘 수 없음.

### 새 래치를 세울 자리 후보
| 후보 | 편집 | census | 장단 |
|---|---|---|---|
| ① `runInterlock` 안 — 청산 전용 결과 갈래에서 `facts.entry.Block(…)` | runInterlock(키 아님) + `gateFacts` 구조체 칸 + **NewContext 의 gateFacts 리터럴 한 줄**(`entry: wiring.entry`) | NewContext 키 1(어차피 B16 때문에 갱신) | 판정 · 감사 · 래치가 한 함수 — runInterlock 단위 시험이 래치까지 단언. **권고** |
| ② `NewContext` 안 — runInterlock 뒤 `if automation.<표지> { wiring.entry.Block(…) }` | NewContext 만 | 키 1 | 판정과 집행이 두 함수로 갈림 — 변이 「래치 삭제」가 NewContext 수준 시험에서만 잡힘 |
| ③ 새 함수(새 파일)를 NewContext 가 부름 | NewContext 한 줄 + 새 함수 | 키 1 | ②와 같고 함수 하나 더 |
| ✗ `buildGateway` | — | 키(buildGateway) 추가 | 판정 전이라 불가 |

순서: 래치는 Context 가 반환되기 **전**에 서야 하고(루프 · 전략 조립 · 보조 실행자는 전부 그 뒤), ①·②·③ 모두 만족. NewContext 의 `opts.BudgetFactory.Budget` 자리(sites.json)는 무변.

### 재시작 뒤
- `buildGateway` 가 매 기동 `execgw.NewEntryGate(in.clock, nil)`(engine gateway.go:265) — 빈 `latches`. 기동 복원은 알림 미전달(원장) · 운영 모드(원장) · RECONCILE(원장)뿐.
  그래서 메모리 래치(새 사유 포함)는 재시작으로 사라지고, 새 사유는 그 기동의 인터록이 다시 판정해 다시 세운다(낡았으면). 이 사유를 `Clear` 하는 코드는 0(grep).
- 결과: attestation 을 고친 뒤 재기동 → 래치 없음 · 수락 감사(정본 Scenario 「attestation 을 갱신한 뒤의 재기동」), 고치지 않고 재기동 → 다시 섬(「낡은 채 재기동」).
