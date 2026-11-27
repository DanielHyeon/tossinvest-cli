# a208 design — 편입 전용 보호선 정책과 60초 유예

근거 표기: `Cn` = `analysis/ast-evidence/claims.md` 의 AST 근거 주장, `S§n` = `analysis/stockos-reference.md`. 분기 ID(`B12@355` 등)는 base
`baeddfe31` 의 `tools/logic-map` 산출물 기준이다. 생산 코드 변경은 이 단계에 없다.

## 1. 사실(F)

### 1.1 편입이 정해지는 자리와 정책이 정해지는 자리(오늘)

| 단계 | 자리 | 하는 일 |
|---|---|---|
| 편입 후보 판정 | `ReconcileDriver.adopt`(internal/app/engine/adoption.go:215) | 신선 조건 · 대사 · include/exclude 통과 후보를 `adoptOne` 으로 |
| 합성 손절 | `adoptOne` adoption.go:374 → `exitpolicy.SyntheticStop(observed, DefaultStopPct)` | 손절 = 편입 관측가 × (1 − pct), pct ∈ [0.02, 1) (exitpolicy/adoption.go:46-72) |
| **정책 선택** | `adoptOne` adoption.go:392 `ExitPolicyID: d.opts.CommonPolicy` | 엔진 공통 정책을 그대로 복사 — 편입 갈래 0(C1) |
| 영속 | `Journal.AdoptPosition`(journal/adoption.go:178) | 기록 + `positions.adoption_id` 한 트랜잭션, set-once |
| exit state 종류 | `OpenAdoptedExitState` → `policyKindForID`(journal/adoption.go:468-473) | 빈 ID → RATCHET, 아니면 LADDER(C2) |
| 정체성 고정 | `seedPolicyIdentity`(journal/exit_state.go:206) → `LegacyLadderPolicyIdentity(id, adopted)` | RUNNER ∧ 편입만 `1.0.0-adopted.1`(C3) |
| 판정 시 표 | `ExitObserver.ladderFor`(exitloop.go:1324) → `CommonLadderForPosition(id, Adopted())` | RUNNER ∧ 편입만 부분익절 0(C4) |
| 판정 | `judge` B7@1047 → `judgeLadder` / `judgeRatchet`(C5) → `EvaluateLadder`(C6) | 손절 B25@439, 익절 B28~B31 |

### 1.2 오늘 편입 포지션이 받는 것

| 공통 정책 | 편입 포지션 정책 | 부분익절 | 본전 · 보호선 상향 |
|---|---|---|---|
| 없음(RATCHET) | RATCHET | +1.0R 에 40%(C8, ratchet.go:438-448) | +0.8R 실질 본전(C8) 등 |
| BALANCED | BALANCED | 2.5% · 4.0% 에 25%, 6.0% 전량(common_policy.go:34-40) | rung 잠금 |
| HYBRID_50 | HYBRID_50 | 3.0% 25% · 4.8% 1/3(:61-67) | rung 잠금 + runner trail |
| RUNNER | **편입 RUNNER 사본** | 없음(C4 · C7) | rung 잠금 0/2.0/3.5% at +2.5/4.5/7.0% |

운영 공통 정책 값은 이 change 에서 읽지 않았다(사람 확인 Q6).

### 1.3 StockOS(S§1~§3)

- 부분익절 제거는 **RUNNER 편입에만**(`aee.py:6796-6807`, `external_runner_v1`). 표는 TossOS RUNNER 와 같다.
- 60초 settle window: 상태 행 `created_at`(영속, 벽시계 UTC) 기준 `age < 60`(`:6978-6998`). 창 안에서 익절 · 보호 상향 action 을
  HOLD 로 낮추고(`:7001-7024`, MOVE_STOP_TO_BREAKEVEN 포함 `:307`), 사다리를 끄고(`:6868`), 고점을 동결한다(`:2483-2492`).
  hard stop · time exit 는 통과.
- 손절 기준은 평균단가 × (1 − 1.2%)(`:1867-1871`) — 사용자가 가져오지 말라고 한 결함.

## 2. 결정 요약

| ID | 결정 | 대안(기각 사유) |
|---|---|---|
| D1 | 토글이 참이면 자동 편입은 `COMMON_LADDER_RUNNER` 를 기록 → 기존 편입 RUNNER 사본 재사용 | (b) 공통 정책마다 편입 사본 신설 — 정체성 · digest 2 개 + RATCHET 사본(부분 비율 0 은 `RatchetConfig.Validate` 범위 밖, 검증 안 함)이 더 필요, 같은 사용자 결정에 비해 표면 3 배. (c) RATCHET 유지 — 40% 부분익절이 남아 결정 위반 |
| D2 | 설정 키 `engine.adoption.floor_only_exit`, `omitempty` | exit_policy 블록 — 의미가 「편입」이고 편입 블록 seam 이 이미 콘솔 · audit 에 있음 |
| D3 | 60초 유예 = 평가 입력 `LadderInput.ProfitSideHeld`, 동결형(G1) | 판정 생략 · 가격 고정 — 4.3 |
| D4 | 유예 기준 시각 = 편입 기록 `observed_at`, 비교 대상 = 시세 `FetchedAt` | 단조 시계 — 재시작을 못 넘음 |
| D5 | 편입 알림 문구를 토글에 따라 둘로 | 문구 하나 — 참일 때 「중간 익절 동일 적용」이 거짓이 됨 |

