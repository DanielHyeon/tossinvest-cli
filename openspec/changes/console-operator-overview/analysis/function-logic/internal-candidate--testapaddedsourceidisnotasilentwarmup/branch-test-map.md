# Branch Test Map: `TestAPaddedSourceIdIsNotASilentWarmUp`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 정규 철자로 관측을 기록한다 | 이 테스트 | — (동작 무변경) | yes |
| B2 | 앞뒤 공백·대문자 세 철자를 모두 시도한다 | 이 테스트 | — (동작 무변경) | yes |
| B3 | 조회가 에러 없이 돌아야 한다 | 이 테스트 | — (동작 무변경) | yes |
| B4 | 0행 + 에러 없음은 배선 결함을 WARMING_UP으로 위장한다 | 이 테스트 | 기록됨(§3 P2 — `SourceObservations`가 source를 trim하지 않던 결함) | yes |
| B5 | 패딩된 id로도 쓸 수 있어야 한다 | 이 테스트 | — (동작 무변경) | yes |
| B6 | 정규 철자 조회가 에러 없이 | 이 테스트 | — (동작 무변경) | yes |
| B7 | 쓰기 쪽 정규화가 읽기 쪽과 같다 — 아니면 아무 읽기도 찾지 못할 철자로 저장된다 | 이 테스트 | 기록됨(§3 P2) | yes |
| B8 | 빈 source id는 침묵이 아니라 에러 | 이 테스트 | 기록됨(§1 7절 — 조용한 성공이 실패보다 더 완전히 감춘다) | yes |
| B9 | `SourceSeries`도 같은 규칙 | 이 테스트 | 기록됨(§3 P2) | yes |
