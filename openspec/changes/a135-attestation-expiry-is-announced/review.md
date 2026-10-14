# a135 review

Function Logic Map: not-applicable — `internal/app/engine/attestation_watch.go` 는 새 파일(새 leaf 함수만), `internal/obs/event.go` 는 상수 선언 2개(함수 · 등급표 무편집).
수정된 기존 함수 `engineRuntime` · `TestProductionRuntimeStartsExactlyOneAlertDeliverer` 는 번들을 만듦(`analysis/function-logic/`, 분기 변화 0).

## 구현 리뷰 (2026-10-14, 분리된 컨텍스트 · 적대적 · 읽기 전용) — APPROVE-WITH-FIXES

리뷰어 실행: `go test -race -run TestA135` · 엔진 패키지 전체 · obs 패키지 · `go vet` 통과, overlay 변이 · 프로브(저장소 무변경).

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| R1 | P1 | 판정의 `recover` (그리고 `runAuxiliaryBody` 의 회수)가 알림기 안의 패닉까지 삼키면, `Notifier.claimAndDeliver` 가 `n.mu` 를 defer 없이 놓으므로 잠금이 영구히 잡혀 exit 관측기의 `alert` · `Acknowledge` · `Flush` 가 조용히 멈춤(overlay 프로브로 교착 재현). 이 change 가 처음으로 Notify 를 회수 경계 아래에서 부름 | **수용.** 회수는 `prepare`(이 파일의 읽기 · 판정 · 문구)에만, Notify 는 `notifyOutsideRecovery` 의 별도 goroutine 에서 — 패닉은 어떤 회수에도 닿지 않고 프로세스를 내림(감독 루프 호출자와 같은 결과). 고정: 자식 프로세스 시험 `TestA135NotifyPanicIsNotRecovered` + 변이 M9 CAUGHT. `claimAndDeliver` 의 패닉 안전 해제(근본 수리)는 알림 경로 기존 함수 편집이라 **후속**(Manager 보고) |
| R2 | P2 | 신원 재설정이 시험에 안 박힘(M7 생존) | 수용 — `TestA135SameStageOnANewAttestationIsAnnouncedAgain`(같은 단계 · 다른 신원 · 발급만 다른 신원), M7 CAUGHT |
| R3 | P2 | 신원이 outbox 키에 있는지 안 박힘(M8 · M11 생존) | 수용 — 같은 시험이 두 신원의 키가 다름을 단언, M8 CAUGHT |
| R4 | P3 | 종료 중 Notify 가 끊겨 메모리 래치 · 오류 줄 | 수용 — `ctx.Err()` 면 기록을 시작하지 않음. 기록 도중 취소의 잔재는 알림기 공통 동작으로 남음 |
| R5 | P3 | `ErrIncomplete` 를 「파일이 없다」 문구로 알림 | 수용 — `incomplete` 단계 · 문구 분리 |
| R6 | P3 | 빈 계좌가 `account_mismatch` 로 붙음 | 수용 안 함 — 게이트 켜진 실행 중 엔진에서는 생기지 않음(인터록 5 가 기동을 거절). 라벨 차이 |
| R7 | P3 | `supervised_proof_at` 이 같은 endpoint 의 오래된 증거로 앞당겨짐 | 수용 — endpoint 별 최신의 최소(`binding`), `TestA135SupervisedProofAtIsTheBindingProof`, M11 CAUGHT |
| R8 | P3 | 만료 · 범위 상실이 프로세스당 한 번만 알려짐(다른 지속 critical 은 리마인드) | 수용 — 재시작이 거절될 상태는 KST 날짜를 키에 넣어 하루 한 번(`TestA135InvalidStatesAreRemindedDaily`, M10 CAUGHT). spec 문장 갱신 |
| R9 | P3 | 재시작이 잦으면 normal 통지가 재시작마다 | 수용 안 함(design D2 의 받아들인 비용) |
| R10 | P3 | 계좌 없음 시험이 `unreadable` · 불일치의 `Key` 를 안 봄 | 수용 — `TestA135UnreadableAlertCarriesNoAccountOrPath`(경로 포함), 불일치 검사에 Key 추가 |

변이(무변이 대조군 GREEN 먼저): 11/11 CAUGHT — `analysis/mutation/receipt.txt`.

## 후속(이 change 밖)

- `obs.Notifier.claimAndDeliver` 의 잠금 해제를 패닉 안전하게(defer) — 지금은 이 change 가 회수 경계 밖 호출로 피해 감. 다른 보조
  실행자가 critical Notify 를 회수 경계 아래에서 부르기 시작하면 같은 교착이 생김.
- 콘솔 72시간 경고와 알림 D-2 의 기준 통일(design D5).
- a134 의 호스트 알림 입구가 서면 attest 타이머 실행 뒤 결과 보고 검토(design D3).

## 사람 승인 base 재고정 (2026-10-14)

- 승인: 사용자(2026-10-14, Manager 제시안 「1. 둘 다 승인」 선택).
- 옛 base `421bf53c` → 새 base `870baed68359`(이 change 첫 Go 커밋의 부모).
- 귀속 실측: 옛 base..새 base 에서 이 change 디렉터리를 만지고 `.go` 를 고친 비병합 커밋 0.
- 옛 base 에서 required 28 함수 — 거의 전부 형제 착지 몫(a128 1.6·1.7·a134(85752cd6·0d16eaa5)). 새 base 격리 워크트리 실측: check_analysis rc 0, evidence complete.