## 3. D1 · D2 — 편입 전용 정책의 선택과 설정

### 3.1 선택

`adoptOne` 의 정책 ID 한 자리(adoption.go:392)만 바꾼다:

```go
// 편입 전용 보호선 정책이 켜져 있으면 공통 정책과 무관하게 RUNNER 편입 사본을 기록함(a208 D1)
exitPolicyID := d.opts.CommonPolicy
if d.opts.Adoption.FloorOnlyExit {
	exitPolicyID = exitpolicy.CommonLadderRunner
}
```

그 뒤 경로는 전부 기존 코드다: C2(LADDER) → C3(`1.0.0-adopted.1`) → C4(부분익절 0) → C7(익절 갈래 B29 · B30 불가). 편입 드라이버는
설정 블록을 이미 통째로 받으므로(`Context.ReconcileDriver` reconcileloop.go:383, C15) 배선 변경이 없다.

**편입 포지션에만 적용되는 근거.** 이 자리는 `adoptOne` 안에 있고, `adoptOne` 의 유일한 호출자는 `ReconcileDriver.adopt` 다(CodeGraph
callers). 엔진 진입 포지션의 정책은 `openState` 의 비편입 갈래(exitloop.go:820-825 `protection.Policy` / `o.opts.CommonPolicy`)에서 정해지고
이 change 는 그 갈래를 고치지 않는다. 운영자 재편입(READOPT)은 `position_policy_command.go:359-366` 에서 따로 정하고 역시 고치지 않는다.

**토글 뒤집기는 살아 있는 포지션을 바꾸지 않는다.** 정책은 편입 기록(`position_adoptions.exit_policy_id`)과 exit state 에 스냅숏되고 판정은
그 값을 읽는다(`ladderFor` 는 `m.state.PolicyID` — 설정이 아님). 기존 SHALL 「전역 정책 변경은 활성 포지션을 재해석하지 않는다」와 같은 구조.

### 3.2 설정 표면(C12 · C13)

| 자리 | 변경 |
|---|---|
| `config.Adoption` | `FloorOnlyExit bool \`json:"floor_only_exit,omitempty"\`` — 거짓이면 직렬화 키가 없어 콘솔 저장의 파일 바이트가 지금과 같음 |
| `rawAdoption` | `FloorOnlyExit bool \`json:"floor_only_exit"\`` |
| `mergeAdoption` | `next.FloorOnlyExit = raw.FloorOnlyExit`. 거부된 블록은 지금처럼 전면 0(B3@304) → 거짓 |
| `Service.LoadRawEngineAdoption` | 필드 이월(66-70 줄 조립부) |
| `Service.SaveEngineAdoption` | `next` 조립(81 줄)에 필드 이월 |
| `Console.handleSettingsSave` | 버리던 `Load` 값을 받아 `next.FloorOnlyExit = current.FloorOnlyExit`(폼에 없는 값을 저장이 끄지 않게) |
| `Adoption.validate` | 변경 없음 — 토글만 참이고 편입이 꺼져 있으면 편입이 일어나지 않아 의미가 없을 뿐 위험하지 않음(새 거부 규칙 = 새 가드, 도입 안 함) |

`handleSettingsInclude` 는 이미 `current` 를 읽어 고친 뒤 저장하므로 `SaveEngineAdoption` 이월만으로 값이 남는다.

### 3.3 audit(§0.5)

- 엔진 기동: `recordGateSettings` 표 **끝**에 `{audit.ActionAdoptionToggle, "engine.adoption.floor_only_exit", strconv.FormatBool(adoption.FloorOnlyExit), ""}`.
  `RecordChange` 가 직전 값과 같으면 쓰지 않으므로(a163 시험이 그 동작을 단언) 배포 뒤 첫 기동에 `false` 한 행, 이후 바뀔 때만 한 행.
  `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` 의 「마지막 행 = source_age_report」 단언은 「끝에서 둘째」로 고친다.
- 콘솔 저장 시점 기록(`recordAdoptionSave`)은 고치지 않는다: 콘솔은 이 값을 바꾸지 못하고(3.2 이월), 바꾸는 길은 설정 파일 편집뿐이며
  그 변경은 다음 기동 기록이 남긴다(한도 값의 수동 편집과 같은 trail). 사람이 콘솔 체크박스를 원하면(Q3) 그때 이 함수에 행을 더한다.

