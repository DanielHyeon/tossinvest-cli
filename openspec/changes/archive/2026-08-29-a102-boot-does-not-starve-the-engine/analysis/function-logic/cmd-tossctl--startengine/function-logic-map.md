# Function Logic Map: `startEngine`

> **이 change는 이 함수를 편집하지 않는다.** D7의 범위 밖 선언과 design의 「범위 밖」이
> `engineStartProbe`의 결과 기반 전환을 별도 change로 미뤘다. 이 문서는 **proposal 근거용
> 분석**이고, 분기 열거는 전부 `ast.json`이 낸 것이다.

- Source: `cmd/tossctl/engineproc.go` (177-227)
- AST evidence: `ast.json` — AST 기준 branches **7** / returns 8 / calls 21
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256 (현재 HEAD): `502362c0b925e09f57b5253ad92e1ce1bbcfe0d1f601869e76c44ff6f9d13439`
- 작성 사유: a102 proposal이 **"엔진 자동 시작이 3초 probe로 끝나고, 그 직후 서베이가
  같은 rate 예산을 때린다"**를 주장하면서 이 함수의 `engineStartProbe` 분기를 근거로 썼다.
  그 주장의 AST 근거가 이 묶음이다.

## 이 함수가 하는 일

콘솔의 `StartEngine` seam. 이 프로필의 엔진을 **없을 때만** 띄우고, 3초(`engineStartProbe`)
동안 즉시 종료를 지켜본 뒤 한 줄을 돌려준다.

**a102가 이 함수에서 읽은 사실 하나**: `:217`의 `select`는 "3초 안에 죽지 않았다"까지만
확인한다. 그것은 **재시작 복구가 끝났다는 뜻이 아니다** — 복구는 실측 ~50초가 걸렸다
(2026-08-13 01:35:02→01:35:52). 겹2가 존재하는 이유가 이 간극이다. 함수 주석은 "every
refusal happens before the loops start"라고 적고 있는데, 그 문장은 **거절**에 대해서만
참이고 **복구 완료**에 대해서는 참이 아니다. design은 그 주석을 고치는 것을 별도 change로
선언했다(범위 밖).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root.configDir` | 임의 | CLI 플래그 | B1 — 경로 해석 실패는 즉시 return |
| 실행 파일 경로 | — | `binstamp.SelfPath()` | B2 — 즉시 return |
| `pids, findErr` | 이 프로필의 엔진 프로세스 | `engineFindProcesses(dir)` | 열거 실패는 **거절을 강화**한다 |
| 마커 신선도 | — | `enginelock.Read` | **단독으로는 거절하지 못한다** (a056, `markerRefusesStart`) |
| `engineStartProbe` | 3s 상수 | `engineproc.go` | B6 — 그 안에 죽으면 로그 tail과 함께 실패 |

> **불변식**: 마커 하나로 거절하지 않는다. 컨테이너 재생성·SIGKILL·호스트 재부팅이 전부
> 프로세스를 지우고 파일을 남긴다. 거절에는 **프로세스 관측 또는 열거 실패**가 함께 필요하다.

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:179` | `engineJournalDir` 오류 | — | `:180` |
| B2 | `:183` | `binstamp.SelfPath` 오류 | — | `:184` |
| B3 | `:199` | `markerRefusesStart(status.Running, observed, findErr != nil)` | — | `:201` "이미 실행 중이다 (pid …)" |
| B4 | `:204` | `observed` | — | `:205` "엔진 프로세스가 이미 있다" |
| B5 | `:210` | `engineSpawnDetached` 오류 | 자식 없음 | `:211` |
| B6 | `:217` | `select`: `<-wait` / `<-time.After(engineStartProbe)` | 로그 tail 읽기 | `:221`·`:223` 또는 통과 |
| B7 | `:220` | probe 안에 죽었고 `exitErr == nil` | — | `:221` "곧바로 종료했다 (오류 없이)" |

정상 이탈: `:226` "엔진을 시작했다 — 로그 …".

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineFindProcesses(dir)` `:191` | **한 번만** 열거 | 두 번 세면 한 판정 안에서 두 답이 나온다 | ast.json |
| `enginelock.Read(...)` `:199` | 자문 마커 | 관대한 reader — 실패는 전부 "엔진 없음" | ast.json |
| `markerRefusesStart` `:199` | a056의 판정 | 마커 단독 거절 금지 | `engineproc.go:169` |
| `engineSpawnDetached` `:209` | 자식 기동 | `wait` 채널을 돌려준다 | ast.json |
| `time.After(engineStartProbe)` `:224` | **3초 관측 창** | **복구 완료와 무관하다** | ast.json |

live binding — 호출자는 둘: `runConsole`의 `StartEngine` seam(`console.go:520`)과
`runConfiguredEngineAutostart`가 부팅에 쓰는 closure(`console.go:343`). **후자가 a102의
문맥이다** — 부팅 자동 시작이 3초 뒤 돌아오고, 그 직후 같은 함수 안에서 서베이가 시작된다.

## State mutations and fallbacks

- 프로세스 밖 상태: 자식 프로세스 하나와 그 로그 파일.
- fallback 없음 — 모든 실패 방향이 `(note, error)`로 돌아간다.

## Safety conclusion

- Safe edit boundary: **없음 — a102는 이 함수를 편집하지 않는다.**
- High-risk impact: yes(엔진 기동). a102는 이 함수의 계약을 **읽기만** 한다.
- a102가 이 함수에 대해 남기는 선언된 생략: `engineStartProbe` 주석의 거짓 전제
  ("probe를 통과한 엔진은 모든 거절을 지났다" → 복구 완료를 뜻하지 않는다)는 **고치지
  않는다.** 겹2가 서베이를 신호에 묶으면 probe는 서베이 기동과 무관해지므로, 그 주석 수정은
  별도 change다(design 「범위 밖」). **침묵한 생략이 아니라 선언된 생략이다.**
