# 변이 원장 — time namespace 로트 2단계

- 하네스: `mutation-harness.py`(저장소 무변, `go test -overlay` 사본, 시작 sha 단언 · 판마다 원본 sha 재단언 · 무변이 대조군 GREEN 확인 뒤 한 번에 한 판).
- 대상 sha256: `bootstamp.go` `82c39b91edca…`(편집 뒤), `bootstamp_linux.go` `02d2cdb2cc46…`(무변).
- 실행 패키지: `./internal/clock` 전체 + `./internal/accountbudget -run 'CrossNamespace|HintMaxAndRestore|ExpiredSlots'`.
- 원문: 1차 `mutation-run-1.log`, 최종 `mutation-run-2.log`.

## 최종(2차) — CAUGHT 13 · SURVIVED 1 · 대조군 GREEN

| ID | 변이 | 결과 | 잡은 시험 |
|---|---|---|---|
| M0 | 무변이 대조군 | GREEN | — |
| M1 | B3 uptime 항 삭제 | CAUGHT | `TestBootRemainingSameProvenance…`, `TestBootZeroDurationDoesNotHideInvalidClock` |
| M2 | B4 uptime 항 삭제 | CAUGHT | `TestBootCrossNamespaceOrderingStillRejected`(duration 0 판) |
| M3 | B2 uptime 항 삭제 | CAUGHT | `TestBootCrossNamespaceOrderingStillRejected`, `TestBootRemainingRejectsAnchorBeforeSameBootProvenance` |
| M4 | B3 에 ns 항 복원 | CAUGHT | 7개(저장소 2 포함) |
| M4a | B2 에 ns 항 복원 | CAUGHT | `TestBootCrossNamespaceAnchorRestart`, `TestBootRemainingRejectsAnchorBeforeSameBootProvenance` |
| M4b | B4 에 ns 항 복원 | CAUGHT | 같음 |
| M4c | B8 에 ns 항 복원 | CAUGHT | 6개(저장소 2 포함) |
| M5 | B7(다른 boot) 를 같은 boot 처럼 차감 | CAUGHT | 5개 |
| M6 | B8 분기 삭제 | **SURVIVED** | — |
| M8 | B5(duration 0 반환)를 B2 앞으로 | CAUGHT | `TestBootCrossNamespaceOrderingStillRejected`, `TestBootZeroDurationDoesNotHideInvalidClock` |
| M7 | `Snapshot` offset 0 검사 삭제(L152) | CAUGHT | `TestBootSnapshotFailureCases` |
| M7a | `Snapshot` current≠children 검사 삭제(B6 L145) | CAUGHT | `TestBootSnapshotFailureCases` |
| M7b | `Snapshot` 표본 앞뒤 ns 비교 삭제(B12 L168) | CAUGHT | `TestBootSnapshotFailureCases` |

M7 · M7a · M7b: 이번 수정이 `Snapshot` 의 표본 안 가드들의 못을 뽑지 않았음을 보인다.

### SURVIVED M6 — 대신 막는 분기(Manager 결정 2)

B8(편집 뒤 L87, `now.Uptime < base.Uptime`)에 닿는 입력은 B7(L84)을 지났으므로 now 와 base 가 같은 boot 다.
- anchor = nil → base = saved → 같은 입력을 **B3**(L71 `now.BootID == saved.BootID && now.Uptime < saved.Uptime`)이 먼저 거절.
- anchor ≠ nil → base = anchor → 같은 입력을 **B4**(L74 `now.BootID == anchor.BootID && now.Uptime < anchor.Uptime`)가 먼저 거절.
따라서 B8 은 도달 불가이고 M6 은 행동 동등이다. 구조 시험은 만들지 않았다(결정 2).

### 1차에서 드러난 반대 방향 의존(M2)

1차(`mutation-run-1.log`)에서는 M2(B4 uptime 삭제)가 **SURVIVED** 였다. duration > 0 이면 B8 이 같은 입력을 거절해 B4 를
받쳐 주었기 때문이다. B4 만 막는 입력은 duration 0(B5 가 B8 전에 반환)이다. 새 시험 파일의
`TestBootCrossNamespaceOrderingStillRejected` 에 「duration 0 + anchor 역행 → 거절」 판을 넣어 2차에서 CAUGHT. 같은 1차에서 M7b 는
변이 문장이 변수 미사용 빌드 오류를 내 무효였고, 사용을 남기는 형태(`false && (…)`)로 고쳐 2차에서 CAUGHT.

## 고친 기존 시험 3개가 편집 뒤 죽이는 변이(통과 이유 재계수)

| 시험 | 편집 뒤 죽이는 변이 |
|---|---|
| `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor` | M1 · M4 · M5 |
| `TestBootZeroDurationDoesNotHideInvalidClock` | M1 · M4 · M8 |
| `TestBootRemainingRejectsAnchorBeforeSameBootProvenance` | M3 · M4a · M4b · M4c |

편집 전 이 셋은 ns 불일치 거절(지금은 M4 계열이 되돌리는 동작)을 단언했다. 편집 뒤에는 그 판이 「ns 차이를 거절하지 않음」을
단언해 M4 계열을 죽이고, 남은 판이 uptime · anchor 순서(M1 · M3) 와 duration 0 순서(M8) · 다른 boot 보류(M5)를 지킨다.
