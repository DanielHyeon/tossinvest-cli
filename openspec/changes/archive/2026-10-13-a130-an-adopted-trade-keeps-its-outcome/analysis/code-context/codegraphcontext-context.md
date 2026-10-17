# CodeGraphContext / GBrain 보조 문맥 (a130)

- GBrain: `gbrain_project.py search` → `[gbrain-project] busy: owner pid=23201`(다른 세션 소유 — advisory, WORKFLOW 「GBrain 단일 프로세스」). 결과 없음.
- 파일 기억 `scripts/memory-recall.sh "trade outcome adopted backfill"` → 0건.
- CodeGraphContext: `make sdd-sync` 때 갱신(2.4). 후보 문맥으로 본 것(HEAD 직접 읽기로 확인):
  - `internal/journal/position_adjustments.go` — 조정 행의 칼럼(`prev_quantity` · `new_quantity` · `new_avg_price` · `created_at`) 과 「'' = 브로커 미보고」 규약.
  - `internal/journal/adoption.go` — `position_adoptions.observed_at` 은 편입 직전 관측 시각, 편입 기록은 set-once.
  - `internal/journal/readonly.go` — `OpenReadOnly`(mode=ro + query_only), `*ReadOnly` 메서드 집합 시험.
  - `internal/journal/outbox.go` — 미전달 경보가 신규 진입을 막음 → 거절 기록 경로로 부적합(design D3).
  - `internal/console/templates_portfolio.go` 안내문 · `portfolio_test.go:752` 문장 고정.
  - `tools/a087-market-sell-probe` — 미리보기 → 확인 토큰 → 실행 도구 선례.
