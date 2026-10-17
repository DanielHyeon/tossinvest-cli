# a091 구현 리뷰 i1 — Claude 보이스 A(적대 · §0.3 · 게이트) 원문 요지

- 2026-10-01, Opus, 독립 컨텍스트, 스냅숏 `/tmp/claude-1000/a091-i1-tree`(= `97a6f717`) read-only, 증명은 자기 스크래치 사본.

VERDICT: REJECT — 안전 불변식 위반은 없음(§0.3 · §0.9 반환값 B1~B7 편집 전 AST 와 같음, 패닉 경로 없음, 게이트는 B2 · B7 공유 술어, 생산 배선 덮기 단일 생성자).

1. P1 — 알림 꺼짐 B2 의 로그 SHALL 이 생산에서 성립하지 않음: `logZeroFloor` 는 `o.opts.Log` 에 쓰는데 생산은 그것을 넣지 않음(`cmd/tossctl/engine.go:665-673` 은
   `UnobservedLog` 만, `exitloop.go:191` 「Log 는 생산에서 nil」). 생산 모양 탐침(`o.Log=nil`)에서 꺼짐 B2 의 로그 줄 0. 하네스가 `o.Log = logger` 를 넣어서 시험이 통과.
2. P2 — 관측마다의 원인 로그 줄을 포지션에 이을 수 없음: 생산의 유일한 줄은 `RecordOnly` 의 `logEvent(withoutFields(e))` — 본문에 종목 · 포지션 · 키 없음.
3. P2 — 원인 경계 변이 생존: Holdings 판정에 `|| LocalSells`(A), `|| Stale || NoSnapshot`(A2) — 둘 다 SURVIVED. 한정 항 다섯 표 시험 필요.
4. P3 — 익절 0주 본문이 「손절이 나가지 않았다」(거짓).
5. P3 — 생존 변이: 부재/낡음 문구 맞바꿈 · critical B2 가린 줄 생략 · 기록 실패 줄 생략 · payload `cause`/`position_id`/`floor_bound` 제거 · 종료 문구(죽은 코드) ·
   `Unwrap()` nil 을 참으로.
6. P3 — 벽시계 수락 시험의 CI 불안정 가능성(연결 풀 칸).
확인된 참: H2 한 종류(critical) · 계좌 필드 없음 · `escalate` 두 줄 · 한국어 `이름(코드)` · `cancellationOnly` 모양들 · 종료 순서(루프 합류 뒤 원장 닫힘).
