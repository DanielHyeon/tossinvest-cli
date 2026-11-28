# a208 design — 편입 전용 보호선 정책(공통 정책 보호선 그대로, 부분익절 없음, 유예 없음)

근거 표기: `Cn` = `analysis/ast-evidence/claims.md` 의 AST 근거 주장, `S§n` = `analysis/stockos-reference.md`. 분기 ID(`B12@355` 등)는 base
`baeddfe31` 의 `tools/logic-map` 산출물 기준이다.

**개정 1(2026-11-28, 사람 결정 반영).** 초판의 60초 보호선 상향 유예(D3 · D4, `LadderInput.ProfitSideHeld`)를 사람 결정 Q1(b)로 **삭제**했다.
그에 따라 사다리 평가기 · 입력 digest · 관측기 · 판정 경로 편집과 롤백 순서 제약이 모두 사라졌다. 결정 기록은 10 절. 초판의 C9 · C10 · C11 ·
C16 AST 근거는 「왜 유예를 뺐는가」의 근거로만 남는다(10 절 Q1).

**개정 2(2026-11-28, 사용자 결정 「F3 2」 — 리뷰 F3 처분).** 개정 1 의 D1(「모든 편입이 RUNNER 편입 사본」)은 운영 공통 정책 HYBRID_50 보다
보호선을 느슨하게 만든다(본전 잠금 +1.8% → +2.5%, 고점 −6.5% 러너 트레일 없음) — 보수 방향이 아니었다. 사용자는 선택지 2 를 골랐다:
**토글이 참이면 새 편입은 「설정된 공통 정책」의 보호선 표(단계 트리거 · 보호선 · 러너 트레일)를 그대로 쓰고 부분익절 비율만 0 인 사본으로 판정된다.**
D1 · 3.1 · 3.3 · 3.4 · 7 · 8 · 9 · 11 을 고쳤고 D6(사본 없는 공통 정책은 토글 무시)을 더했다. 새 AST 근거는 `analysis/function-logic/` 의
F3 번들 5 개(CommonLadderForPosition · LegacyLadderPolicyIdentity · AdoptionRequest.record · Journal.OpenExitState · mergeEngine).

## 1. 사실(F)

### 1.1 편입이 정해지는 자리와 정책이 정해지는 자리(오늘)

| 단계 | 자리 | 하는 일 |
|---|---|---|
| 편입 후보 판정 | `ReconcileDriver.adopt`(internal/app/engine/adoption.go:215) | 신선 조건 · 대사 · include/exclude 통과 후보를 `adoptOne` 으로 |
| 합성 손절 | `adoptOne` adoption.go:374 → `exitpolicy.SyntheticStop(observed, DefaultStopPct)` | 손절 = 편입 관측가 × (1 − pct), pct ∈ [0.02, 1) (exitpolicy/adoption.go:46-72) |
| **정책 선택** | `adoptOne` adoption.go:392 `ExitPolicyID: d.opts.CommonPolicy` | 엔진 공통 정책을 그대로 복사 — 편입 갈래 0(C1) |
| 영속 | `Journal.AdoptPosition`(journal/adoption.go:178) | 기록 + `positions.adoption_id` 한 트랜잭션, set-once. 정책 ID 는 등록된 공통 정책인지만 검사(journal/adoption.go:379-382) |
| exit state 종류 | `OpenAdoptedExitState` → `policyKindForID`(journal/adoption.go:468-473) | 빈 ID → RATCHET, 아니면 LADDER(C2) |
| 정체성 고정 | `seedPolicyIdentity`(journal/exit_state.go:206) → `LegacyLadderPolicyIdentity(id, adopted)` | RUNNER ∧ 편입만 `1.0.0-adopted.1`(C3) |
| 판정 시 표 | `ExitObserver.ladderFor`(exitloop.go:1314) → `CommonLadderForPosition(m.state.PolicyID, Adopted())` | RUNNER ∧ 편입만 부분익절 0(C4). 설정이 아니라 exit state 의 정책 ID 를 읽음 |
| 판정 | `judge` B7@1047 → `judgeLadder` / `judgeRatchet`(C5) → `EvaluateLadder`(C6) | 손절 B25@439, 익절 B28~B31 |

### 1.2 오늘 편입 포지션이 받는 것

| 공통 정책 | 편입 포지션 정책 | 부분익절 | 본전 · 보호선 상향 |
|---|---|---|---|
| 없음(RATCHET) | RATCHET | +1.0R 에 40%(C8, ratchet.go:438-448) | +0.8R 실질 본전(C8) 등 |
| BALANCED | BALANCED | 2.5% · 4.0% 에 25%, 6.0% 전량(common_policy.go:34-40) | rung 잠금 |
| HYBRID_50 | HYBRID_50 | 3.0% 25% · 4.8% 1/3(:61-67) | rung 잠금 + runner trail |
| RUNNER | **편입 RUNNER 사본** | 없음(C4 · C7) | rung 잠금 0/2.0/3.5% at +2.5/4.5/7.0% |

