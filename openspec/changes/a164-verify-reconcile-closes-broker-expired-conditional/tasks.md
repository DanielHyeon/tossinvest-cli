# a164 tasks

> **High-risk 인접**(검증 기록의 미결 판정 → `--include-trigger` HOLD·정리 프롤로그). 계약은 「브로커의 양성 증거로만 닫는다」(design D2).
> 사용자 결정 2026-10-22 (a) 가 a121 Q6 귀결을 대체한다(design D0). 주문 경로 편집 0. 이 로트(0.1~0.5)는 문서·AST 뿐이다.
> 실기록 대사(6)는 `mutating: true` 명령이라 **사람이 터미널에서** 돌린다.

## 0. 게이트 선행

- [x] 0.1 PM: `STORY-TOS-a164` 등록(Feature `FEAT-TOS-002` — 검증 Story `STORY-TOS-003` 과 같은 Feature) · Feature `stories` ·
      `_registry.yaml` 연결 → `python3 tools/pm/generate_master_tracker.py --check` — 2026-10-22, 다른 세션의 미커밋 PM 편집(a162)과
      섞지 않으려고 HEAD 의 격리 worktree 에서 재생성하고 a164 몫만 가져옴
- [x] 0.2 `python3 tools/sdd/capture_change_base.py --change a164-verify-reconcile-closes-broker-expired-conditional` — base
      `4cd7af19db013435576e44302c964ebc5c2f0557`(`feat/a112-four-family-runtime` HEAD)
- [x] 0.3 `openspec validate a164-verify-reconcile-closes-broker-expired-conditional --strict --no-interactive` — 2026-10-22 rc=0
- [x] 0.4 함수 내부 주장의 AST 선행 — `analysis/ast/` 29개 + `enumerate-output.md`(분기 169 · 반환 138 · STALE 0)
- [x] 0.5 CodeGraph hard evidence — `analysis/codegraph-affected*.txt`(`--filter '*_test.go'`, codegraph 1.6.0) ·
      `analysis/code-context/{codegraph-baseline,codegraphcontext-context,evidence-reconciliation}.md`
- [x] 0.6 **사람 읽기 전용 확인 — 이 change 의 진행 여부를 정함.** 엔진 프로필 자격으로, 장 시간 무관:
      ① `tossctl --config-dir <엔진 프로필> order conditional get DJwYn8P_dD9lQMEeYc-5l5_yDTzu2fo55FEJkhU8WVg --output json`
      ② `tossctl --config-dir <엔진 프로필> order conditional list --status CLOSED --symbol 333430 --limit 100 --output json`(hasNext 면 커서로 끝까지)
      ③ 같은 명령 `--status OPEN` · `tossctl … order list --status OPEN --symbol 333430`(있으면).
      두 명령 모두 `mutating` 표지가 없는 GET 이다(`cmd/tossctl/order.go` list·get). 결과를 계좌 정보 제거 뒤
      `analysis/receipts/expiry-probe-<date>.json` 에 남긴다(top-level status · first.status · triggeredOrderId · second 유무 · 404 본문 code·requestId).
      판정: ①이 200 `EXPIRED` 이거나 ②에 그 id 가 `EXPIRED` 로 있으면 계속. 둘 다 아니면(404 + 부재) **여기서 멈추고 사용자 재결정**(H3).
      ③에 행이 있으면 H4 가 막는다 — 사용자에게 보고
      — **2026-10-22 Manager 실행(사용자 승인 「둘 다 예」), 판정: 계속.** ① 200 `EXPIRED`(SINGLE · first `EXPIRED` · second·발동 id 없음) ·
      ② CLOSED 1 행 = 그 id `EXPIRED` · ③ OPEN 조건 0. 영수증 `analysis/receipts/expiry-probe-2026-10-22.json`(수량 제거, 계좌 키 0 확인).
      ③의 일반 주문 쪽은 공식 읽기로 재지 못함(`orders list` 는 WTS pending — design 「0.6 영수증 대조」 C4) → 6.2 실행이 판정.
      영수증은 CLI 도메인 출력이라 원 바이트 골든이 아님(C1) · D3 키 존재 요구 정정(C2)
