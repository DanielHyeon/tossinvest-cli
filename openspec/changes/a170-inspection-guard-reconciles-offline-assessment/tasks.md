# a170 tasks

Story: `STORY-TOS-a170` · **Small**(시험 전용) — Function Logic Map: not-applicable(생산 함수 편집 0). 시험 + lint + 경량 리뷰.

## 0. 계약

- [x] 0.1 원인 조사 · 판정(proposal Why) · Story 등록(FEAT-TOS-013) · base 고정

## 1. RED → GREEN

- [x] 1.1 RED — 새 시험(허용 파일 통과 · 다른 경로 거절 · `offlinecontinuation` import 거절)과 기존 실패 영수증 `analysis/red/red.txt`
- [x] 1.2 GREEN — `authority_guard_test.go` 허용 목록 + `offlinecontinuation` import 거절, `analysis/red/green.txt`
- [x] 1.3 변이 — 새 규칙을 끄면 시험이 잡는지(`analysis/red/mutation.txt`): M1~M5 CAUGHT(M2 는 1 판 생존 → 형제 파일 시험 추가 뒤 CAUGHT), 무변이 대조군 GREEN

## 2. 검증 · 리뷰 · 종결

- [x] 2.1 `go test ./internal/campaignstateevidence/ ./internal/offlinecontinuation/` · `make test`(이 change 에 승인됨) · `make lint` — rc 0 · 0 · 0, ok 115(`analysis/red/verify.txt` · `green.txt`)
- [x] 2.2 경량 리뷰 → `review.md`
- [ ] 2.3 sdd-sync → gate → archive → Story 경로 갱신 — 완료 gate 실행은 이 change 에 아직 승인되지 않음(2026-10-23 승인 범위 = make test). 사람 승인 대기