## 4. D3 · D4 — 60초 유예

### 4.1 의미(G1, StockOS 동결형)

토글이 참인 관측기에서 편입 포지션(`p.Adopted()`)의 사다리 평가가 **유예 중**이면 `EvaluateLadder` 는:

| 단계(C6) | 유예 아님(지금) | 유예 중 |
|---|---|---|
| B12@355 워터마크 흡수 | 관측가 > 고점이면 고점 = 관측가 | **안 함**(고점 = 이전 고점) — StockOS `:2490` 고점 동결과 같음 |
| B13~B15 도달 rung | 고점 수익률로 계산 | **안 함**(newIndex = 이전 rung) |
| B16~B21 잠금가 합성 | rung 이 있으면 합성 | **안 함**(기준선 = 이전 기준선) |
| B22@420 승격 | rung 이 오르면 승격 | 없음(위 결과) |
| B23@427 완료 | 그대로 | 그대로 |
| **B25@439 손절** | 관측가 < 기준선 → 전량 손절 | **그대로** — 같은 관측가, 기준선 = 이전 기준선 |
| B26@441 손절 미해소 억제 | 그대로 | 그대로 |
| B27@453 승격 없음 return | — | 항상 여기서 return |

floor-only RUNNER 에는 익절 발의가 없으므로(C7) 유예가 실제로 막는 것은 **rung 승격으로 인한 보호선 상향**뿐이다(4.5).

구현 모양(편집 예정, 최소):

```go
// LadderInput 에 더함 — 거짓이면 직렬화하지 않아 저장 JSON 이 a208 이전과 같음
ProfitSideHeld bool `json:",omitempty"`

// EvaluateLadder 1~4 단계: 유예 중에는 워터마크 · rung · 기준선을 이전 값으로 둠(손절 비교는 5 단계에서 그대로)
probe := previousHigh
if !in.ProfitSideHeld && observed.Cmp(probe) > 0 { probe = observed }
...
newIndex := in.State.ActivatedRung
if !in.ProfitSideHeld { /* 기존 rung 루프 */ }
```

`newIndex == ActivatedRung` 이면 B16 합성은 이미 합성된 기준선과 같은 잠금을 다시 넣을 뿐이고 B22 는 거짓이므로, 루프와 흡수 두 곳만 막으면
나머지는 기존 분기가 그대로 처리한다(편입 직후에는 `ActivatedRung = NoRung` 이라 B16 자체가 거짓).

`canonicalLadderInput` 은 `in.ProfitSideHeld` 가 참일 때만 토큰 `"adoption-grace"` 를 끝에 덧붙인다 → 거짓이면 입력 digest · snapshot id ·
decision id 가 지금과 같고, 참이면 「같은 숫자인데 다른 판정」이 다른 입력으로 식별된다(C10). 복구 재평가는 저장된 입력 JSON 의 필드를 그대로
읽어 같은 답을 낸다(C9).

### 4.2 누가 · 언제(D4)

`judgeLadder` 가 평가 입력을 만들기 직전(exitloop.go:1150 앞)에 정한다:

```
held := o.opts.AdoptionFloorOnly && m.position.Adopted() && m.state.ActiveRung == NoRung &&
        o.inAdoptionGrace(ctx, m, quote.FetchedAt)
```

- **시계.** 기준 = 편입 기록 `observed_at`(adoption.go:391 `journal.RFC3339(d.clk.Now())`, `position_adoptions` 에 영속, set-once).
  비교 대상 = 이 관측의 `quote.FetchedAt`(시세 읽기 시각, 판정 기록 `ObservedAt` 과 같은 값). 유예 ⇔ `0 ≤ FetchedAt − observed_at < 60s`.
  둘 다 같은 벽시계(C16). 단조 시계는 재시작을 넘지 못하므로 쓰지 않는다 — StockOS 도 영속 `created_at` 벽시계다(S§2).
- **재시작.** 기준 시각이 원장에 있으므로 재시작 뒤 첫 관측도 같은 창을 본다: 남은 유예만 이어지고 새로 생기지 않는다. 고점을 유예 중에
  옮기지 않으므로(G1) 재시작 전후로 맞출 상태가 없다.
- **벽시계 되돌림.** `FetchedAt < observed_at` 이면 유예 아님(보호선 상향 허용 쪽). 미래 시세는 이미 `quoteUsable` B1@1233 이 버린다.
  앞으로 튀면 유예가 일찍 끝날 뿐(더 보수적).
- **읽기 실패.** 편입 기록 읽기 · 시각 해석이 실패하면 유예 아님 + 경고 로그 한 줄. **조기 return 을 만들지 않는다** — 판정은 계속되고 손절은
  그대로 평가된다. 새 가드 없음(실패는 보호선 상향을 허용하는 쪽으로만 감).
