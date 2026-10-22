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
- [ ] 0.8 **proposal-freeze 리뷰** — 적대적 Eng 보이스 포함(High-risk 인접), 결과 `review.md`.
      필수 질문: 「이 diff 의 어느 줄이 살아 있는 조건주문을 닫힘으로 적을 수 있나」
- [ ] 0.9 편집 착수 직전 `make sdd-sync` → 0.5 질의 재실행 · `evidence-reconciliation.md` 갱신(fingerprint 신선)

## 1. 함수 내부 증거 (편집 전, 면제 없음)

- [ ] 1.1 FLM/BTM 번들 — design 「FLM/BTM 대상」의 기존 함수 전부(`scaffold_analysis.py` → `ast.json` · `function-logic-map.md` ·
      `branch-test-map.md` · `risk-pattern-report.md`). `verifyReconcileApprove` 가 logic-map 에 안 잡히면 `Function Logic Map: not-applicable`
      사유 + 행동 시험
- [ ] 1.2 부재 모드 불변 고정 — 플래그 없는 실행의 거절 코드·순서를 기존 시험(`reconcile_g1/g2/g3/codex/structure_test.go`)으로 확인하고,
      빠진 갈래가 있으면 편집 **전에** 시험을 더해 GREEN 으로 둔다
- [ ] 1.3 `official` 클라이언트가 성공 응답 requestId 를 노출하는지 코드로 확정(design D5 R3) — 없으면 응답 `Digest` 로 대신함을 design 에 적음
- [ ] 1.4 Pre-Edit 선언(`review.md`)

## 2. RED

- [ ] 2.1 by-id 대사 읽기 해독: statement · silent(404 `conditional-order-not-found`) · 거절(404 다른 code·해독 불능 본문 · 400 · 401/403 · 429 · 5xx ·
      전송 · null result · required 필드 결측/null/형 불일치 · `first` 결측 · `triggeredOrderId`/`second` 형 불일치) + 수락: `triggeredOrderId`·`second`
      키 없음과 null 이 같게 「없음」(design D3 정정) — httptest. 원 바이트 골든은 OpenAPI 스키마 + 0.6 영수증 값으로 만든다(영수증은 CLI 출력, C1)
- [ ] 2.2 증거 규칙 수락 두 갈래: by-id 만 · CLOSED 만(by-id silent) · 둘 다
- [ ] 2.3 증거 규칙 거절 — design D2 거부 1~5 와 D3 표의 행마다 하나, **가드별 거절 코드 단언**(공유 문구 아님):
      `expiry-evidence-absent` · `expiry-target-live` · `expiry-target-fired` · `expiry-target-status` · `expiry-read-error` · `expiry-reads-differ` ·
      `expiry-chain-open` · `open-conditional` · `open-plain-order` · `closed-fired-trace` · `closed-status`(어휘 비어 있음) · by-id id/심볼/시장 불일치
- [ ] 2.4 순서: 승인 전 거절(후보·계좌·Q3·어휘) · 승인 뒤 읽기 · `--expired` 에서 Q1 미호출 · 플래그 없음에서 Q1 호출(기존)
- [ ] 2.5 기록: 줄 하나 추가 · 기존 바이트 불변 · `terminal()` 넷째 종결 · `Outstanding`/`PendingCleanup`/HOLD 두 곳(cmd·`New`)이 따름 ·
      구 바이너리 시뮬레이션(필드 버린 해독)에서 다시 outstanding · 계좌 원문·응답 원문 없음
- [ ] 2.6 봉인: 대사 리더 메서드 집합이 GET 넷(목록 2 · 종목 · by-id)뿐 — `reconcile_seal_test.go` 세 층 갱신
- [ ] 2.7 cmd: `--expired` 플래그 · 승인 화면/질문의 모드 표시 · `mutating: "true"` 유지 · 부재 모드 Q6 문구(D8)
- [ ] 2.8 report/status 의 「reconciled expired」 구별 표시

## 3. GREEN · REFACTOR

- [ ] 3.1 새 파일 `internal/official/reconcile_detail_reads.go`(by-id 대사 읽기, 엄격 해독)
- [ ] 3.2 새 파일 `internal/verifylive/reconcile_expired.go`(모드 판정·진술 수집·A3~A7) + 공통 보조 함수 추출(D4 — 부재 모드 순서 불변)
- [ ] 3.3 `record.go`·`reconcile_record.go`·`projection_reconciled.go`·`report.go`·`reconcile.go`·`reconcile_check.go`·`cmd/tossctl/verify_reconcile.go` 최소 편집
- [ ] 3.4 FLM/BTM 재추출 · 분기 수 대조 · `check_analysis.py --change …`

## 4. 변이 목록(각각 시험이 잡아야 함 — 실행 기록은 `analysis/mutation/`)

- [ ] 4.1 by-id 404 를 statement 로 취급 → 2.3 `expiry-evidence-absent` 가 잡음
- [ ] 4.2 A3 `|S| ≥ 1` 삭제(빈 진술 집합 통과)
- [ ] 4.3 A4 `EXPIRED` 비교를 `COMPLETED` 허용으로 넓힘 / triggeredOrderId 검사 삭제 / second 검사 삭제
- [ ] 4.4 by-id 둘째 읽기 삭제 또는 두 응답 비교 삭제
- [ ] 4.5 A5 세 판정 각각 삭제(만료 모드에서만)
- [ ] 4.6 A7 체인 검사 삭제
- [ ] 4.7 `--expired` 에서 Q1 을 다시 부름(통과 경로가 죽어야 함) / 플래그 없음에서 Q1 을 건너뜀(부재 모드가 열려야 함 — 기존 시험이 잡음)
- [ ] 4.8 `terminal()` 에서 `ReconciledExpired` 삭제 → 2.5 HOLD·정리 시험이 잡음
- [ ] 4.9 by-id 의 id·심볼·시장 일치 검사 각각 삭제
- [ ] 4.10 만료 어휘를 비어 있지 않은 기본값으로 바꿈(영수증 없이 통과)
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
