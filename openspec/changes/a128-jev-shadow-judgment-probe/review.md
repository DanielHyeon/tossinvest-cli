# a128 리뷰 기록

## 1.8 (state_rev a128-s3 — WTS 조회 입력) — 경량 리뷰, 2026-10-15

- 범위: `tools/a128-jev-shadow-probe/` 만(새 `wts.go`·`wts_test.go` + 배선). `internal/` 무접촉, High-risk 경로 0, 게이트 판정 변경 0
  → 비례 원칙상 Function Logic Map·변이 원장·다중 적대 리뷰는 `not-applicable`. 시험 + lint + 경량 리뷰 1 회.
- 리뷰어: 독립 code-reviewer 서브에이전트 1(읽기 전용, offline `go vet`·`go test -race` 재실행 — 94 PASS). 판정 **APPROVE-WITH-FIXES**,
  blocker 0.
- 확인된 것(리뷰어): 표면 = 조회 셋(AST 고정) · 피드 범위 계좌 맥락 없음 · `AuthError`/`StatusError` 메시지에 본문·쿠키·헤더 없음 ·
  WTS 계수가 `broker_reads` 와 분리 · 거절 상태는 사이클별(시장 간 공유 없음) · 보관소 `newsMu` 보호 · `created > at` 선행 차단.

| # | 지적 (non-blocker) | 처리 |
|---|---|---|
| 1 | 시간대 없는 createdAt = KST 가정이 틀리면 선행 차단이 그만큼 약해짐 | 부분 수리: 받은 순간 + 5분보다 늦은 기사는 버리고 세며 전부면 결손. 반대 방향(실제 UTC 를 KST 로 읽어 9h 이르게 보임)은 기계적으로 못 가림 — 새는 폭은 `at`~피드 수신 사이(수 초, 피드는 사이클 초반 한 번)에 만들어진 기사뿐. **첫 s3 사이클 로그 `newest raw` 를 벽시계와 대조해 시간대를 확정할 것(사람/Manager).** |
| 2 | 0 건 피드가 news=present·빈 목록 | 수리 — 쓸 수 있는 기사 0 = 결손(`TestEmptyFeedAndEmptyFlowsAreGaps`) |
| 3 | 직전 거래일 수급 0 행이 하루 캐시되어 「0 행」으로 나감 | 수리 — 오류로 돌려 이름·재시도(같은 시험) |
| 4 | 값 0 인 지수를 이름 없이 버림 | 수리 — `major_indices: <이름> has no positive value` |
| 5 | US 종목 기사 이름 일치가 영문/한글 불일치면 언제나 빈 목록 | 미수리·기록: s2 state 실측상 공식 `/stocks` US 이름은 한글(예 AAPL=애플) → 일치 가능. 일치율은 첫 US s3 세션 state 로 잴 것 |
| 6 | 5xx·429 에 종목마다 실패 요청 반복 | 수리 — 서버 쪽 연속 3 실패면 그 사이클 수급 멈춤(`TestConsecutiveServerFailuresStopTheCycleFlowReads`·`TestFlowFailureCountIsConsecutive`) |
| 7 | 24h 창이 정리 시장의 at 기준이라 다른 시장에선 조금 넘을 수 있음 | 수리 — 고르는 쪽에서도 이 판단의 at 으로 창(`TestHeadlinesApplyTheWindowAtTheirOwnJudgment`) |
| 8 | `day_cache_hits` 에 WTS 수급 캐시 적중이 섞임 | 미수리·기록(표시용 수, 예산 판단은 `broker_reads`·`wts_requests` 로 함) |

- 자기 리뷰에서 찾은 것(리뷰 전): dry run 에서 피드 53/53 이 RFC3339 해석 실패 → news=present·빈 목록 → 시간대 없는 모양 해석 +
  결손 처리로 수리(`TestFeedWithNoParseableTimeIsAGapNotEmptyNews`).
- 남은 위험: 뉴스 시간대 가정(1번) · 뉴스 포함 입력 토큰 미실측 · US 기사 일치율 미실측 · 영숫자 KR 코드는 수급·공식 읽기 모두 결손(후속 change).
