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

## 5. 배포와 운영

- [x] 5.1 **첫 배포 완료 (2026-08-12 08:19:55 KST, image `e3cc624401ae`).**
  ~~"서베이가 자동으로 돌아오는지 확인한다"~~ — **이 문장은 첫 배포에 적용할 수 없다.**
  5.2가 말하는 대로 이 배포 시점에는 키가 없어 OFF이므로 돌아오지 않는 것이 정상이다.
  원문은 5.2와 모순됐고, 여기서 고친다. 이 배포가 실제로 검증한 것은 **OFF 경로**다.
- [x] 5.2 **첫 배포에서는 켜지지 않는다 — 실운영에서 확인했다.** 재생성 뒤 콘솔 로그에
  `엔진 자동 시작: 엔진을 시작했다`는 있고 **soak 줄은 없다.** 컨테이너 안에도 console(pid 7)과
  `engine run`(pid 16)뿐이고 soak 프로세스는 없다. `runConfiguredSoakAutostart`의 OFF 분기가
  빈 문자열을 반환한 것이며 `TestConfiguredSoakAutostartOffDoesNotStart`와 같은 동작이다.
- [ ] 5.4 **ON 경로는 아직 미검증이다.** 운영자가 대시보드에서 **[soak 재시작]을 한 번**
  눌러야 승인이 기록되고(audit `soak.autostart`), 그 **다음 배포**에서야 자동 복구가
  관측된다. 이 change의 목적 자체는 그때 검증된다 — 지금은 절반만 확인됐다.
- [ ] 5.5 **리뷰 수정은 아직 배포되지 않았다.** 돌고 있는 이미지 `e3cc624401ae`에는
  `bootSurvey`가 없다. 다만 **컨테이너 배포에서는 R1이 도달 불가능하다** — 컨테이너를
  재생성하면 서베이도 함께 죽으므로 부팅 시점에 "이미 돌고 있는 서베이"가 존재할 수 없다.
  R1은 콘솔만 재기동되는 host 설치에서 발동한다. 따라서 지금 버튼을 눌러 승인을 기록하는
  것은 안전하고, 다음 배포가 곧 5.4의 검증이 된다.
- [x] 5.3 운영 문서 — `docs/operations.md`의 「LIVE 주문과 엔진 자동 시작」 바로 뒤에
  「capability 서베이 자동 시작」 절을 넣었다. 재시작 버튼이 승인을 영속시킨다는 것,
  끄려면 config를 고친다는 것, 엔진 다음이라는 순서와 그 이유를 적었다.

## 범위 밖

- soak 정지 버튼과 끄기 UI. 콘솔에는 원래 정지 표면이 없다.
- 서베이가 재는 내용(endpoint·주기·판정). a100이 endpoint를 늘렸고 이 change는 그것을 모른다.
- attestation 재발급 절차 자체(a100 D11).
