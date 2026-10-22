# a164 review

## 0.8 proposal-freeze 리뷰 — 2026-10-22

- 리뷰: 독립 적대 Eng(설계 비작성자), 읽기 전용. 원문 전체는 `analysis/freeze-review-2026-10-22.md`(정본 — 이 절은 판정과 처분만 옮김).
- 대상: `636be0080` + `f4dc1559d`(change 문서·영수증). base `4cd7af19d` 이후 `internal/`·`cmd/` 변경은 `internal/filldetect/unknown_status_test.go`
  하나라 change AST 29 는 신선(리뷰어 실측).
- **판정: APPROVE-WITH-CHANGES** — P0 0 · P1 1 · P2 6 · P3 6. 방향(사용자 결정 (a), H1~H5)은 재론하지 않음.
- 처분 로트: 2026-10-22 문서 로트(production 편집 0). 각 지적의 인용 줄은 반영 전에 HEAD 로 재확인했다
  (`reconcile_reads.go:136-150` 행 해독이 `apiConditionalOrder` 전체를 풂 · `verify_reconcile.go:228-240` `ensureIdentity` 래퍼 둘 ·
  `reconcile.go:231` `start := now()` · `reconcile_check.go:390-394` 발동 흔적 루프가 대상을 안 뺌(AST `checkReconcileAbsence` B6·B7) ·
  `reconcile_check.go:415` 승인 머리 · `verify_reconcile.go:260` 질문 · `projection_reconciled.go:26` 필터 · `cmd/tossctl/soak.go:438`
  `SucceededEndpoints` 호출 · 정본 spec `:799-806`(결속 괄호)·`:860-883`(감사·증명 경계) · AST `validateM0TriggerMode` B4~B7 `:417`·`:419`).

### 필수 질문 — 「이 diff 의 어느 줄이 살아 있는 조건주문을 닫힘으로 적을 수 있나」

리뷰 답 요약: 생존 판정은 A5(같은 심볼 OPEN 조건 0 · OPEN 일반 0 · CLOSED 발동 흔적 0) 하나가 짊어진다. A5 가 서 있는 한 by-id·CLOSED 증거의
어떤 조합으로도 살아 있는 조건주문을 닫는 입력 순서를 찾지 못했다. A2~A4 는 「만료」라는 이름의 정확도 가드다. 그래서 P1·P2 는 A5 를 약하게 만들
구현 자유도와 정확도 가드의 구멍을 겨눈다. 처분 뒤: A5 를 design D2 머리에 「생략·완화 불가」로 못 박고(P2-6), spec 에 같은 문장과 시나리오,
변이 4.5·4.17 로 판정·읽기 생략을 둘 다 죽인다.

### 처분표

