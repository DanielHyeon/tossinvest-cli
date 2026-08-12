# Tasks — a101-the-soak-outlives-a-deploy

## 0. 착수 전 조건

- [x] 0.1 `base-commit.txt` 고정과 `make sdd-sync`. base는 **a100 커밋 `4c6927ea`**다 —
  a101은 a100 위에 쌓였고, 처음 잡았던 `ce78b0db`(a100의 문서 커밋)로 두면 a100의 코드
  변경이 a101의 diff로 새어 들어온다. 재고정은 커밋 **후**에 한다.
- [x] 0.2 **편집 전 Function Logic Map.** 기존 함수 내부를 바꾸는 대상은
  `cmd/tossctl.runConsole`(`console.go:211`) 하나다 — 기동 시 autostart 호출 한 줄과
  `RestartSoak` seam 래핑이 그 안에 들어간다. 나머지는 전부 새 파일의 새 함수다.
- [x] 0.3 High-risk 판정. **인증 경로다**(§0-5) — 이 배선의 산출물이 attestation을 갱신하는
  프로세스이기 때문이다. Pre-Edit 선언을 남긴다.

## 1. 설정 키 (RED 먼저)

- [x] 1.1 `LoadSoakAutostart` — 파일 없음·`soak` 블록 없음·키 없음이 **전부 false**다.
  파싱 실패는 에러다(false를 조용히 반환하지 않는다).
- [x] 1.2 `SaveSoakAutostart` — `soak.autostart` **하나만** 쓴다. 기존 키를 건드리지 않는다.
  `soak` 블록이 없으면 만든다.
- [x] 1.3 구버전 호환 — 파싱이 strict가 아님을 확인했다(`DisallowUnknownFields` 없음). 구버전
  코드가 `soak` 블록을 만나도 무시한다.
- [x] 1.4 `engine.autostart`와 독립임을 고정한다 — 한쪽을 저장해도 다른 쪽이 바뀌지 않는다.

## 2. 기동 배선

- [x] 2.1 `runConfiguredSoakAutostart(load, start) string` — `runConfiguredEngineAutostart`의
  형태를 그대로 따른다. 분기 넷: load 없음 / 읽기 실패 / OFF / 기동 실패.
- [x] 2.2 **기동 실패가 콘솔을 막지 않는다**를 테스트로 고정한다.
- [x] 2.3 `runConsole`에 호출을 넣는다. 엔진 autostart 바로 다음이다 — 순서가 뒤바뀌면
  서베이가 엔진보다 먼저 rate budget을 쓴다.

## 3. 승인 영속

- [x] 3.1 `RestartSoak` seam을 감싸 **성공했을 때만** `SaveSoakAutostart(true)`를 부른다.
- [x] 3.2 audit 항목 `soak.autostart`를 추가하고 기록한다.
- [x] 3.3 **기록 실패가 재시작 결과를 뒤집지 않음**을 고정한다. 실패는 반환 문자열에 덧붙인다.
- [x] 3.4 재시작이 실패하면 승인을 기록하지 않음을 고정한다.

## 4. 게이트

- [x] 4.1 편집한 기존 함수의 FLM 재생성과 SHA 대조.
- [x] 4.2 `check_analysis.py --change a101-the-soak-outlives-a-deploy` — **통과**
  (`evidence complete or diff-proven exempt`). a100을 `4c6927ea`로 커밋하고 이 change의
  `base-commit.txt`를 그 커밋으로 다시 고정한 뒤에야 통과한다. 그 전에는 a100의 미커밋 Go
  변경이 a101의 diff에 섞여 들어왔다(review.md 이탈 3).
- [x] 4.3 `go test` + `-race` + `go vet` — `internal/config`, `internal/console`, `internal/audit`,
  `cmd/tossctl`.
- [x] 4.4 `openspec validate --all --strict` — 86 passed, 0 failed
- [x] 4.5 `make sdd-sync` → `make sdd-check` (**exit 0**). CodeGraph hard-evidence index가
  worktree와 일치한다. `sdd-sync` 자체는 exit 1인데 사유는 advisory `codegraphcontext
  update`의 300s timeout 하나뿐이고 hard evidence는 기록된다.
