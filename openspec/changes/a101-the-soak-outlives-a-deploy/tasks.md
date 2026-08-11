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
- [ ] 4.5 `make sdd-sync` → `make sdd-check`
- [ ] 4.6 gstack 독립 리뷰.
- [ ] 4.7 PM 동기화(`STORY-TOS-a101`).

## 5. 배포와 운영

- [ ] 5.1 **배포 자체가 이 change의 첫 검증이다.** 재빌드 + 컨테이너 재생성 후
  서베이가 자동으로 돌아오는지 확인한다. 돌아오지 않으면 이 change는 실패다.
- [ ] 5.2 **첫 배포에서는 아직 켜져 있지 않다.** 키가 없으므로 false이고, 운영자가 soak
  재시작을 한 번 눌러야 승인이 기록된다. 그 다음 배포부터 자동으로 살아난다.
- [ ] 5.3 운영 문서에 한 줄 — 재시작 버튼이 승인을 영속시킨다는 것과, 끄려면 config를
  고친다는 것.

## 범위 밖

- soak 정지 버튼과 끄기 UI. 콘솔에는 원래 정지 표면이 없다.
- 서베이가 재는 내용(endpoint·주기·판정). a100이 endpoint를 늘렸고 이 change는 그것을 모른다.
- attestation 재발급 절차 자체(a100 D11).
