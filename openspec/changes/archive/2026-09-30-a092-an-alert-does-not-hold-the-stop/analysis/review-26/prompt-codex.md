(prompt-common.md 전문을 먼저 읽는다 — codex 보이스 초점.)

## codex — 전면 적대 리뷰

frozen 24판 델타의 SHALL 문장을 구현과 한 줄씩 대조하라. 25라운드에서 당신은 「소진 뒤 반납의 행 없음을 선점으로 처리」(P0)를 찾았고 `55963f29` 가 고쳤다 — 그 수리의 완전성과, 같은 부류(결과 분류가 선점/실패를 잘못 가르는 자리)가 남아 있는지 `deliver` · `alertDeliverer` · `Flush` · `ReleaseAlertClaim` 결과 처리 전부에서 확인하라. 그리고 저자 주장 1~6.