운영 공통 정책 값은 이 change 에서 읽지 않았다(Manager 전달: 운영은 `COMMON_LADDER_HYBRID_50`).

### 1.3 StockOS(S§1~§3)

- 부분익절 제거는 **RUNNER 편입에만**(`aee.py:6796-6807`, `external_runner_v1`). 표는 TossOS RUNNER 와 같다.
- 60초 settle window 가 있다(`:6978-7024`). 이 change 는 이식하지 않는다(Q1 → (b)).
- 손절 기준은 평균단가 × (1 − 1.2%)(`:1867-1871`) — 사용자가 가져오지 말라고 한 결함.

## 2. 결정 요약

| ID | 결정 | 대안(기각 사유) |
|---|---|---|
| D1 | (개정 2) 토글이 참이면 자동 편입은 **설정 공통 정책의 편입 보호선 사본 ID** 를 기록 — HYBRID_50 → 새 사본 `COMMON_LADDER_HYBRID_50_ADOPTED_FLOOR_ONLY`(등록부 밖, 고정 version/digest), RUNNER → 기존 편입 RUNNER 사본(이미 같은 의미). ~~기존 편입 RUNNER 사본 재사용~~ 은 F3 로 철회 | (b) 공통 정책마다 편입 사본 신설 — 사람 결정 Q2 로 기각. (c) RATCHET 유지 — 40% 부분익절이 남아 결정 위반 |
| D2 | 설정 키 `engine.adoption.floor_only_exit`, `omitempty`, 설정 파일로만 켬(Q3 확정) | exit_policy 블록 — 의미가 「편입」이고 편입 블록 seam 이 이미 콘솔 · audit 에 있음. 콘솔 체크박스 — Q3 로 이 change 밖 |
| ~~D3~~ | ~~60초 유예~~ — **삭제**(Q1 → (b)) | 10 절 Q1 |
| ~~D4~~ | ~~유예 기준 시각~~ — **삭제**(D3 와 함께) | — |
| D5 | 편입 알림 문구를 토글에 따라 둘로 | 문구 하나 — 참일 때 「중간 익절 … 동일하게 적용」이 거짓이 됨 |
| D6 | (개정 2) 공통 정책에 사본이 없으면(BALANCED · 없음) 토글 무시 = OFF 와 같은 기록 · 알림, 기동 audit 값 `ignored` + 사유 | BALANCED 사본(최종 rung 전량 익절이 남아 「보호선만」이 아님) · FinalTakeFull 까지 끈 사본(「부분익절만 0」 을 넘는 표 변경 — 사용자 결정 밖) · RATCHET 대체(사다리 표 없음) — 기각 |

## 3. D1 · D2 — 편입 전용 정책의 선택과 설정

### 3.1 선택(개정 2)

**정체성 해소 위치 — 새 ID 를 고정 version/digest 로.** 현재 코드에서 정책 정체성은 다음 자리에서 정해지고 대조된다(rg 전수 + AST 번들):

| 자리 | 하는 일 | F3 편집 |
|---|---|---|
| `journal.AdoptionRequest.record`(adoption.go:380) | 편입 기록 정책 ID 는 등록 공통 정책이어야 함 | 사본 ID 하나만 더 받음(조건 확장, 분기 수 무변) |
| `journal.Journal.OpenExitState`(exit_state.go:139) | LADDER 정책 ID 는 등록 공통 정책 · default_v1 | 사본 ID 하나만 더 받음(조건 확장) |
| `seedPolicyIdentity` → `exitpolicy.LegacyLadderPolicyIdentity(id, adopted)`(exit_state.go:215) | 새 exit state 의 version/digest 결정(편입 여부는 `positions.adoption_id`) | 사본 ID case 추가 → 고정 정체성, **편입 아니면 거부** |
| `exitpolicy.CommonLadderForPosition(id, adopted)` ← `ExitObserver.ladderFor`(exitloop.go:1324) | 판정 시 표 | 사본 ID 갈래를 맨 앞에 추가(기존 갈래 무편집), 편입 아니면 거부 |
| `managedPolicyIdentity`(exitloop.go:768) · `legacyPolicyIdentity`(exit_snapshot.go:718 · 645) | 저장 정체성이 비었을 때의 옛 행 해소 | 무편집 — 새 행은 OpenExitState 가 version/digest 를 항상 저장(v10 열) |
| `EvaluateLadderSnapshot` 의 상태 정체성 ↔ 표 정체성 대조 | 판정 직전 대조 | 무편집 — 사본 표의 `Identity()` 가 고정 digest 를 다시 계산 |
| `CommonPolicyByID` 를 쓰는 설정 검증(config/engine.go:81) · 엔진 옵션(exitloop.go:383 · reconcileloop.go:338) · override/READOPT(positionpolicy/model.go:164 · 179) · 콘솔 목록(`RegisteredCommonPolicies` — optimization.go:108 · position_policy.go:325, `CommonPolicyFieldDescriptor`) | 선택 가능한 정책 | 무편집 — 사본은 등록부 밖이라 **선택지로 노출되지 않음**(`TestA208FloorOnlyCopyIsNotSelectable` · `TestA208FloorOnlyCopyIsNotAConfigurableCommonPolicy`) |

