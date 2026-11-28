# a208 분기 주장의 AST 근거

proposal · design 이 근거로 쓰는 기존 함수 내부 분기 · early return · 호출은 이 디렉터리의 `tools/logic-map` 산출물로 열거함
(.claude/CLAUDE.md 필수 진입 5). 이 단계(제안 · 설계)는 기존 함수를 편집하지 않으므로 FLM 번들이 아니라 주장의 근거로 이 자리에 둠
(선례 a203 · a206 · a207 `analysis/ast-evidence/`). 구현 단계에서 고치는 함수는 편집 **전에** `analysis/function-logic/` 번들을
따로 만듦(tasks.md 1.1).

- 기준: worktree `~/tossos-work/a208` HEAD `baeddfe31`(= origin/main), 2026-11-28.
- 생성: `GOFLAGS=-trimpath go build -o <scratch>/logicmap ./tools/logic-map` 뒤 `logicmap --file <파일> --func <함수> > <패키지>--<함수>.ast.json`.
- 줄 번호 ↔ 원문 대조는 산출물의 `branches[].at.line` · `returns[].at.line` · `calls[].at.line` 을 같은 HEAD 원문 줄과 맞춰 읽음.
  분기의 조건식 문구는 그 원문 줄을 인용한 것이고(logic-map 은 조건식을 싣지 않음), 분기 · return · 호출의 **존재와 위치**가 AST 근거임.

| 산출물 | 파일:줄 | source_sha256(앞 12) | 분기 | return | 쓰임 |
|---|---|---|---|---|---|
| `exitpolicy--EvaluateLadder` | internal/exitpolicy/ladder.go:307-481 | ee9c91aa8a5d | 32 | 23 | 편집 예정(유예 입력) |
| `exitpolicy--canonicalLadderInput` | internal/exitpolicy/snapshot.go:272-289 | 5aaed0f721c2 | 3 | 2 | 편집 예정(digest 토큰) |
| `exitpolicy--EvaluateLadderSnapshot` | snapshot.go:96-167 | 5aaed0f721c2 | 11 | 7 | 인용 |
| `exitpolicy--ValidateRecoveryDerivation` | internal/exitpolicy/recovery.go:99-130 | 292a78a44ead | 8 | 5 | 인용 |
| `exitpolicy--CommonLadderForPosition` | internal/exitpolicy/common_policy.go:106-127 | 7f44c4c04a10 | 5 | 4 | 인용(재사용) |
| `exitpolicy--LegacyLadderPolicyIdentity` | internal/exitpolicy/identity.go:51-76 | 48162e44dfb5 | 8 | 2 | 인용(재사용) |
| `exitpolicy--EvaluateRatchet` | internal/exitpolicy/ratchet.go:335-453 | 9aaa342d5ff8 | 22 | 14 | 인용(안 고침) |
| `exitpolicy--ratchetCandidate` | ratchet.go:462-511 | 9aaa342d5ff8 | 15 | 14 | 인용(안 고침) |
| `engine--ReconcileDriver.adoptOne` | internal/app/engine/adoption.go:373-435 | 2a60c39990e2 | 3 | 3 | 편집 예정(정책 선택 · 알림 문구) |
| `engine--ExitObserver.judge` | internal/app/engine/exitloop.go:1010-1053 | 0a799bac4f3b | 9 | 7 | 인용 |
| `engine--ExitObserver.judgeLadder` | exitloop.go:1128-1195 | 0a799bac4f3b | 8 | 7 | 편집 예정(유예 판정) |
| `engine--ExitObserver.ladderFor` | exitloop.go:1314-1329 | 0a799bac4f3b | 3 | 4 | 인용 |
| `engine--managedPolicyIdentity` | exitloop.go:758-771 | 0a799bac4f3b | 3 | 4 | 인용 |
| `engine--ExitObserver.openState` | exitloop.go:797-844 | 0a799bac4f3b | 10 | 7 | 인용 |
| `engine--ExitObserver.openAdoptedState` | exitloop.go:855-873 | 0a799bac4f3b | 3 | 3 | 인용 |
| `engine--ExitObserver.quoteUsable` | exitloop.go:1228-1237 | 0a799bac4f3b | 1 | 2 | 인용 |
| `engine--ExitObserver.workingSet` | exitloop.go:596-756 | 0a799bac4f3b | 31 | 3 | 인용(롤백 위험) |
| `engine--recordGateSettings` | internal/app/engine/interlock.go:774-848 | 7d2ce7348681 | 6 | 2 | 편집 예정(audit 행) |
| `engine--Context.ExitObserver` | internal/app/engine/exitwiring.go:319-368 | 9cea5b87e5f7 | 6 | 4 | 편집 예정(배선 한 줄) |
| `engine--Context.ReconcileDriver` | internal/app/engine/reconcileloop.go:369-400 | a47a346fb26f | 6 | 4 | 인용(안 고침) |
| `journal--Journal.OpenAdoptedExitState` | internal/journal/adoption.go:279-291 | 6adc78fcb71d | 1 | 2 | 인용 |
| `journal--policyKindForID` | adoption.go:468-473 | 6adc78fcb71d | 1 | 2 | 인용 |
| `journal--seedPolicyIdentity` | internal/journal/exit_state.go:206-236 | a210bdabf1f8 | 6 | — | 인용 |
| `journal--StoredExitSnapshot.validate` | internal/journal/exit_snapshot.go:178-201 | 09af632ae584 | 7 | 7 | 인용(롤백 위험) |
| `config--mergeAdoption` | internal/config/engine.go:293-309 | 0f545b846c3c | 3 | 2 | 편집 예정 |
| `config--Service.LoadRawEngineAdoption` | internal/config/adoption_io.go:47-73 | 6f1df446b991 | 4 | 4 | 편집 예정 |
| `config--Service.SaveEngineAdoption` | adoption_io.go:80-126 | 6f1df446b991 | 9 | 9 | 편집 예정 |
| `console--Console.handleSettingsSave` | internal/console/settings.go:238-273 | f326a5647a5f | 4 | 4 | 편집 예정(값 이월) |
| `main--recordAdoptionSave` | cmd/tossctl/adoptionsettings.go:75-92 | e8ae8ab98af8 | 2 | 2 | 인용(안 고침) |

