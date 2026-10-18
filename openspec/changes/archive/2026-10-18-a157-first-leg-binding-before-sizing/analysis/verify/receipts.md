# a157 2.1 검증 영수증

- 대상 코드: 커밋 `1e58d3b56`(GREEN · 1.1~1.3). 시험이 돈 HEAD 는 `170ce9144` — 그 사이 커밋은 다른 세션의 `.claude/CLAUDE.md` · `.codex/agents.md`
  문서 편집 하나뿐이고 `git diff --stat 1e58d3b56 HEAD -- internal/ cmd/` 가 비어 있음(Go 소스 동일).
- 도구: `go1.26.5 linux/amd64`. 시간 제한은 Makefile 값(`test` 30m · `test-seams` 45m) — 기본 10m 로는 `internal/journal`(약 13~14 분)과
  태그 `internal/app/engine`(약 13 분)이 부하 아래서 시간 초과로 끊겼음(코드 실패 아님, 같은 코드가 30m/45m 에서 통과).
- 원문: `lint.txt` · `tests.txt` · `validate.txt` · `check-analysis.txt` · `invariants.txt`(1.3).

| 명령 | 결과 |
|---|---|
| `make lint` (gofmt -l cmd internal tools/logic-map + `go vet ./...` + `go vet -tags tossos_testseams ./...`) | **rc 0**, gofmt 위반 0, vet 출력 0 (4m17s) |
| `go vet ./...` | `make lint` 안에서 실행 — 위와 같음 |
| `go test -count=1 -timeout 30m ./internal/execgw/ ./internal/app/engine/ ./internal/risk/ ./internal/riskbucket/ ./internal/journal/` | **ok ×5** (execgw 270s · engine 551s · risk · riskbucket · journal 741s) |
| 같은 패키지 `-tags tossos_testseams -timeout 45m` | **ok ×5** (execgw 332s · engine 806s · risk · riskbucket 7s · journal 824s) |
| `go test -race -count=1 -tags tossos_testseams ./internal/execgw/ -run 'TestA157\|TestQFinalCampaignFirstLeg\|TestA136QFinalCampaignFirstLeg\|TestA138QFinalCampaignFirstLeg\|TestA137'` | **ok** (409s) |
| `go test -race -count=1 -tags tossos_testseams ./internal/app/engine/ -run 'TestA157\|TestA137\|TestA146\|TestStrategyFirstLegAdmission\|TestTheMarketThatLeadsAWaveAlwaysPublishesIt\|TestAFailedWaveCarriesNoAssemblyForEitherMarket'` | **ok** (100s) |
| `go test -count=1 -timeout 30m ./cmd/tossctl/` (조립 감독자 dormant 시험 — 패키지 함수 조립 · Run) | **ok** (76s) |
| `openspec validate a157-first-leg-binding-before-sizing --strict` | **valid** |
| `python3 tools/logic-map/check_analysis.py --change a157-first-leg-binding-before-sizing` | **rc 0** — required 3 함수, `evidence complete`. 착지 기록은 다른 세션의 미커밋 추적 파일 때문에 지금 못 함(Manager 몫) |
| 불변(1.3): `riskguardian_qfinal.go` 파일 sha · `IssuePrecheckedQFinalCampaignFirstLeg` 본문 sha | 편집 전과 같음 — `6fcace26…` · `28c40c8a…`(`invariants.txt`) |

## a157 시험 목록(새 파일 5)

- `internal/execgw/a157_first_leg_binding_testseam_test.go`(태그) — 9 시험: 결속 단독 10 행 · 진입 단독 5 행 · 상한 0 + Bn 9 행 · 버킷/표지 없는 입력 불가 + B2/B8 ·
  일일 손실(한도 · 측정 불가) + Bn 18 행 · (f) 2 행 · 통과 봉인 값(KR q_final 10 · US 4 + 원자 발급) · nil Guardian · 실제 거절의 판별.
- `internal/execgw/a157_first_leg_structure_test.go` — AST 구조 시험(design D6).
- `internal/execgw/a157_daily_loss_kind_test.go` — 판별 표 13 행.
- `internal/app/engine/a157_first_leg_admit_testseam_test.go`(태그) — 5 시험: B2/B8 단독 · 일일 손실 한도/측정 불가 × (없음 · B2 · B8) · 상한 0 단독 · 상한 0 + B2/B8.
- `internal/app/engine/a157_assembly_supervisor_test.go` — 3 시험: 알림 칸 둘 · 기록자 → `c.Log` · EntryGate nil/실제(생산 감독자 대조군).

## 생산 파일 diff stat (base `7b1748481` → HEAD, 시험 제외)

```text
internal/app/engine/strategy_entry_supervisor.go   |   9 +-
internal/app/engine/strategy_first_leg_admission.go |  17 ++--
internal/execgw/a157_daily_loss_kind.go            |  62 +++++++++++++   (새 파일)
internal/execgw/riskguardian_first_leg.go          | 103 +++++++++++++--------  (대부분 닫힌 함수 안으로 들여쓰기)
4 files changed, 141 insertions(+), 50 deletions(-)
```

## round 2 — 2.3 구현 리뷰 수리 뒤 (커밋 `b8d6d4cee`)

수리 내용(Manager 전달 1~7): 구조 시험 실행 모양 고정 · 통과 봉인 정확 대조 · 판별 주석 범위 축소 + tasks 후속 · 판별 이름 `DailyLossRefusalKind*` ·
엔진 강화 실패 행 · 원장 M14c 동등 표기 · 첫 다리 주석을 불변식으로. 원문: `tests-round2.txt` · `lint-round2.txt` · `validate-round2.txt` · `check-analysis-round2.txt`.

| 명령 | 결과 |
|---|---|
| `make lint` | **rc 0** (gofmt 위반 0 · vet 무태그/태그 출력 0, 2m26s) |
| `go test -count=1 -timeout 30m ./internal/execgw/ ./internal/app/engine/` | **ok ×2** (execgw 138s · engine 314s) |
| 같은 패키지 `-tags tossos_testseams -timeout 45m` | **ok ×2** (execgw 201s · engine 389s) |
| `-race` 대상 execgw(`TestA157\|TestQFinalCampaignFirstLeg\|TestA136…\|TestA138…\|TestA137`) | **ok** (385s) |
| `-race` 대상 engine(`TestA157\|TestA137\|TestA146\|TestStrategyFirstLegAdmission\|두 파도 시험`) | **ok** (107s) |
| `openspec validate --strict` | **valid** |
| `check_analysis.py` (번들 셋 재추출 뒤) | **rc 0**, evidence complete |
| 불변: `riskguardian_qfinal.go` 파일 sha · `IssuePrecheckedQFinalCampaignFirstLeg` 본문 sha | `6fcace26fd7b2ddc…` · `28c40c8ab49a1cb6…` — 편집 전과 같음 |
| 변이 round2 (`analysis/mutation/round2.md`) | **36/36 CAUGHT**(M15b · M15c · M15d · M18 · M19 추가) |

risk · riskbucket · journal · cmd/tossctl 은 round 2 에서 Go 편집이 없어(`git diff --stat 1e58d3b56 b8d6d4cee` 는 internal/execgw · internal/app/engine 안의 8 파일뿐, 그 밖 패키지의 컴파일은 `make lint` 의 vet ./... 가 확인) 시험을 다시 돌리지 않음 — round 1 결과를 그대로 씀.