| # | 요지 | 처분 | 반영 위치 |
| --- | --- | --- | --- |
| P1-1 | by-id silent(404) + CLOSED 가 대상을 이름하면 수락 — 목록-상세 모순이고 「옛 id 상세 404」는 정정 서명. D7 논증이 이 갈래에서 거꾸로 | **반영** — A3 를 「by-id 두 응답 둘 다 statement」로. silent ×2 + CLOSED 대상 → 새 거절 `expiry-sources-disagree`, silent ×2 + 부재 → `expiry-evidence-absent`. H1~H5 보다 엄격한 쪽으로만 움직이므로(수락 갈래 하나를 거절로) 새 사람 결정 없음 — 근거는 design D2 「사람 결정과의 관계」. 333430 영향 없음(0.6 by-id 200 `EXPIRED`) | design D2 A3·「통과시키는 것」·거부 1 · D3 표 · D7 · proposal 「무엇을」2 · spec 시나리오 「only the CLOSED list…」를 거절로 · tasks 2.2·2.3 · 변이 4.2·4.12 |
| P2-1 | spec 이 by-id 를 승인 앞에서 금지하지 않음(「list read」만) | **반영** — ADDED 문장을 「any account-scoped order or conditional-order read, including the by-id reads」로. 결속 쪽은 정본 `:799-806` 괄호가 이미 막음(by-id 는 허용 목록에 없음) | spec ADDED · 시나리오 「the approval is declined」 · design D4 · tasks 2.4 · 변이 4.20 |
| P2-2 | a121 「audit and attestation boundaries」 짝 없음 — by-id 200 을 Call 로 남기면 `SucceededEndpoints` 가 셀 수 있음 | **반영(ADDED 에 복제)** — MODIFIED 로 정본 requirement 를 넓히는 대신 ADDED requirement 안에 같은 문장(멱등 · Calls 없음 · 원 실패 보존 · 판정·취소·체결·endpoint·soak·attestation·interlock 불산입 · Q2 식별자 규칙)과 시나리오 둘(repeated · attestation unchanged). 이유: 정본 블록을 통째 MODIFIED 로 복사하면 델타가 커지고 a121 문장 drift 위험이 늘어남 | spec ADDED · design D5 「감사·증명 경계」 · tasks 2.5 · 변이 4.21 |
| P2-3 | CLOSED 행에 leg status 검사 없음(행 타입에 `FirstStatus` 없음) | **닫음(행 타입이 실을 수 있음)** — 해독은 이미 `apiConditionalOrder` 를 통째 푼다(`reconcile_reads.go:141`), 그래서 리뷰 안(새 해독 파일) 대신 `ReconcileConditionalRow.FirstStatus` 한 필드 + 대입 한 줄. 같은 스키마 검증을 두 벌로 베끼는 것보다 작고(판정 한 자리), 부재 모드는 그 필드를 읽지 않아 동작 불변. `DecodeReconcileConditionalPage` 는 FLM 대상에 추가. A4 가 대상 CLOSED 행에도 leg 어휘를 요구(`expiry-target-status`) | design D2 A4 · R1 · FLM 대상 · spec ADDED 둘째 조건 · tasks 2.3·3.1 · 변이 4.13 |
| P2-4 | 새 by-id 리더 메서드의 `ensureIdentity` 경유가 어디에도 없음 | **반영** | design D4 · tasks 2.6·3.1 · 변이 4.20 |
| P2-5 | 변이 목록이 거절 이유 하나당 변이 하나를 못 채움(①~⑧) | **반영** — ①→4.13 · ②→4.14 · ③→4.15 · ④→4.16 · ⑤→4.17(+2.4 호출 수 단언, design A1 에 `readReconcileSet` 이름으로 고정) · ⑥→4.18 · ⑦→4.19(+2.5) · ⑧→4.12·4.20 | tasks 4.12~4.21 · design D2 A1 |
| P2-6 | A5 를 안전 등뼈로 적지 않음 | **반영** — D2 머리 「위계」 문단, spec 에 「these liveness checks SHALL NOT be skipped or relaxed」 + 시나리오 「expiry evidence never relaxes the liveness checks」, tasks 2.3 명시 사례 (a) | design D2 · spec ADDED · tasks 2.3 |
| P3-1 | 신선도 창 시작이 by-id ① 뒤(`reconcile.go:231`) | **반영** — 만료 모드는 by-id ① 직전이 창 시작 | design D2 A1 · D4 · spec ADDED · tasks 2.4 · 변이 4.20 |
| P3-2 | H2 서술의 조건 누락(pending-create 체크포인트면 B8 건너뜀) | **반영** — AST B5(`:417`)·B7(`:419`) 재확인. proposal 비목표를 「pending-create 체크포인트가 없을 때」로 조건화, 결론은 감사 사실에 의존(미검증)으로 명시 | proposal 「비목표」 |
| P3-3 | by-id `type` 미검사 | **반영** — `type == "SINGLE"`, 어긋나면 `expiry-target-status` | design D2 A4 · spec ADDED · tasks 2.3 · 변이 4.13 |
| P3-4 | RED 컴파일 전략 | **반영** — 골격 커밋 먼저(새 심볼, 만료 모드는 늘 거절), RED 는 단언에서 실패 | tasks 2 머리 |
| P3-5 | 공유 발동 흔적 루프에서 대상 행 `COMPLETED` 일 때 어느 코드가 나는가 | **반영(고정)** — 루프는 대상을 빼지 않으므로 `closed-fired-trace`. 이유: A5 가 등뼈라 먼저 돌고, 보조 함수에 대상 제외 인자를 넣지 않아 부재 모드와 같은 함수 그대로 씀. `expiry-target-fired` 는 by-id `COMPLETED` + 목록 흔적 없음에서만 | design D2 A5 · tasks 2.3 (b)(c) |
| P3-6 | 승인 화면이 by-id 읽기를 말하지 않음 | **반영** — 만료 모드 화면·질문에 「상세 조회 2회」 · 별도 하위 명령 금지(리뷰 질문 5 권고) | design D4 · tasks 2.7 |