## 주장

### 편입 판정과 정책이 정해지는 자리 (오늘)

- **C1 — 편입 기록의 정책 ID 는 엔진 전역 공통 정책을 그대로 복사함.** `ReconcileDriver.adoptOne` 은 분기 3 개(B1@375 합성 손절 유도 실패 →
  return@377, B2@394 `AdoptPosition` 거절 → return@396, B3@399 exit state 개설 실패 → 기록만 하고 계속)뿐이고, 정책 ID 는 분기 밖
  `adoption.go:392` `ExitPolicyID: d.opts.CommonPolicy` 한 자리에서 정해짐. 편입만의 정책을 고르는 갈래는 0 개.
- **C2 — 편입 포지션의 exit state 종류는 정책 ID 의 유무 하나로 정해짐.** `Journal.OpenAdoptedExitState`(B1@281 기록 읽기 실패 →
  return@282) 는 `policyKindForID(adoption.ExitPolicyID)` 로 종류를 고르고, `policyKindForID` 는 B1@469 빈 ID → RATCHET(return@470),
  그 밖 → LADDER(return@472).
- **C3 — LADDER 편입 포지션의 정책 정체성은 편입 여부를 인자로 받아 고정됨.** `seedPolicyIdentity` B1@212 RATCHET / B2@214 LADDER →
  `LegacyLadderPolicyIdentity(policyID, adopted)`. 그 함수의 B5@63 `case CommonLadderRunner` 안 B6@65 `if adopted` 만 버전 ·
  digest 를 `1.0.0-adopted.1` · `adoptedRunnerPolicyDigest` 로 바꿈. BALANCED(B4@61) · HYBRID_50(B7@68) 에는 편입 갈래가 없음.