`CommonLadderForPosition(id, adopted)` 의 기존 `adopted && RUNNER` 갈래를 넓히는 방식(예: `adopted && HYBRID_50` 도 부분익절 0)은 **쓰지 않는다** —
이미 `COMMON_LADDER_HYBRID_50` 으로 기록된 편입 행의 의미가 바뀐다. 대신 새 ID 를 편입 기록에 적고 그 ID 만 새 갈래로 해소한다. 같은 ID 에
새 version 을 붙이는 방식도 기각: 편입 기록(`position_adoptions`)은 ID 만 저장하므로 스키마 변경이 필요하다.

**사본 표.** `internal/exitpolicy/a208_adopted_floor_only.go` 의 `adoptedFloorOnlyLadder` 가 등록된 HYBRID_50 의 복사본에서 `PartialRatio` 만 "0" 으로
바꾸고 ID · version(`1.0.0-adopted.1`) · 고정 digest(`sha256:cb8039a3…98998c5`)를 붙인다. 표 값을 따로 적지 않으므로 트리거(1.8/3.0/4.8/6.5%) ·
보호선(0/1.2/2.5/3.8%) · 러너 트레일(6.5%) · `FinalTakeFull`(거짓)이 공통 정책과 갈릴 수 없고, 원본 표가 바뀌면 `Identity()` 가 digest 불일치로
판정을 멈춘다(fail-closed). 러너 트레일은 평가기에서 최종 rung 도달 시 붙으며 부분익절과 무관하다(ladder.go:392 — 무편집).

**공통 정책 → 사본(`AdoptedFloorOnlyPolicyID`).** HYBRID_50 → 새 사본, RUNNER → `COMMON_LADDER_RUNNER`(기존 편입 RUNNER 사본이 이미 「RUNNER 표 +
부분익절 0」 — RUNNER 는 트레일 · 전량 익절이 없어 의미가 정확히 같으므로 새 정체성을 만들지 않음), BALANCED · 없음 → 사본 없음(D6).

**엔진.** 갈래는 새 파일 `internal/app/engine/a208_adoption_floor_only.go` 의 함수에 둔다(`adoptOne` 새 분기 0):

```go
// 토글이 참이고 공통 정책에 사본이 있을 때만 사본 ID 와 참
func adoptionFloorOnlyPolicy(opts ReconcileDriverOptions) (string, bool) {
	if !opts.Adoption.FloorOnlyExit {
		return "", false
	}
	return exitpolicy.AdoptedFloorOnlyPolicyID(opts.CommonPolicy)
}
```

`adoptionExitPolicyID` 는 그 ID(없으면 공통 정책 그대로), `adoptionAlertBody(opts)` 는 그 불리언으로 문구를 고른다. `adoptOne` 은 알림 호출 인자
한 줄만 바뀐다. 그 뒤 경로: 편입 기록(사본 ID) → `OpenAdoptedExitState` → LADDER(C2) → `seedPolicyIdentity`(편입 → 고정 정체성) →
`ladderFor` → 사본 표 → `EvaluateLadder`(무편집; 부분익절 갈래는 비율 0 이라 불가, 최종 전량 갈래는 `FinalTakeFull` 거짓이라 불가).

**편입 포지션에만 적용되는 근거.** 이 자리는 `adoptOne` 안에 있고, `adoptOne` 의 유일한 호출자는 `ReconcileDriver.adopt` 다(CodeGraph
callers). 엔진 진입 포지션의 정책은 `openState` 의 비편입 갈래(exitloop.go:820-825)에서 정해지고 이 change 는 그 갈래를 고치지 않는다.
원장 쪽에서도 사본 정체성은 편입 포지션에만 나온다(`LegacyLadderPolicyIdentity` · `CommonLadderForPosition` 이 비편입을 거부 —
`TestA208FloorOnlyCopyNeedsAnAdoptedPosition`). 운영자 재편입(READOPT)은 `position_policy_command.go:359-366` 에서 따로 정하고 고치지 않는다.

**토글 뒤집기는 살아 있는 포지션을 바꾸지 않는다.** 정책은 편입 기록(`position_adoptions.exit_policy_id`)과 exit state 에 스냅숏되고 판정은
그 값을 읽는다(`ladderFor` 는 `m.state.PolicyID` — 설정이 아님). 이미 편입된 포지션은 지금 정책 그대로 남고, 켠 뒤의 새 편입만 토글을 따른다.