- **비용.** 편입 기록은 set-once 라(journal/adoption.go:105-108) 포지션 ID → `observed_at` 를 관측기 메모리에 한 번만 캐시한다. 조건의 앞
  세 항(토글 · 편입 · 첫 rung 전)이 거짓이면 읽지 않는다 → 토글 OFF 는 원장 읽기 0. 브로커 호출 0, rate limit 예산 변화 0. 캐시는 관측
  고루틴 하나만 만진다(`Run` 루프) — 구현 단계에서 `make test-race` 대상 목록과 대조(tasks 2.6).
- **새 함수는 새 파일**(`internal/app/engine/a208_adoption_grace.go`)에 둔다 — 기존 함수 편집 표면을 `judgeLadder` 의 한 줄로 줄임.

### 4.3 왜 평가 입력인가 — 기각한 두 모양

| 모양 | 기각 사유 |
|---|---|
| 유예 중 판정을 건너뜀(손절선 위일 때만) | 「성공한 exit 관측은 … snapshot 을 최신화한다」(exit-policy SHALL) 위반, 그리고 「관측 두절은 보유 포지션 단위로도 판정한다」의 임계가 60초라 유예 끝 무렵 거짓 critical 알림을 낼 수 있음. 손절선 비교를 엔진에 한 번 더 두게 됨(판정 둘) |
| 유예 중 관측가를 `min(관측가, 고점)` 으로 넣음(평가기 무변경) | 판정은 같게 나오지만 저장 스냅숏의 `ObservedPrice` 가 실제 관측가가 아님 — 증거 위조. 콘솔 · 판정 기록이 그 값을 보임 |

평가 입력 필드는 판정기 하나가 결과를 정하고(손절 비교는 기존 B25 하나), 증거가 사실대로 남고, 복구가 재현한다. 대가는 롤백 순서(7).

### 4.4 손절 · 비상 청산이 늦어지지 않는다는 근거

1. **평가기 안.** 유예는 B12 · B13 루프만 막는다. 손절 갈래 B25@439 는 무편집이고, 같은 관측가와 **이전 기준선**을 비교한다. 편입 직후 이전
   기준선은 합성 손절가(편입 관측가 × (1 − pct))이며 이는 유예가 없을 때 그 관측에서 쓰였을 기준선과 같다 — 유예가 없었다면 그 관측에서 rung
   승격이 일어났을 경우에만 기준선이 더 높았을 것이고, 그 경우 관측가는 기준선 **위**(승격하려면 고점이 +2.5% 이상이어야 하고 잠금은 t0)라
   손절이 아니다. 단 「유예 없었다면 같은 관측에서 승격과 손절이 같이 났을」 경우는 없다: 승격은 관측가가 t0 × 1.025 이상, 손절은 관측가가
   잠금(t0) 미만이라 같은 관측에서 동시에 참일 수 없다(고점 흡수는 같은 관측가로 하므로). → 유예 중 손절 발의는 유예 없는 평가의 발의와
   action · level · ratio · 수량이 같다(RED T3 가 표로 단언).
2. **판정 앞뒤.** `judge` 의 조기 return(B1 · B2 · B3 · B5 · B6)과 `judgeLadder` 의 오류 return 은 그대로이고, 유예 판정은 새 return 을
   더하지 않는다(4.2). `record` · `submit` · a207 압축은 무편집.
3. **다른 경로.** 사다리 평가의 생산 호출자는 `judgeLadder` 와 복구 재검증 둘뿐이다(evidence-reconciliation). 청산 일괄(`internal/flatten`) ·
   브로커 보호(`internal/protection`) · 관측 두절 사다리(`checkOutage`)는 사다리 평가를 부르지 않으므로 유예의 영향 밖이다.
4. **재난 손절.** 이 저장소에 편입 전용 재난 손절은 아직 없다(사용자 항목 2, 다음 change). 이 change 는 그것을 만들지도 막지도 않는다.

### 4.5 TossOS 에서 유예가 실제로 하는 일(정직한 평가)

StockOS 유예의 원인 — 평균단가 기준선 + 묵은 브로커 스냅숏 시세로 편입 순간 이미 「이익」 — 은 TossOS 에 없다: TossOS 는 편입 관측가로
기준선 · 고점을 열어 R = 0 에서 시작한다(S§4). 남는 효과는 하나다: 편입 관측(최대 15초 묵음 허용)과 첫 60초 관측 사이에 가격이 +2.5% 이상
움직였을 때 첫 rung 승격(보호선을 t0 로 올림)을 막는다.

- 득: 묵은 편입 관측가 기준으로 바로 잠금을 걸었다가 작은 되돌림에 t0 에서 팔리는 「휩쏘」를 피한다.
- 실: 그 60초 동안 보호선은 합성 손절가(−pct)에 머문다 — 유예가 없을 때보다 **덜 보수적**이다. 유예가 끝난 뒤 G1 은 유예 중 고점을 잊으므로
  그 상승분에 대한 잠금은 다시 +2.5% 를 넘어야 생긴다.
