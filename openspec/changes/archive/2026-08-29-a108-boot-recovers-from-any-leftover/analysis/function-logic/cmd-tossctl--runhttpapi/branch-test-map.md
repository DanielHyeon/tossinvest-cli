# Branch Test Map: `runHTTPAPI`

- Source: `cmd/tossctl/httpapi.go` (121-217)
- **gstack Fix 라운드에서 이 함수의 전략 블록이 통째로 나갔다.** 부팅의 전략 판정은
  이제 `strategyRuntimeReaderFor`(같은 파일의 새 함수)가 지고, 이 함수는 그 결과를
  받아 두 곳(`reader.strategyRuntime` · `httpapi.NewRouter`)에 넘긴다. 그래서 분기가
  22 → 16 으로 줄었고, 사라진 여섯은 전부 「descriptor 를 어떻게 읽는가」였다.
  뽑아낸 이유는 그 판정을 **직접 측정**하기 위해서다 — 데몬 안에 묻혀 있으면
  「어느 실패가 어느 화면 값이 되는가」를 재려고 TLS·인증·HTTP 를 다 지나야 한다.
- 편집 이력: 원 라운드가 B9(dial 실패)를 fatal → 강등으로, Fix 라운드(6.8①)가
  B11(조사 불가)을 fatal → 강등으로 뒤집었다. 그 두 분기는 지금 이 표에 없다 —
  옮겨 간 곳의 행이 아래 「옮겨 간 판정」 절이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil ctx | 간접 — cobra 가 항상 ctx 를 준다 | no | no |
| B2 | 경계 구성 실패(공인 bind·CIDR·TLS 모호) | `TestHTTPAPIBoundaryRejectsPublicAndAmbiguousTLSConfiguration` `httpapi_test.go:49` | no | no |
| B3 | 직접 TLS 모드 | `TestHTTPAPIDefaultsStayLoopbackNoTokenReadOnlyAndRequireTLS` `httpapi_test.go:27` | no | no |
| B4 | 인증서 로드 실패 | — (미고정) | no | no |
| B5 | 리소스 열기 실패 | — (미고정) | no | no |
| B6 | 스냅샷 캐시 실패 | — (미고정) | no | no |
| B7 | 스트림 생성 실패 | — (미고정) | no | no |
| B8 | mutation 라우트 구성 실패 | `TestHTTPAPIMutationsStayAbsentUntilBothSecurityArtifactsAreConfigured` `httpapi_mutation_test.go:19` | no | no |
| B9 | 라우터 생성 실패 | — (미고정) | no | no |
| B10 | 서버 생성 실패 | — (미고정) | no | no |
| B11 | 직접 TLS 면 TLSConfig 설정 | — (미고정) | no | no |
| B12 | 직접 TLS 면 ListenAndServeTLS | — (미고정) | no | no |
| B13 | serve 종료 vs ctx 종료 | `TestADeadDescriptorDoesNotStopTheDaemon` `a108_the_daemon_treats_absence_and_failure_alike_test.go:151` (배너를 본 테스트가 ctx 를 끊고, graceful shutdown 이 nil 로 돌아온다) | no | yes |
| B14 | `http.ErrServerClosed` | — (미고정) | no | no |
| B15 | Shutdown 실패 | — (미고정) | no | no |
| B16 | serve 오류가 ErrServerClosed 아님 | — (미고정) | no | no |

## 옮겨 간 판정 — `strategyRuntimeReaderFor`

이 함수가 더 이상 직접 갖지 않는 네 상태의 핀은 그대로 살아 있고, 대상만 바뀌었다.
(그 함수는 이 change 가 **새로 만든** 것이라 자기 FLM 을 요구하지 않는다 —
check_analysis 는 base 에 존재하던 함수의 변경만 증거를 요구한다.)

| 상태 | 결과 | Test |
|---|---|---|
| 엔진 디렉터리 해석 실패 | 경고 + nil | — (미고정: `engineJournalDir` 는 `--config-dir` 가 있으면 실패하지 않는다) |
| descriptor 부재 | 조용한 nil = `NOT_CONFIGURED` | `TestAnAbsentDescriptorAndADeadOneBootTheSame` · `TestADialFailureRendersUnavailableRatherThanNotConfigured/descriptor_가_없다` |
| descriptor 조사 불가(비-NotExist) | 경고 + nil | `TestAnUninspectableDescriptorDegradesLikeTheConsole` |
| dial 실패 | 경고 + **항상-에러 sentinel** = `RUNTIME_UNAVAILABLE` | `TestADeadDescriptorDoesNotStopTheDaemon` · `TestADialFailureRendersUnavailableRatherThanNotConfigured` · `TestASocketFileWithNoOwnerDegradesTheDaemon` |
| dial 성공 | 클라이언트 | — (미고정: 살아 있는 엔진 소켓이 필요하다. 뮤테이션 M16·M18 이 양방향을 잰다) |

## RED 을 실제로 관측한 순서 (원 라운드, B9 이던 분기)

1. 사고 당일의 디스크 상태를 그대로 만드는 fixture 를 썼다: control 디렉터리 0700 +
   `endpoint.json` 0600(schemaVersion v1·socket·token·pid=16) + **socket 없음**.
2. 그 상태에서 `runHTTPAPI` 를 돌렸더니
   `httpapi: strategy runtime projection unavailable: strategy projection runtime:
   socket is invalid` 로 **오류 반환**했다 — 컨테이너가 `Restarting (1)` 로 돌던 그 모양.
3. 강등을 구현하니 4건 전부 GREEN.

## Fix 라운드(6.8①)에서 RED 을 관측한 순서 — **커밋으로 남았다**

테스트만 담은 커밋 `d8b27021` 이 GREEN 커밋 `aecc03e0` 보다 먼저 있다.

1. `TestAnUninspectableDescriptorDegradesLikeTheConsole` —
   `httpapi = httpapi: inspect strategy runtime projection: … not a directory`.
2. `TestASocketFileWithNoOwnerDegradesTheDaemon` — 「S3 에서 강등이 발동하지 않았다」.
   그때의 `Dial` 은 socket 의 Lstat·모드·perm 만 봐서 S3 를 성공으로 읽었다.
   T1-fix 의 connect probe 가 병합되자 **손대지 않고** GREEN 이 됐다 — 그 RED 가
   병합 검증이었다(정산: `mutation-ledger-t2.md`).

## gstack 라운드에서 관측한 RED

`TestADialFailureRendersUnavailableRatherThanNotConfigured` 의 두 행이
`KR/US 판정 = "NOT_CONFIGURED", want "RUNTIME_UNAVAILABLE"` 로 실패했다. 같은 표의
부재 대조군은 그때도 통과했다 — 즉 잰 것은 「전부 실패한다」가 아니라 **구분**이다.

## 미고정으로 남긴 분기 (사유)

B4~B7·B9~B12·B14~B16 은 이 change 가 편집하지 않았고, 전부 「생성 실패 → return」
또는 TLS 종단 분기다. 고정하려면 TLS fixture 와 각 생성자의 실패 주입이 필요하며
a108 의 범위 밖이다 — 선언된 생략.
