너는 TossOS(Go 자동매매, 실제 돈) 저장소의 **적대적 설계 리뷰어**다(2회차, 마지막). 읽기 전용 — 파일을 쓰거나 고치지 말고, 네트워크 · `tossctl` · `~/.config/tossctl` 접근 · 시험 실행을 하지 마라.
작업 디렉터리는 저장소 루트다.

change: `openspec/changes/a199-a-stale-capability-proof-blocks-entry-not-exit/`. 1회차(`analysis/review/proposal-freeze-r1-codex.md`)의 발견 R1~R6 에 대해 Manager 가 결정 12건을 내렸고
그것이 `proposal.md` · `design.md` · `specs/engine-safety/spec.md`(delta) · `review.md` 「Manager 결정」 표 · `analysis/edit-surface.md` · `analysis/claim-check.md`(B2 절)에 반영됐다. 먼저 그것들을 읽어라.
현재 코드: `internal/app/engine/interlock.go`, `internal/app/engine/engine.go`(NewContext), `internal/app/engine/attestation_watch.go`, `internal/attest/attest.go`, `internal/soak/attest.go`(BuildAttestation),
`internal/execgw/retry.go`, `internal/execgw/gateway.go`, `internal/app/engine/alertdelivery.go`, `internal/app/engine/runtime.go`, `cmd/tossctl/engine.go`. 정본: `openspec/specs/engine-safety/spec.md`.

이번 초점(이것만):
1. **결정 3(빈 endpoint 항목 · 형식 번호 < 1 은 청산 전용 아님)과 결정 7(「청산 전용으로 돌고 있다」 알림의 조건을 파일이 아니라 기동 표지에 — 파일 회복이면 일반 등급 「재시작하면 정상으로 뜬다」)이 새로 연 구멍.**
   예: 결정 7 알림이 a196 의 키 · 에피소드 · 「알림 꺼짐이면 critical 도 일반」 규칙 · 알림 전달 실패 → 진입 래치/모드 승격 규칙과 부딪히는가, 하루 한 번이 깨지는가, 기동 표지가 정상인 엔진에서 새 알림이 새는가.
2. **닫힌 집합의 경계** — 아직 청산 전용으로 잘못 내려가는 attestation(형식 · 필드 · 시각 경계 · 계좌 표기), 또는 잘못 거부되는 정상 발급 파일. 발급자 인용(claim-check B2)이 맞는가.
3. **감시 문장이 거짓이 되는 상태가 남았는가** — 단계(healthy · notice · imminent · expired · scope_lost · account_mismatch · unreadable · incomplete · proof_overdue) × 기동 표지(정상 · 청산 전용) 조합마다 delta 와 design 이 말하는 문장이 사실인가.

출력(한국어): 발견마다 `[P0|P1|P2|P3] 제목`, 근거(파일:줄 · 입력 · 순서), 영향, 제안. P0 = 진입이 열리거나 손절이 막힘, P1 = 정본 위반 · 안전 논증의 구멍, P2 = 빠진 소비자 · 시험 공백, P3 = 문구. 각 발견에 「설계 결정을 바꿔야 하나(예/아니오)」를 적어라. 마지막에 「확인했지만 문제없음」 짧게. 추측은 추측이라고 적어라.