- §0.9(손절 · 익절 변경은 더 보수적 방향만, 불명확하면 금지)와 긴장한다. 사용자 승인 사항이지만 **사람 결정 Q1** 로 다시 묻는다.
  대안: (a) 유예 없음(가장 보수적, 코드 최소), (b) G2 — rung 승격만 막고 고점은 흡수(유예 끝 첫 관측에서 밀린 승격이 바로 일어남; 휩쏘 회피
  효과는 사라짐).

### 4.6 D5 — 편입 알림

토글이 참이면 편입 알림 본문의 「보호선 승격, 중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게 적용된다」(adoption.go:412-413)를
「보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다. 편입 직후 60초 동안은 보호선을 올리지 않는다(손절은 그대로)」로 바꾼다. 필드 ·
key · 제목은 같다(`exit_policy_id` 필드가 RUNNER 를 보인다). 거짓이면 지금 문구 그대로.

## 5. 토글 OFF = 지금 동작

| 표면 | OFF 에서 | 근거 |
|---|---|---|
| 편입 기록 정책 ID | `d.opts.CommonPolicy` 그대로 | 3.1 갈래가 거짓 |
| 편입 알림 | 지금 문구 | 4.6 |
| 원장 읽기 | 유예용 읽기 0 | 4.2 조건 첫 항 |
| 평가 입력 | `ProfitSideHeld=false` → 평가 결과 동일 | 4.1 갈래가 거짓 |
| 저장 스냅숏 JSON | 새 키 없음(`omitempty`) | 4.1 |
| 입력 digest · id | 토큰 없음 → 같음 | 4.1 |
| 설정 파일 | 콘솔 저장이 새 키를 쓰지 않음(`omitempty`) | 3.2 |
| audit | 기동 기록에 `engine.adoption.floor_only_exit=false` 한 행(첫 기동 1 회) | 3.3 — 동작이 아닌 기록 |

RED T16 · T18 · T5 · T4 가 위 줄을 바이트로 단언한다.

## 6. a207(막힌 청산 재시도 압축)과의 상호작용

- a207 은 같은 (포지션, 세대, action, 수준)의 직전 무장이 해제됐을 때 다음 무장 행을 압축한다. 유예는 무장을 만들지 않는다(익절 발의 없음,
  승격은 상태 전용 `ActionLadderHoldStopPromoted` — `Orderable()` 거짓). 유예 중 손절 무장은 유예 없는 경우와 같으므로 a207 의 압축 판정 ·
  해소 알림 · 재시도 빈도가 그대로다.
- 유예 경계를 넘는 첫 관측은 입력 digest 가 바뀌어 `requiresFullJudgement` 가 한 번 전체 판정을 고를 수 있다(snapshot id 가 입력에 묶임).
  그 행은 무장이 아니므로 a207 압축 대상이 아니다. 구현 단계에서 T11 이 그 주기의 행 수를 잰다.

## 7. 스키마 · 롤백 · 순서

- 스키마 · journal migration 없음. 설정 키는 additive(없으면 거짓).
- **롤백 위험(C11).** 유예 중 저장된 스냅숏 JSON 에는 `"ProfitSideHeld":true` 가 있다. a208 이전 이미지는 저장 스냅숏을
  `DisallowUnknownFields` 로 읽으므로(exit_snapshot.go:253) 그 행을 손상으로 보고 격리한다 → 그 포지션은 **손절까지 판정되지 않는다**
  (exitloop.go:673-686 → `judge` B3@1022). 유예가 끝나면 다음 관측이 유예 없는 스냅숏으로 덮으므로(「성공한 exit 관측은 snapshot 을
  최신화한다」) 위험 창은 「마지막 편입 뒤 60초 + 한 관측 주기」다.
- **롤백 순서(사람 절차):** ① 설정에서 `floor_only_exit` 를 끄고 a208 이미지로 엔진을 재시작 ② 마지막 편입 뒤 60초 이상, 관측 주기 1 회 이상
  지남을 확인(원장 읽기 전용 질의: 열린 exit state 의 effective snapshot JSON 에 `ProfitSideHeld` 가 없음) ③ 그 뒤에만 옛 이미지로 교체.
  `docs/operations.md` 롤백 절에 이 순서를 적는다(tasks 4.2). 토글이 한 번도 켜진 적 없으면 제약 없음(held 행이 생길 수 없음).
- 앞으로 가는 배포: 옛 행에는 새 키가 없고 `omitempty` 거짓으로 읽혀 같은 판정 — 제약 없음.

## 8. 시험 계획(High-risk)