리뷰 본문 「질문별 확인」의 나머지(같은 id 재사용 잔여 · `m0ManualReconcileIDs` 우회는 안전 방향 · scrub 이 미지 키 보존 · 구 바이너리 안전 방향)는
비차단 확인으로 받아들이고 별도 처분 없음. 거절한 지적: **없음**. P2-2·P2-3 은 리뷰가 제시한 두 안 중 다른 쪽을 골랐고 이유를 위 표에 적었다.

### 반영 뒤 상태

- `openspec validate a164-verify-reconcile-closes-broker-expired-conditional --strict --no-interactive` — valid(2026-10-22).
- production 편집 0. FLM/BTM 번들은 구현 로트 1.1(면제 아님). Function Logic Map: not-applicable — 이 로트는 문서뿐이고 분기 주장은
  `analysis/ast/enumerate-output.md` 의 B 번호로만 함.
- 미검증 잔여: 원 응답 키 존재(C1) · 공식 OPEN 일반 주문 목록(C4 — 6.2 실행이 처음 판정) · 기록의 ChainID·M0 체크포인트·전제 fail 셋.

## 구현 로트 1 — 함수 내부 증거 · Pre-Edit (2026-10-22, worktree `a164-impl` @ `aec9bde67`)

### 1.1 FLM/BTM 번들

`analysis/function-logic/` 13 번들 — 편집 대상 11(`Reconcile` · `checkReconcileAbsence` · `checkReconcileWindow` · `ReconcileApproval.WriteText` ·
`Artifact.terminal` · `Artifact.MarshalJSON` · `ReconciledArtifacts` · `writeReconciled` · `DecodeReconcileConditionalPage` · `newVerifyReconcileCmd` ·
`runVerifyReconcile`) + 근거 대상 2(`readReconcileSet` · `reconcileReads.ensureIdentity` — 편집 없음, 설계 주장 A1·P2-4 의 근거).
BTM 의 시험 칸은 손으로 고르지 않았다: 시험 함수 279 개를 하나씩 `-coverprofile` 로 돌려 분기 본문 블록이 실행된 시험만 적었다
(`analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py` · `notes.json` · `btm-cov-base.json`). `check_analysis.py --change …` rc=0.

- `encodeReconcileLine`: design FLM 대상 목록에 있었으나 **편집하지 않음** — 만료 줄은 새 함수 `encodeReconcileExpiredLine`(새 파일)이 씀.
  Function Logic Map: not-applicable — 기존 함수 무편집.
- `verifyReconcileApprove`: Function Logic Map: not-applicable — 패키지 변수에 담긴 함수 리터럴이라 `tools/logic-map` 이 찾지 못함
  (`function "verifyReconcileApprove" not found`). 편집은 질문 문구의 모드 라벨 하나. 행동 시험: 기존 `TestReconcileDefaultApproval*` 셋 +
  새 만료 모드 질문 시험(2.7).
- 측정이 드러낸 기존 공백(행동 시험 0): `Reconcile` B1~B4(생산 호출자가 늘 채우는 nil 기본값) · B24 · B27(사실상 도달 불가) · B28 · B29 · B31,
  `ensureIdentity` B2, `runVerifyReconcile` B1·B5, `DecodeReconcileConditionalPage` B3. 각 번들의 「Required test」 칸에 사유.

### 1.2 부재 모드 불변 고정

편집 전 새 시험 둘(`internal/verifylive/reconcile_a164_absent_test.go`) — 기준선에서 GREEN:
`TestReconcileAbsenceGuardsRefuseInTheirContractOrder`(네 가드가 겹친 입력에서 하나씩 벗김 → open-conditional → open-plain-order →
closed-target-expired → closed-fired-trace) · `TestReconcileRefusesAtTheWriteBoundaryGateItself`(B29 행동 — 넷째 `now()` 에서만 창 초과).
기존 a121 시험(`reconcile_g1/g2/g3/codex/structure/green/rollback/projection/seal_test.go`)은 기준선 634 전부 통과.

### 1.3

design R3 에 기록 — 성공 응답 requestId 는 클라이언트가 노출하지 않음(`client.go:213-227` `unwrapAndDecode`), statement 는 `result` 바이트 지문.

