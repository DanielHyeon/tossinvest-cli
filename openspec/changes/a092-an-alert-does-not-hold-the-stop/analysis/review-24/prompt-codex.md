# a092 24라운드 — 좁은 확인 (codex 한 보이스)

읽기 전용 리뷰다. 파일을 만들거나 고치지 말고 git 상태를 바꾸지 마라. LIVE 주문·엔진 기동·`mutating: true` 명령 금지. 운영 원장·자격 증명 파일은 열지 마라.
`openspec/changes/a092-an-alert-does-not-hold-the-stop/review.md` 와 `analysis/review-2*/` 의 다른 원문은 열지 마라. 이 트리는 git 저장소가 아니다(git archive `9fa0bb90`).

## 범위 — 좁게

23라운드(세 보이스 모두 BLOCK — 남은 결함은 전부 문서, 안전 불변식 · exit goroutine 원격 대기 0 · 게이트 오개방 0은 성립)의 처분 M1~M16 을 24판이 **적은 대로 반영했는지만** 확인한다.
새 설계 전면 재리뷰가 아니다. 반영분이 새 모순을 만들었으면 그것은 판정한다.

대상 문서: `openspec/changes/a092-an-alert-does-not-hold-the-stop/` 의 `specs/engine-safety/spec.md`, `specs/exit-policy/spec.md`, `design.md` 의 **D0.3i**,
`tasks.md` 의 「24.」 절, `proposal.md` 머리 24판 표. 정본 `openspec/specs/engine-safety/spec.md` · `openspec/specs/risk-management/spec.md`, 코드는 필요한 곳만.

## 확인 항목 (각각 PASS / FAIL + 파일:줄)

- **M1**: 통지 SHALL 이 「통지자(announcer)를 받는 전이」로 한정되고, 전달 실패(`CRITICAL_ALERT_UNDELIVERED` — 동기 `internal/obs/notifier.go:382-383` · 실행자 `internal/app/engine/alertdelivery.go:451-452`, 둘 다 announcer nil)와 durable 기록 실패 강화가 무통지 예외로 명시됐는가. 같은 델타의 다른 SHALL 이나 Scenario 와 새로 모순되는가.
- **M2**: K3 구조 핀이 「커밋 성공 경로」로 한정돼 HEAD 양성 대조군(`internal/journal/operating_mode.go:468-476`, 커밋 실패 반환 `:469`)을 통과하는 문언인가. 투영기 몸체 `go` 0 핀이 있는가.
- **M3**: archive 게이트(a066 입구 이행 커밋 인용 전 archive 금지)가 tasks 에 게이트로 적혔고, 델타에 「세울 자기 사유가 없는 기록자는 입구를 써야 한다」가 있는가.
- **M4 · M9**: 기록 부류가 「알림기의 배제 잠금 아래에서 기록하는 모든 경로」(기록 전용 입구 + 동기 claim `ClaimAlertForDelivery`)로 한 요구 안에서 일관되고, census 에 `ClaimAlertForDelivery` 가 있는가. 재무장 문장이 「재알림 창에 의한 재무장」으로 한정됐는가.
- **M5**: 「모든 발송자」 문단에 a124 두 조항(승격 미포함 판정은 승격 금지 · 승인 시각으로 순서 추정 금지)이 있고, 설계의 (ii) 수단이 `notifyCritical`(`notifier.go:223-228`) 자리이며 RED 가 `:484` · `:571` 로 한정됐는가.
- **M6**: flatten 잔여가 「도달 0」으로 정정되고 근거 사슬이 코드와 맞는가(`internal/reconcile/recovery.go:351` · `internal/app/engine/runtime_wiring.go:184` · `cmd/tossctl/flatten.go:199-272` · `internal/execgw/replay.go:249-259` · `internal/app/engine/gateway.go:312-315`). 정적 핀 둘이 있고, 델타의 입구 · 셈~해제 문단이 「엔진 프로세스」로 한정됐는가.
- **M7 · M8**: 역할 동일성 핀(`Entry` 식별자 == `newNotifier` 게이트 인자)과 생산 조립의 비순환 정의(`tossctl` main 에서 도달하는 비시험 경로).
- **M10**: 통지 신원이 `rec.ID`(TEXT PRIMARY KEY), rowid 는 울타리 순서 전용.
- **M11**: MODIFIED 「배달 실행자의 정지가 …」 블록에 새 코드 좌표가 없고(이름으로), 남아 있던 `risk-management :102-108` 이 요구 이름으로 병기됐으며, **비인용(규범) 줄이 정본과 같은가**(직접 diff).
- **M12 · M13 · M14 · M15 · M16**: 통지 실패 정의(= 통지 기록 실패, 재읽기 상태) · K8 서술 정정 + RED · 잔여(정본 `:221-224`, 처리 주체 a092 구현 로트) · 커버리지 셋 · exit-policy 한정어에서 수단 「버퍼」 제거.

## 출력

맨 위에 **판정 한 줄: PASS / FAIL**(FAIL 은 반영 누락·반영이 만든 새 모순이 있을 때만). 그 아래 항목별 PASS/FAIL 표(`항목 | 판정 | 근거(파일:줄) | 남은 결함`),
그 아래 새로 찾은 결함이 있으면 등급(P0/P1/P2, 문서만이면 (T))과 함께. 마지막 줄 `Recommendation: <action> because <reason>`.
