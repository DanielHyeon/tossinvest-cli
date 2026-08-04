# a060 · 콘솔의 soak 재시작은 콘솔의 프로필을 쓴다

- **Feature**: `FEAT-TOS-004` — Operator console controls and visibility
- **Story**: `STORY-TOS-a060`
- **Spec**: `operator-console`

## Why

**콘솔의 [soak 재시작] 버튼이 프로덕션에서 100% 실패하고 있다.** 2026-08-03 a059
`issues.md` I1이 "soak에도 같은 구멍이 있고, spawn에 `--config-dir`를 넘기는 변경이
생기면 재현된다"고 예측했는데, 확인해 보니 예측이 아니라 이미 벌어진 일이었다.
방향만 반대다 — 플래그를 **넘기지 않아서** 깨졌다.

컨테이너의 `soak.log`는 같은 줄만 반복한다.

```
Error: soak: no Open API credentials — the survey measures whether they renew
themselves unattended, so it cannot start without them.
Run `tossctl openapi login` or set TOSSCTL_OPENAPI_KEY/TOSSCTL_OPENAPI_SECRET
```

자격증명은 있다. `/var/lib/tossos/config/openapi-credentials.json`에.
콘솔이 spawn한 자식이 거기를 안 볼 뿐이다.

```go
// 콘솔은 자기 프로필로 기록 경로를 정하고
soakRecord, _ := resolveSoakRecord(root, "")     // /var/lib/tossos/config/…
RestartSoak: func() (string, error) { return restartSoak(soakRecord, …) }

// 자식은 플래그 없이 뜬다
exec.Command(binary, "soak", "run")              // → 기본 프로필 /var/lib/tossos/data
```

## 관측된 결과

| | 콘솔이 보는 곳 | 자식이 쓰는 곳 |
|---|---|---|
| 자격증명 | `config/openapi-credentials.json` (있음) | `data/` (없음) → **즉시 종료** |
| 기록 | `config/capability-soak.jsonl` (없음) | `data/capability-soak.jsonl` |
| 로그 | `config/soak.log` (실패 반복) | 같은 파일 |

콘솔 화면은 이렇게 말한다.

> 아직 기록이 없다. `tossctl soak run`으로 시작하라. (/var/lib/tossos/config/capability-soak.jsonl)

정직한 문장이다. 다만 그 아래 [soak 재시작] 버튼을 눌러도 그 기록은 영원히 생기지
않는다. `data/`에는 405 사이클짜리 기록(507KB, 7/30까지)이 있는데, 그건 컨테이너화
이전에 **호스트** 바이너리가 쓴 것이다(마지막 사이클의 `binary.path`가
`/home/daniel/.local/bin/tossctl`). 지금 콘솔과는 무관한 유물이다.

## 왜 a059와 한 몸인가

a059는 엔진에서 **플래그를 넘기기 때문에** 패턴이 안 맞았다. soak은 **플래그를 안 넘겨서**
프로필이 안 맞는다. 같은 비대칭의 양면이고, 한쪽을 고치면 다른 쪽이 따라온다 —
soak spawn이 플래그를 물려받는 순간 `soakProcessPattern = "tossctl soak run"`이
a059 이전의 엔진 패턴과 똑같이 깨진다.

그래서 셋은 나눌 수 없다.

1. spawn이 프로필을 물려받는다 (지금 깨진 것을 고친다)
2. 패턴이 그 명령줄에 맞는다 (1이 만드는 결함을 막는다)
3. 소유 판정이 남의 서베이를 걸러 낸다 (2가 만드는 위험을 막는다)

## What Changes

`engineArgs`와 같은 모양의 `soakArgs`를 만들어 spawn이 콘솔의 `--config-dir`·
`--session-file`을 물려받게 한다. 패턴을 `tossctl( .*)? soak run`으로 넓히고,
a059가 만든 소유 판정을 두 프로세스가 **공유하는 헬퍼**로 올려 soak에도 적용한다.
soak의 소유 기준은 journal 디렉터리가 아니라 **기록 경로**다 — 서베이의 정체를
정하는 것은 그것이 append하는 기록이기 때문이다.

그리고 a059 `issues.md` I1이 요구한 계약 테스트를 넣는다: 패턴을 셸 스크립트가 아니라
**spawn 경로가 만드는 argv**에 묶는다.

## Non-Goals

- `data/`의 405 사이클 기록을 옮기거나 지우지 않는다. 콘솔은 이미 그것을 읽지 않고,
  옮기면 호스트 시절 기록을 컨테이너 프로필의 것인 양 만든다 (`issues.md` I2).
- `soak run`의 동작·주기·판정 기준을 바꾸지 않는다. 조회 전용 성질도 그대로다.
- 구현 범위에서는 설치된 `~/.local/share/tossos/bin/soak-autostart.sh`를 고치거나 지우지
  않는다. 저장소 밖 운영 산출물이고 손대려면 사람 승인이 필요하다 (`issues.md` I3).
- 자격증명을 자동으로 옮기거나 복사하지 않는다.

### 2026-08-03 운영자 승인 후속 처분

구현과 실측이 끝난 뒤 운영자가 저장소 밖의 만료된 1회성
`~/.local/share/tossos/bin/soak-autostart.sh`를 은퇴시키도록 별도로 지시했다. 실행·등록
참조가 없음을 확인한 뒤 제거했으며, 이는 위 구현 Non-Goal을 운영자 승인으로 확장한 후속
운영 처분이다. 코드 동작은 바꾸지 않는다. 처분 근거는 `issues.md` I3, 검토 기록은
`review.md`의 "운영자 승인 후속 처분"에 남긴다.

## 덤으로 드러난 것 — soak drift 테스트는 대조할 상대가 없다

`TestTheProcessPatternMatchesTheAutostartScript`는 이렇게 생겼다.

```go
if soakProcessPattern != "tossctl soak run" {
        t.Errorf("the pattern is %q; soak-autostart.sh greps for \"tossctl soak run\"", …)
}
```

`tools/soak-autostart.sh`는 **이 저장소에 없다.** 엔진 쪽 drift 테스트는 실제로 파일을
읽어 대조하지만(`autostartScript(t)`), soak 쪽은 상수를 자기가 하드코딩한 리터럴과
비교할 뿐이다. 두 반쪽을 묶는다고 주장하지만 한쪽이 저장소에 존재하지 않는다.

그래서 이 change는 그 테스트를 **실재하는 계약**으로 바꾼다: 패턴을 리터럴이 아니라
`soakArgs`가 만드는 argv에 묶는다. 설치된 스크립트와의 관계는 테스트가 아니라
`issues.md` I3에 사실로 기록한다.

## Impact

- `cmd/tossctl/soakproc.go` — `soakArgs` 신설, `spawnDetachedSoak`, `soakProcessPattern`,
  `pgrepSoak`, `restartSoak`의 호출부.
- `cmd/tossctl/engineproc.go` — 소유 판정을 공유 헬퍼로 추출(동작 불변).
- `cmd/tossctl/console.go` — `restartSoak`에 root를 넘긴다.
- `cmd/tossctl/soakproc_test.go` — drift 테스트를 실재하는 계약으로 교체.
- `operator-console` spec — ADDED 1건. 기존 요구사항은 MODIFY하지 않는다
  (a055 `issues.md` I1의 미아카이브 MODIFY 부채를 늘리지 않는다 — a059와 같은 이유).