### 1.4 Pre-Edit Gate

```text
Pre-Edit Gate:
- change id / task id: a164-verify-reconcile-closes-broker-expired-conditional / 2.x·3.x
- 대상 심볼: verifylive.Reconcile · checkReconcileAbsence(A5 보조 함수 추출) · checkReconcileWindow(모드 인자) ·
  ReconcileApproval.WriteText · Artifact.terminal · Artifact.MarshalJSON · ReconciledArtifacts · writeReconciled ·
  official.DecodeReconcileConditionalPage · cmd newVerifyReconcileCmd · runVerifyReconcile · verifyReconcileApprove(문구) ·
  타입 Artifact(필드 ReconciledExpired) · ReconcileConditionalRow(FirstStatus) · ReconcileReader(메서드 하나) · ReconcileParams(Expired) ·
  ReconcileApproval(Mode). 새 파일: internal/verifylive/reconcile_expired.go · internal/official/reconcile_detail_reads.go
- CodeGraph definition/callers/callees/impact: analysis/code-context/evidence-reconciliation.md 「0.9 재확인」 · codegraph-affected-0.9.txt(1,208)
- CodeGraphContext 후보와 evidence reconciliation: CGC incomplete·옛 체크아웃(R5) — advisory 미반영, HEAD grep 으로 대조
- 기존 동작 파악 근거: HEAD aec9bde67 · FLM 13 번들 · 시험별 coverage 279 · 기준선 스위트 verifylive 634 / official 455 / cmd Reconcile 62 통과
- Function Logic Map / Branch Test Map: analysis/function-logic/ (위 1.1), not-applicable 둘은 사유 위
- upstream 상속 테스트 영향: no — verify reconcile 은 TossOS 고유(a121). 부재 모드 시험 전부 무수정 통과가 회귀 기준(Q6 문구 시험 하나는 D8 로 문구만 갱신)
- 실패 테스트 선행 작성: yes — 골격 커밋(만료 모드 = 늘 expiry-evidence-absent) 뒤 RED, 단언에서 실패 확인
- 설정·DB·journal 변경과 rollback: 없음 — 검증 기록(jsonl)에 additive omitempty 필드 하나·새 StepID. 구 바이너리는 그 줄을 비-종결로 읽음(D10)
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문·취소·정정 경로 편집 0, 리더는 GET 만(봉인 시험 갱신), 새 공식 GET 2회는 기존 lease 안(D6),
  토글·운영 설정 변경 0, 실기록 대사 실행은 사람(6.2)
```

구현 결정(설계 안의 선택, safe local — design 의도 명백):

- 만료 모드는 `Reconcile` 안의 `p.Expired` 분기로 넣는다. 잠금·재해독·한 번의 쓰기는 한 벌만(복사 금지 — 구조 핀 셋이 그 한 벌을 고정).
- D4 의 「만료 어휘 확정」 정적 거절 코드가 design 에 이름이 없어 `expiry-vocabulary-unfixed` 로 둠(Q3 의 `freshness-unfixed` 와 같은 모양).
- 어휘 값: top-level 은 `EXPIRED` 단일 비교(A4), leg 어휘 `reconcileExpiryLegVocabulary = {EXPIRED}`, A6 CLOSED 어휘
  `reconcileExpiryClosedVocabulary = {EXPIRED}` — H5(OpenAPI enum ∩ 0.6 영수증). 영수증 사본은 `internal/verifylive/testdata/` 에 두고 골든 시험이 대조
  (change 디렉터리는 archive 때 옮겨지므로 시험이 그 경로를 읽지 않게).

## 5.1 구현 리뷰 — gstack `review` (2026-10-22, 대상 `aec9bde67..b7193f036`, 수리 뒤 `f4e71ad1d`)

- 보이스: (1) Claude 독립 적대 서브에이전트(code-reviewer, 신선한 문맥, 읽기 전용 — 생산 diff 896줄 전독, 시험은 요약),
  (2) Codex 적대 패스(`codex exec -s read-only`, 같은 diff), (3) 구현자 셀프 critical pass(체크리스트 Pass 1·2 — SQL·LLM·셸 해당 없음).
