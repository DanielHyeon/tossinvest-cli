# Issues — a060-soak-restart-uses-this-profile

## I1. 세 change가 같은 결함의 세 얼굴이었다

**분류**: 기록용 — 패턴을 남긴다

| change | 증상 | 원인 |
|---|---|---|
| a059 | 콘솔이 자기가 띄운 **엔진**을 못 찾는다 | spawn이 플래그를 **넘겨서** 패턴이 안 맞음 |
| a060 | 콘솔이 띄운 **서베이**가 아무것도 못 찾는다 | spawn이 플래그를 **안 넘겨서** 프로필이 안 맞음 |

한 문장으로: **부모가 계산한 것과 자식이 보는 것이 다르다.** 콘솔은 자기
`--config-dir`로 경로를 계산해 화면에 그리는데, 자식 프로세스가 그 프로필을 물려받는지
확인하는 계약이 어디에도 없었다.

이 저장소에서 자식을 spawn하는 자리는 이제 둘 다 계약을 갖는다
(`TestTheProcessPatternMatchesWhatTheConsoleSpawns`,
`TestTheSoakSpawnCarriesThisProfile`). 세 번째가 생기면 같은 계약을 먼저 쓸 것.

## I2. `data/`의 405 사이클 기록은 건드리지 않는다

**분류**: 대안 검토 결과, 채택하지 않음

`/var/lib/tossos/data/capability-soak.jsonl`에 405 사이클(507KB)이 있다. 마지막
사이클의 `binary.path`가 `/home/daniel/.local/bin/tossctl`이므로 **컨테이너화 이전
호스트 바이너리**가 쓴 기록이다.

콘솔 프로필(`config/`)로 옮기자는 대안을 검토하고 버렸다.

- 옮기면 호스트 시절 관측을 컨테이너 프로필의 것인 양 만든다. capability attestation은
  게이트 입력이고, 그 증거의 출처를 흐리는 것은 방향이 반대다.
- 콘솔은 이미 그 기록을 읽지 않고 "아직 기록이 없다"라고 말한다. 이 change 뒤에
  버튼이 만드는 기록이 바로 콘솔이 가리키던 그 경로에 생기므로, 화면과 현실이 처음으로
  일치한다.
- 기존 파일은 그대로 남아 있으므로 나중에 사람이 판단할 수 있다.

## I3. 저장소 밖 `soak-autostart.sh`를 운영자 승인으로 은퇴시켰다

**분류**: ~~권고 남김~~ → **처분 완료 (2026-08-03, 운영자 지시로 제거)**

> **처분**: `~/.local/share/tossos/bin/soak-autostart.sh`를 제거했다. 제거 전에 확인한
> 것: 실행 중이 아니고(`pgrep -af soak` 없음), systemd unit·cron·shell rc·
> `console-launch.sh` 어디에도 활성 실행·등록 참조가 없다. 저장소에는 설계 역사와 이전
> 구현 근거를 설명하는 텍스트 참조가 남아 있다.
> 애초에 최대 24시간짜리 1회성 감시자라 7/30 이후로는 만료 상태였다.
> `tools/engine-autostart.sh`가 이 스크립트를 "선례"로 인용하던 주석도 함께 고쳤다.
> 핵심 동작은 아래 기록에 남겼다.
>
> **주의**: 이 제거는 아래 I7의 갱신 공백을 만들지 않는다 — 이 스크립트는 이미
> 죽어 있었고, I7은 그것과 무관한 별개의 경로 불일치다.

제거 전 `~/.local/share/tossos/bin/soak-autostart.sh`의 핵심 동작:

```sh
if ! pgrep -f "tossctl soak run" >/dev/null 2>&1; then
  nohup setsid "$BIN" soak run >> "$LOG" 2>&1 &
```

이 스크립트는 자기 자식을 플래그 없이 띄웠으므로 자기 패턴으로 자기 자식을 봤다.
어긋남은 한 방향뿐이었다: **콘솔이 띄운(플래그 달린) 서베이를 이 스크립트는 못 봤다.**

제거 전 결과는 "호스트 기본 프로필에서 서베이를 하나 더 시작할 수 있다"였다. 두
서베이는 서로 다른 기록에 쓰므로 기록 경합은 없고, 공유되는 것은 계좌의 rate
budget이었다.

### 처분 근거