- **C4 — 판정 시점의 사다리 표도 같은 규칙.** `ExitObserver.ladderFor` 는 B1@316 legacy `default_v1` 이 아니면
  `CommonLadderForPosition(id, m.position.Adopted())`(호출 @1324) 로 표를 받음. `CommonLadderForPosition` 의 편입 갈래는 B2@111
  `adopted && policy.ID == CommonLadderRunner` 하나이고, 그 안 B3@112 range 가 모든 rung 의 `PartialRatio` 를 `"0"` 으로 바꿈.
  RUNNER 표는 `FinalTakeFull` 이 없음(common_policy.go:44-54). → **오늘 TossOS 에서 부분익절 없는 편입 포지션은 공통 정책이 RUNNER 일 때뿐.**
  BALANCED · HYBRID_50 편입 포지션은 엔진 진입 포지션과 같은 부분익절 사다리를 받음.
- **C5 — 판정 분기는 정책 종류 하나로 갈림.** `ExitObserver.judge` B7@1047 switch: B8@1048 LADDER → `judgeLadder`(return@1049),
  B9@1050 그 밖 → `judgeRatchet`(return@1051). 그 앞의 B1@1012 시세 사용 불가 · B2@1019 묵은 발의 정리 · B3@1022 정체성 오류 ·
  B5@1036 재판정 표시 실패 · B6@1042 본전가 계산 실패는 모두 판정 전 조기 return 이고 이 change 는 손대지 않음.

### 사다리 평가 안에서 익절 · 보호선 · 손절이 갈리는 자리

- **C6 — `EvaluateLadder` 의 단계와 분기.** B12@355 워터마크 흡수(관측가 > 이전 고점) · B13@374 range + B15@379 도달한 최고 rung ·
  B16@386 rung 이 있으면 잠금가 합성(B18@392 마지막 rung 의 runner trail) · B22@420 rung 승격 · B23@427 완료면 return@430 ·
  **B25@439 손절(관측가 < 합성 기준선) → B26@441 같은 손절이 미해소면 억제 return@443, 아니면 `ActionLadderStop` 비율 1 발의 return@448** ·
  B27@453 승격 없으면 return@454 · B28@457 switch: B29@458 마지막 rung ∧ FinalTakeFull → 전량 익절, B30@461 `PartialRatio > 0` → 부분익절,
  B31@466 그 밖 → 상태 전용(보호선만 올림) return@469 · B32@475 미해소 발의가 있으면 익절 억제.
  → 손절(B25)은 승격 판정(B22) **뒤**, 익절 갈래(B28) **앞**에 있음. 손절은 그 관측에서 합성된 기준선과 비교함.
- **C7 — floor-only RUNNER 에서는 B29 · B30 이 참이 될 수 없음.** C4 의 B3 가 `PartialRatio` 를 전부 `"0"` 으로 만들고 RUNNER 는
  `FinalTakeFull` 이 거짓이므로, 남는 결과는 B31(상태 전용 승격) · B25(손절) · B27(무변화) 셋뿐.
- **C8 — 비율 래칫(RATCHET)에도 부분익절과 본전 승격이 있음.** `EvaluateRatchet` B18@438 `wantsPartial` → B22@444 기본 갈래가
  `ActionRatchetPartial`(비율 `cfg.PartialRatio`) 발의. 본전 승격은 `ratchetCandidate` B10@496 · B12@498 breakeven trigger 갈래(return@499). 손절은 B16@422.
  → 공통 정책이 비어 있는 엔진의 편입 포지션도 오늘은 40% 부분익절을 받음. design D1 이 편입 포지션을 RATCHET 으로 보내지 않는 근거.

### 기록 · 복구가 평가 입력을 다시 읽는 자리

