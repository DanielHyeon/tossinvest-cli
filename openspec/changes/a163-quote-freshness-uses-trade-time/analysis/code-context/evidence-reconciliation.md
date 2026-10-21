# a163 증거 조정 (SDD 5단계)

| 질문 | CodeGraph 1.6.0 | CGC | AST / HEAD 원문 | 결론 |
| --- | --- | --- | --- | --- |
| `adaptPrices` 의 호출자 | `Client.Prices` 하나 | 다른 사본만 | `official--client.prices.ast.json` `:149` | 일치 — 어댑터 편집의 생산 도달은 `Prices` 경유뿐 |
| `adaptPrices` 가 브로커 시각을 읽나 | callees `parseDecimal` 만 | — | AST 호출 텍스트에 `Timestamp` 없음, `:175 time.Now` | 일치 — 읽지 않는다 |
| `quoteUsable` 호출자 | 6 | 줄 번호 다른 사본 | 6 자리 모두 AST 분기 좌표로 확인(`quote-time-evidence.md` C3) | 일치 |
| `observe` 호출자·피호출자 | 진짜 1·5 + 거짓 엣지 3·4 | — | AST `calls` 에 거짓 엣지 대상 없음 | 거짓 엣지 제외 후 일치 |
| `stableObservationID` 가 `FetchedAt` 을 쓰나 | 호출자 `snapshotContext` | — | AST `:1208` `FetchedAt.UTC` · B2 `:1209` | 일치 — D2 의 근거 |

불일치 0(거짓 엣지는 도구 한계로 분류, 편집 차단 사유 아님). CGC 는 이 체크아웃을 보지 못했다(advisory 미반영).
편집 착수 시(tasks 1.x) `make sdd-sync` 뒤 같은 질의를 다시 돌려 이 표를 갱신한다.