콘솔이 컨테이너로 옮겨가며 이 스크립트의 역할은 끝났다. 초기 a060 구현 범위에서는
저장소 밖 운영 산출물이어서 손대지 않았지만, 구현·실측 뒤 운영자가 별도로 제거를
지시했다. 실행·등록 참조가 없고 이미 만료된 1회성 감시자임을 확인한 뒤 제거했다.

## I4. `pgrep -a` 의존이 soak에도 생겼다

**분류**: 기록용 (수용) — a059 I4와 같다

`-a`는 procps-ng와 BusyBox 양쪽에 있고 컨테이너의 BusyBox 1.37에서 실측했다. 없는
환경에서는 pgrep이 오류를 반환하고 그것은 열거 실패로 처리된다 — 부재가 아니므로
`restartSoak`은 오류를 돌려주고 조용히 성공하지 않는다.

## I5. 기본 프로필 콘솔의 시야가 좁아졌다

**분류**: 의도된 결과, 기록용

변경 전에는 어떤 프로필의 콘솔이든 플래그 없는 서베이를 자기 것으로 봤다. 이제는
기록 경로가 일치할 때만 본다.

`--config-dir`를 명시한 콘솔과 생략한 autostart가 **같은 기본 경로**를 가리키는 경우는
같은 서베이로 판정된다 — 양쪽을 `resolveSoakRecord`라는 같은 함수에 통과시키기 때문이다
(design D2). 플래그 문자열을 비교했다면 이 경우가 틀렸을 것이다.

## I6. 장중에 서베이와 엔진이 rate budget을 나눠 쓴다 — 실측으로 처음 보였다

**분류**: 신규 관측, 이 change의 범위 밖 — 후속 판단 필요
**발견**: 2026-08-03 09:26 KST, tasks 5.7 실측 중

서베이가 기동하자마자 `GET /api/v1/orders`가 **429(rate limited)**로 거부됐다. 두 번의
기동 모두 같았다.

```
2026-08-03T00:26:02Z  GET /api/v1/orders가 429로 거부되었다. 15s 기다린 뒤 다시 읽는다 (1/2)
2026-08-03T00:26:37Z  GET /api/v1/orders가 429로 거부되었다. 15s 기다린 뒤 다시 읽는다 (1/2)
```

서베이는 설계대로 물러섰다 재시도해 사이클을 완주했다. 결함이 아니라 **경합의 증거**다.

이 사실은 a060 **이전에는 관측될 수 없었다.** 컨테이너에서 서베이가 뜨지도 못했기
때문이다. 버튼을 고친 첫 결과가 "고쳤더니 예산이 빠듯하다"를 보여 준 셈이다.

### 왜 여기서 다루지 않는가

`docs/WORKFLOW.md` §0.4는 "공식 API 호출을 추가하면 rate limit 예산(retry matrix)에
반드시 계상한다"고 정한다. 이 change는 호출을 **추가하지 않았다** — 원래 있어야 했던
호출이 처음으로 실제 발생했을 뿐이다. 그래도 예산 재계상이 필요한지는 별도 판단이며,
답은 서베이 주기·엔진 폴링 주기·브로커 한도를 함께 봐야 나온다.

### 측정 후 상태

측정이 끝난 뒤 서베이를 SIGINT로 정지시켜 측정 전 상태로 되돌렸다. 장중에 새 API
소비자를 조용히 남기지 않기 위해서다. 기록 2 사이클은 남아 있고 콘솔이 그것을 읽는다.

서베이를 상시 가동할지는 운영 판단이다. ~~지금 `gate_verified=true`이고 콘솔의
capability attestation은 이미 충족 상태이므로 급하지 않다.~~ 가동한다면 폐장 시간대에
시작하는 편이 이 429를 피한다.

> **2026-08-03 정정**: "급하지 않다"는 틀렸다. attestation은 **2026-08-29에 만료**되고
> 갱신은 3일 연속 서베이를 요구한다. 상시 가동은 선택이 아니라 요건이다. 다만 지금
> 서베이를 켜도 갱신되지 않는다 — 아래 I7. 429 자체는 새 위험이 아니다: 지금 쓰이는
> attestation도 `114 throttled call(s)`를 포함한 기록에서 발급됐다. 서베이는 물러섰다
> 재시도하도록 설계돼 있고 판정 기준은 throttle을 실패로 세지 않는다.

## I7. attestation 갱신 timer가 콘솔과 다른 파일을 읽는다 — 같은 결함의 네 번째 얼굴

