# a170 tasks

Story: `STORY-TOS-a170` · **Small**(시험 전용, 생산 함수 편집 0). 게이트가 요구하는 시험 파일 함수 FLM 번들 둘(1.4). 시험 + lint + 경량 리뷰.

## 0. 계약

- [x] 0.1 원인 조사 · 판정(proposal Why) · Story 등록(FEAT-TOS-013) · base 고정

## 1. RED → GREEN

- [x] 1.1 RED — 새 시험(허용 파일 통과 · 다른 경로 거절 · `offlinecontinuation` import 거절)과 기존 실패 영수증 `analysis/red/red.txt`
- [x] 1.2 GREEN — `authority_guard_test.go` 허용 목록 + `offlinecontinuation` import 거절, `analysis/red/green.txt`
- [x] 1.3 변이 — 새 규칙을 끄면 시험이 잡는지(`analysis/red/mutation.txt`): M1~M5 CAUGHT(M2 는 1 판 생존 → 형제 파일 시험 추가 뒤 CAUGHT), 무변이 대조군 GREEN
- [x] 1.4 FLM/BTM 번들 — 게이트는 시험 파일의 기존 함수도 수정 집합으로 센다(`allowedImport` revision current · `TestGuardRejectsRuntimeCaptureUse` revision base, 2026-10-23 실측)

## 2. 검증 · 리뷰 · 종결

- [x] 2.1 `go test ./internal/campaignstateevidence/ ./internal/offlinecontinuation/` · `make test`(이 change 에 승인됨) · `make lint` — rc 0 · 0 · 0, ok 115(`analysis/red/verify.txt` · `green.txt`)
- [x] 2.2 경량 리뷰 → `review.md`
- [x] 2.3 sdd-sync → gate → archive → Story 경로 갱신 — 2026-10-23 사용자 승인(Manager 전달): 완료 gate → archive, base 창 형제 착지 시 base 재고정(review §4)
  - 종결(2026-10-28, 사용자 승인 "1" — base 재고정 금지): 재고정 뒤 창에 든 형제 함수 26 을 a172 아카이브 번들 복사로 덮음(review §5), gate ⑤ 사전 실행 rc 0. 완료 게이트 결과는 review.md 「완료 게이트 실행 영수증」.