**손절은 그대로.** 합성 손절(adoption.go:374)과 손절 판정(B25@439)은 편집하지 않는다. 사본의 rung 잠금 · 러너 트레일은 공통 정책과 같은 값으로
손절선을 올리기만 한다 — 손절 기준선을 낮추는 갈래는 없다(C6). 유예가 없으므로 편입 직후 관측부터 rung 승격이 평소대로 일어난다.

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

`handleSettingsInclude` 는 이미 `current` 를 읽어 고친 뒤 저장하므로 `LoadRawEngineAdoption` · `SaveEngineAdoption` 이월만으로 값이 남는다.

### 3.3 audit(§0.5)

- 엔진 기동: `recordGateSettings` 표 **끝**에 `{audit.ActionAdoptionToggle, "engine.adoption.floor_only_exit", floorOnlyAuditValue(adoption), adoption.FloorOnlyIgnored}`.
  (개정 2) 값은 `true` · `false`, 토글이 참인데 사본이 없어 무시되면 `ignored`(detail = 사유). 사유는 로더가 `mergeEngine` 에서 편입 · 공통 정책
  두 블록을 병합한 뒤 `Adoption.floorOnlyIgnoredReason` 으로 파생하는 `json:"-"` 필드(`Adoption.FloorOnlyIgnored`) — 엔진과 같은
  `exitpolicy.AdoptedFloorOnlyPolicyID` 하나로 판정해 정의가 둘이 되지 않는다. `recordGateSettings` 시그니처 · `NewContext` 호출 줄은 무변경(a166 census).
  값을 `ignored` 로 따로 두는 이유: `RecordChange` 는 같은 값이면 쓰지 않으므로, 토글은 참으로 두고 공통 정책만 바꾼 기동이 묻히지 않게 함.
  `RecordChange` 가 직전 값과 같으면 쓰지 않으므로(a163 시험이 그 동작을 단언) 배포 뒤 첫 기동에 `false` 한 행, 이후 바뀔 때만 한 행.
  거부된 블록은 전면 0 이라 `false` 로 남고, 거부 이유는 지금처럼 `engine.adoption.enabled` 행의 detail 에 남는다.