- [x] 0.7 사람 결정 H1~H5 기록(design 「사람 결정」 표에 날짜·선택) — 2026-10-22 사용자 「둘 다 예」: 다섯 다 기본안으로 확정
- [x] 0.8 **proposal-freeze 리뷰** — 적대적 Eng 보이스 포함(High-risk 인접), 결과 `review.md`.
      필수 질문: 「이 diff 의 어느 줄이 살아 있는 조건주문을 닫힘으로 적을 수 있나」
      — 2026-10-22 독립 적대 리뷰 APPROVE-WITH-CHANGES(P0 0 · P1 1 · P2 6 · P3 6, 원문 `analysis/freeze-review-2026-10-22.md`).
      같은 날 문서 로트로 13건 전부 반영(처분표 `review.md`). 답: 생존 판정은 A5 하나가 짊어진다 — A5 가 서 있는 한 살아 있는 조건주문을
      닫는 입력 순서는 리뷰가 찾지 못함. P1-1 로 by-id silent + CLOSED 수락 갈래를 거절로 바꿈
- [x] 0.9 편집 착수 직전 `make sdd-sync` → 0.5 질의 재실행 · `evidence-reconciliation.md` 갱신(fingerprint 신선)
      — 2026-10-22 worktree `a164-impl`: CodeGraph fingerprint 기록, affected 1,208(`analysis/codegraph-affected-0.9.txt`), CGC incomplete(R5)

## 1. 함수 내부 증거 (편집 전, 면제 없음)

- [x] 1.1 FLM/BTM 번들 — design 「FLM/BTM 대상」의 기존 함수 전부(`scaffold_analysis.py` → `ast.json` · `function-logic-map.md` ·
      `branch-test-map.md` · `risk-pattern-report.md`). `verifyReconcileApprove` 가 logic-map 에 안 잡히면 `Function Logic Map: not-applicable`
      사유 + 행동 시험
      — 13 번들, BTM 은 시험별 coverage 측정(review.md 「구현 로트 1」), `check_analysis.py` rc=0
- [x] 1.2 부재 모드 불변 고정 — 플래그 없는 실행의 거절 코드·순서를 기존 시험(`reconcile_g1/g2/g3/codex/structure_test.go`)으로 확인하고,
      빠진 갈래가 있으면 편집 **전에** 시험을 더해 GREEN 으로 둔다
      — 새 시험 둘 기준선 GREEN(review.md 1.2)
- [x] 1.3 `official` 클라이언트가 성공 응답 requestId 를 노출하는지 코드로 확정(design D5 R3) — 없으면 응답 `Digest` 로 대신함을 design 에 적음
      — 노출 안 함, design R3
- [x] 1.4 Pre-Edit 선언(`review.md`)
      — review.md 「1.4 Pre-Edit Gate」

## 2. RED

> 컴파일 전략(freeze 리뷰 P3-4): 첫 커밋은 **골격**이다 — 새 필드 `ReconciledExpired`·`FirstStatus`, 새 거절 코드 상수, 새 리더 메서드·
> `--expired` 플래그를 **판정 없이**(만료 모드 진입 = 항상 `expiry-evidence-absent` 거절, by-id 메서드 = 미구현 오류) 세운다. 골격은 생산 동작을
> 바꾸지 않는다(플래그 없는 경로 불변, 플래그 있는 경로는 늘 거절). 그 위에서 RED 시험이 **컴파일되고 단언에서 실패**해야 RED 로 친다.

- [x] 2.1 by-id 대사 읽기 해독: statement · silent(404 `conditional-order-not-found`) · 거절(404 다른 code·해독 불능 본문 · 400 · 401/403 · 429 · 5xx ·
      전송 · null result · required 필드 결측/null/형 불일치 · `first` 결측 · `triggeredOrderId`/`second` 형 불일치) + 수락: `triggeredOrderId`·`second`
      키 없음과 null 이 같게 「없음」(design D3 정정) — httptest. 원 바이트 골든은 OpenAPI 스키마 + 0.6 영수증 값으로 만든다(영수증은 CLI 출력, C1)
      — RED 2026-10-22: `internal/official/reconcile_detail_reads_test.go` 5 시험(골든 = OpenAPI 모양 + 0.6 값) — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)
