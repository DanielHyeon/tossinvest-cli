# a208 design — 편입 전용 보호선 정책(부분익절 없음, 유예 없음)

근거 표기: `Cn` = `analysis/ast-evidence/claims.md` 의 AST 근거 주장, `S§n` = `analysis/stockos-reference.md`. 분기 ID(`B12@355` 등)는 base
`baeddfe31` 의 `tools/logic-map` 산출물 기준이다.

**개정 1(2026-11-28, 사람 결정 반영).** 초판의 60초 보호선 상향 유예(D3 · D4, `LadderInput.ProfitSideHeld`)를 사람 결정 Q1(b)로 **삭제**했다.
그에 따라 사다리 평가기 · 입력 digest · 관측기 · 판정 경로 편집과 롤백 순서 제약이 모두 사라졌다. 결정 기록은 10 절. 초판의 C9 · C10 · C11 ·
C16 AST 근거는 「왜 유예를 뺐는가」의 근거로만 남는다(10 절 Q1).

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
| D1 | 토글이 참이면 자동 편입은 `COMMON_LADDER_RUNNER` 를 기록 → 기존 편입 RUNNER 사본 재사용(Q2 확정) | (b) 공통 정책마다 편입 사본 신설 — 사람 결정 Q2 로 기각. (c) RATCHET 유지 — 40% 부분익절이 남아 결정 위반 |
| D2 | 설정 키 `engine.adoption.floor_only_exit`, `omitempty`, 설정 파일로만 켬(Q3 확정) | exit_policy 블록 — 의미가 「편입」이고 편입 블록 seam 이 이미 콘솔 · audit 에 있음. 콘솔 체크박스 — Q3 로 이 change 밖 |
| ~~D3~~ | ~~60초 유예~~ — **삭제**(Q1 → (b)) | 10 절 Q1 |
| ~~D4~~ | ~~유예 기준 시각~~ — **삭제**(D3 와 함께) | — |
| D5 | 편입 알림 문구를 토글에 따라 둘로 | 문구 하나 — 참일 때 「중간 익절 … 동일하게 적용」이 거짓이 됨 |

## 3. D1 · D2 — 편입 전용 정책의 선택과 설정

### 3.1 선택

`adoptOne` 의 정책 ID 한 자리(adoption.go:392)와 알림 본문 한 자리만 바꾼다. 갈래는 새 파일
`internal/app/engine/a208_adoption_floor_only.go` 의 함수 둘에 두어 `adoptOne` 에 새 분기를 만들지 않는다(분기 재번호 없음):

```go
// 편입 전용 보호선 정책이 켜져 있으면 공통 정책과 무관하게 RUNNER 편입 사본을 기록함(a208 D1)
func adoptionExitPolicyID(opts ReconcileDriverOptions) string {
	if opts.Adoption.FloorOnlyExit {
		return exitpolicy.CommonLadderRunner
	}
	return opts.CommonPolicy
}
```

그 뒤 경로는 전부 기존 코드다: C2(LADDER) → C3(`1.0.0-adopted.1`) → C4(부분익절 0) → C7(익절 갈래 B29 · B30 불가). 편입 드라이버는
설정 블록을 이미 통째로 받으므로(`Context.ReconcileDriver` reconcileloop.go:383, C15) 배선 변경이 없다.

**편입 포지션에만 적용되는 근거.** 이 자리는 `adoptOne` 안에 있고, `adoptOne` 의 유일한 호출자는 `ReconcileDriver.adopt` 다(CodeGraph
callers). 엔진 진입 포지션의 정책은 `openState` 의 비편입 갈래(exitloop.go:820-825)에서 정해지고 이 change 는 그 갈래를 고치지 않는다.
운영자 재편입(READOPT)은 `position_policy_command.go:359-366` 에서 따로 정하고 역시 고치지 않는다(Q3 — READOPT 는 토글을 따르지 않음).

