# Branch Test Map: `openHTTPAPIResources`

- Source: `cmd/tossctl/httpapi.go` (592-661)
- 이 change 가 편집한 분기는 **B9 하나**(그 `else` 절)다. gstack 리뷰 A3 — 무경고 강등 금지.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | journal 경로 해석 실패 | — (미고정) | no | no |
| B2 | 원장을 열었다 / 없다 | `TestADeadDescriptorDoesNotStopTheDaemon` `a108_the_daemon_treats_absence_and_failure_alike_test.go:151` (원장 없는 tmpdir 에서 데몬이 뜬다) | no | yes |
| B3 | 성과 저장소를 열었다 / 없다 | 간접 — 같은 테스트 | no | yes |
| B4 | 최적화 registry 실패 | — (미고정) | no | no |
| B5 | 최적화 DB 열기 실패 | — (미고정) | no | no |
| B6 | accountRef 클로저의 오류 전달 | 간접 | no | no |
| B7 | journalReader 가 있으면 performanceSource | 간접 | no | no |
| B8 | adoption seam 이 있다 | 간접 | no | no |
| B9 | **엔진 디렉터리를 못 정했다 → 경고 + 강등** (편집) | — (미고정, 아래 사유) | no | no |
| B10 | `reader.validate()` 실패 | — (미고정) | no | no |

## B9 를 테스트로 고정하지 않은 이유 (선언된 생략)

`engineJournalDir` 은 `root.configDir` 가 비어 있지 않으면 **실패할 수 없고**(그대로
돌려준다), 비어 있으면 `journal.DefaultPath()` 로 간다. 그 함수를 실패시키려면 XDG
환경변수와 홈 디렉터리 해석을 프로세스 전역으로 망가뜨려야 하는데, 그것은 병렬로 도는
다른 테스트의 경로까지 오염시킨다(`testenv.Isolate` 가 막으려는 바로 그 상태).

a108 의 모든 데몬 테스트는 `--config-dir` 를 넘기므로 이 분기에 **도달하지 않는다.**
그래서 이 행은 「고정했다」가 아니라 「고정하지 않았다」로 남기고, 대신 편집이
안전한 이유를 코드 모양으로 적는다: 새로 생긴 side effect 는 stderr 한 줄뿐이고,
반환값·reader 상태·거절 조건은 하나도 바뀌지 않았다(`git diff` 로 확인 가능한 크기).

같은 원인이 `strategyRuntimeReaderFor` 에서도 같은 모양으로 경고를 찍고, 그쪽 분기도
같은 이유로 미고정이다 — 숨기지 않고 두 곳 모두에 적는다.

## 미고정으로 남긴 나머지 (사유)

B1·B4·B5·B10 은 이 change 가 편집하지 않았고, 전부 「열기 실패 → 닫고 거절」이다.
고정하려면 파일시스템 오류 주입이 필요하며 a108 의 범위 밖이다 — 선언된 생략.