- [x] 4.6 gstack 독립 리뷰 — **결함 1건을 잡았고 고쳤다**(review.md R1·R2). 서브에이전트가
  아니라 다른 모델(codex, read-only)로 돌렸다. 같은 모델의 두 번째 통과는 독립이 아니다.
- [x] 4.8 **리뷰 수정**(2026-08-12). 부팅 경로가 버튼의 seam을 그대로 쓰면 안 된다.
  `bootSurvey` 신설(RED 선행, 커버리지 100.0%), 부팅에서 `PrepareSpawn` 제거,
  `runConsole` FLM·branch-test-map·pre-edit-gate 재생성(분기 44 유지, SHA `08562e7541ddf0d8`).
- [x] 4.7 PM 동기화(`STORY-TOS-a101`) — story 파일, `_registry.yaml`, `FEAT-TOS-004`의
  역링크. 역링크를 빠뜨리면 `generate_master_tracker.py`가 `feature reverse link missing`으로
  거부한다.
- [x] 4.9 **`make gate`를 끝까지 돌리고 결과를 기록했다 — 통과가 아니다.** 9단계 중 8단계가
  통과하고 **7/9 `make test`만 실패**한다. 실패 4건은 전부 a099의 RED 핀
  (`a099_claim_excludes_the_second_sender_test.go`, 커밋 `7f3cbb03`)이고 a101이 만든 것이
  아니다. **`7f3cbb03` 이후 이 브랜치에서는 어느 change로도 게이트가 통과하지 않는다** —
  a100 7.5가 열려 있는 이유도 같고, 그 사실이 어디에도 적혀 있지 않았다.
  닫는 조건은 a099의 GREEN이다. 상세는 review.md 「게이트 최종 결과」.
  도중에 두 가지를 고쳤다: CodeGraph 지문 재sync, PM 트래커 재생성(`fc501c81`).

## 5. 배포와 운영

- [x] 5.1 **첫 배포 완료 (2026-08-12 08:19:55 KST, image `e3cc624401ae`).**
  ~~"서베이가 자동으로 돌아오는지 확인한다"~~ — **이 문장은 첫 배포에 적용할 수 없다.**
  5.2가 말하는 대로 이 배포 시점에는 키가 없어 OFF이므로 돌아오지 않는 것이 정상이다.
  원문은 5.2와 모순됐고, 여기서 고친다. 이 배포가 실제로 검증한 것은 **OFF 경로**다.
- [x] 5.2 **첫 배포에서는 켜지지 않는다 — 실운영에서 확인했다.** 재생성 뒤 콘솔 로그에
  `엔진 자동 시작: 엔진을 시작했다`는 있고 **soak 줄은 없다.** 컨테이너 안에도 console(pid 7)과
  `engine run`(pid 16)뿐이고 soak 프로세스는 없다. `runConfiguredSoakAutostart`의 OFF 분기가
  빈 문자열을 반환한 것이며 `TestConfiguredSoakAutostartOffDoesNotStart`와 같은 동작이다.
- [x] 5.4 **승인 기록 확인 (2026-08-12 08:44:51 KST).** 운영자가 [soak 재시작]을 눌렀고
  세 곳이 전부 일치한다.
  - config: `{"soak": {"autostart": true}, "engine": {…, "autostart": true, …}}` — **splice가
    `engine` 블록을 건드리지 않았다.** 두 키의 독립성이 실파일에서 확인됐다(1.4의 단위 테스트가
    주장하던 것).
  - audit: `{"action":"soak.autostart","setting":"soak.autostart","old":"false","new":"true"}`
  - 프로세스: `tossctl --config-dir … --session-file … soak run` (pid 79) — **프로필 플래그가
    붙어 있다.** a060이 고친 결함(플래그 없는 자식)이 재현되지 않았다.

  **audit 줄이 둘이다.** 버튼이 3.6초 간격으로 두 번 눌렸고 두 번째는 `old:"true" new:"true"`다.
  그 사이 첫 서베이는 정상 종료되고 새 서베이가 섰다 — 버튼의 설계된 동작이지만, **누를 때마다
  승인을 다시 쓰므로 값이 바뀌지 않아도 audit 줄이 늘어난다.** 결함은 아니고 소음이다.