**토글 뒤집기는 살아 있는 포지션을 바꾸지 않는다.** 정책은 편입 기록(`position_adoptions.exit_policy_id`)과 exit state 에 스냅숏되고 판정은
그 값을 읽는다(`ladderFor` 는 `m.state.PolicyID` — 설정이 아님). 기존 SHALL 「전역 정책 변경은 활성 포지션을 재해석하지 않는다」와 같은 구조.
이미 편입된 포지션은 지금 정책 그대로 남고, 켠 뒤의 새 편입만 토글을 따른다(Q3).

**손절은 그대로.** 합성 손절(adoption.go:374)과 손절 판정(B25@439)은 편집하지 않는다. RUNNER 사본의 첫 rung 잠금(+2.5% → t0)은 손절선을
올리기만 한다 — 손절 기준선을 낮추는 갈래는 없다(C6). 유예가 없으므로 편입 직후 관측부터 rung 승격이 평소대로 일어난다.

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

- 엔진 기동: `recordGateSettings` 표 **끝**에 `{audit.ActionAdoptionToggle, "engine.adoption.floor_only_exit", strconv.FormatBool(adoption.FloorOnlyExit), ""}`.
  `RecordChange` 가 직전 값과 같으면 쓰지 않으므로(a163 시험이 그 동작을 단언) 배포 뒤 첫 기동에 `false` 한 행, 이후 바뀔 때만 한 행.
  거부된 블록은 전면 0 이라 `false` 로 남고, 거부 이유는 지금처럼 `engine.adoption.enabled` 행의 detail 에 남는다.
