# Branch Test Map: `runHTTPAPI`

- Source: `cmd/tossctl/httpapi.go` (120-242)
- 이 change 가 편집한 분기는 **B9·B10** 이고, **B11 은 의도적으로 그대로 둔 것**이라
  「유지」를 재는 행이 따로 있다. RED/GREEN 열은 이 change 에서 관측했는가를 뜻한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil ctx | 간접 — cobra 가 항상 ctx 를 준다 | no | no |
| B2 | 공인 bind·CIDR·TLS 모호 | `TestHTTPAPIBoundaryRejectsPublicAndAmbiguousTLSConfiguration` `httpapi_test.go:49` | no | no |
| B3 | 직접 TLS 모드 | `TestHTTPAPIDefaultsStayLoopbackNoTokenReadOnlyAndRequireTLS` `httpapi_test.go:27` | no | no |
| B4 | 인증서 로드 실패 | — (미고정) | no | no |
| B5 | 리소스 열기 실패 | — (미고정) | no | no |
| B6 | 엔진 디렉터리를 해석했다 | `TestTheDaemonReadsTheDescriptorBesideItsOwnJournal` `a108_the_daemon_treats_absence_and_failure_alike_test.go:202` | no | yes |
| B7 | descriptor 가 stat 된다 (죽은 잔재) | `TestADeadDescriptorDoesNotStopTheDaemon` `a108_the_daemon_treats_absence_and_failure_alike_test.go:128` | yes | yes |
| B8 | descriptor 부재 — 설계된 강등 (대조군) | `TestAnAbsentDescriptorAndADeadOneBootTheSame` `a108_the_daemon_treats_absence_and_failure_alike_test.go:148` | no (기존 동작) | yes |
| B9 | dial 실패 → 강등 (**이 change 가 바꾼 분기**) | `TestADeadDescriptorDoesNotStopTheDaemon` · `TestAnAbsentDescriptorAndADeadOneBootTheSame` | yes | yes |
| B10 | dial 성공 → 클라이언트 대입 (**이 change 가 만든 else**) | — (미고정: 살아 있는 엔진 소켓이 필요하고 그것은 겹1 T1 범위다. 뮤테이션 M8 이 이 분기를 fatal 쪽으로 되돌리면 B9 테스트 둘이 죽는다) | no | no |
| B11 | stat 오류가 NotExist 가 아니다 → **fatal 유지** | `TestAnUninspectableDescriptorIsStillFatal` `a108_the_daemon_treats_absence_and_failure_alike_test.go:179` | 해당 없음(유지 핀) | yes |
| B12 | 스냅샷 캐시 실패 | — (미고정) | no | no |
| B13 | 스트림 생성 실패 | — (미고정) | no | no |
| B14 | mutation 라우트 구성 실패 | `TestHTTPAPIMutationsStayAbsentUntilBothSecurityArtifactsAreConfigured` `httpapi_mutation_test.go:19` | no | no |
| B15 | 라우터 생성 실패 | — (미고정) | no | no |
| B16 | 서버 생성 실패 | — (미고정) | no | no |
| B17 | 직접 TLS 면 TLSConfig 설정 | — (미고정) | no | no |
| B18 | 직접 TLS 면 ListenAndServeTLS | — (미고정) | no | no |
| B19 | serve 종료 vs ctx 종료 | `TestADeadDescriptorDoesNotStopTheDaemon` (ctx 만료 → graceful shutdown → nil) | no | yes |
| B20 | `http.ErrServerClosed` | — (미고정) | no | no |
| B21 | Shutdown 실패 | — (미고정) | no | no |
| B22 | serve 오류가 ErrServerClosed 아님 | — (미고정) | no | no |

## RED 을 실제로 관측한 순서 (B9)

1. 사고 당일의 디스크 상태를 그대로 만드는 fixture 를 썼다: control 디렉터리 0700 +
   `endpoint.json` 0600(schemaVersion v1·socket·token·pid=16) + **socket 없음**.
   seam 은 쓰지 않았다 — 이 실패는 `Dial` 자신의 socket 검사이고 T1 이 편집 중인
   회수 규칙과 무관하다.
2. 그 상태에서 `runHTTPAPI` 를 돌렸더니
   `httpapi: strategy runtime projection unavailable: strategy projection runtime:
   socket is invalid` 로 **오류 반환**했다 — 컨테이너가 `Restarting (1)` 로 돌던 그 모양.
   같은 실행에서 `TestAnUninspectableDescriptorIsStillFatal` 는 통과했고(유지 핀은 처음부터
   초록이어야 한다), `TestAnAbsentDescriptorAndADeadOneBootTheSame` 의 **부재 절반은
   통과**한 뒤 죽은 절반에서 실패했다 — harness 가 실제로 데몬을 세운다는 증거다.
3. 강등을 구현하니 4건 전부 GREEN.

## 미고정으로 남긴 분기 (사유)

B4·B5·B12·B13·B15~B18·B20~B22 는 이 change 가 편집하지 않았고, 전부 「생성 실패 → return」
또는 TLS 종단 분기다. 고정하려면 TLS fixture 와 각 생성자의 실패 주입이 필요하며
a108 의 범위 밖이다 — 선언된 생략.

B10 은 이 change 가 만든 `else` 인데도 미고정이다. 이유를 숨기지 않고 적는다:
살아 있는 projection 소켓이 필요하고, 그것을 만드는 것은 겹1(T1)이 지금 재작성 중인
`strategyprojectionrpc.Start` 다. 대신 뮤테이션 M8 이 이 분기를 사고 당시 모양으로
되돌렸을 때 B9 테스트 둘이 죽는 것을 확인했다(`mutation-ledger-t2.md`).