- [x] 5.6 **자동 복구 검증 완료 (2026-08-12).** `bootSurvey`의 **두 분기가 모두** 실운영에서
  발동했고, 둘 다 의도한 쪽으로 갔다.

  | 시각 (KST) | 사건 | 콘솔 로그 |
  |---|---|---|
  | 08:51:08 | 콘솔 재시작(`execve`, pid 7 유지) | `soak 자동 시작: 이미 실행 중이다 (pid 260). 새로 시작하지 않았다` |
  | 09:08:13 | **컨테이너 재생성**(release override 적용) | `soak 자동 시작: 실행 중인 soak을 찾지 못했다. 새로 시작했다` |

  두 번째가 이 change가 존재하는 이유다 — 2026-08-11 14:00Z에 측정한 결함(배포가 서베이를
  조용히 죽인다)이 닫혔다. 새 컨테이너에 `soak run`(pid 27)이 **아무도 버튼을 누르지 않았는데**
  서 있다.

  첫 번째는 **리뷰 수정이 없었다면 반대로 갔을 자리다.** 바로 위 줄에서 엔진이 같은 판정을
  내리고 있다(`엔진 자동 시작 거부: 엔진이 이미 실행 중이다 (pid 16, …)`) — 두 autostart가
  이제 같은 규칙을 따른다.
- [x] 5.7 **재시작이 3일 streak을 0으로 되돌린다는 것을 실측했다(운영 주의사항).**
  09:04:54의 [soak 재시작]이 만든 사이클이 `credentials FAILED (rate_limited)`로 끝났고
  (accounts·buying-power·holdings가 429, 뒤의 여섯은 ok) `credential streak 0 of 3`가 됐다.
  버튼은 `PrepareSpawn`으로 토큰 캐시를 지우고 즉시 새 사이클의 read 버스트를 낸다 — 직전
  order walk(30요청·64초)가 연 429 penalty window와 겹치면 자격증명 판정이 실패한다.
  ⇒ **autostart가 켜진 뒤로는 재시작이 필요 없고, 누르면 시계가 0으로 돌아간다.**
  운영 문서에 반영했다.
- [x] 5.5 **리뷰 수정 배포 완료** (2026-08-12 08:43:07 KST, image `14b567319e16`).
  같은 안전 창(두 시장 다 닫힘) 안에서 두 번째 재생성을 했다. 콘솔·엔진 재기동 확인,
  soak 줄은 여전히 없다(키가 없으므로 OFF — 5.2와 같은 결과).
  참고로 **컨테이너 배포에서는 R1이 애초에 도달 불가능하다** — 컨테이너를 재생성하면
  서베이도 함께 죽으므로 부팅 시점에 "이미 돌고 있는 서베이"가 존재할 수 없다.
  R1이 발동하는 곳은 콘솔만 재기동되는 host 설치다.
- [x] 5.3 운영 문서 — `docs/operations.md`의 「LIVE 주문과 엔진 자동 시작」 바로 뒤에
  「capability 서베이 자동 시작」 절을 넣었다. 재시작 버튼이 승인을 영속시킨다는 것,
  끄려면 config를 고친다는 것, 엔진 다음이라는 순서와 그 이유를 적었다.

## 범위 밖

- soak 정지 버튼과 끄기 UI. 콘솔에는 원래 정지 표면이 없다.
- 서베이가 재는 내용(endpoint·주기·판정). a100이 endpoint를 늘렸고 이 change는 그것을 모른다.
- attestation 재발급 절차 자체(a100 D11).
