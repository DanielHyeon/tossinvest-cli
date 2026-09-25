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

| R1 | chmod **뒤** 재확인의 결과를 버림(`_, _ = sameVerifiedSocket(…)`) — post-review P1-1 | **사망** (2판) | 1판(ed3cb1d7): **생존** — 순서 핀이 호출만 셌다. 2판: `TestTheStaleProbeRefusesANameSwappedDuringTheChmod`(chmod seam 으로 경합을 결정적으로) · 강화된 순서 핀(재확인이 반환을 가르는가) |
| R2 | chmod **앞** 재확인의 결과를 버림 | **사망** (2판) — 구조 핀 | 1판 생존. 2판: `TestTheStaleProbeChecksTheNameOnBothSidesOfTheChmod`(`a113IsGatingRecheck` 2개 요구). 행동으로는 뒤 재확인이 받는다 |
| R3 | chmod 의 ENOENT 도 생존으로 | **사망** (2판) | 1판 생존. 2판: `TestTheStaleProbeReadsAChmodFailureConservatively/사라졌으면_사망` |
| R3b | chmod 의 모든 실패를 사망으로(위험 방향) | **사망** | `…/권한_오류면_생존` |
| R4 | seam 의 운영 값을 no-op 으로 | **사망** | `TestStartRecoversFromUnwritableSocketLeftover` · RED 테스트 (`TestTheProductionChmodIsTheRealOne` 은 포인터 비교) |

2판 합계(2026-09-25, 커밋 전 워킹트리): **생존 0 / 14**. 1판(ed3cb1d7)은 9/9 로 적었지만 구현 후
독립 리뷰가 R1·R2·R3 세 생존을 찾았다 — 1판의 "생존 0" 은 **내가 고른 변이 집합**에 대한 진술이었다.

알려진 생존(측정 불가, 기존): socket uid 절(`!ownedByEffectiveUser(socketStat.Uid)`) 삭제 — 비root 가
남의 소유 socket 을 못 만든다(post-review R5, a108 이래 동일). nlink 절 삭제는 사망(R4 리뷰 판,
`TestStartRefusesUnsafeLeftoverShapes/socket에_hard_link가_걸려_있다`). 등가 변이: nil 가드 단독 제거(`os.SameFile(nil, x)` 는 false — freeze 리뷰어 판정).

root 에서는 EACCES 행이 `t.Skip` 되므로 N1·N4·N5 의 일부 사망 근거가 사라진다 — CI(ubuntu-latest,
비root)와 이 워크스테이션(비root)에서 잰 값이다.
