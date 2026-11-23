# a201 Issues

## I1 — tasks 1.13 이후 [blocked]: 「세우지 않는다」 쪽 설계 둘이 리뷰 반례로 기각 (2026-11-24)

- **분류**: blocking(설계 결정이 실제 반례로 기각되어 freeze 불가) — Manager 결정 2026-11-24 「a201 을 보류한다」.
- **막힌 이유**: 낡은 비교가 차단을 **세우지 않게** 하는 설계 둘이 각각 codex 동결 리뷰의 결정급 P1 로 기각됐다.
  - v1(낡은 불일치를 비교에서 뺀다 · 다시 읽은 로컬 대 옛 스냅숏): r1 R1-1 · R1-2(`analysis/review/proposal-freeze-r1-codex.md`).
  - v2(체결이 지나간 주기 전체를 보류): r2 R2-1 · R2-2(`analysis/review/proposal-freeze-r2-codex.md`).
  - 공통 뿌리: 대사 추적기(`Tracker.Observe`)는 생략된 관측이 남겼어야 할 증거(a192 결속 · 연속 횟수 · 크레딧 반박 · 건강 계수)에 기댄다. 세우지 않는 쪽의 수리는 종목 단위든 주기 단위든 그 증거를 버린다.
- **재개 방향(후보 — 설계 아님)**: 해제 쪽에서 닫는다(design §6).
- **재개 조건**(하나):
  1. a194 의 「차단이 10분 넘게 남았다」 알림이 이 모양(체결이 로컬 읽기를 앞지른 주기에 선 수량 불일치)으로 한 번이라도 온다.
  2. 사용자가 재개를 지시한다.
- **재개 때**: 증거 조사부터 다시 한다(design §6 의 질문). codex 회차는 새로 센다(이 change 의 2회는 소진 — 설계가 바뀌므로 새 동결 리뷰).
- **보존**: 1단계 산출물(`analysis/`)과 재현 시험 `internal/app/engine/a201_stale_quantity_comparison_repro_test.go`(「현재 동작」 6)는 재개 때의 RED 재료다. 이 브랜치는 main 에 합치지 않는다. 재개 때 base 이후 main 이 `RunOnce` 등을 고쳤으면 AST · FLM 을 다시 뽑는다.
- **지금의 운영 안전망**: 운영 이력에 이 고착의 흔적은 0(`analysis/ops-aggregates.md`). 선 차단은 a194 가 알리고 운영자 해제(`reconcile-resolve`)로 풀린다.