- 필수 질문 「이 diff 의 어느 줄이 살아 있는 조건주문을 닫힘으로 적을 수 있나」:
  - Claude: 없음 — 닫는 자리는 `reconcile.go` 의 `reconciled.ReconciledExpired = true` 하나이고 A5(`checkReconcileNoLiveOrder` ·
    `checkReconcileNoFiredTrace`, 부재 모드와 같은 함수)·A3(silent 면 무조건 거절)·A4·A7 를 모두 통과해야 닿음. 남는 길은 design R1·R2(명칭 오류, 생존 은폐 아님).
  - Codex: 정확히 해독된 LIVE 응답·보이는 OPEN 후속·보이는 발동 흔적으로는 우회 없음. 단 **해독이 부정확해지는 길 하나**(아래 C-P1).
- 판정: 수리 뒤 P0·P1 잔여 0. 잔여 P2 셋은 전부 a121 공용 경로(전송·쓰기 경계)에서 물려받은 것이라 이 change 범위 밖 — Manager 결정 항목.

| # | 보이스 | 요지 | 처분 |
| --- | --- | --- | --- |
| C-P1 | Codex | by-id 해독이 `map[string]json.RawMessage` 로 접어 **중복 키의 마지막 값**만 남김 — `"triggeredOrderId":"CHILD","triggeredOrderId":null` 이 미발동이 되어, 목록이 비고 CLOSED 가 축출된 경우 발동한 주문을 만료로 닫을 수 있음 | **수리 `f4e71ad1d`** — `reconcileObject` 가 토큰 단위로 읽어 중복 키·뒤따르는 데이터를 거절(result · first 두 층). 시험 `TestReconcileByIDRefusesSchemaHoles` 의 `dup-status`·`dup-first`·`dup-triggered`·`dup-second`(RED 확인 뒤 GREEN), 변이 4.22 |
| C-P1 잔여 | Codex(파생) | 같은 성질이 a121 목록 행 해독(`DecodeReconcileConditionalPage` → `apiConditionalOrder` 구조체)에도 있음 — CLOSED 의 **다른 id** 행 발동 흔적을 중복 키로 숨기면 A5 셋째가 못 봄(두 모드 공통) | **미수리 · Manager 결정** — a121 공용 해독 경로이고 부재 모드 동작을 바꾸므로 a164 범위 밖. 대상 자신의 발동은 by-id(중복 거절)가 막음. 후속 change 권고 |
| C-P2a | Codex | 공용 `Client.send` 가 모든 2xx 를 받음 — design D3 의 「HTTP 200」 과 다름 | **미수리 · Manager 결정** — 공용 전송(High-risk) 편집이 필요. 본문은 여전히 엄격 스키마·A4 를 통과해야 함. design D3 와 코드의 차이로 기록 |
| C-P2b | Codex | 401 이면 토큰 갱신 뒤 한 번 재시도 — 재시도는 `ensureIdentity` 를 다시 하지 않고, 갱신이 디스크의 다른 자격 토큰을 쓸 수 있음(`token.go`) | **미수리 · 상속** — a121 목록 읽기와 같은 전송(CG-6 은 첫 범위 읽기 전 한 번만 확인). 계좌 헤더 seq 는 고정. 후속 change 권고(기억 「공유 credential 토큰 전쟁」 과 같은 계열) |
| C-P2c | Codex | 마지막 신선도 재검이 직렬화 **전**에 잡은 `at` 을 씀 — 직렬화·봉인·fsync 지연은 15 s 창 밖 | **미수리 · 상속** — a121 쓰기 경계 구조 그대로이고 구조 핀(`TestReconcileFinalGateSitsAtTheWriteBoundary`)이 고정. 만료 모드가 더한 직렬화(지문·관측)는 마이크로초 단위 |
| C-P3 = F1 | 둘 다 | 부재 모드 신선도 거절 문구와 report 대사 절 머리 바이트가 바뀜(D1 「D8 예외만」 위반) | **수리 `f4e71ad1d`** — 부재 모드 a121 바이트 복원, 만료 모드만 새 문구. 시험 `TestReconcileAbsenceKeepsItsA121Wording`, 변이 4.23 |
| F2 | Claude | 줄의 `started_at` 은 창 시작(by-id ① 직전)인데 D5 는 「첫 목록 읽기」 라고 적음 | **문서 사안 · Manager** — 코드는 A1·P3-1(창 시작 = by-id ① 직전)과 맞음. D5 문장이 낡음 — frozen design 이라 고치지 않고 여기 기록 |
| F3 | Claude | D10 의 「구 바이너리는 다시 outstanding·정리 대상(DELETE 404 반복)」 은 부정확 — 확정형이 HeldUntil 을 싣지 않아 구 바이너리는 **다시 보유**(정리 대상 아님) | **문서 사안 · Manager** — 안전 방향(Outstanding·HOLD 에 남음). 시험 `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` 이 「보유」 를 단언. a121 과 같은 성질 |
| F4 | Claude | `reconcile.expired.by_id` 값 `"statement,statement"` 상수 | **수리 `f4e71ad1d`** — 응답 분류에서 유도 |

