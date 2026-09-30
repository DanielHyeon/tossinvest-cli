# 보이스 A (code-reviewer, Claude) — 25라운드 원문 요약 보존 (2026-09-29, 트리 git archive 22db26e7)

에이전트가 보고를 넘긴 직후 주간 한도(429)로 종료됨 — 보고 본문은 전달됨.

**판정: APPROVE** — P0 없음, 이번 변경이 만든 안전 불변식 위반 없음. P1 1건은 기존 동작·선언된 잔여.

실행(사본 `/tmp/claude-1000/a092-r25-A-1568849`, `git rev-parse` 실패 확인 뒤, 끝나고 삭제):
- `go test -race -count=1 ./internal/obs/` ok (255.5s), race 0.
- `go test -race -run 'A092|RecordAlert' ./internal/journal/` ok. `./internal/journal/` 전체 `-race` 는 600s 기본 한도 초과(기존 성질).
- `go test -race -run 'A092|A066|RiskRelaxation|Relax' ./internal/app/engine/` ok. `./cmd/tossctl` TestA092 PASS.

| # | 등급 | 파일:줄 | 무엇 | 제안 |
|---|---|---|---|---|
| 1 | P1(기존 · 선언된 잔여 C8) | exitloop.go:1345→:1536 · record_only.go:49-51 | 일반 등급 `EventExitProposalCapped` 가 **축소된 청산 제출 바로 앞**에서 동기 `publishBestEffort`(최대 `DefaultPublishTimeout` 10s), capped 제안마다 반복 | 단위 ⑤ 에서 가장 먼저 — 기록 전용 · 제출 뒤로 · 비동기 실행자. 잔여 문구를 「RECONCILE capped 청산 제출 경로 앞 동기 publish」로 좁힐 것 |
| 2 | P2 | notifier.go:710-715 · a099_round4_test.go | 「죽은 발송자 신호 = logClaimStolen」은 틀림 — 생산에서 만료 임차를 먼저 가져가는 배달 실행자는 탈취를 `EventAlertClaimHeld`(WARN, `stole_from`)로 씀(`alertdelivery.go:305-309`). 같은 이벤트 이름이 알림기 INFO / 실행자 WARN | 실행자 탈취 줄을 `EventAlertClaimStolen` 으로, 또는 문서 · 시험 메시지 정정 |
| 3 | P2(잠재) | notifier.go:807-906 | `Flush` 는 여전히 `n.mu` 를 쥔 채 `Publish` — 생산 호출자 0 | 삭제 · 생산 호출 금지 AST 핀 · 잠금 밖 전송 중 하나(K19) |
| 4 | P2 | record_only.go:88 (25.6) | 완화 통지가 새 구조화 로그 줄(`target` 계좌 ref · operator · approval)을 남김 — 이행 전에는 없던 줄 | 불변식 8 해석 확정, 허용하지 않으면 로그에는 `published` |
| 5 | P2(기존 · 범위 밖) | replay.go:553 | `parkAlert` 의 잠금 밖 삽입이 셈~해제 사이에 끼면 PENDING 1개를 남긴 채 게이트가 열림 — 단 자기 사유(UNRESOLVED)로 먼저 잠김, 전달 실패 전이라 경계 | 겹침 표의 「주장하지 않는 것」에 명시 |

끼어듦 판정(동기×2 · 동기+기록 전용 · 동기+Acknowledge · 실행자+동기): 이중 발송 · 오개방 · 영구 잠김 · 교착 없음. 잠금 순서 `n.mu → g.mu` 하나. `readVerdict` 위치 세 자리 모두 근거 확정 뒤(반납 오류 갈래 포함 — 틀리는 방향은 잠그는 쪽).

저자 주장 1~6: 전부 참(1 k3 합성 성립 — `ConfirmedFloor` 조회 둘 다 `f.retrier.Query`, 다른 announcer 경로 없음 / 2 L19 원장 읽기 / 3 겹침 표 여섯 행 코드와 일치 / 4 / 5 행 바이트 동일 · 부수 효과 변경 / 6 주장 범위 안).

Recommendation: 승인. 단위 ⑤(C8)에서 `EventExitProposalCapped` 동기 publish(#1)를 가장 먼저, #2 탈취 이벤트 이름 정렬.