### 8.1 Branch Test Map — 편집 예정 함수(base 분기 ID 기준, 편집 뒤 재번호는 difflib 정렬로 — 기억 「분기 번호는 위치다」)

**`exitpolicy.EvaluateLadder`(32 분기)** — 이 change 가 만지는 갈래와 지키는 갈래:

| Branch | 시나리오 | 시험 |
|---|---|---|
| B12@355 + 새 held 조건 | 유예 중 관측가 > 고점 → 고점 불변 / 유예 아님 → 흡수 | T1 · 기존 ladder_test 워터마크 시험 |
| B13~B15 + 새 held 조건 | 유예 중 rung 목표 초과 → newIndex 불변 / 유예 아님 → 승격 | T1 · `TestExternalRunnerHasNoAutomaticPartials` |
| B16@386 | 유예 중 NoRung → 거짓 | T1 |
| B22@420 | 유예 중 거짓 | T1 |
| B25@439 | 유예 중 관측가 < 기준선 → `ActionLadderStop` 비율 1 | T2 · T3 |
| B26@441 | 유예 중 손절 미해소 → 억제(유예 없음과 같음) | T3 의 한 행 |
| B27@453 | 유예 중 손절 아님 → 무변화 return | T1 |
| B29 · B30 | floor-only RUNNER 에서 불가 | `TestExternalRunnerHasNoAutomaticPartials`(기존) · T10 |
| 그 밖 B1~B11 · B17~B21 · B23 · B24 · B28 · B31 · B32 | 무편집 — 기존 ladder_test · common_policy_test 통과 확인 | 선택 실행 |

**`exitpolicy.canonicalLadderInput`(3 분기)** — 새 갈래 1(held 토큰): T4(거짓 → base digest 고정값과 같음) · T6(참 → 다름, 재검증 통과).

**`engine.ExitObserver.judgeLadder`(8 분기)** — 새 입력 한 줄(분기 0 추가, 판정은 새 파일 함수): T11~T16. 기존 B1~B8 무편집.

**`engine.ReconcileDriver.adoptOne`(3 분기)** — 새 갈래 1(토글) + 알림 문구 갈래 1: T7 · T8 · T9. 기존 B1~B3 무편집.

**새 함수 `ExitObserver.inAdoptionGrace`(새 파일)** — 갈래: 캐시 적중 / 읽기 실패 / 해석 실패 / 음수 경과 / 경계 59.999s · 60.000s: T12 · T14 · T15.

설정 · 콘솔 · audit 함수(`mergeAdoption` · `LoadRawEngineAdoption` · `SaveEngineAdoption` · `handleSettingsSave` · `recordGateSettings` ·
`Context.ExitObserver`)는 분기 추가 없이 대입 한 줄씩 — T18~T22.

### 8.2 RED(구현 전에 실패해야 함)

