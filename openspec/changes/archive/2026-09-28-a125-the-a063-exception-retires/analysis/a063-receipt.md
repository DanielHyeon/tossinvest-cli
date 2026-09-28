# a063 재고정 조건 ① 영수증 — 오늘의 측정 (a125 4.2)

- 하네스 `harness/a063_receipt.py`(양성 단언 4종 — codex freeze C1). 고정 워크트리 `a125-probe`, HEAD `76d0816a8569a88e31b946e027dab67993aa1395`(a063 전환 커밋 `76d0816a`), 도구 = 특례 없는 `5440edad` 판.
- **판정: PASS** (rc 0, `failures` 0)
- 창: base `da80ce31b6a1ab5d443016768f970a82bab102db`(P) → 워킹트리, landing `""`, required 341, 오류 332
- 귀속 커밋(P 뒤 자기 Go 비병합): `676bd4b46d52`
- 귀속 함수 9 — 번들 대응 · revision 이 창의 요구와 전부 일치, 그 번들들의 오류 **0**:

| 함수 | 번들 | 번들 revision | 창이 요구하는 revision |
|---|---|---|---|
| `cmd/tossctl/soak.go:newSoakAttestCmd` | `cmd-tossctl--newsoakattestcmd` | current | current |
| `cmd/tossctl/soak.go:runSoakAttest` | `cmd-tossctl--runsoakattest` | current | current |
| `cmd/tossctl/soak_test.go:TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing` | `cmd-tossctl--testsoakattestrefusesanunfinishedsoakandwritesnothing` | base | base |
| `cmd/tossctl/soak_test.go:TestSoakAttestWritesAVerifiableAttestation` | `cmd-tossctl--testsoakattestwritesaverifiableattestation` | base | base |
| `internal/console/console_test.go:TestTheDashboardReportsAnUnstartedMachineWithoutFailing` | `internal-console--testthedashboardreportsanunstartedmachinewithoutfailing` | base | base |
| `internal/console/data.go:Console.readAttestation` | `internal-console--console.readattestation` | current | current |
| `internal/soak/attest.go:BuildAttestation` | `internal-soak--buildattestation` | current | current |
| `internal/soak/attest.go:Summary.Evaluate` | `internal-soak--summary.evaluate` | current | current |
| `internal/soak/attest_test.go:TestBuildAttestationRefusesAnIncompleteSoak` | `internal-soak--testbuildattestationrefusesanincompletesoak` | current | current |

- 남은 오류 332 은 전부 `missing evidence for modified function` — 귀속 밖(형제 change) 함수의 누락이다(분류 안 된 오류 0).
- **이것은 재고정이 아니다.** 결정 (나): 재고정은 a063 게이트 직전에 한 번, 이 하네스로 P → 그때 HEAD 를 다시 잰다(a063 tasks 4.4.1).
  오늘의 '9' 는 오늘의 측정값이다. 귀속의 알려진 한계(문서 커밋과 Go 커밋을 나눈 경우 · 병합 자신의 변경)는 사람이 확인한다.
- 정정 경위: 첫 실행은 soak_test 두 base 번들이 E 블롭 해시를 적어 `AST hash does not match modified function revision` 둘로 FAIL 했다 →
  P 블롭 해시로 고친 `76d0816a` 에서 PASS(design D4-2).
