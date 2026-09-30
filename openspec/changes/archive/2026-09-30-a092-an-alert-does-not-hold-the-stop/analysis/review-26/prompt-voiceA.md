(prompt-common.md 전문을 먼저 읽는다 — 보이스 A 초점.)

## 보이스 A — 동시성 · 잠금 · 원칙 E · 새 실행자

- 좁힌 `n.mu` · 배달 실행자 · 기록 전용 입구 · `Acknowledge` · `mode-release`(TransitionOperatingMode → 투영 → 기록 전용 통지) · `NormalRelay` 사이의 모든 끼어듦에서 (a) 이중 발송 (b) 게이트 오개방 (c) 영구 잠김 (d) 교착 (e) 모드 투영 역행을 판정하라. 잠금 순서(`n.mu` · `g.mu` · journal 트랜잭션 · `j.modeMu`)를 그려라.
- `NormalRelay.Run` 의 종료 · 패닉 · 가득 참 경로와 런타임 배수 순서(`runtime.go` Run 의 `wg.Wait` 와 journal close 순서)를 확인하라.
- 필요하면 사본에서 `go test -race` 를 돌려라.
