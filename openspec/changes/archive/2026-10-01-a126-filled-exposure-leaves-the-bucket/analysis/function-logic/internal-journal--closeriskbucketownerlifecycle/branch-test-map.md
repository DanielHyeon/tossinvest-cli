# Branch Test Map: `closeRiskBucketOwnerLifecycle`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:811` `if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHER…` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B2 | `:815` `if orders == 0 {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B3 | `:818` `if _, err := j.db.Exec(`UPDATE positions SET state='CLOSED',quantity='0',closed_at='2026-03-30T00:35:00Z' WHERE account_ref=? AND market=…` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B4 | `:821` `if _, err := j.db.Exec(`UPDATE position_campaigns SET state='CLOSED',entry_blocked=1,version=version+1,updated_at='2026-03-30T00:35:00Z' …` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B5 | `:824` `if _, err := j.db.Exec(`DELETE FROM position_campaign_claims WHERE campaign_id=(SELECT campaign_id FROM risk_bucket_owners WHERE account_…` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B6 | `:827` `if _, err := j.db.Exec(`UPDATE risk_reservations SET state='RELEASED',released_at='2026-03-30T00:36:00Z',release_reason='BROKER_TERMINAL'…` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B7 | `:831` `if reconcile {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