**분류**: 신규 발견, 이 change의 범위 밖 — **STORY-TOS-a063으로 이관 (기한 2026-08-29)**
**발견**: 2026-08-03, I6의 "상시 가동 여부"를 확인하던 중

I1은 "부모가 계산한 것과 자식이 보는 것이 다르다"였다. 여기서는 부모와 자식이
합의한 뒤에도 **갱신하는 쪽이 딴 데를 본다.**

```
콘솔 [soak 재시작] → --config-dir 상속 → ~/.config/tossctl/capability-soak.jsonl   (2 cycle, 오늘)
tossos-attest.timer → 플래그 없음     → ~/.local/share/tossos/capability-soak.jsonl (405 cycle, 7/30 정지)
```

`resolveSoakRecord`는 `--config-dir`가 없으면 `journal.DataDir()`로 떨어지는데,
attestation 경로는 같은 조건에서 `config.DefaultPaths().ConfigDir`로 간다. 그래서 host
timer는 **data의 기록을 읽어 config에 attestation을 쓴다.** 컨테이너 mount가 그
비대칭을 그대로 옮긴다 — `config`는 `~/.config/tossctl`, `data`는
`~/.local/share/tossos`.

timer는 6시간마다 돌고 있고 7/30 이후 **매번 실패한다**(`attest-autorun.log`).

```
2026-08-03T06:25:46+09:00 attest attempt (systemd timer)
Error: soak: the capability soak is not complete:
  - the newest cycle in the record is 3d 6h old and at most 2d 0h is accepted;
    the attestation would be dated today while describing an API nobody has
    looked at since 2026-07-30 — restart `tossctl soak run`
  - unattended credential refresh is proven for 0 consecutive day(s); 3 are required
  - GET /api/v1/accounts was never exercised inside the window
  … (6개 엔드포인트 전부)
```

거부는 올바른 동작이다 — 판정 기준이 정확히 이런 상황을 막으려고 있다. 문제는
**아무도 그 거부를 안 본다**는 것이다.

### 만료가 무엇을 막는가

`automation_gate.enabled=true`, `trading.place=true`인 실계좌 프로필이다. 그래서
interlock 조항 4(attestation 유효)는 살아 있다. `runInterlock`은 engine context를
만들 때 **한 번만** 돈다(`internal/app/engine/engine.go:505`).

- 지금 도는 엔진(pid 16)은 8/29에 멈추지 않는다. 재평가가 없다.
- 8/29 이후 **첫 기동**이 거부된다. `docker compose up -d` 한 번, 컨테이너 recreate
  한 번이면 그 순간이 온다 — a056이 존재하는 이유가 바로 recreate가 엔진을
  재기동한다는 사실이다.
- 사전 경고가 없다. `attest.ExpiresWithin`은 "alerting"용으로 쓰였지만 production
  호출자가 없다(테스트만 참조). 첫 증상이 "배포했더니 엔진이 안 뜬다"가 된다.

### 갱신에 필요한 것 (실측된 판정 기준)

| 기준 | 값 |
|---|---|
| 최신 사이클 나이 | ≤ 2일 |
| 무인 자격증명 갱신 연속일 | ≥ 3일 |
| 토큰 만료 관측 사이클 | ≥ 2회 |
| 커버할 GET 엔드포인트 | 6종 전부 |

즉 **3일 이상 연속 가동한 서베이**가 attest 시점 기준 2일 안에 있어야 한다.
간헐 실행으로는 만들 수 없는 조건이다.

### 후속 계약 (a063에서 수행 — 운영 승인 항목)

두 가지가 함께 가야 하고, 순서가 있다.

1. timer를 콘솔 프로필로 맞춘다: `tossctl soak attest --config-dir "$HOME/.config/tossctl" …`.
   attestation 출력 경로는 지금과 같은 곳(`~/.config/tossctl`)이므로 엔진 게이트가
   보는 파일은 바뀌지 않는다. a060이 세운 "콘솔 프로필이 정본" 원칙과 같은 방향이다.
2. 그 다음 콘솔 [soak 재시작]으로 서베이를 띄우고 **계속 둔다.** 순서를 바꾸면
   3일치를 모으고도 timer가 못 본다.

`~/.config/systemd/user/tossos-attest.service`는 저장소 밖 운영 산출물이라 손대지
않았다. 시작 시점은 I6대로 폐장 시간대가 낫다. 구현과 운영 검증은
`a063-align-attestation-renewal-profile`에서 추적한다.