셀프 critical pass(체크리스트): 새 상태 값 `reconciled_expired`·`reconcile-expired` 의 소비자를 HEAD grep 으로 전수 — 종결은 `terminal()` 한 곳
(`outstandingLines` 두 자리 · `ReconciledArtifacts` · `joinChain`), `reconciled_absent` 를 직접 읽는 생산 코드 0, StepID 비교 소비자 넷은 카탈로그 밖 값이라 영향 없음
(`TestReconcileExpiredStepIDIsOutsideTheCatalogueAndTheAbsenceStep`). 추가 발견 없음.

## 5.2 영향 범위 검증 (2026-10-22, `f4e71ad1d`)

- 실행: `go test -count=1 ./internal/verifylive/ ./internal/official/ ./cmd/tossctl/`(무태그, 패키지 전체) 통과 ·
  `go test -count=1 -tags tossos_testseams` 같은 세 패키지 2080 통과 · `make lint`(gofmt + `go vet ./...` + 태그 vet) rc=0 ·
  `openspec validate a164-… --strict --no-interactive` rc=0 · `check_analysis.py --change a164-…` rc=0.
- 선택 이유: 편집 파일은 이 세 패키지에만 있음. CodeGraph affected 1,208 은 cmd/tossctl 전이 의존 때문이고, 그 밖 패키지는 바뀐 심볼
  (`Artifact.terminal`·`ReconcileReader` 등)을 `go build ./...`·`go vet ./...` 로만 확인(컴파일 통과).
- **미실행**: `make test-seams`(= `go test -tags tossos_testseams ./...`, 45분) 전체 · `make test` · `make test-race` · `make gate`.
  정본 규칙(전체 시험·게이트는 사용자 명시 요청 때만)과 tasks 5.2(「`make test-seams` 중 대사 seam」)에 따라 영향 세 패키지로 한정.
  콘솔(`internal/console` 의 `PendingCleanup` 소비)은 시험을 돌리지 않음 — 종결 술어 한 곳 편집이라 같은 판정을 받음(코드 근거만).
- 관측: 리뷰 수리 직후 verifylive 한 판이 어느 시험인지 남기지 못한 채 1 회 실패 → 곧바로 세 판(단독 ×1, official 과 함께 ×2) 연속 통과.
  원인 미확인(편집 직후 컴파일 캐시 경합으로 추정 — 미검증). 재현 안 됨.
- 운영 메모: 리뷰 시점(18:20 KST)에 이 worktree 의 git admin 디렉터리에 0바이트 `index.lock` 이 남아 커밋이 막힘 — 잡은 프로세스 없음,
  `git fsck --connectivity-only` 정상 확인 뒤 제거(이 worktree 전용, 30분 넘은 빈 파일만).

## 5.3 sdd-sync (2026-10-22, HEAD `8246ad1e2`) — 미완(체크 안 함)

- `make sdd-sync` rc=2: CodeGraph fingerprint 는 `8246ad1e2` 로 기록됨(`.sdd/index-state.json`), CodeGraphContext `incomplete`(R5 그대로).
- `python3 tools/sdd/check_index_freshness.py` rc=1 두 판 연속: `CodeGraph status probe failed: … timed out after 15 seconds`.
  같은 명령을 직접 재면 `codegraph status .` → "✓ Index is up to date" 인데 **real 27 s**(user 1.3 s — 디스크 대기, `/mnt/D` NTFS 위 195 MB db).
  판정: 색인은 신선하나 게이트 프로브가 시간 초과로 실패 — 환경 사안. gate 전에 부하 낮은 때 `make sdd-sync` → `check_index_freshness.py` 재실행 필요.
  이 절을 쓰는 커밋 자체가 fingerprint 를 다시 낡게 하므로, 어차피 gate 직전 재동기화가 필요함.