- `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` 의 「마지막 행 = source_age_report」 단언은 「a208 행 바로 앞」으로 고친다.
  a197 은 같은 이유로 `NewContext` 에 따로 기록했지만, `NewContext` 본문은 a166 배선 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`)가
  고정하므로 이 change 는 census 밖인 `recordGateSettings` 표를 고친다(호출 줄 무변경 → census 무영향).
- 콘솔 저장 시점 기록(`recordAdoptionSave`)은 고치지 않는다: 콘솔은 이 값을 바꾸지 못하고(3.2 이월), 바꾸는 길은 설정 파일 편집뿐이며
  그 변경은 다음 기동 기록이 남긴다(한도 값의 수동 편집과 같은 trail).

### 3.4 D5 — 편입 알림

토글이 참이면 편입 알림 본문의 「지금부터 보호선 승격, 중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게 적용된다.」(adoption.go:412-413)를
「지금부터 보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다(편입 전용 보호선 정책).」로 바꾼다. 나머지 문장 · 필드 · key · 제목은 같다
(`exit_policy_id` 필드가 RUNNER 를 보인다). 거짓이면 지금 문구 그대로(바이트 동일).

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

없음. a207 은 무장 행 압축이고, 이 change 는 편입 기록의 정책 ID 만 고른다. 편입 RUNNER 사본의 판정은 이미 있는 경로(공통 정책이 RUNNER 인
엔진의 편입)와 같다.

## 7. 스키마 · 롤백 · 순서(재유도)

- 스키마 · journal migration · 저장 스냅숏 필드 없음. 설정 키는 additive(없으면 거짓).
- **롤백(a208 → 옛 이미지)에 순서 제약 없음.** 근거 셋:
  1. 옛 이미지의 설정 읽기는 `json.Unmarshal`(internal/config/service.go:250 · adoption_io.go:61)이고 `DisallowUnknownFields` 를 쓰지 않는다
     (`rg DisallowUnknownFields internal/config` 0 건) → `floor_only_exit` 키는 무시되고 편입은 공통 정책을 따른다(= OFF).
  2. 토글이 켜져 있던 동안 편입된 포지션의 기록 · exit state 정책 ID 는 `COMMON_LADDER_RUNNER` 다. 편입 RUNNER 사본(`CommonLadderForPosition` 의
     `adopted && RUNNER` 갈래, 정체성 `1.0.0-adopted.1`)은 `a633ecff3`(2026-08-01)부터 있어 현재 롤백 핀 이미지에도 있다 → 옛 이미지도 같은
     정체성 · 같은 표로 판정한다. journal 은 정책 ID 가 등록된 공통 정책인지만 본다(journal/adoption.go:379-382).
  3. 저장 스냅숏에 새 필드가 없으므로 옛 이미지의 `decodeStoredSnapshot`(`DisallowUnknownFields`)이 손상으로 읽을 행이 생기지 않는다(초판의
     롤백 위험 C11 은 유예와 함께 사라짐).
- **롤백 뒤 다시 올릴 때 한 가지(절차, 코드 가드 없음):** 옛 이미지의 콘솔 저장은 편입 블록을 자기가 아는 필드로 다시 쓰므로 `floor_only_exit`
  키를 지운다. 다시 a208 이미지로 올린 뒤 설정 파일의 키를 확인하고, 기동 audit 의 `engine.adoption.floor_only_exit` 행(값이 바뀌었으면
  true→false 한 행)으로 확인한다. `docs/operations.md` 에 적는다(tasks 4.2).
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
| T1 | `TestA208FloorOnlyAdoptionRecordsTheRunnerCopy`(internal/app/engine/a208_adoption_floor_only_test.go) | ON + BALANCED → 기록 · exit state 정책 RUNNER, 정체성 `1.0.0-adopted.1`, 그 표의 부분익절 비율 전부 0, 합성 손절 = 관측가 × (1 − pct) |
| T2 | `TestA208FloorOnlyAdoptionReplacesRatchet` | ON + 공통 정책 없음 → LADDER RUNNER 사본(RATCHET 아님) |
| T3 | `TestA208OffAdoptionIsUnchanged` | OFF + BALANCED → 기록 정책 = BALANCED, 알림 본문 = base 문구 바이트 |
| T4 | `TestA208FloorOnlyAlertSaysNoPartialTakeProfit` | ON → 알림 본문이 「중간 익절은 하지 않는다」를 말하고 「중간 익절, 손절이 … 동일하게」를 말하지 않음, `exit_policy_id` 필드 = RUNNER |
| T5 | `TestA208FloorOnlyAdoptedPositionNeverTakesProfit` | RUNNER 편입 기록(T1 이 쓰는 값)의 포지션을 관측기로 구동, 가격이 +2.5/4.5/7.0/10% 를 지남 → 제출 0, 기준선은 단조 상승, 첫 rung 위 첫 관측에서 바로 승격(유예 없음) |
| T6 | `TestA208TurningTheToggleOffKeepsTheAdoptedPolicy` | ON 으로 편입 → OFF 드라이버로 다음 편입 → 앞 포지션의 exit state 정책은 RUNNER 그대로, 새 편입은 공통 정책 |
| T7 | `TestA208FloorOnlyExitRoundTrips`(internal/config/a208_floor_only_exit_test.go) | 키 없음 → 거짓 · 저장 바이트에 키 없음, 참 → merge · LoadRaw · Save 모두 보존 |
| T8 | `TestA208ARefusedBlockClearsFloorOnly` | 손절폭 범위 밖 + 참 → 전면 0(거짓) |
| T9 | `TestA208ConsoleSaveKeepsFloorOnly`(internal/console) | 참인 블록 → 폼 저장 → 저장된 블록 참 유지 |
| T10 | `TestA208FloorOnlyIsAuditedAtStartup`(internal/app/engine) | 기동 기록 끝 행이 `engine.adoption.floor_only_exit`(바로 앞은 a163 행), 같은 값 재기동은 행 0, 값이 바뀌면 한 행 |

T5 · T6 은 이미 있는 경로(편입 RUNNER 사본 · 정책 스냅숏)의 끝-끝 확인이라 base 에서도 통과할 수 있다 — RED 대상은 T1~T4 · T7~T10.
갱신: `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` — 「마지막」 → 「a208 행 바로 앞」(시험 함수 본문 편집 → gate ⑤ 번들).

### 8.3 선택 시험(전체 시험 · 전체 gate 는 사용자 명시 요청 시에만)

`go test ./internal/config/... ./internal/exitpolicy/...`, `go test ./internal/console/`, `go test ./internal/app/engine/ -run 'Adopt|A208|A163|A074|A197|Reconcile|Include|Driver'`,
`go test ./internal/accountbudgetruntime/`, `make lint`, `check_analysis.py`. 이유: 편집 파일의 직접 호출자 · 같은 패키지(8.1) + a166 census.
미실행: 전 저장소 `make test` · `make gate` · 실계좌 검증.

### 8.4 변이 원장(High-risk — 결정 줄)

| M | 변이 | 잡아야 할 시험 |
|---|---|---|
| M1 | `adoptOne` 이 토글을 무시(항상 공통 정책) | T1 · T2 |
| M2 | `adoptOne` 이 OFF 에서도 RUNNER | T3 |
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
| `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | internal/app/engine/a163_source_age_audit_test.go | 시험 |