- `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` 의 「마지막 행 = source_age_report」 단언은 「a208 행 바로 앞」으로 고친다.
  a197 은 같은 이유로 `NewContext` 에 따로 기록했지만, `NewContext` 본문은 a166 배선 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`)가
  고정하므로 이 change 는 census 밖인 `recordGateSettings` 표를 고친다(호출 줄 무변경 → census 무영향).
- 콘솔 저장 시점 기록(`recordAdoptionSave`)은 고치지 않는다: 콘솔은 이 값을 바꾸지 못하고(3.2 이월), 바꾸는 길은 설정 파일 편집뿐이며
  그 변경은 다음 기동 기록이 남긴다(한도 값의 수동 편집과 같은 trail).

### 3.4 D5 — 편입 알림

사본이 적용되면(토글 ∧ 공통 정책에 사본 있음) 편입 알림 본문의 「지금부터 보호선 승격, 중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게 적용된다.」(adoption.go:412-413)를
「지금부터 보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다(편입 전용 보호선 정책).」로 바꾼다. 나머지 문장 · 필드 · key · 제목은 같다
(`exit_policy_id` 필드가 사본 ID 를 보인다). 적용되지 않으면(토글 거짓 · 무시) 지금 문구 그대로(바이트 동일).

## 4. (삭제) 60초 유예

사람 결정 Q1 → (b)로 삭제. 초판 4 절의 설계(평가 입력 필드 · digest 토큰 · 관측기 판정 · 캐시)는 구현하지 않는다. 삭제 사유는 10 절 Q1.

## 5. 토글 OFF = 지금 동작

| 표면 | OFF 에서 | 근거 |
|---|---|---|
| 편입 기록 정책 ID | `d.opts.CommonPolicy` 그대로 | 3.1 갈래가 거짓 |
| 편입 알림 | 지금 문구(바이트 동일) | 3.4 |
| 사다리 평가 · 저장 스냅숏 · digest | 무편집 | 이 change 가 손대지 않음 |
| 설정 파일 | 콘솔 저장이 새 키를 쓰지 않음(`omitempty`) | 3.2 |
| audit | 기동 기록에 `engine.adoption.floor_only_exit=false` 한 행(첫 기동 1 회) | 3.3 — 동작이 아닌 기록 |

RED T3 · T7 이 위 줄을 바이트로 단언한다.

## 6. a207(막힌 청산 재시도 압축)과의 상호작용

없음. a207 은 무장 행 압축이고, 이 change 는 편입 기록의 정책 ID 만 고른다. 편입 보호선 사본의 판정은 이미 있는 판정 경로(공통 정책이 RUNNER 인
엔진의 편입)와 같다.

## 7. 스키마 · 롤백 · 순서(개정 2 — 제약 생김)

- 스키마 · journal migration · 저장 스냅숏 필드 없음. 설정 키는 additive(없으면 거짓). `Adoption.FloorOnlyIgnored` 는 `json:"-"`(파일 무관).
- **롤백(a208 → 옛 이미지)에 순서 제약이 있다(개정 1 의 「제약 없음」 철회).** 토글이 켜져 있던 동안 HYBRID_50 공통 정책으로 편입된 포지션의
  편입 기록 · exit state 정책 ID 는 `COMMON_LADDER_HYBRID_50_ADOPTED_FLOOR_ONLY` 다. 이 ID 를 모르는 옛 이미지는 `CommonLadderForPosition` 이
  「unknown common policy」를 돌려주므로 `judgeLadder` 가 `alertRefused` 로 **그 포지션의 판정 전체(손절 포함)를 건너뛴다**
  (exitloop.go:1130-1133 — 손절 판정은 같은 `EvaluateLadder` 안에 있음). exit state 가 아직 안 열린 편입은 옛 이미지의 `OpenExitState` 가 거부한다.
  즉 그 포지션은 옛 이미지에서 엔진 손절이 없다(거부 알림은 남음). 새 가드는 만들지 않고 절차로만 다룬다(`docs/operations.md`):
  1. 롤백 전에 이 정책 ID 로 열린 exit state 수를 엔진 journal 을 읽기 전용으로 직접 조회해 확인한다(`exit_states.policy_id`, completed=0).
  2. 0 이면 제약 없음. 0 이 아니면 그 이미지로 롤백하지 않는다(또는 사람이 그 포지션을 직접 관리할 것을 받아들인 뒤에만).
  3. a208 이 운영에 오른 뒤에는 롤백 핀을 a208 을 포함한 이미지로 올리는 것이 이 제약을 없애는 정석이다.
  공통 정책이 RUNNER 인 동안의 편입은 기존 `COMMON_LADDER_RUNNER` 라 옛 이미지도 같은 표 · 정체성으로 판정한다(`a633ecff3`, 2026-08-01 부터).
- 옛 이미지의 설정 읽기는 `json.Unmarshal`(엄격 아님)이라 `floor_only_exit` 키는 무시된다(= OFF). 옛 이미지의 콘솔 저장은 이 키를 지운다 —
  다시 a208 이미지로 올린 뒤 설정 파일의 키와 기동 audit 행을 확인한다(R6).
- 앞으로 가는 배포: 옛 설정에는 키가 없고 거짓으로 읽혀 같은 동작 — 제약 없음.

## 8. 시험 계획(High-risk)

### 8.1 Branch Test Map — 편집 예정 함수(base 분기 ID 기준)

**`engine.ReconcileDriver.adoptOne`(3 분기)** — 새 갈래 2(정책 선택 · 알림 문구, 둘 다 토글): T1 · T2 · T3 · T4. 기존 B1~B3(합성 손절 실패 ·
편입 거절 · exit state 열기 실패) 무편집 — 기존 reconcileloop_test 편입 시험군 통과 확인.

설정 · 콘솔 · audit 함수(`mergeAdoption` · `LoadRawEngineAdoption` · `SaveEngineAdoption` · `handleSettingsSave` · `recordGateSettings`)는 분기 추가
없이 대입 · 표 행 한 줄씩 — T7~T10. `TestA163…` 은 단언 대상 행 위치만 바뀜.

### 8.2 RED(구현 전에 실패해야 함)

| ID | 시험(파일) | 단언 |
|---|---|---|
| T1 | `TestA208FloorOnlyAdoptionRecordsTheHybridCopy`(internal/app/engine/a208_adoption_floor_only_test.go) | (개정 2) ON + HYBRID_50 → 기록 · exit state 정책 = 사본 ID, 고정 정체성, 사본 표의 트리거 · 보호선 · 트레일 · 전량 여부 = 공통 HYBRID_50, 부분익절 전부 0, 합성 손절 = 관측가 × (1 − pct) |
| T1b | `TestA208OnWithRunnerKeepsTheExistingAdoptedRunnerCopy` | (개정 2) ON + RUNNER → 기록 RUNNER, 정체성 `1.0.0-adopted.1`, 알림 = 사본 문구 |
| T2 | `TestA208FloorOnlyIsIgnoredWithoutAFloorCopy` | (개정 2) ON + BALANCED · ON + 없음 → 기록 = 공통 정책(없음이면 RATCHET), 알림 = base 바이트 |
| T3 | `TestA208OffAdoptionIsUnchanged` | OFF + BALANCED · OFF + HYBRID_50 → 기록 정책 = 공통 정책(사본 ID 아님), 알림 본문 = base 문구 바이트 |
| T4 | `TestA208FloorOnlyAlertSaysNoPartialTakeProfit` | ON + HYBRID_50 → 알림 본문이 「중간 익절은 하지 않는다」를 말하고 「중간 익절, 손절이 … 동일하게」를 말하지 않음, `exit_policy_id` 필드 = 사본 ID |
| T5 | `TestA208FloorOnlyAdoptedPositionNeverTakesProfit` | (개정 2) 사본 ID 편입 기록의 포지션을 관측기로 구동, +1.9/3.1/4.9/6.6/15% → 제출 0, 기준선 10000 · 10120 · 10250 · 10380 · 10752.5(러너 트레일), 첫 rung 위 첫 관측에서 바로 승격(유예 없음) |
| T6 | `TestA208TurningTheToggleOffKeepsTheAdoptedPolicy` | ON + HYBRID_50 으로 편입 → OFF 드라이버로 다음 편입 → 앞 포지션의 exit state 는 사본 정체성 그대로, 새 편입은 HYBRID_50 |
| T11 | `TestA208FloorOnlyCopyKeepsTheHybridFloors`(internal/exitpolicy) | (개정 2) 사본 표 = HYBRID_50 표(부분익절만 0), 정체성 고정, 시드 정체성 = 표 정체성 |
| T12 | `TestA208FloorOnlyCopyIsAdoptionOnly` · `TestA208FloorOnlyCopyNeedsAnAdoptedPosition`(journal) | (개정 2) 비편입이면 표 · 정체성 · exit state 개시 모두 거부 |
| T13 | `TestA208AdoptedHybridRowsKeepTheirMeaning` | (개정 2) 편입 HYBRID_50 행 = 등록 HYBRID_50 표(부분익절 포함) · 1.0.0 정체성 — 기존 갈래 무변 고정 |
| T14 | `TestA208FloorOnlyCopyIsNotSelectable` · `TestA208FloorOnlyCopyIsNotAConfigurableCommonPolicy`(config) | (개정 2) 사본은 등록부 · 필드 선택지 밖, 설정 공통 정책으로 쓰면 거부 |
| T15 | `TestA208JournalRefusesOtherUnregisteredPolicies`(journal) | (개정 2) 원장은 사본 ID 만 더 받고 다른 미등록 ID 는 거부 |
| T16 | `TestA208FloorOnlyIgnoredReasonFollowsTheCommonPolicy`(config) · `TestA208IgnoredFloorOnlyIsAuditedAtStartup`(engine) | (개정 2) 무시 사유는 BALANCED · 없음일 때만, 토글 값은 그대로; audit 값 true → ignored + 사유 |
| T7 | `TestA208FloorOnlyExitRoundTrips`(internal/config/a208_floor_only_exit_test.go) | 키 없음 → 거짓 · 저장 바이트에 키 없음, 참 → merge · LoadRaw · Save 모두 보존 |
| T8 | `TestA208ARefusedBlockClearsFloorOnly` | 손절폭 범위 밖 + 참 → 전면 0(거짓) |
| T9 | `TestA208ConsoleSaveKeepsFloorOnly`(internal/console) | 참인 블록 → 폼 저장 → 저장된 블록 참 유지 |
| T10 | `TestA208FloorOnlyIsAuditedAtStartup`(internal/app/engine) | 기동 기록 끝 행이 `engine.adoption.floor_only_exit`(바로 앞은 a163 행), 같은 값 재기동은 행 0, 값이 바뀌면 한 행 |

개정 1 RED: T5 · T6 은 base 에서도 통과 가능(당시 RUNNER 경로). 개정 2 RED(`analysis/tdd/red-f3.txt`): FAIL 10 — T1 · T2 · T4 · T5 · T6 · T11 · T12(journal) ·
T15 · T16 둘; PASS 기대 6 — T1b · T3 · T12(exitpolicy, 미지 ID 라 이미 거부) · T13 · T14 둘(가드).
갱신: `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` — 「마지막」 → 「a208 행 바로 앞」(시험 함수 본문 편집 → gate ⑤ 번들).

### 8.3 선택 시험(전체 시험 · 전체 gate 는 사용자 명시 요청 시에만)

`go test ./internal/config/... ./internal/exitpolicy/... ./internal/journal/... ./internal/positionpolicy/...`(개정 2 — journal · positionpolicy 추가), `go test ./internal/console/`, `go test ./internal/app/engine/ -run 'Adopt|A208|A163|A074|A197|Reconcile|Include|Driver'`,
`go test ./internal/accountbudgetruntime/`, `make lint`, `check_analysis.py`. 이유: 편집 파일의 직접 호출자 · 같은 패키지(8.1) + a166 census.
미실행: 전 저장소 `make test` · `make gate` · 실계좌 검증.

### 8.4 변이 원장(High-risk — 결정 줄)

| M | 변이 | 잡아야 할 시험 |
|---|---|---|
| M1 | `adoptOne` 이 토글을 무시(항상 공통 정책) | T1 · T2 |
| M2 | (개정 1) `adoptOne` 이 OFF 에서도 사본 ID | T3 |
| M3 | 알림 문구를 토글과 무관하게 ON 문구 | T3 |
| M4 | 알림 문구를 토글과 무관하게 OFF 문구 | T4 |
| M5 | `mergeAdoption` 이월 누락 | T7 |
| M6 | `LoadRawEngineAdoption` 이월 누락 | T7 · T9 |
| M7 | `SaveEngineAdoption` 이월 누락 | T7 |
| M8 | `omitempty` 제거 | T7 |
| M9 | 거부 블록이 토글을 남김 | T8 |
| M10 | `handleSettingsSave` 이월 누락 | T9 |
| M11 | 기동 audit 행 누락 | T10 |
| M12 | 기동 audit 행 값을 상수 `false` 로 | T10 |
| M13 | (개정 2) 사본의 `PartialRatio` 를 원본 값으로 되돌림(0 대입 삭제) | T1 · T5 · T11 |
| M14 | (개정 2) OFF 갈래가 사본 ID 를 적음(토글 검사 삭제) | T3 |
| M15 | (개정 2) 기존 `adopted && RUNNER` 갈래를 HYBRID_50 까지 넓힘(편입 HYBRID_50 행이 부분익절 0 표로 판정) | T13 |
| M16 | (개정 2) 사본 갈래의 편입 검사 삭제(비편입도 사본 해소) | T12 |
| M17 | (개정 2) 원장이 사본 ID 를 받지 않음(record 조건 원복) | T1 · T15 |
| M18 | (개정 2) BALANCED 도 사본 있음으로 대응(무시 갈래 삭제) | T2 · T16 |
| M19 | (개정 2) 사본이 러너 트레일을 버림 | T1 · T5 · T11 |
| M20 | (개정 2) audit 값이 무시돼도 FormatBool | T16 |

### 8.5 리뷰

구현 뒤 gstack review(적대적 패스 포함, 가능하면 작성자와 다른 계열), 2 회차 상한, 리뷰 발견에서 새 가드를 만들지 않음.

## 9. 편집 표면과 gate ⑤ 번들(편집 **전**에 만듦)

| 함수 | 파일 | 등급 |
|---|---|---|
| `ReconcileDriver.adoptOne` | internal/app/engine/adoption.go | High-risk |
| `recordGateSettings` | internal/app/engine/interlock.go | audit |
| `mergeAdoption` | internal/config/engine.go | 설정 |
| `Service.LoadRawEngineAdoption` · `Service.SaveEngineAdoption` | internal/config/adoption_io.go | 설정 |
| `Console.handleSettingsSave` | internal/console/settings.go | 콘솔(화면 무변경 — UI 스킬 · 브라우저 확인 대상 아님) |
| (개정 2) `CommonLadderForPosition` | internal/exitpolicy/common_policy.go | High-risk |
| (개정 2) `LegacyLadderPolicyIdentity` | internal/exitpolicy/identity.go | High-risk |
| (개정 2) `AdoptionRequest.record` | internal/journal/adoption.go | High-risk |
| (개정 2) `Journal.OpenExitState` | internal/journal/exit_state.go | High-risk |
| (개정 2) `mergeEngine` | internal/config/engine.go | 설정 |
| `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | internal/app/engine/a163_source_age_audit_test.go | 시험 |