- [x] 2.2 증거 규칙 수락 두 갈래: by-id statement ×2 + CLOSED 대상 행 없음(축출) · by-id statement ×2 + CLOSED 대상 행 `EXPIRED`.
      (첫 판의 「CLOSED 만(by-id silent)」 수락 갈래는 freeze 리뷰 P1-1 로 2.3 거절 갈래로 옮김)
      — RED 2026-10-22: `TestExpiryReconcileAcceptsTwoStatementsWhenTheClosedListEvictedTheTarget` · `…AndAnExpiredClosedRow` (+ A6·A7 양성 대조 둘) — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)
- [x] 2.3 증거 규칙 거절 — design D2 거부 1~5 와 D3 표의 행마다 하나, **가드별 거절 코드 단언**(공유 문구 아님):
      `expiry-evidence-absent` · `expiry-target-live` · `expiry-target-fired` · `expiry-target-status` · `expiry-read-error` · `expiry-reads-differ` ·
      `expiry-chain-open` · `expiry-sources-disagree`(by-id silent ×2 + CLOSED 대상 행 — P1-1) · `open-conditional` · `open-plain-order` ·
      `closed-fired-trace` · `closed-status`(어휘 밖) · by-id id/심볼/시장/type≠`SINGLE`(P3-3) 불일치 · by-id/CLOSED 행 `first.status` 어휘 밖(P2-3).
      명시 사례: (a) by-id 200 `EXPIRED` ×2 + CLOSED `EXPIRED` 인데 OPEN 조건 1건 → `open-conditional`(P2-6, A5 는 증거로 완화 불가) ·
      (b) 만료 모드에서 대상 CLOSED 행이 `COMPLETED` → `closed-fired-trace`(공유 발동 흔적 루프가 먼저 — P3-5) ·
      (c) by-id `COMPLETED` + 목록 발동 흔적 없음 → `expiry-target-fired` · (d) by-id 한 번 statement·한 번 silent → `expiry-reads-differ`
      — RED 2026-10-22: `TestExpiryReconcileRefusesEachGuardWithItsOwnCode` 하위 38 + `TestExpiryReconcileRefusesAnOpenChainSibling` · 어휘 둘 — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)
- [x] 2.4 순서: 승인 전 거절(후보·계좌·Q3·어휘) · 승인 뒤 읽기 · **승인 거절이면 by-id 호출 0**(P2-1) · `--expired` 에서 Q1 미호출 ·
      플래그 없음에서 Q1 호출(기존) · 창 시작이 by-id ① 직전(P3-1 — by-id ① 뒤 지연을 주입해 15 s 초과 거절) ·
      리더가 받은 `ReconcileOpenOrdersPage` 호출이 두 읽기 만큼(P2-5 ⑤ — 만료 모드가 `readReconcileSet` 을 그대로 씀)
      — RED 2026-10-22: `TestExpiryReconcileStaticRefusalsComeBeforeApproval` · `…DeclinedApprovalReadsNothing` · `…DoesNotNeedTheQ1Measurement` · `…ReadsInTheContractOrder` · `…WindowStartsBeforeTheFirstByIDRead` · 구조 핀 — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)