| ID | 시험(파일) | 단언 |
|---|---|---|
| T1 | `TestA208HeldLadderNeitherAbsorbsNorPromotes`(internal/exitpolicy/a208_adoption_grace_test.go) | held 입력, 관측가 > t0×1.07 → HighWater · Baseline · ActiveRung 불변, Proposal 0 |
| T2 | `TestA208HeldLadderStillStopsOut` | held 입력, 관측가 < 합성 손절 → `ActionLadderStop`, Ratio 1 |
| T3 | `TestA208HeldStopEqualsUnheldStop` | 손절선 아래 가격 · 미해소 action 표 전수에서 held/unheld 의 Action · Level · Ratio · Suppressed · CancelPendingFirst 같음 |
| T4 | `TestA208UnheldInputDigestIsTheBaseDigest` | 고정 입력의 InputDigest · SnapshotID · DecisionID 가 base 에서 잰 상수와 같음(상수는 구현 **전** base 에서 먼저 씀 — 기억 「서명 상수는 산출물을 먼저 쓰고」) |
| T5 | `TestA208UnheldRecoveryJSONHasNoHeldKey` | `NewLadderRecoveryPolicy` JSON 에 `ProfitSideHeld` 없음 |
| T6 | `TestA208HeldSnapshotRecoversOnlyWithItsFlag` | held 스냅숏은 `ValidateRecoveryDerivation` 통과, 저장 입력에서 플래그를 지우면 실패 |
| T7 | `TestA208FloorOnlyAdoptionRecordsTheRunnerCopy`(internal/app/engine/a208_adoption_floor_only_test.go) | ON + BALANCED → 기록 정책 RUNNER, exit state 정체성 `1.0.0-adopted.1` |
| T8 | `TestA208FloorOnlyAdoptionReplacesRatchet` | ON + 공통 정책 없음 → LADDER RUNNER 사본 |
| T9 | `TestA208OffAdoptionIsUnchanged` | OFF → 기록 정책 = 공통 정책, 알림 본문 = base 문구 바이트 |
| T10 | `TestA208FloorOnlyAdoptedPositionNeverTakesProfit` | 관측기 구동, 가격이 모든 rung 을 지남 → 제출 0, 기준선만 상승 |
| T11 | `TestA208GraceHoldsTheRaiseButNotTheStop` | 유예 중 +3% → 승격 0; 이어 손절선 아래 → **같은 주기** 손절 제출 |
| T12 | `TestA208GraceEndsAtSixtySeconds` | 59.999s held, 60.000s 승격 |
| T13 | `TestA208GraceAnchorSurvivesRestart` | 새 관측기 인스턴스(캐시 없음) + 같은 원장 → 남은 유예만 |
| T14 | `TestA208AClockBeforeTheAdoptionIsNotGrace` | FetchedAt < observed_at → 승격 |
| T15 | `TestA208AnAnchorReadFailureIsNotGraceAndStillJudges` | 읽기 실패 주입 → 승격 허용, 손절 평가됨, 판정 return 없음 |
| T16 | `TestA208OffObserverNeverReadsTheAdoption` | OFF → `AdoptionOf` 호출 0, 저장 스냅숏 JSON 이 base 와 같음 |
| T17 | `TestA208TurningTheToggleOffKeepsTheAdoptedPolicy` | ON 편입 → OFF 재기동 → 같은 정체성으로 판정 |
| T18 | `TestA208FloorOnlyExitRoundTrips`(internal/config/a208_floor_only_exit_test.go) | 키 없음 → 거짓 · 저장 바이트에 키 없음, 참 → merge · LoadRaw · Save 모두 보존 |
| T19 | `TestA208ARefusedBlockClearsFloorOnly` | 손절폭 범위 밖 + 참 → 전면 0(거짓) |
| T20 | `TestA208ConsoleSaveKeepsFloorOnly`(internal/console) | 참인 파일 → 폼 저장 → 참 유지 |
| T21 | `TestA208FloorOnlyIsAuditedAtStartup`(internal/app/engine) | 기동 기록 끝 행이 `engine.adoption.floor_only_exit`, 같은 값 재기동은 행 0 |
| T22 | `TestA208TheExitObserverReceivesTheToggle` | `Context.ExitObserver` 가 설정 값을 옵션으로 넘김(생산 배선 경로로 구동 — 기억 「시험이 매번 조종석에 앉으면」) |

갱신: `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` — 「마지막」 → 「a208 행 바로 앞」(시험 함수 본문 편집 → gate ⑤ 번들).

### 8.3 선택 시험(전체 시험 · 전체 gate 는 사용자 명시 요청 시에만)

`go test ./internal/exitpolicy/... ./internal/config/... ./internal/journal/...`, `go test ./internal/console/ -run 'Settings|A208'`,
`go test ./internal/app/engine/ -run 'Adopt|A208|A207|A163|A074|Ladder|Exit'`, 해당 태그 시험이 있으면 `make test-seams` 의 같은 패키지 부분,
`make lint`. 이유: 편집 파일의 직접 호출자 · 같은 패키지(8.1). 미실행: 전 저장소 `make test` · `make gate` · 실계좌 검증.

### 8.4 변이 원장(High-risk — 비례 원칙상 의무)

사본 하네스(`analysis/a208_mutation.py`, 시작 sha 단언 · 무변이 대조군 GREEN 선확인 · 한 번에 한 판 — 기억 「뮤테이션 원복은 baseline」 ·
「뮤테이션은 대상에 닿아야」).

| M | 변이 | 잡아야 할 시험 |
|---|---|---|
| M1 | held 를 무시(흡수 조건에서 `!held` 제거) | T1 · T11 |
| M2 | held 를 무시(rung 루프 조건 제거) | T1 · T11 |
| M3 | held 가 손절도 막음(B25 를 `!held &&` 로) | T2 · T3 · T11 |
| M4 | 경계 `< 60s` → `<= 60s` | T12 |
| M5 | 음수 경과를 유예로 | T14 |
| M6 | 읽기 실패를 유예로 | T15 |
| M7 | 읽기 실패에서 조기 return | T15 |
| M8 | 토글 조건 제거(항상 유예 판정) | T16 |
| M9 | `adoptOne` 이 토글을 무시 | T7 · T8 |
| M10 | `adoptOne` 이 OFF 에서도 RUNNER | T9 |
| M11 | digest 토큰을 거짓에서도 붙임 | T4 |
| M12 | `omitempty` 제거 | T5 · T16 |
| M13 | 직렬화 제외(`json:"-"`) | T6 · T13 |
| M14 | `SaveEngineAdoption` 이월 누락 | T18 |
| M15 | `handleSettingsSave` 이월 누락 | T20 |
| M16 | 기동 audit 행 누락 | T21 |
| M17 | `Context.ExitObserver` 배선 누락 | T22 |
| M18 | 알림 문구를 토글과 무관하게 | T9 |

### 8.5 리뷰

