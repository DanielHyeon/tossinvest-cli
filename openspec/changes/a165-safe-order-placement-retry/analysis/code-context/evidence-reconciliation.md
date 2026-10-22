# a165 evidence reconciliation (0.3)

| 질문 | CodeGraph 1.6.0 | CodeGraphContext | 현재 HEAD(소스·AST) | 결론 |
|---|---|---|---|---|
| 주문 mutation 의 브로커 호출 자리 수 | submit 이 classifyMutation 의 유일한 caller | 쓸 수 없음(다른 체크아웃) | AST `gateway.submit`: for/range/select/go 0, `plan.call` 708 한 자리, `call` 은 B44(712)/B45(717) 배타 | attempt 당 1회. 재시도는 새 구조를 넣어야 생긴다 |
| strategy 경로의 호출 횟수 | withStrategyEntryGateAuthority caller=submit 1 | — | AST: fn 은 반환문 80 한 번, 63 에서 `defer g.entry.mu.Unlock()` | 1회, EntryGate 잠금 안 |
| 401 재전송 | send callers 3(verb 공유) | — | AST `client.send` B4 for(344) 조건 `attempt<2 && code==401`, B8 `!adopted` break | 401 일 때만, 최대 2회. 본문 바이트는 postWithHeaders 가 루프 밖에서 1회 인코딩 |
| 429·5xx·전송 실패의 분류 | — | — | AST `classifymutation` B8 → `statusOf` B4/B5 → `ClassifyHTTPMutation` B7 default; 전송 오류는 98 → B1/B2 | 429·5xx = AMBIGUOUS, 전송 전 실패 = NOT_SENT, 전송 후 = AMBIGUOUS |
| 409 already-processing(취소) 분류 | — | — | `classifyRefusalCode` 목록에 없음 → B8 → 409 → B7 | AMBIGUOUS(IN_DOUBT). 엔진 경로에 재시도 없음 |
| Amend 생산 호출자 | ops.modifyHandler·시험(modifyHandler 는 이름 충돌 — trading.Service.Amend 를 부름, 1.3 S11) | — | grep 0(엔진) | 엔진 경로에서 409 정정 재시도는 닿지 않음(proposal 의 「CodeGraph 로 재확인」 해소) |
| Cancel 생산 호출자 | (CodeGraph Cancel 은 trading.Service.Cancel 등과 섞임) | — | grep: exitloop.go:1686 · flatten.go:382 | 둘 다 노출 감소 경로 — design D7 |

불일치: CodeGraphContext 만 다른 체크아웃을 가리킨다(편집 차단 사유 아님 — advisory). CodeGraph 와 HEAD 사이 불일치 0.
