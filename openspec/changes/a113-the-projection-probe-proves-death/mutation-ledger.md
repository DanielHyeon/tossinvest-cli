# a113 뮤테이션 원장

측정 2026-09-25, 비root(uid 1000), Go `go test -count=1`, 하네스 `analysis/harness/run.sh`(저장소를
건드리지 않고 `go.mod`·`internal/strategyprojection{,rpc}` 사본에서 변이 — 병행 세션 보호).
무변이 대조군 `none` rc=0 을 매 판 먼저 확인했다(사본 하네스가 눈먼 경우 가르기).

| id | 변이 | 결과 | 죽인 테스트 |
| --- | --- | --- | --- |
| N1 | 순수 probe 에 owner-write 절 재도입 | **사망** rc=1 | `TestProjectionLivenessClausesEachDecideOnTheirOwn/쓰기_비트가_깎여도_수락_중이면_생존`·`/쓰기_비트가_깎인_죽은_socket_은_묻지_못한다` |
| N2 | 회수 probe 의 chmod 제거 | **사망** | `TestStartRecoversFromUnwritableSocketLeftover`(영구 거부 재발) · RED 테스트 · 판정 표 2행 |
| N3 | chmod **뒤** SameFile 재확인 제거 | **사망** — 단 구조 핀 하나로만 | `TestTheStaleProbeChecksTheNameOnBothSidesOfTheChmod`. 1차 판에서는 **생존**했다: 바뀐 파일 테스트가 chmod **앞** 재확인에서 먼저 걸려 뒤의 확인을 가렸다(뒤의 확인은 chmod–connect 경합에서만 일하고 그 경합은 결정적으로 못 만든다) → 순서 AST 핀 추가 |
| N3b | chmod **앞** 재확인 제거 | **사망** — 구조 핀 | 같은 핀. 행동 테스트는 뒤의 확인이 받아 초록(둘이 서로를 덮는다 — 그래서 행동이 아니라 순서를 고정) |
| N4 | F1-N1 원형판 = chmod 제거 + owner-write 절 재도입(병의 복원) | **사망** | RED `TestTheReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped` · 순수 표 2행 · 회수 표 1행 |
| N5 | chmod 모드 `0o600` → `0o666` (freeze P1-3) | **사망** | RED 테스트(0600 단정) · 회수 표 산 0400 행 · `TestStartRefusesLiveProjectionOwnerWithoutRemovingIt`(0666 socket 의 Dial 이 정확-0600 거부) |
| N6 | 회수가 verify 의 FileInfo 대신 새 Lstat 을 넘김 (freeze P2-2) | **사망** — 구조 핀 | `TestTheReclaimHandsTheProbeTheInodeItVerified` |
| N7 | 회수가 새 함수를 버리고 순수 probe 로 되돌아감(배선) | **사망** | `TestStartRecoversFromUnwritableSocketLeftover` · RED 테스트 · 배선 AST 핀 |

생존 0 / 9. 등가 변이: nil 가드 단독 제거(`os.SameFile(nil, x)` 는 false — freeze 리뷰어 판정).

root 에서는 EACCES 행이 `t.Skip` 되므로 N1·N4·N5 의 일부 사망 근거가 사라진다 — CI(ubuntu-latest,
비root)와 이 워크스테이션(비root)에서 잰 값이다.