proposal-freeze: Manager 가 정하는 보이스 구성(High-risk → 적대적 Eng 관점 포함, 작성자와 다른 계열). 구현 뒤 gstack review 1 회, 2 회차 상한
(기억 「SDD 작업 방식」). review.md 는 proposal-freeze 리뷰 때 Manager 가 만든다.

## 9. 편집 표면과 gate ⑤ 번들(구현 단계에서 편집 **전**에 만듦)

| 함수 | 파일 | 등급 |
|---|---|---|
| `EvaluateLadder` | internal/exitpolicy/ladder.go | High-risk |
| `canonicalLadderInput` | internal/exitpolicy/snapshot.go | High-risk |
| `ExitObserver.judgeLadder` | internal/app/engine/exitloop.go | High-risk |
| `ReconcileDriver.adoptOne` | internal/app/engine/adoption.go | High-risk |
| `Context.ExitObserver` | internal/app/engine/exitwiring.go | 배선(a166 census 밖 — evidence-reconciliation) |
| `recordGateSettings` | internal/app/engine/interlock.go | audit |
| `mergeAdoption` | internal/config/engine.go | 설정 |
| `Service.LoadRawEngineAdoption` · `Service.SaveEngineAdoption` | internal/config/adoption_io.go | 설정 |
| `Console.handleSettingsSave` | internal/console/settings.go | 콘솔(화면 무변경 — UI 스킬 · 브라우저 확인 대상 아님) |
| `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | internal/app/engine/a163_source_age_audit_test.go | 시험 |

새 코드: `internal/app/engine/a208_adoption_grace.go`(유예 판정 · 캐시), 새 시험 파일 4 개. 타입 필드 추가(`LadderInput` ·
`ExitObserverOptions` · `config.Adoption` · `rawAdoption`)는 함수 편집이 아니다.

## 10. 사람 결정(Open questions)

1. **Q1 — 유예를 그대로 넣을까?** floor-only 정책에서 유예가 막는 것은 보호선 상향뿐이라 그 60초는 덜 보수적이다(4.5). 선택지:
   (a) 승인대로 G1(StockOS 동형) · (b) 유예 없음 · (c) G2(승격만 미루고 고점은 흡수). 설계 기본값은 (a), 대신 사람 재확인 필수.
2. **Q2 — 표는 RUNNER 사본으로 충분한가?** 모든 편입이 RUNNER 잠금(+2.5% → t0, +4.5% → +2.0%, +7.0% → +3.5%)을 받는다. 공통 정책별
   편입 사본(BALANCED · HYBRID 잠금을 유지하고 익절만 뺌)을 원하면 D1(b) — 정체성 · digest 추가로 표면이 커진다.
3. **Q3 — 토글을 콘솔에서 켜고 끄게 할까?** 기본안은 설정 파일 편집 + 콘솔 저장 시 보존 + 기동 audit. 체크박스는 항목 3(콘솔 작업) 때 함께.
4. **Q4 — 이미 편입된 포지션은?** 지금 정책 그대로 남는다(정책 스냅숏 규칙). 바꾸려면 항목 3 의 해제 · 재편입이 필요하고, 이 change 의
   READOPT 는 토글을 따르지 않는다.
5. **Q5 — 롤백 순서(7) 수용?** 토글을 켠 적이 있으면 「끄고 60초 + 1 주기 뒤 이미지 교체」.
6. **Q6 — 운영 공통 정책 값 확인.** 이미 RUNNER 라면 이 change 의 체감 효과는 60초 유예뿐이다(부분익절은 이미 없음). 이 change 는 운영 설정을
   읽지 않았다.

## 11. 위험

- R1 롤백 순서(7)를 어기면 유예 중이던 편입 포지션이 손절까지 멈춘다 — 절차로만 막음(코드 가드 없음, 새 가드 금지 지시).
- R2 G1 은 유예 중 상승분을 잊는다(4.5) — 사람 결정 Q1.
- R3 `ActiveRung == NoRung` 조건 때문에 ON 전환 전 편입 · 이미 승격된 포지션은 유예 판정 자체를 건너뜀 — 의도(유예는 편입 직후 60초에만 의미).
- R4 READOPT 는 새 편입 기록을 만들지 않으므로(journal/apply_hook.go:709-740 은 exit_states 만 갱신) 재편입 뒤에는 유예가 없다 — 범위 밖, 항목 3.

## 12. 관련 / 다음

- **항목 2(다음 change) — 사용자 결정 기록:** 편입 전용 손절 설정. 유예 기본 **0분**, **재난 손절(disaster stop) 항상 켬**. 이 change 는 그 설정을
  만들지 않는다.
- 항목 3(이후): 종목별 정책 선택 · 편입 해제 · 즉시 청산 버튼 · 콘솔 토글(Q3).
- a207(막힌 청산 재시도 압축) — 6 절, 상호작용 없음.