- **C9 — 저장 스냅숏은 평가를 다시 돌려 검증됨.** `ValidateRecoveryDerivation` B4@114(else, LADDER) 가 저장된
  `LadderSnapshotInput` 으로 `EvaluateLadderSnapshot`(호출 @116) 를 다시 돌리고 B8@126 `!reflect.DeepEqual` 이면 오류 return@127.
  → 평가 입력에 새 값을 더하면 그 값이 저장 입력(JSON)에 실려야 재평가가 같은 답을 냄.
- **C10 — 입력 digest 는 정해진 필드 목록만 덮음.** `canonicalLadderInput` 은 B1@276 range(숫자 6 개 정규화, B3@281 오류 return@282)
  뒤 rung · 완료 · 미해소 action · 미해소 rung 을 덧붙이고 return@288. 새 입력은 여기 넣지 않으면 digest 가 그것을 모름.
- **C11 — 옛 이미지가 모르는 키를 가진 저장 스냅숏은 손상으로 읽힘.** `decodeStoredSnapshot`(exit_snapshot.go:250-264, AST 미생성 —
  분기 주장이 아니라 `dec.DisallowUnknownFields()` 한 줄의 존재 인용) 와 `StoredExitSnapshot.validate`(178-201). 손상은
  `ExitObserver.workingSet` 의 `result.Corruption != nil` 갈래(exitloop.go:673-686)로 가서 격리 · `refused` 로 빠지고, `judge` B3@1022
  가 경보만 내고 return@1024 — **그 포지션은 손절까지 판정되지 않음.** design 7 의 롤백 순서 근거.

### 설정 · audit

- **C12 — 편입 설정 블록은 네 필드만 읽고 씀.** `mergeAdoption`(B2@298 enabled 포인터, B3@304 거부 → 전면 0 return@306),
  `LoadRawEngineAdoption`(블록을 66-70 줄에서 필드별로 다시 조립), `SaveEngineAdoption`(81 줄 `next := Adoption{Enabled, DefaultStopPct}`
  + 82-83 목록, 분기 B1@85 검증 거부). 새 필드는 세 자리 모두에 손으로 옮겨야 함.
- **C13 — 콘솔 저장은 블록을 폼에서 새로 만듦.** `Console.handleSettingsSave` 는 B2@245 읽기 실패만 보고 읽은 값을 버림(`_, _, err :=`
  @244), 250-254 줄에서 폼 네 값으로 `config.Adoption` 을 새로 만들어 B4@268 `Save` 함. → 새 필드를 이월하지 않으면 콘솔 저장 한 번이
  그 값을 조용히 끔.
- **C14 — 엔진 기동 audit 은 끝에 덧붙이는 표.** `recordGateSettings` B5@842 range 가 표의 순서대로 기록함. 마지막 행은
  a163 `engine.exit_source_age_report`(interlock.go:840) 이고, `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten`
  이 그것이 마지막임을 단언함 → 새 행을 덧붙이면 그 시험 본문을 고쳐야 함(gate ⑤ 번들 대상).
- **C15 — 편입 드라이버는 편입 설정 전체를 이미 받음.** `Context.ReconcileDriver` 는 분기 밖 383 줄 `opts.Adoption = c.Config.Engine.Adoption`
  으로 블록을 통째로 넘김 → 정책 선택 쪽 배선은 0 줄. 관측기 `Context.ExitObserver` 는 343 줄에 `CommonPolicy` 만 받음 → 유예 쪽 배선 한 줄 필요.

### 시세 시각

- **C16 — 관측 시세의 시각은 벽시계 UTC 이고 미래 시세는 버림.** `ExitObserver.quoteUsable` B1@1233 `now.UTC().Before(quote.FetchedAt)`
  → 사용 불가 return@1234, 아니면 임대 경과 비교 return@1236. 편입 기록의 시각은 `adoption.go:391` `journal.RFC3339(d.clk.Now())` 로 같은
  벽시계에서 옴.
