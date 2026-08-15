# a109 구현 중 발견 — 분류와 처리


Teammate가 구현 중 발견한 계약 결함을 기록한다. 분류는 ① blocking(안전·동작 모순 —
구현 중단 후 보고) ② safe local(스펙 의도가 명백한 사소한 보완 — 구현하며 사후 기록)
③ editorial(즉시 수정) 셋이다.

**blocking 0건.** (있으면 여기 맨 위에 적고 구현을 멈춘다.)

---

## T1 — ③ editorial

### E1. tasks §1a의 "자기 표면 9개"는 슬러그 수를 잘못 셌다 (2026-08-15)

tasks.md §1a는 T1의 FLM 표면을 "9개: prepare/stat/validate×2계열, Start×3, Close×3"으로
적었다. 실제 scaffold 된 T1 슬러그는 **17개**다:

- `internal/positionpolicyrpc` 8 — PreparePrivateSocket · PrepareRuntimeSocket ·
  statPrivateSocket · ValidatePrivateSocket · ValidateRuntimeSocket ·
  ValidatePrivateControlDirectory · ValidateRuntimeControlDirectory · validatePrivateDirectory
- `internal/app/engine` 9 — Start×3 · Close×3 · descriptor 발행×3

§1a 문구를 17로 고치고 17개를 전부 완성했다. (9는 "prepare 2 + stat 1 + validate 소켓 2 +
Start 3 + Close 3 = 11"과도 맞지 않아 어느 셈으로도 성립하지 않는다.)

### E2. `validatePrivateDirectory` scaffold의 Source 경로가 실제 정의 파일과 다르다

`analysis/function-logic/internal-positionpolicyrpc--validateprivatedirectory/`의
`function-logic-map.md` scaffold는 Source를 `internal/positionpolicyrpc/private_fs_unix.go`로
적었으나, 같은 디렉터리의 `ast.json`은 `internal/positionpolicyrpc/client.go`를 가리킨다
(실제 정의도 client.go:140). `check_analysis.py`가 ast.json의 경로를 정본으로 요구하므로
맵을 client.go로 정정했다. `risk-pattern-report.md`는 처음부터 client.go로 옳게 적혀 있었다.

---

## T1 — ② safe local

### S1. `PreparePrivateSocket`/`PrepareRuntimeSocket`을 **삭제하지 않고 남긴다**

design D2는 두 함수의 동작(probe 없는 unlink)을 거부로 바꾼다고만 하고, 함수의 존폐를
말하지 않는다. a109 GREEN 이후 두 함수는 기동 경로에서 호출되지 않는다.

**남기기로 했다.** 근거: ① `PrepareRuntimeSocket`에는 기존 테스트
(`TestPrepareRuntimeSocketNeverDeletesNonSocket`)가 있어 삭제는 기존 테스트 삭제를 동반한다.
② 둘 다 공개 API이고 `_other.go` 플랫폼 스텁까지 짝이 있다. ③ 삭제하면 FLM이
`revision: base`로 바뀌어 게이트 형태가 달라진다.

대신 **본문은 한 줄도 고치지 않는다** — 고치면 공유 helper(`statPrivateSocket`)를 통해
완화가 클라이언트 검증으로 번진다(freeze P1-3). 남은 위험: 다음 사람이 이 함수를 다시
기동 경로에 배선할 수 있다. 그 위험은 §1.2의 "산 주인 거부" 핀이 잡는다 — 배선을 되돌리면
그 테스트가 죽는다.

### S2. 회수는 잔재 descriptor의 **모양**을 검사한다(내용은 관용)

design D2는 "descriptor는 rename이 덮는다 / 0바이트·잘린 잔재는 관용"만 말하고, 최종
descriptor 이름 자리의 **모양**을 어떻게 다룰지는 말하지 않는다(P1-7②가 "descriptor 자리의
이물은 회수하지 않는다"고만 적었다).

`ValidatePrivateFile`(0600 정규 파일·우리 uid·nlink 1)을 쓰기로 했다. 근거: ① a108의
회수도 `openVerifiedDescriptor`로 같은 형식을 요구한다(이식 원형과 일관). ② 우리 발행
경로는 언제나 chmod 0600 뒤 rename 하므로 이 요구가 자기 잔재를 거부할 수 없다.
③ "이물은 회수하지 않는다"의 구현이 곧 거부다(지우지 않고 기동을 실패시키면 D3 강등이
받는다). 내용은 **파싱하지 않는다** — 0바이트 관용은 그대로다.

뮤테이션 M14가 이 절을 지웠을 때 1차에서는 아무 테스트도 죽지 않았고,
`TestReclaimRefusesADescriptorOfTheWrongShape`를 추가해 죽였다.

### S3. `SweepPrivateStagingLeftovers`는 오류를 돌려주지 않는다

design D2a는 "낯선 엔트리는 오늘처럼 무시한다"까지만 말한다. 위생이 **실패를 보고할지**는
정하지 않았다. 오류 없는 형태로 만들었다 — 이 endpoint에 새 실패 경로를 만들지 않는다는
D2a의 이유(이물 하나가 격리 해제 표면을 매 부팅 지운다)가 "디렉터리를 못 읽었다"에도 같은
힘으로 적용되기 때문이다. 못 치운 잔재는 다음 부팅이 다시 시도한다.

### S4. 회수 후 control 디렉터리는 **언제나 이번 기동이 만든 것**이다

a108 모형을 이식하면 회수가 디렉터리째 지우고 Start가 다시 만든다. 그 결과 두 socket
transport의 `createdControlDir` 플래그가 의미를 잃어 제거했다(실패 정리가 조건 없이
디렉터리를 지운다). command endpoint는 회수를 넣지 않았으므로 그 플래그를 그대로 뒀다.

이것은 동작 변화다: 편집 전에는 "남이 만든 디렉터리는 실패해도 남긴다"였고, 편집 후에는
"회수를 통과한 디렉터리만 존재하므로 지워도 우리 것"이다. 회수가 낯선 엔트리를 거부하는
것이 그 전제이고, 그 거부를 §1.3 핀과 뮤테이션 M1b가 지킨다.

---

## T1 §1-fix — A1 적대 리뷰 판결 구현 중 남긴 것 (2026-08-15)

### I1. a108 원형(`projectionSocketAccepts`)도 같은 owner-write 추정을 가지고 있다

§1-fix F1은 `privateSocketAccepts`의 owner-write 사망 추정을 chmod-then-probe로 바꿨다.
그 절은 a108 `internal/strategyprojectionrpc/transport_unix.go`의 `projectionSocketAccepts`에서
이식한 것이고, **원형에는 그대로 남아 있다.**

**후속 change 등록 후보.** a108 확정 코드는 이 change의 표면이 아니다(pre-edit-gate T1의
선언된 무변경). 병의 모양은 동일하다 — 쓰기 비트가 깎인 **수락 중인** projection socket을
죽었다고 읽고 지운다. 다만 노출은 형제보다 좁다: 그 endpoint의 최종 이름은 이미 D1 의례를
지나 0600으로만 나타나므로, 그 상태에 이르려면 외부에서 chmod가 일어나야 한다.

수정은 기계적이다(chmod 0600 → dial, owner-write 절 삭제). 같은 뮤테이션(F1-N1의 원형판)이
그것을 지킬 수 있다.

### I2. `SweepPrivateStagingLeftovers`에는 F5의 probe가 없다 — 범위를 주석에 적었다

§1-fix F5는 "수락 중인 socket은 이름과 무관하게 unlink하지 않는다"를 **회수**
(`ReclaimStalePrivateEndpoint`)에 통일했다. 위생(`SweepPrivateStagingLeftovers`)에는
그 probe를 넣지 않았다.

**근거:** 이 함수를 쓰는 유일한 endpoint(policy command)는 loopback TCP라 socket을 발행하지
않는다 — 자기 접두의 socket을 만들 수 있는 경로가 없다. 그리고 `private_staging.go`는
플랫폼 무관 파일이라 probe(unix 전용)를 부르려면 `_other.go` 스텁 짝을 새로 들여야 하는데,
그것은 Manager 판결의 범위 밖이다("임의 확장 금지").

**남은 위험:** 다음 사람이 socket을 발행하는 endpoint에 이 위생을 배선하면 그 방어가 없다.
함수 doc에 그 조건을 명시했다("socket을 발행하는 endpoint에서 쓰려면 probe를 먼저 들여야
한다"). 코드가 아니라 문서로 막은 것이므로 **약한 방어**다.

### I3. 원장 §B에 남은 M26(Close의 listener 직접 닫기)은 잠정 생존이다

M22("경합이라 못 죽인다")가 A1에게 반증된 이상, 같은 사유를 든 M26도 그 지위를 잃는다.
지금 원장은 M26을 "경합(잠정)"으로 적고 있다 — "죽지 않는다"가 아니라 "내가 쓴 방법으로는
안 죽었다"는 뜻이다.

**후속 후보.** `Shutdown`이 대개 listener를 닫아 주므로 "누가 닫았는가"를 파일 관측으로
가르는 방법을 아직 찾지 못했다. M22를 죽인 A1의 수법(정리 대상 자리에 관측용 파일을 두고
그 생사를 본다)이 여기에도 통할 가능성이 있다.

### I4. P2-F(산 주인 + 이물 공존의 오귀속)는 잔존 수용이다

A1 P2-F: 산 주인이 살아 있는 디렉터리에 이물이 함께 있으면, 회수는 열거 루프에서 먼저
만나는 **이물**을 이유로 거부하므로 운영자에게는 "이물을 치워라"만 보인다 — 실제로는
주인이 살아 있는 것이 더 근본적인 원인인데도.

**수용한다.** §1-fix F6으로 거부 오류가 위반 엔트리를 지목하므로, 안내는 이제 **실행
가능하다**: 운영자가 그 이물을 치우고 재시작하면 다음 거부가 "the endpoint owner is still
alive"로 바뀐다. 두 원인을 한 번에 보여 주려면 열거를 끝까지 돌고 오류를 모아야 하는데,
그것은 "낯선 엔트리를 만나면 아무것도 건드리지 않고 즉시 멈춘다"는 D2의 성질을 약화한다.
두 번의 부팅으로 두 원인이 순서대로 보이는 쪽을 택했다.

---

## T2 트랙 발견 (병합 편입)

## T2-1 (safe local) — D4 가 열거한 nil 검사 2곳 밖에 production REST 경로가 하나 더 있다

- **발견**: design D4 와 freeze P1-4 는 소비자측 nil 검사를 두 곳으로 열거했다 —
  `cmd/tossctl/httpapi_reader.go:566` 과 `internal/httpapi/strategy_runtime.go:18`.
  실제로는 **세 곳**이고, 셋 중 SSE helper 쪽(`strategy_runtime.go:18`)은
  `StrategyRuntimeSnapshotFunc` 의 것인데 **production 호출자가 없다**
  (`rg StrategyRuntimeSnapshotFunc` = 정의 1 + contract test 1).
  열거에서 빠진 `internal/httpapi/router.go:154` 가 REST `/api/v1/strategy-runtime`
  경로의 진짜 부재 판정이다:

  ```go
  case "strategy-runtime":
      if r.strategyRuntime == nil {
          return strategyprojection.DormantSnapshot(r.now().UTC()), nil
      }
  ```

- **왜 결함인가**: 재부착 wrapper 는 정의상 non-nil 이므로 이 검사가 영원히 거짓이 된다.
  전략 화면을 안 쓰는 배포에서 REST 응답이 **dormant 스냅샷 → 오류**로 바뀐다. 그것이
  http-api-service delta 의 「재부착 전의 응답 값은 기존 부재·unavailable 구분을 그대로
  유지해야 한다(SHALL)」를 REST 경로에서 거짓으로 만든다. a108 D4-2 가 금지한 접힘의
  같은 모양이다.
- **분류**: safe local. 설계 의도(부재/unavailable 구분 보존)를 **바꾸지 않고 완성**하는
  기계적 동일 수정이다 — 세 자리 모두 같은 공유 판정 하나로 교체한다.
- **처리**: `internal/httpapi/router.go` 를 T2 표면에 더하고(1줄), 판정은
  `internal/httpapi/strategy_runtime.go` 에 **한 벌만** 둔다
  (`StrategyRuntimeAbsent` + `StrategyRuntimePresence`). 복사한 검사는 어긋나기 시작한
  검사다(a098 D7.1) — 세 곳에 세 벌을 두지 않는다.
- **Manager 확인 요청**: T2 표면 목록에 `internal/httpapi/router.go` 추가.

## T2-2 (editorial) — 격리 해제 오귀속 문구는 한 곳이 아니라 세 곳이다

- **발견**: design D3a-2 는 `internal/console/exit_quarantine.go:227–229` 한 자리를
  인용했다. 같은 문자열이 **세 벌** 있다 — `:161`(release preview),
  `:196`(release apply), `:229`(`writeQuarantineError` 의 `ErrUnwired` 가지).
- **왜 결함인가**: 인용된 file:line 만 고치면 같은 값의 사본 둘이 살아남고, 운영자가
  실제로 먼저 만나는 것은 preview 경로(:161)다.
- **분류**: editorial — 정정의 단위는 줄이 아니라 **값**이다. 즉시 세 곳 모두 고친다.
- **처리**: 세 문자열을 같은 문구로 바꾸고, 핀 테스트가 **세 경로 전부**를 확인한다.

## T2-3 (editorial) — proposal 의 fatal 줄번호는 freeze 판결대로 :274/:279/:315 다

- proposal 본문의 `:274/:279/:315` 는 현재 HEAD 와 일치함을 AST 로 재확인했다
  (`analysis/function-logic/cmd-tossctl--runenginerun/ast.json` returns[8]=274,
  returns[9]=279, returns[11]=315). freeze P1-9 가 정정한 :294 표기는 남아 있지 않다.
  **수정 불요** — 확인만 기록한다.

## T2-5 (safe local) — 재부착 wrapper 의 로그는 두 goroutine 이 함께 쓴다

- **발견**: 첫 `-race` 전체 실행이 `strategyRuntimeAttachment` 의 로그 writer 에서 데이터
  경합을 잡았다 — 부착 전이는 **시도 goroutine**이, 탈착 전이는 **요청 goroutine**이
  말하는데 둘 다 잠금 없이 `fmt.Fprintf(a.log, …)` 였다.
- **왜 결함인가**: 설계는 「보고는 상태 전이 시 1회」만 정했고 그 보고가 **두 goroutine
  에서 나온다**는 것은 D4 의 백그라운드 시도 결정이 만든 결과다. 설계에 없던 상태다.
- **분류**: safe local — 구현이 만든 경합이고 수정도 구현 안에서 끝난다.
- **처리**: 로그 전용 잠금(`logMu`)과 `report` 한 곳. 상태 잠금(`mu`)을 쓰지 않은 이유는
  느린 writer 하나가 모든 요청의 상태 조회를 멈추게 하면 안 되기 때문이다.

## T2-6 (safe local) — a098 테스트가 alerts 문구의 부분 문자열을 핀하고 있었다

- **발견**: `cmd/tossctl/a098_the_operator_command_names_a_person_test.go:341`
  (`TestTheCommandsRefuseWhenNoEngineIsRunning`)가
  `strings.Contains(err.Error(), "엔진이 없다")` 를 요구한다. D3a-2 의 첫 문구 시안
  (「…엔진이 없거나, …강등 부팅했다」)은 그 부분 문자열을 깨뜨려 이 테스트가 **실제로
  실패**했다.
- **왜 중요한가**: a098 의 의도는 「운영자가 자기 경로를 의심하게 만들지 마라」이고,
  a109 의 요구는 「엔진 부재를 **단정**하지 마라」다. 둘은 충돌하지 않는다 — 충돌한 것은
  단정문을 통째로 지운 첫 시안이었다.
- **분류**: safe local (문구 재작성으로 해소, 기존 테스트는 손대지 않았다).
- **처리**: 문구를 **조건문 두 갈래**로 바꿨다 — 「엔진이 없다면 `engine run` 을 살려라 —
  엔진이 돌고 있다면 강등 부팅한 것이고 원인은 엔진 로그에 있다」. 단정이 사라지면서
  부재일 때의 행동 안내(a098 의 요구)는 그대로 남는다.
- **남는 결합**: ~~a098 은 여전히 부분 문자열로 이 문구에 묶여 있다.~~ **해소됨
  (§2-fix F6)** — a098 의 단정을 `errors.Is(err, errEngineAlertsUnavailable)` 로
  바꿨다. 이제 a098 은 거부의 **정체**를, a109 는 그 거부가 무슨 말을 해야 하는지를
  각자 잰다. 의도가 계속 측정되는지는 M27b 로 확인했다(sentinel 을 날것의 오류로
  되돌리면 a098 이 FAIL).

## T2-7 (safe local) — 표면 밖 신규 테스트 파일 둘

- `internal/httpapi/a109_absence_is_a_state_not_a_nil_test.go` (T2-1 의 REST 자리를 잰다)
- `internal/console/a109_a_degraded_surface_is_not_an_old_build_test.go` (§2.5 문구 핀)
- 둘 다 **신규 파일**이고 기존 파일을 고치지 않는다. 그 패키지의 코드를 바꾼 이상 핀도
  그 패키지에 있어야 한다 — cmd/tossctl 에서 문자열로 흉내 내면 그것은 동작이 아니라
  소스를 재는 테스트가 된다. T1 표면(`internal/positionpolicyrpc`,
  `internal/app/engine`)과는 겹치지 않는다.

## T2-8 (safe local) — 재부착 wrapper 는 `cmd/tossctl` 의 **새 파일**에 둔다

- **발견**: wrapper 를 `httpapi.go` 안에 두자 `check_analysis.py` 가 삽입 지점 바로 위의
  `unavailableStrategyRuntime.Read` 를 「수정됨」으로 잡았다 — 아무도 편집하지 않은
  함수의 증거 묶음을 요구한다(`--unified=0` 의 인접 hunk 가 그 함수의 줄 범위에 닿는다).
- **분류**: safe local. a102·a108 이 같은 이유로 세운 관례(「새 코드는 새 파일에」)를
  따른다.
- **처리**: `cmd/tossctl/httpapi_strategy_attach.go` 신설. `httpapi.go` 에는 부팅 1회
  해석(`resolveStrategyRuntimeReader`)과 그것을 감싸는 `strategyRuntimeReaderFor` 만
  남는다. T2 표면 목록에 없는 **신규 production 파일**이므로 여기 기록한다.

## T2-9 (기록) — 뮤테이션이 테스트 결함 둘을 잡았다

- M10(single-flight)과 M22(SSE 부재 판정)가 **생존**했고, 둘 다 「테스트가 다른 이유로
  초록」이었다. 테스트를 고친 뒤 재측정해 CAUGHT 로 바꿨다 — 상세는
  `mutation-ledger-t2.md` §5.
- 이것이 원장의 실제 수확이다. 두 테스트는 이 change 안에서 **한 번도 실패한 적이
  없었고**, 그래서 그 초록은 증거가 아니었다.

## T2-4 (기록) — D5b 실측은 편집 전에 끝냈고 결과는 "사전 wipe 불요"다

- tasks 3.3 은 §3 소속이라 T2 가 체크하지 않는다. 측정 결과는 완료 보고에 담고
  Manager 가 배포 절차 prose 에 반영한다.
- 요약: 구버전(HEAD, pre-a109) 세 Start × 잔재 6모양 = 18/18 START OK, 잔재는 그대로
  생존. 코드 근거: `internal/positionpolicyrpc`·`internal/app/engine` production 코드에
  `os.ReadDir` 호출 0건(테스트 3건뿐) — 구버전 회수는 디렉터리를 열거하지 않으므로
  `.s-*` 잔재를 **볼 수 없다.**

## T2-10 (정산) — A2 적대 리뷰 P2 여덟 건의 §2-fix 처리

Manager 판결(2026-08-15) 후 브랜치 `feat/a109-t2-degrade-and-redial` 위에서 F1~F9 로
구현했다. RED 가 의미 있는 항목은 RED 커밋을 선행했다.

| 항목 | A2 발견 | 처리 | 측정 |
|---|---|---|---|
| F1 | P2-1 요청자 ctx 취소를 endpoint 실패로 읽는다 | `observe` 가 취소 계열 오류 + **호출자 ctx 종료**를 상태 무변경으로 가린다 | RED 실측(시도 1회·탈착 보고·live client 교체) → GREEN, 뮤테이션 M29 CAUGHT |
| F2 | P2-2 늦은 read 실패가 새 부착을 뒤엎는다 | 읽기가 쓴 **자리의 세대**(seat)가 현재일 때만 전이 | RED 실측(탈착 보고·재-dial 2회) → GREEN, M30 CAUGHT |
| F3 | P2-3 재부착 wake 가 집계 성공에 얹혀 있다 | publisher 루프가 `Refresh` 밖에서 부재 판정을 먼저 부른다(주기 유지·새 ticker 없음) | RED 실측(2초 안에 시도 0회) → GREEN, M31 CAUGHT |
| F4 | P2-4 주석 핀의 공허한 단정 2개 | 고유 구절로 교체 | M24·M25 생존(재현) → M24b·M25b·M26 CAUGHT |
| F5 | P2-5 격리 거부의 **제목**이 빌드를 탓한다 | `quarantineUnwiredTitle` 상수 1벌 + a079 핀을 값 단위로 | RED(세 경로 FAIL) → GREEN, 핀 둘 |
| F6 | P2-6 a098 의 우연한 문자열 결합 | `errors.Is(err, errEngineAlertsUnavailable)` | M27b CAUGHT(의도 유지 확인) |
| F7 | P2-10 T1 병합 대비 nil-safe 계약 | 소비자 쪽 계약 테스트 1개 | M28 CAUGHT |
| F8 | P2-7 부작용 있는 술어 | `StrategyRuntimePresence`·`StrategyRuntimeAbsent` 주석에 명문화 | 문서 |

## T2-11 (잔존 선언) — `exitquarantine.ErrUnwired` 문자열은 고치지 않았다

`internal/exitquarantine/model.go:49` 의

```go
ErrUnwired = errors.New("exit quarantine: the engine control plane does not offer quarantine release")
```

는 F5 가 고친 **운영자 화면 문구**와 같은 오귀속("control plane 이 제공하지 않는다" =
사실상 「빌드가 낡았다」)을 담고 있다. 그럼에도 손대지 않았다:

- 이 값은 화면에 **도달하지 않는다**. 콘솔은 `errors.Is` 로 갈래를 고르고 자기 제목·
  본문(`quarantineUnwiredTitle`·`quarantineUnwiredDetail`)을 쓴다. 운영자가 이 문자열을
  보는 곳은 로그·디버그 출력뿐이다.
- `internal/exitquarantine` 은 T2 표면이 아니고, sentinel 문자열 변경은 그 값을 문자열로
  잡는 다른 패키지의 테스트를 건드릴 수 있다.

**잔존 위험**: 로그를 읽는 사람은 여전히 「control plane 이 제공하지 않는다」를 본다.
후속 change 후보로 남긴다.

## T2-12 (기록) — publisher 서명이 하나 늘었다

`publishHTTPAPISnapshots` 에 `strategyRuntime httpapi.StrategyRuntimeReader` 인자를
더했다(F3). 호출자는 둘뿐이다 — `runHTTPAPI` 와 그 테스트. 재부착 wake 를 집계 밖으로
꺼내려면 루프가 그 값을 **알아야** 하고, 대안(캐시가 source 를 되돌려주게 하기)은
`httpAPISnapshotCache` 에 이 change 와 무관한 결합을 새로 만든다.

## T2-13 (판결 이행 방식 보고) — F2 의 "같은 reader 인가"는 **세대 번호**로 물었다

Manager 판결의 문면은 「`Read` 가 사용한 reader 를 `observe` 에 넘겨 현재 `a.reader` 와
**동일할 때만** 전이」다. 구현은 reader 값을 넘기되 비교를 `==` 가 아니라 **자리의 세대
번호**(`seat uint64`)로 한다. 이유는 안전이다:

- 인터페이스 값의 `==` 는 두 동적 타입이 같고 그 타입이 **비교 불가**면 런타임 패닉이다.
- `httpapi.StrategyRuntimeReader` 의 구현 중 값 타입이 스냅샷을 들고 있으면 정확히 그
  경우가 된다 — `strategyprojection.Snapshot` 은 `map[Market]MarketProjection` 을 들고
  있고, 이 저장소 안에 이미 그 모양의 reader 가 있다(테스트의 `a109Projection`).
- 즉 문면대로 쓰면 「전략 읽기가 실패했을 때 패닉하는 조회 데몬」이 될 수 있다.

세대 번호는 같은 질문에 더 강한 답을 준다: 같은 값이 **다시 앉은** 경우도 새 자리로
세므로, 옛 실패는 언제나 버려진다(보수 방향). 판정 자체는 바뀌지 않는다.
