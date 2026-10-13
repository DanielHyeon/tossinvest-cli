# a135 Tasks

Story: `STORY-TOS-a135` · 관측 · 알림 경로(주문 · 손절 · 사이징 · 인터록 판정 무변경). 비례 원칙: 변이 원장 · 다중 적대 리뷰 의무 아님.

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a135.yaml`, `_registry.yaml`, `FEAT-TOS-009` 역링크) + tracker 재생성
- [x] 0.2 proposal · design · spec 델타(`engine-safety` 요구 1) · tasks, `openspec validate a135-attestation-expiry-is-announced --strict --no-interactive` rc 0
- [x] 0.3 `capture_change_base.py` 로 base 고정(`421bf53c`)
- [x] 0.4 운영 attestation · 갱신 타이머 실측(읽기 전용 · 메타데이터만) → `analysis/verify/live-attestation.md`
- [x] 0.5 근거로 인용한 기존 함수 분기의 AST 열거(`verifyGate` · `acceptSupervised` · `BuildAttestation`) → `analysis/code-context/`
      — 편집 대상 아님. `EventCredentialExpiring` 생산 발행자 0(비시험 grep), `RequiredEndpoints` 호출자(codegraph 1.6.0 `callers`) 확인

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — `internal/app/engine/a135_attestation_watch_internal_test.go`(판정 · 실행자 부재로 빌드 실패) ·
      `cmd/tossctl` 생산 보조 실행자 집합 기대(`attestation-watch` 없음) → `analysis/red/red.md`
- [x] 1.2 GREEN — 새 `internal/app/engine/attestation_watch.go`(판정 `judgeAttestation` · 실행자 `attestationWatch` ·
      `Context.AttestationWatch`), `internal/obs/event.go` 상수 2(`EventCredentialExpiryNotice` · `EventAttestationWatchFailed`, 등급표 무편집)
- [x] 1.3 배선 — `cmd/tossctl/engine.go` `engineRuntime` 의 `Auxiliary` 에 `ectx.AttestationWatch(clk)` 한 원소 +
      `cmd/tossctl/a098_the_engine_registers_a_sender_test.go` 기대 집합. a134 커밋(`85752cd6`) 뒤에 넣음.
      `engineRuntime` · `TestProductionRuntimeStartsExactlyOneAlertDeliverer` 의 FLM 번들(`analysis/function-logic/`, 분기 변화 0)
- [x] 1.4 경량 변이(무변이 대조군 GREEN 먼저) — 판정 · 실행자 핵심 줄 + 리뷰 지적 자리 → `analysis/mutation/receipt.txt` 11/11 CAUGHT

## 2. 검증

- [x] 2.1 gofmt 빈 출력 · `go vet ./...` · `go vet -tags tossos_testseams`(engine · obs · cmd/tossctl) rc 0
- [x] 2.2 `go test -count=1` — `internal/app/engine` · `internal/obs` · `internal/attest` · `internal/soak` 전부 ok
      (`TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence` 포함 — 「기간 밖 감독 증거로 갱신하면 POST 누락」 실행 고정)
- [x] 2.3 배선 · 리뷰 반영 뒤 `go test -count=1`(engine · obs · cmd/tossctl) + `-tags tossos_testseams`(같은 셋) → `analysis/verify/receipts.md`
- [x] 2.4 구현 리뷰(분리된 패스, 적대적) → `review.md` — APPROVE-WITH-FIXES(P1 1 · P2 2 · P3 7), P1 · P2 전부 수용, P3 5 수용 · 2 기각(사유)
- [ ] 2.5 `make sdd-sync` · `make sdd-check`

## 3. 사람 · Manager

- [ ] 3.1 Manager `make gate CHANGE=a135-attestation-expiry-is-announced` + 독립 검증
- [ ] 3.2 배포(사람) — 이미지 재빌드 + 엔진 재시작, KST 05:00~09:00. D-7 경계 10-16T01:39Z 전에 해야 D-7 통지가 나감
- [ ] 3.3 운영(사람, 이 change 밖) — 10-23T01:39Z 전에 감독하 실주문 검증 재실행(먼저) + soak 판독 완전성 회복 뒤 갱신.
      a063 4.2(거절을 가리지 않는 단위 설치)
