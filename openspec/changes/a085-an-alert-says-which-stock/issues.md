# a085 · Issues

## I1 — 콘솔 절반은 a086이다

`operatorview.BuildExitLine`이 `snapshot_quarantined`에서 저장값을 버리는 문제는
a077의 승인된 요구사항과 충돌하므로 이 change에서 뺐다. `proposal.md`의 "분리 결정"
참조. a086은 `operator-console`의 해당 요구사항을 MODIFIED로 고치고 그 수정분에 대한
리뷰를 받아야 한다.

## I2 — 차단 evidence 문자열이 첫 관측에 고정된다 (a083 I4와 동일 건)

포지션 화면의 "the engine believes 2 of 103590, the account says 3"은 차단이 처음
기록된 시각의 값이고, 같은 화면 아래의 원장 수량 4와 다르다. `EnterReconcile`이
멱등이라 첫 evidence를 보존하기 때문이다.

a085는 알림 문구만 다루므로 범위 밖이다. a086이 화면 표현을 다룰 때 함께 판단한다.

## I3 — 알림 본문의 원문 에러는 여전히 영어다

한국어 설명 뒤에 `원인: <원문>`으로 붙인다. 의도된 것이다 — 운영자가 그대로 검색하거나
붙여넣을 수 있어야 하고, 브로커·Go 에러 문자열을 번역하면 원본과 대조할 수 없다.

## I4 — `make gate`가 병행 세션의 편집으로 막혀 있다 (a084 I8과 동일 건)

`.claude/CLAUDE.md`와 `.codex/agents.md`를 다른 세션이 편집 중이라 agent-config sync가
drift를 보고한다. 다른 세션의 진행 중 안전 부트스트랩 편집을 덮어쓰지 않았다.

a085 자체의 증거는 전부 통과했다.

```
go test ./...   8016 passed, 0 failed, 99 packages
go vet ./...    clean
openspec        valid --strict
logic-map       evidence complete (24 함수)
```