새 시험 파일 3 개(+ console 시험 1 개는 새 파일). 타입 필드 추가(`config.Adoption` · `rawAdoption`)는 함수 편집이 아니다.
초판 편집 예정이던 `EvaluateLadder` · `canonicalLadderInput` · `judgeLadder` · `Context.ExitObserver` 는 편집하지 않는다.

개정 2 새 파일: `internal/exitpolicy/a208_adopted_floor_only.go`(사본 ID · 정체성 · 표 · 대응), `internal/config/a208_floor_only_ignored.go`(무시 사유).
번들은 F3 편집 전(커밋 233897096)에 만들고 편집 뒤 `revision: current` 로 재추출함.

## 10. 사람 결정(2026-11-28, Manager 전달)

1. **Q1 — 유예 → (b) 넣지 않음.** 사유(Manager 기록): floor-only 정책에서 유예가 막는 것은 보호선 상향뿐이라 그 60초는 유예가 없을 때보다
   덜 보수적이다 — §0.9(손절 · 익절 변경은 더 보수적 방향만)와 충돌. 또 새 스냅숏 필드는 롤백 위험: `internal/journal/exit_snapshot.go`
   `decodeStoredSnapshot` 이 `DisallowUnknownFields()` 를 쓰므로(C11) 옛 이미지는 그 행을 손상으로 읽어 그 포지션의 판정(손절 포함)을 멈춘다.
   → `LadderInput.ProfitSideHeld` · 새 스냅숏 필드 없음.
