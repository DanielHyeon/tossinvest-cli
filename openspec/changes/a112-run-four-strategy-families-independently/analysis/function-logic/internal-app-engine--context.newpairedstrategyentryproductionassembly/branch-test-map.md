# Branch Test Map: `Context.NewPairedStrategyEntryProductionAssembly`

- Source SHA-256: `6f1f6804cfd437116c16a48d1526327c28360433f09e0ba3567eb3537442ee5b`; AST branch locations are authoritative.
- Revision: **modified (a112 6.3, 2026-10-01).** 편집 전 9 분기 → 8: 재검증 클로저 안의 drift 판정 if(편집 전 B6 자리)를 순수 함수 `strategyScheduleStillMatchesAdmission`(`strategy_schedule_revalidation.go`)로 **의미 무변경 이동** — 클로저는 수집 한 문장 + 그 함수 호출 한 문장. 조건식 철자 동일(이동 영수증). 나머지 분기는 번호만 당겨졌다.
- 편집 전 번들: `analysis/measurements/lot-6.3/pre-edit/internal-app-engine--context.newpairedstrategyentryproductionassembly/`. 변이 원장 `analysis/measurements/lot-6.3/mutation-6.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 294:2 — nil Context → 오류 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B2 | if at 300:2 — 원장 경로 있음 → 후보 · 원장 경로 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B3 | if at 311:2 — 원장 경로 → 근거 경로 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B4 | if at 321:2 — 레인 세우기 실패 → 오류 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B5 | if at 342:2 — 시계 있음 → dispatch now 배선 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B6 | range at 358:2 — KR · US worker 만들기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B7 | if at 363:2 — 감독자 생성 실패 → 오류 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B8 | if at 370:2 — 투영 발행 실패 → 오류 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