새 시험 파일 3 개(+ console 시험 1 개는 새 파일). 타입 필드 추가(`config.Adoption` · `rawAdoption`)는 함수 편집이 아니다.
초판 편집 예정이던 `EvaluateLadder` · `canonicalLadderInput` · `judgeLadder` · `Context.ExitObserver` 는 편집하지 않는다.

## 10. 사람 결정(2026-11-28, Manager 전달)

1. **Q1 — 유예 → (b) 넣지 않음.** 사유(Manager 기록): floor-only 정책에서 유예가 막는 것은 보호선 상향뿐이라 그 60초는 유예가 없을 때보다
   덜 보수적이다 — §0.9(손절 · 익절 변경은 더 보수적 방향만)와 충돌. 또 새 스냅숏 필드는 롤백 위험: `internal/journal/exit_snapshot.go`
   `decodeStoredSnapshot` 이 `DisallowUnknownFields()` 를 쓰므로(C11) 옛 이미지는 그 행을 손상으로 읽어 그 포지션의 판정(손절 포함)을 멈춘다.
   → `LadderInput.ProfitSideHeld` · 새 스냅숏 필드 없음.
2. **Q2 — 예, D1(a).** 모든 편입이 기존 RUNNER 사본을 쓴다(부분익절 없음; 잠금 +2.5% → t0, +4.5% → +2.0%, +7.0% → +3.5%). 공통 정책별 편입 사본 없음.
3. **Q3 — 예, 기본안.** 토글은 설정 파일로 켜고(기본 false), `handleSettingsSave` 가 값을 이월해 콘솔 저장이 조용히 끄지 않으며, 기동 audit 을 남긴다.
   이 change 에 콘솔 체크박스 없음. 이미 편입된 포지션은 정책을 유지하고 새 편입만 토글을 따르며, READOPT 는 토글을 따르지 않는다.
4. Q4 — Q3 에 포함(이미 편입된 포지션은 그대로).
5. Q5 — 롤백 순서: 유예 삭제로 제약이 없어졌다(7 절 재유도). 남는 것은 롤백 뒤 재배포 때 설정 키 확인 절차 하나.
6. Q6 — 운영 공통 정책은 HYBRID_50(Manager 전달). 이 change 는 운영 설정을 읽지 않았다.

## 11. 위험

- R3 READOPT 는 새 편입 기록을 만들지 않고 토글도 따르지 않는다(journal/apply_hook.go:709-740 은 exit_states 만 갱신) — 범위 밖, 항목 3.
- R5 토글을 켜면 공통 정책이 HYBRID_50 이어도 새 편입은 RUNNER 잠금(+2.5% → t0)을 받는다. HYBRID 의 runner trail 은 편입 포지션에 없다 —
  사용자 결정(Q2) 그대로.
- R6 롤백 뒤 옛 이미지의 콘솔 저장이 키를 지운다(7 절) — 재배포 절차로만 확인(새 가드 없음).
- (초판 R1 · R2 는 유예 삭제로 해당 없음.)

## 12. 관련 / 다음

- **항목 2(다음 change) — 사용자 결정 기록:** 편입 전용 손절 설정. 유예 기본 **0분**, **재난 손절(disaster stop) 항상 켬**. 이 change 는 그 설정을
  만들지 않는다.
- 항목 3(이후): 종목별 정책 선택 · 편입 해제 · 즉시 청산 버튼 · 콘솔 토글.
