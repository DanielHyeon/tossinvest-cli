# Function Logic Map: `startEngine`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a056-autostart-survives-container-recreate
- Risk scan: `risk-pattern-report.md`

이 map은 편집 **전에** 작성했다 (tasks.md 1.1). `startEngine`은 엔진 기동 경로이므로
면제 대상이 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root.configDir` | journal 디렉터리로 해석 가능한 경로 | `engineJournalDir` | 오류 반환, 기동 없음 |
| 자기 실행 파일 경로 | 존재하는 실행 파일 | `binstamp.SelfPath` | 오류 반환, 기동 없음 |
| 엔진 활성 마커 | mtime이 `enginelock.StaleAfter`(5분) 안이면 fresh | `enginelock.Read` | **자문 신호.** 단독으로 거부 근거가 될 수 없다 |
| 엔진 프로세스 목록 | `tossctl engine run` 매칭 PID | `engineFindProcesses` (pgrep) | 오류면 부재를 주장할 수 없음 → 보수적으로 기존 거부 유지 |
| journal flock | 단일 writer | spawn된 `engine run`이 첫 동작으로 획득 | **정본 배타.** 두 번째 런타임은 여기서 죽는다 |

불변식: 이 함수는 브로커·자격증명·토큰을 쥐지 않는다. 게이트·인터록·락 판정은 전부
spawn된 `engine run`이 자기 안에서 한다. 이 함수가 하는 것은 **사전 안내**뿐이다.

## Branches and early returns

이 change가 바꾸는 것은 정확히 한 칸이다: **마커 fresh + 프로세스 미관측**.

| 마커 | 프로세스 열거 | 변경 전 | 변경 후 |
|---|---|---|---|
| fresh | 관측됨 | 거부 (마커 문구) | 거부 (마커 문구) — 동일 |
| fresh | 없음 | **거부** | **진행** — 유령 마커 |
| fresh | 열거 오류 | 거부 | 거부 — 동일 (부재 미증명) |
| stale/없음 | 관측됨 | 거부 (프로세스 문구) | 거부 (프로세스 문구) — 동일 |
| stale/없음 | 없음 | 진행 | 진행 — 동일 |
| stale/없음 | 열거 오류 | 진행 | 진행 — 동일 |

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `engineJournalDir` 오류 | 없음 | 오류 | 기존 커버 |
| B2 | `binstamp.SelfPath` 오류 | 없음 | 오류 | 기존 커버 |
| B3 | 마커 fresh **AND** (프로세스 관측 **OR** 열거 오류) | 없음 | "엔진이 이미 실행 중이다 (pid, 갱신시각)" | `TestAGhostMarkerDoesNotRefuseAStart`, `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestEnumerationFailureKeepsTheRefusal` |
| B4 | 프로세스 관측 (마커 무관) | 없음 | "엔진 프로세스가 이미 있다 (pids)" | `TestStartingIsRefusedWhenAProcessIsAlreadyThere` |
| B5 | `engineSpawnDetached` 오류 | 없음 | 오류 | 기존 커버 |
| B6 | probe 창 안에 즉시 종료 | spawn됨 | 로그 tail + 오류 | `TestARefusedStartReportsTheEnginesOwnLog` |
| B7 | 즉시 종료가 오류 없이 | spawn됨 | 로그 tail + "오류 없이" | 기존 커버 |
| — | 위 어느 것도 아님 | **spawn** | "엔진을 시작했다 — 로그 <path>" | `TestStartingSpawnsTheEngineWithThisProfilesConfigDir`, `TestAStaleMarkerDoesNotBlockAStart`, `TestAGhostMarkerDoesNotRefuseAStart` |

B3이 이 change의 전부다. 변경 전 B3은 `status.Running` 하나였고, 그래서 바로 아래
B4(실제 프로세스 검사)에 **도달하지 못했다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` | journal 디렉터리 결정 | 오류 전파, 재시도 없음 | AST |
| `binstamp.SelfPath` | spawn할 실행 파일 | 오류 전파 | AST |
| `enginelock.Read` | 마커 신선도와 PID·갱신시각 | 파일 없음/파손은 `Running=false` | AST + `internal/enginelock` |
| `engineFindProcesses` | 실제 엔진 프로세스 관측 | 오류는 **부재 아님** | AST, seam |
| `engineSpawnDetached` | 엔진 기동 | 오류 전파 | AST, seam |
| `readLogTail` | 즉시 종료 시 엔진 자신의 사유 | 없음 | AST |

호출 계약은 이 change로 바뀌지 않는다. 바뀌는 것은 `engineFindProcesses`를 **B3보다
먼저 한 번** 호출해 두 판정이 같은 관측을 공유한다는 점뿐이다 (두 번 세면 두 답이 나올
수 있다).

## State mutations and fallbacks

- 도메인 변경 없음. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어느 것도 이
  함수에서 바뀌지 않는다.
- 유일한 side effect는 `engine run` 프로세스 spawn이며, 그 프로세스가 자기 게이트·
  인터록·flock을 스스로 검사한다.
- fallback: 프로세스 열거 실패는 부재로 읽지 않는다.

## Safety conclusion

- Safe edit boundary: 기동 **전 안내** 판정. 배타 자체는 건드리지 않는다.
- 이 change는 거부를 **없애지 않는다.** 거부의 근거를 자문 파일에서 관측된 프로세스로
  옮기고, 최종 배타는 spec이 이미 정한 flock에 남긴다.
- 완화 방향 위험: 살아 있는 엔진을 pgrep이 못 보는 경우(예: 호스트 native 엔진 +
  컨테이너 콘솔, 서로 다른 PID namespace) 두 번째 spawn이 일어난다. 그 spawn은 공유된
  config 디렉터리의 flock에서 죽으므로 journal 단일 writer는 유지되고, 잃는 것은 안내
  문구의 품질뿐이다. 반대편 비용은 실측된 8분 감시 공백이다.
- 손절·비상 청산의 즉시성을 **강화**하는 방향이다: 엔진이 떠 있어야 exit 루프가 돈다.