2. **Q2 — 예, D1(a) → 개정 2 로 대체(사용자 「F3 2」, 2026-11-28).** ~~모든 편입이 기존 RUNNER 사본을 쓴다~~ — 이제 설정 공통 정책의 보호선 표 + 부분익절 0. 옛 기록:(부분익절 없음; 잠금 +2.5% → t0, +4.5% → +2.0%, +7.0% → +3.5%). 공통 정책별 편입 사본 없음.
3. **Q3 — 예, 기본안.** 토글은 설정 파일로 켜고(기본 false), `handleSettingsSave` 가 값을 이월해 콘솔 저장이 조용히 끄지 않으며, 기동 audit 을 남긴다.
   이 change 에 콘솔 체크박스 없음. 이미 편입된 포지션은 정책을 유지하고 새 편입만 토글을 따르며, READOPT 는 토글을 따르지 않는다.
4. Q4 — Q3 에 포함(이미 편입된 포지션은 그대로).
5. Q5 — 롤백 순서: 유예 삭제로 제약이 없어졌다(7 절 재유도). 남는 것은 롤백 뒤 재배포 때 설정 키 확인 절차 하나.
6. Q6 — 운영 공통 정책은 HYBRID_50(Manager 전달). 이 change 는 운영 설정을 읽지 않았다.

