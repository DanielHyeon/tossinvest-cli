# a168 tasks

> 정의 2026-10-22, 착수 2026-11-25(base `e57d06cd6`, 브랜치 `feat/a168-verify-read-decode-hardening`). 출처: a164 `review.md` 「5.1」 C-P1 잔여 · C-P2a ·
> C-P2b · C-P2c 와 「Manager 처분」.

## 0. 착수 준비

- [x] 0.1 proposal 열린 질문 Q1~Q3 에 사람 답을 받고 proposal 에 기록한다.
- [x] 0.2 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a168-verify-read-decode-hardening` 로 base 고정.
- [x] 0.3 `make sdd-sync` 후 CodeGraph hard evidence(`DecodeReconcileConditionalPage` · `DecodeReconcileOrderPage` ·
  `decodeReconcileEnvelope` · `unwrapAndDecode` · `reconcileObject` · Q1 이 (b) 면 `Client.send`)의 callers/callees/impact 와
  FLM/BTM — 해독 판정은 게이트 판정 경로라 면제 없음. proposal 의 줄 번호를 AST 열거로 재확인.
  - 진행(2026-11-25): sdd-sync(CodeGraph 신선·CGC incomplete·GBrain 시간 초과) · CodeGraph 1.6.0 질의 · AST 34개 · 측정 ·
    정상 입력 census 완료(`analysis/`). 편집 대상 기존 함수 4개 FLM/BTM 번들 — 편집 전 커밋 `3bfe3c41c`, 편집 뒤 갱신(`check_analysis.py` rc=0, required 4).

## 1. 계약 (design freeze)

- [x] 1.1 design.md — 중복 키·뒤따르는 데이터 거절 규칙(층별: 봉투 · 목록 행 · 중첩 `first`/`second`), 재사용할 해독기, 거절 코드,
  성공 상태 계약(Q1 선택), 부재 모드 거절 순서(a121 「가드 순서 계약」) 불변 여부.
  - 진행: design.md 작성·리뷰 반영 완료. 동결 2026-11-25 — Manager 결정 1~4 확정, 4 = (나) → D6.
- [x] 1.2 spec delta 확정(현재 초안: `specs/order-execution/spec.md`).
  - 진행: 두 Requirement 로 다시 씀, `openspec validate --strict` rc=0. D6 Requirement 추가 뒤 strict rc=0 — 확정.
- [x] 1.3 proposal-freeze 리뷰(적대적 Eng 보이스 포함) → `review.md`.

## 2. 구현 (RED → GREEN → REFACTOR → VERIFY)

- [x] 2.1 Pre-Edit 선언(review.md 「2.1 Pre-Edit Gate」).
- [x] 2.2 RED: design §8 표(R1~R10 · 대조 C1~C6 · R-T). 정정(design §1 P5): CLOSED 다른 id 행의 발동 흔적 숨김은 **만료 모드**
  (`EXPIRED` 행, R8a), 부재 모드는 행 통째 숨김(R8b). 뒤따르는 데이터는 이미 거절 — 특성 고정(R-T)이지 RED 아님.
  - 실측(편집 전, `analysis/red-green/red-*-before.txt`): official 하위 시험 54 실패(모호한 본문 30 · 200 아닌 2xx 20 · 401→201 4),
    verifylive 10 실패 — 그중 9 건 「줄 씀」(R8a · R8b ×2 · R8c · R9 · R11 ×3 · R12 lone `Result`), 1 건 코드 불일치(R12 행 단독 변형 →
    편집 전 `closed-fired-trace`). 처음부터 GREEN 인 대조·특성(C1 · C4 · C5 · R-T · 입력 검사 · 204)은 RED 로 세지 않음.
- [x] 2.3 GREEN 최소 구현(새 파일 `internal/official/reconcile_response.go` + 리더 4 개 편집, `send`·`unwrapAndDecode`·`reconcileObject` 불변). 원문 `analysis/red-green/green-*-after.txt`.

## 3. 검증

- [ ] 3.1 변이 원장 — 각 거절 가드 · 양성 대조군 포함.
- [ ] 3.2 적대 리뷰 + gstack review.
- [ ] 3.3 영향 범위 시험(`internal/official` · `internal/verifylive` · `cmd/tossctl`)과 미실행 범위 기록.
- [ ] 3.4 Manager 독립 검증.

## 4. 닫기

- [ ] 4.1 완료 게이트(사용자 명시 요청 시) → archive → Story 경로·PM sync.
- [ ] 4.2 검증된 교훈만 Hindsight retain.
