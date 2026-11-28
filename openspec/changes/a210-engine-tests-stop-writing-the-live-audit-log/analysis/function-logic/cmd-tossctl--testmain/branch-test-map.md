# Branch Test Map: `TestMain`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없음 — 순서 고정 `i18n.SetLang` → 상자 안 `m.Run` | 가짜 HOME 측정 `analysis/measure.sh` 의 `cmd/tossctl`(무태그 · `tossos_testseams`) + 패키지 시험 전체 | yes — base 에서 상자에 `.local/share/tossos/audit.log` 1 개(`analysis/red/summary.txt`) | 편집 뒤 측정 `analysis/green/summary.txt` |
