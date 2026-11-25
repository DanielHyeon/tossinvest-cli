# a166 tasks 4.2 실환경 실측 — 호스트 · 컨테이너 결속 (2026-11-25)

## 판정

**4.2 는 지금 환경으로는 체크할 수 없다. 미충족 3건(B1 · B2 · B3).**

- 측정: 2026-11-25 09:22~09:30 KST, Manager 읽기 전용 실측.
- 승인: 사용자 「승인」(2026-11-25) — Manager 가 읽기 전용으로 증거를 모으고 사람이 확인하는 범위.
- 운영 변경 0(설정 · 컨테이너 · 타이머 · 바이너리 무변).
- 아래 구분: 「충족」= 측정해서 맞음, 「미충족」= 측정해서 계약과 다름, 「안 봄」= 이번 측정 범위 밖.

## 측정 방법

- `docker inspect`(User · HostConfig · Mounts · Env 는 키 이름만)
- `stat -f` · `stat -c '%d:%i'` · `findmnt`
- `readlink /proc/self/ns/{time,user}`, `/proc/sys/kernel/random/boot_id`, `/proc/self/timens_offsets`, `CLOCK_BOOTTIME`
- `systemctl --user list-timers` · `list-unit-files` · `cat`, `ps`
- 바이너리 `grep -ac TOSSOS_ACCOUNT_BUDGET_MODE`(예산 코드 포함 여부)
- 환경: Docker 서버 29.9.2, 커널 7.0.0-43-generic

## 충족

| 항목 | 실측 |
|---|---|
| UID | 호스트 · `tossos-tossos-1` · `tossos-httpapi-1` 모두 1000:1000. user namespace 는 셋이 같다(`user:[4026531837]`). |
| 파일시스템 | `~/.config/tossctl` · `~/.local/share/tossos` · `~/.local/share/tossos-deploy` 는 `/dev/nvme1n1p3` ext4(rw,relatime), symlink 아님. |
| 동일 객체 | bind mount 된 두 디렉터리의 device:inode 가 호스트와 두 컨테이너에서 같다 — config `66309:11809077`, data `66309:11940865`. |
| boot | `boot_id` 가 셋에서 같다(앞 8자 `ad865048`). 두 컨테이너의 `timens_offsets` 는 monotonic · boottime 모두 `0 0`, uptime 이 호스트 `CLOCK_BOOTTIME` 과 같은 초(134593). |
| 현재 모드 | 두 컨테이너 env 키에 `TOSSOS_ACCOUNT_BUDGET_MODE` · `TOSSOS_ACCOUNT_BUDGET_MANIFEST` 가 없다 = OFF(변경 없음). 컨테이너는 read-only rootfs, restart `unless-stopped`. |

## 미충족

### B1 time namespace 가 참여자마다 다르다

- 실측: 호스트 `time:[4026531834]`, `tossos-tossos-1` `time:[4026533806]`, `tossos-httpapi-1` `time:[4026533867]`.
- 코드: `internal/clock/bootstamp_linux.go` 의 `Snapshot` 은 `/proc/self/ns/time` 의 device · inode 를 `BootStamp.Namespace` 에 넣는다. `internal/clock/bootstamp.go` 의 `RemainingBootHint` 는 같은 `BootID` 에서 `now.Namespace != saved.Namespace`(또는 anchor 와 다름)이면 `ErrBootStamp` 를 돌려준다.
- 계약: design 77행도 「같은 boot/time namespace」를 전제로 적는다.
- 귀결: 한 참여자가 남긴 유효 hint 를 다른 time namespace 의 참여자가 복원하면 fail-closed 가 된다. 지금 배치(호스트 도구 + 컨테이너 둘)는 지원 계약 밖이다.
- 범위 한정: Manager 가 읽은 범위는 위 두 파일의 이 비교뿐이다. `Namespace` 를 비교하는 다른 자리의 전수는 AST 로 세지 않았다(안 봄).

### B2 같은 root 가 참여자마다 다른 경로 문자열로 보인다

- 실측: 호스트 `~/.config/tossctl` ↔ 컨테이너 `/var/lib/tossos/config`, 호스트 `~/.local/share/tossos` ↔ 컨테이너 `/var/lib/tossos/data`.
- 계약: `analysis/operator-cutover-draft.md` 는 단일 v2 manifest 의 root 절대 경로를 모든 참여 namespace 가 **같은 문자열**로 열 것을 요구하고, namespace 별 manifest · 경로 변환을 허용하지 않는다.
- 귀결: 전환 때 호스트와 같은 절대 경로로 붙는 새 bind mount(compose 변경, 4.5 사람 승인)가 필요하다.

### B3 참여자 중 예산 코드가 없는 바이너리와 별도 프로필이 있다

- `~/.local/bin/tossctl`(2026-10-14 빌드) — 예산 심볼 0. `tossos-attest.timer`(00/06/12/18시)가 이 바이너리로 `soak attest` 를 돈다.
- `~/.local/bin/tossctl-heartbeat`(2026-11-24 빌드) — 예산 심볼 있음. `tossos-heartbeat.timer`(매 분) `engine heartbeat-check`. 이 명령이 공식 API 를 부르는지는 안 봤다.
- `~/.local/share/tossos-a184/bin/a184-jev-warning-signal`(2026-11-13 빌드) — 예산 심볼 있음. `tossos-a184-signal.timer` · `tossos-a184-evaluate.timer`. **별도 config dir `~/.config/tossctl-a184`** 를 쓰고 그 안에 자체 openapi 자격 파일이 있다 — 같은 group 에 묶어야 하는 다른 프로필이다(같은 계좌 · 같은 키인지는 안 봤다).
- `go run ./tools/a128-jev-shadow-probe`(`tossos-a128-kr.timer` · `tossos-a128-us.timer`, `--config-dir ~/.config/tossctl`) — 저장소 HEAD 에서 매번 빌드. 측정 시각에 KR 프로브가 실행 중이었다.
- `tossos-a140-kr.timer`(`go run ./tools/a140-breakout-signal-shadow`, `--config-dir ~/.config/tossctl`) — unit 파일은 있다. 활성 여부는 타이머 목록에 안 보였다.
- 컨테이너 안: `tossos-tossos-1` 의 `console` · `engine run` · `soak run`, `tossos-httpapi-1` 의 `httpapi` — 모두 `--config-dir /var/lib/tossos/config`.
- 사람이 손으로 치는 호스트 CLI(`verify run` 등)와 일회용 프로브도 참여자다(2026-11-25 08:56 KST 에 Manager 가 읽기 전용 OPEN 조회 프로브 1회 실행).
- 시스템 타이머 · crontab 에 toss 항목 0.
- 안 봄: 같은 호스트의 다른 제품 컨테이너(arkos · pms-lite · StockOS infra-web · neo4j · postgres · typedb)가 같은 OpenAPI 키를 쓰는지.

## Manager 결정 대기(사람)

- **B1 처분**
  1. 코드: 같은 `boot_id` 이고 각 참여자의 offsets 가 0 일 때 time namespace 신원이 달라도 같은 시계로 받도록 계약 · 코드를 고치는 새 로트(착수 때 `Namespace` 비교 자리 전수를 AST 로 먼저 센다).
  2. 운영: 모든 참여자를 한 time namespace 안에서만 실행하도록 배치를 바꾼다.
  - Manager 권장: (1)을 조사 로트로 시작. 결정 전에는 4.2 를 체크하지 않는다.
- **B2 · B3**: 4.5 전환 절차에 넣는다 — 같은 경로 bind mount, 호스트 공용 바이너리 교체는 `make stage-local-update` 경로, a184 프로필 결속.