- [x] 2.5 기록: 줄 하나 추가 · 기존 바이트 불변 · `terminal()` 넷째 종결 · `Outstanding`/`PendingCleanup`/HOLD 두 곳(cmd·`New`)이 따름 ·
      구 바이너리 시뮬레이션(필드 버린 해독)에서 다시 outstanding · 계좌 원문·응답 원문 없음 · 줄의 `calls` 비어 있음 ·
      `SucceededEndpoints` 결과 불변 · 같은 artifact 재실행 = `no-candidate` · 관측에 대상 외 브로커 id 없음(Q2) — P2-2 ·
      `ReconciledArtifacts`·status 출력에 「reconciled expired」로 보임(P2-5 ⑦ — `projection_reconciled.go:26` 필터)
      — RED 2026-10-22: verifylive 6 + cmd `TestM0TriggerHoldFollowsAReconciledExpiredArtifact` · 구 바이너리 `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)
- [x] 2.6 봉인: 대사 리더 메서드 집합이 GET 넷(목록 2 · 종목 · by-id)뿐 — `reconcile_seal_test.go` 세 층 갱신 ·
      cmd `reconcileReads` 의 by-id 가 첫 계좌 범위 호출일 때 Accounts 재조회가 먼저(P2-4, `ensureIdentity`)
      — RED 2026-10-22: 봉인 2층 bool · 읽기 메서드 목록(골격 커밋) · `TestReconcileByIDReadRechecksIdentityBeforeItIsSent` — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)
- [x] 2.7 cmd: `--expired` 는 같은 명령의 플래그(별도 하위 명령 금지) · 승인 화면/질문의 모드 표시 + 만료 모드에서 「상세 조회 2회」 명시
      (P3-6 — 지금 머리 「공식 목록 조회 전용」 `reconcile_check.go:415`, 질문 「reconciled absent 로 기록할까?」 `verify_reconcile.go:260`) ·
      `mutating: "true"` 유지 · 부재 모드 Q6 문구(D8)
      — RED 2026-10-22: `TestVerifyReconcileExpiredIsAFlagOfTheSameMutatingCommand` · `TestReconcileApprovalQuestionNamesTheMode` · 승인 화면 · Q6 문구 시험 갱신 · 전 경로 seams 시험 — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)
- [x] 2.8 report/status 의 「reconciled expired」 구별 표시
      — RED 2026-10-22: `TestExpiryReconcileTextLabelsReconciledExpired` — 단언에서 실패 확인(`analysis/red-green/red-2026-10-22.md`)

## 3. GREEN · REFACTOR

- [ ] 3.1 새 파일 `internal/official/reconcile_detail_reads.go`(by-id 대사 읽기, 엄격 해독) · `ReconcileConditionalRow` 에 `FirstStatus` 한 필드와
      `DecodeReconcileConditionalPage` 대입 한 줄(P2-3 — 부재 모드는 읽지 않음) · cmd `reconcileReads` 의 by-id 메서드는 `ensureIdentity` 뒤에만 위임(P2-4)
- [ ] 3.2 새 파일 `internal/verifylive/reconcile_expired.go`(모드 판정·진술 수집·A3~A7) + 공통 보조 함수 추출(D4 — 부재 모드 순서 불변)
- [ ] 3.3 `record.go`·`reconcile_record.go`·`projection_reconciled.go`·`report.go`·`reconcile.go`·`reconcile_check.go`·`cmd/tossctl/verify_reconcile.go` 최소 편집
- [ ] 3.4 FLM/BTM 재추출 · 분기 수 대조 · `check_analysis.py --change …`

## 4. 변이 목록(각각 시험이 잡아야 함 — 실행 기록은 `analysis/mutation/`)

- [ ] 4.1 by-id 404 를 statement 로 취급 → 2.3 `expiry-evidence-absent` 가 잡음
- [ ] 4.2 A3 「by-id 두 응답 모두 statement」 요구 삭제(silent ×2 + 빈 진술 집합 통과)
- [ ] 4.3 A4 `EXPIRED` 비교를 `COMPLETED` 허용으로 넓힘 / triggeredOrderId 검사 삭제 / second 검사 삭제
- [ ] 4.4 by-id 둘째 읽기 삭제 또는 두 응답 비교 삭제
- [ ] 4.5 A5 세 판정 각각 삭제(만료 모드에서만)
- [ ] 4.6 A7 체인 검사 삭제
- [ ] 4.7 `--expired` 에서 Q1 을 다시 부름(통과 경로가 죽어야 함) / 플래그 없음에서 Q1 을 건너뜀(부재 모드가 열려야 함 — 기존 시험이 잡음)
- [ ] 4.8 `terminal()` 에서 `ReconciledExpired` 삭제 → 2.5 HOLD·정리 시험이 잡음
- [ ] 4.9 by-id 의 id·심볼·시장 일치 검사 각각 삭제
- [ ] 4.10 만료 어휘를 비어 있지 않은 기본값으로 바꿈(영수증 없이 통과)
- [ ] 4.12 silent+CLOSED 를 수락으로 되돌림(P1-1) → 2.3 `expiry-sources-disagree`
- [ ] 4.13 by-id·CLOSED 행 `first.status` 어휘 검사 각각 삭제(P2-3·P2-5 ①) / `type == SINGLE` 검사 삭제(P3-3)
- [ ] 4.14 silent 분류를 「404 면 전부」로 넓힘(본문 code 비교 삭제) · 본문 해독 불능 404 를 silent 로(P2-5 ②)
- [ ] 4.15 429·5xx·`ErrAuth`·`ErrIPNotAllowed` 를 silent 로 접음(P2-5 ③)
- [ ] 4.16 두 by-id statement 비교의 필드별 삭제 — status · first.status · triggeredOrderId · second 각각(P2-5 ④)
- [ ] 4.17 만료 모드 읽기 집합에서 일반 OPEN 읽기 생략(판정이 아니라 읽기 — P2-5 ⑤)
- [ ] 4.18 만료 모드에서 A6 CLOSED 어휘 검사 자체 삭제(P2-5 ⑥)
- [ ] 4.19 `ReconciledArtifacts` 필터(`projection_reconciled.go:26`)에 넷째 종결 미반영(P2-5 ⑦)
- [ ] 4.20 by-id ① 을 승인 앞으로(P2-1) · by-id 메서드의 `ensureIdentity` 생략(P2-4) · 창 시작을 첫 목록 읽기 직전으로 되돌림(P3-1)
- [ ] 4.21 줄에 by-id Call 을 실음(P2-2 — `SucceededEndpoints` 불변 시험이 잡아야 함)
- [ ] 4.11 무변이 대조군 GREEN · 양성 대조군(변이 켜면 RED) 확인

## 5. 리뷰 · 영향 범위 검증

- [ ] 5.1 gstack review + 적대적 Eng(codex 포함) — 필수 질문 0.8 과 같음. 처분을 `review.md` 에
- [ ] 5.2 관련 시험: `go test ./internal/verifylive/ ./internal/official/ ./cmd/tossctl/ -run 'Reconcile|Outstanding|Cleanup|Hold|Report|M0'`
      + `make test-seams` 중 대사 seam · `go vet` · `make lint`. 선택 이유·미실행 범위를 `review.md` 에(전체 시험은 사용자 명시 요청 때만)
- [ ] 5.3 `make sdd-sync` → fingerprint 신선 확인

## 6. 사람 실행 — 실기록 대사

- [ ] 6.1 배포(이미지 재빌드 + 교체, 사람 실행 — `make image CHANGE=…`) 또는 같은 커밋의 로컬 바이너리
- [ ] 6.2 **사람이** `tossctl --config-dir <엔진 프로필> verify reconcile --market KR --expired` 실행, 화면의 두 마스크·계좌 수 확인 뒤 y.
      에이전트는 실행하지 않는다(`mutating: true`)
- [ ] 6.3 에이전트 읽기 전용 확인(사람 승인 뒤): `tossctl verify status --market KR` 의 outstanding 0 · reconciled expired 1 · 기록 줄 수 +1 ·
      이전 줄 바이트 불변. 결과를 `analysis/receipts/` 에

## 7. 확정

- [ ] 7.1 `make gate CHANGE=a164-verify-reconcile-closes-broker-expired-conditional` — 사용자 명시 요청 때만
- [ ] 7.2 archive → Story 경로 · `generate_master_tracker.py` · `--check`
- [ ] 7.3 Hindsight `tossos-dev-episodes` 에 검증된 교훈 후보만 retain(예: 「by-id ok 는 생존이 아니다」)