## 11. 위험

- R3 READOPT 는 새 편입 기록을 만들지 않고 토글도 따르지 않는다(journal/apply_hook.go:709-740 은 exit_states 만 갱신) — 범위 밖, 항목 3.
- R5 (개정 2로 해소) ~~토글을 켜면 HYBRID_50 이어도 RUNNER 잠금~~ — 사본은 공통 정책의 잠금 · 트레일을 그대로 쓴다.
- R6 롤백 뒤 옛 이미지의 콘솔 저장이 키를 지운다(7 절) — 재배포 절차로만 확인(새 가드 없음).
- R7 (리뷰 1회차 F3 → 개정 2 로 해소) 켜면 본전 잠금이 공통 정책보다 늦던 문제 — 사본 표 = 공통 정책 표(T1 · T11 이 값 대조, M13 · M19 변이).
- R8 (리뷰 1회차, codex P2 · native) 콘솔 저장은 잠금 밖에서 읽은 값을 이월한다 — 콘솔 화면을 연 뒤 저장 사이에 사람이 파일의 토글을
  바꾸면 저장이 옛 값으로 덮는다. include/exclude 도 같은 기존 구조(adoption_io.go 머리 주석 · limits_io.go 의 lost-update 서술). 새 가드
  금지 지시에 따라 고치지 않음 — 토글을 바꾼 뒤에는 기동 audit 행으로 확인(operations.md).
- R9 include_symbols 지정 편입도 `adoptOne` 을 지나므로 토글을 따른다 — 「새 자동 편입」의 의도된 범위(operations.md 에 명시).
- R10 (개정 2) 롤백 순서 제약 — 사본 ID 를 모르는 옛 이미지는 그 포지션의 판정(손절 포함)을 거부한다(7 절). 절차로만 다룸(새 가드 없음).
- R11 (개정 2, D6 fail-closed 가 거부하는 정상 입력) 토글을 켰는데 아래 조합이면 편입은 토글이 꺼진 것과 같다:
  ① 공통 정책 BALANCED(부분익절 25% · 25% + 6.0% 전량 익절이 그대로 남음), ② 공통 정책 없음(RATCHET, +1.0R 40% 부분익절이 남음),
  ③ 공통 정책 설정이 거부된 경우(미등록 ID — 이때는 엔진이 이미 기동을 거부하므로 편입 자체가 없음). ①②는 기동 audit 값 `ignored` 와 사유로만 보인다
  (콘솔 표시 없음 — 사람 결정 Q3). 이 둘을 지원하려면 사용자 결정이 따로 필요하다(①은 FinalTakeFull 처리, ②는 RATCHET 의 사다리 대체).
- R12 (개정 2) 콘솔 「포지션 정책」 화면은 사본 ID 를 선택지로 보이지 않지만, 사본 포지션의 override · inherit 동작은 a208 이전 경로 그대로다
  (override 는 등록 공통 정책으로만 바꿈) — 사본으로 되돌리는 조작은 없다.
- (초판 R1 · R2 는 유예 삭제로 해당 없음.)

## 12. 관련 / 다음

- **항목 2(다음 change) — 사용자 결정 기록:** 편입 전용 손절 설정. 유예 기본 **0분**, **재난 손절(disaster stop) 항상 켬**. 이 change 는 그 설정을
  만들지 않는다.
- 항목 3(이후): 종목별 정책 선택 · 편입 해제 · 즉시 청산 버튼 · 콘솔 토글.
