# a166 tasks 4.2 — B2 · B3 처분안 (2026-11-25)

상태: **처분안(문서)**. 운영 변경 0, Go 제품 코드 · 시험 편집 0, tasks 4.2 미체크. 아래 운영 단계는 모두 사람이 따로 승인한 뒤에만 실행한다.
측정: 2026-11-25 13:55~14:00 KST, 읽기 전용(원문은 같은 디렉터리의 `*-readonly.log`). 기준 소스: `origin/main` `65a274aec`.

## 판정

- **B2 권장 — (가) 같은 절대 경로 bind mount.** 예산 root 와 manifest 디렉터리를 호스트 경로 `<HOME>/.local/share/tossos-budget/{root,launch}` 에 새로 만들고, 두 컨테이너에 **같은 경로 문자열**로 bind 한다(root rw, launch ro). manifest 는 하나만 둔다. 코드 변경은 필요 없다. 컨테이너 재생성(= 엔진 정지)이 필요하다.
- **B3 권장 — 옛 바이너리는 현재 main 빌드로 교체(제외 · 제거 아님), a184 는 「B 20거래일 창이 닫힌 뒤 전환하고 그때 a184 타이머를 끈다」.** a184 바이너리는 required 모드에서 스스로 실행을 거부한다. 그래서 지금 그대로는 참여자가 될 수 없다. 별도 scope 는 계약상 허용되지 않는다.
- 새로 찾은 전환 선행 조건 2건(4.2 처분과 별개):
  - 현재 release 이미지(a197 `e57d06cd6`)와 rollback 이미지(a201 `b93ed133a`)에는 둘 다 B1 수리(`33e4074cf`)가 **없다**. required 로 켜기 전에 B1 이 들어간 새 이미지가 필요하다.
  - 예산 launch 모드를 TossOS audit 에 남기는 코드가 없다.
- 「같은 문자열」은 **코드가 강제하는 조건이 아니다**. 절차 문서(`operator-cutover-draft.md:30`)만 요구한다. 코드가 실제로 요구하는 것은 이것이다: 각 참여자가 자기 mount namespace 에서 manifest 의 root 문자열을 symlink 없이 열 수 있어야 하고, 연 객체의 device:inode 가 manifest identity 와 같아야 한다(아래 B2-1 AST 열거).

## B2 — 같은 root 가 참여자마다 다른 경로 문자열로 보임

### B2-1 강제 지점 (AST 열거)

함수 내부의 분기를 근거로 쓰므로 `tools/logic-map` AST 를 먼저 만들었다. 산출물은 `ast/*.ast.json` 6개이고, 열거는 `ast/enumeration.txt` 다. 만든 도구는 `ast/enumerate.py` 이며 실행 명령은 그 파일 머리에 있다.

| 함수(파일:줄) | root 문자열에 관한 분기 전수 | 의미 |
|---|---|---|
| `accountbudgetruntime.Open` (`internal/accountbudgetruntime/runtime_linux.go:65-103`) | B5 · B7(manifest 경로 유무), B10(L88 manifest 가 root 밖인지), B13(L99 `VerifyIdentity`) | root 문자열은 L95 에서 `OpenObserved` 에 그대로 넘긴다. 다른 참여자나 저장값과 비교하는 분기는 0개다. |
| `readManifest` (`runtime_linux.go:142-220`) | B1(manifest 경로가 절대 경로인지 · 정규형인지), B17(L212 조상 디렉터리가 root identity 인지), B18(L216 `m.Root == ""`) | root 문자열은 「비어 있지 않음」만 검사한다. |
| `accountbudget.checkedDir` (`internal/accountbudget/files_linux.go:29-68`) | B1(L30 절대 · 정규형), B3-B5(L37-46 각 성분 `O_NOFOLLOW`), B7(L55 uid · 0700), B9(L63 ext4/tmpfs) | 경로 어느 단계든 symlink 가 있으면 거절한다. 문자열 자체를 비교하지는 않는다. |
| `Budget.VerifyIdentity` (`internal/accountbudget/identity_linux.go:19-39`) | B4(L35 root · queue · wire 의 dev:ino 가 manifest 와 같은지) | 결속의 실체는 이 dev:ino 비교다. |
| `OpenContext` (`internal/accountbudget/budget_linux.go:105-159`) | L112 `checkedDir(c.Root)`, L116 `root` 는 메모리에만 둠 | root 문자열은 파일에 쓰지 않는다. |
| `Budget.verify` (`budget_linux.go:172-220`) | B4(L187 같은 문자열로 다시 연 객체의 dev:ino 가 처음과 같은지) | 같은 프로세스 안에서만 문자열을 다시 쓴다. |

`rg -n "b\.root"` 를 시험 아닌 파일에서 돌리면 `budget_linux.go:178` 한 곳뿐이다. root 문자열을 영속하는 자리는 없다.

결과를 못 박는 시험:

- `TestIdentityMismatchFails`(`internal/accountbudgetruntime/runtime_test.go:111`) — identity 가 하나만 달라도 거절.
- `TestSameManifestAcrossProcessesRejectsCopiedRoot`(`:161`) — 같은 문자열 자리에 복사한 root 를 두면 거절.
- `TestManifestSymlinkModeAndInsideRoot`(`:187`), `TestParentSymlinkRefused`(`:318`), `TestTwoRealClientsShareGroup`(`:231`).
- `TestA166OfflineCheckRejectsV1MissingAndCopiedEvent`(`offline_test.go:238`).

**「다른 문자열 · 같은 객체」를 여는 경우를 재는 시험은 없다.** 모든 시험은 `setupGroup` 의 `t.TempDir()` 하나(namespace 하나)에서 돈다. 비특권 사용자는 bind mount 를 만들 수 없고 `/proc/self/root` 경로는 `O_NOFOLLOW` 에 막힌다. 그래서 단일 namespace 시험으로는 이 경우를 만들 수 없다.

귀결: namespace 마다 manifest 를 따로 두는 것(다른 root 문자열, 같은 identity)을 **코드는 받는다**. 막는 것은 절차뿐이다(`operator-cutover-draft.md:30`). 그리고 manifest 를 하나로 두는 규칙만이 「참여자가 서로 다른 root(다른 group)로 조용히 갈라지는 것」을 구조로 막는다. manifest 가 둘이면 각 manifest 는 자기 안에서는 일관되게 통과한다. 그래서 운영자 실수로 group 이 갈라져도 검출 지점이 없다.

### B2-2 현재 컨테이너 배치 (실측 `containers-readonly.log` · `container-view-readonly.log`)

- 두 컨테이너가 같은 설정을 쓴다: compose `compose.yaml` + `~/.local/share/tossos-deploy/release-override.yaml`(label `config_files`), working_dir 은 메인 checkout. release-override 는 image 핀만 담는다.
- bind mount 2개(`compose.yaml:42-48`, httpapi `:123-129`):
  - `~/.config/tossctl` → `/var/lib/tossos/config`
  - `~/.local/share/tossos` → `/var/lib/tossos/data`
  - 둘 다 rw, `rprivate`.
- 그 밖에 secret 파일 bind 3개가 있다.
- `--config-dir /var/lib/tossos/config`(Cmd). env 키는 `PATH · TOSSOS_CONTAINER · TOSSOS_DATA_DIR · XDG_CONFIG_HOME · XDG_DATA_HOME` 이고 httpapi 는 `TOSSOS_API_PUBLIC_URL` 이 더 있다. 예산 키는 없다(OFF).
- `read_only: true`, tmpfs `/tmp` · `/run/tossos`, user `1000:1000`, `restart: unless-stopped`.
- `/proc/<pid>/root` 로 컨테이너 mount namespace 를 stat 했다(`docker exec` 없음):
  - config `66309:11809077`, data `66309:11940865` 로 호스트와 같다. bind 는 `st_dev` 를 보존한다.
  - 컨테이너의 `/home` 은 비어 있는 root 0755 디렉터리다(overlay dev 69/71).
- 호스트 경로 조상(`host-paths-readonly.log`, `namei -l`): `/`, `/home`, `<HOME>`, `.local`, `share` 에 symlink 가 없다. `realpath ~` 는 그대로다. 모두 `/dev/nvme1n1p3` ext4 위에 있다.

### B2-3 선택지 비교

| | (가) 호스트 경로와 같은 절대 경로로 bind 추가 **[권장]** | (나) 중립 경로 `/var/lib/tossos-budget` 을 호스트에 새로 만들고 같은 경로로 bind | (나') 호스트 config/data 자체를 `/var/lib/tossos/{config,data}` 로 이동 | (다) 계약을 고쳐 namespace 별 manifest 허용(root 는 기존 data bind 안) |
|---|---|---|---|---|
| 코드 변경 | 없음 | 없음 | 없음 | 없음(코드는 이미 받음 — B2-1). 계약 · 절차 문서만 바뀐다. |
| compose 변경 | 새 override 에 volume 2 + env 2 | 같음(경로만 다름) | volume source 를 모든 곳에서 바꿈 | env 2 만. volume 은 그대로. |
| 호스트 변경 | `<HOME>/.local/share/tossos-budget/{root,launch}` mkdir 0700(sudo 불필요) | sudo 로 `/var/lib/tossos-budget` 생성 · chown 1000 · 0700. 홈 밖 시스템 경로라 백업 · 추적 관행이 없다. | 모든 unit · 스크립트 · 기억의 절대 경로가 고아가 된다(기억 「이름을 바꾼 체크아웃은 절대경로 상태를 고아로 만든다」 8종). 엔진 lock · journal 경로도 바뀐다. | root 를 `~/.local/share/tossos/account-budget` 에, manifest 를 호스트 · 컨테이너용 2개로 둔다. |
| 기존 mount 와 공존 | 새 target 이 기존 target 과 겹치지 않는다(`/var/lib/tossos/*` 밖). 컨테이너 `/home` 은 비어 있다. | 겹치지 않는다. | 기존 target 은 그대로이고 source 만 바뀐다. | 기존 data bind 안에 들어간다. |
| symlink · hardlink · dev:inode(design 101행 · `checkedDir`) | 모든 성분이 실디렉터리다. Docker 가 컨테이너 쪽 `/home/<user>/…` mountpoint 를 만든다. 그 디렉터리들은 root 소유 0755 라 조상 소유는 검사하지 않고 `O_NOFOLLOW` 만 통과하면 된다. root 0700 · uid 1000 · ext4 를 충족한다. | 같다. 호스트 `/var/lib` 는 `/` 와 같은 ext4 다. | 해당 없음 | 충족한다. 다만 data 디렉터리를 통째로 복사 · 백업하면 root 사본이 생긴다(identity 검사가 거절하지만 「복사 디렉터리」 위험이 생긴다). |
| manifest 하나 원칙(`operator-cutover-draft.md:30`) | 지킨다. | 지킨다. | 지킨다. | **어긴다.** group 이 갈라져도 검출 지점이 없다(B2-1 귀결). |
| 롤백 | 새 override 를 빼고 재생성하면 컨테이너가 OFF 가 된다. 호스트 env 도 함께 뺀다(전 참여자 동시, `operator-cutover-draft.md:48`). 디렉터리는 증거로 보존한다. | 같다. | 경로를 되돌리는 것 자체가 큰 작업이다. | env 만 빼면 된다. |
| audit 추적 | override 파일(+ `.bak-<UTC>` 관행), compose label `config_files`, 전환 영수증(아래) | 같고, `/var/lib` 변경은 sudo 로그에만 남는다. | 추적할 변경이 너무 많다. | manifest 2개의 동기화가 사람 손에 남는다. |
| 엔진 정지 | 필요하다(재생성). | 필요하다. | 필요하다(여러 번). | 필요하다(env 주입 때문. env 는 재생성 없이 못 바꾼다). |

(다)의 이점은 volume 줄을 넣지 않는 것뿐이다. env 를 넣으려면 어차피 재생성해야 하므로 엔진 정지는 줄지 않는다. (나)는 (가)와 안전성이 같지만 sudo 로 홈 밖에 시스템 경로를 만들어야 한다. 그래서 **(가)를 권장한다**.

### B2-4 (가)의 compose 초안 — 새 파일 `~/.local/share/tossos-deploy/account-budget-override.yaml` (아직 만들지 않음)

`<HOME>` 은 실제 파일에서 **리터럴 절대 경로**로 쓴다. `${HOME}` 보간은 sudo 나 다른 셸에서 다르게 풀릴 수 있다.

```yaml
# a166 B2: 예산 root 와 manifest 를 호스트와 같은 절대 경로로 붙임(문자열 동일 · 객체 동일).
# 1단계(마운트만): environment 블록 없이 적용 → OFF 와 같음. 2단계(활성화): environment 블록을 추가.
services:
  tossos:
    volumes:
      - type: bind
        source: <HOME>/.local/share/tossos-budget/root
        target: <HOME>/.local/share/tossos-budget/root
        bind: { create_host_path: false }
      - type: bind
        source: <HOME>/.local/share/tossos-budget/launch
        target: <HOME>/.local/share/tossos-budget/launch
        read_only: true
        bind: { create_host_path: false }
    # 2단계에서만:
    # environment:
    #   TOSSOS_ACCOUNT_BUDGET_MODE: required
    #   TOSSOS_ACCOUNT_BUDGET_MANIFEST: <HOME>/.local/share/tossos-budget/launch/manifest.json
  httpapi:
    volumes:   # tossos 와 같은 두 항목
      - type: bind
        source: <HOME>/.local/share/tossos-budget/root
        target: <HOME>/.local/share/tossos-budget/root
        bind: { create_host_path: false }
      - type: bind
        source: <HOME>/.local/share/tossos-budget/launch
        target: <HOME>/.local/share/tossos-budget/launch
        read_only: true
        bind: { create_host_path: false }
```

배치를 이렇게 정한 이유:

- root 와 launch 를 형제로 둔다. manifest 는 root 밖이어야 하기 때문이다(`runtime_linux.go:88` · `:212`, `offline_linux.go:115` · `:161`). 또 provision 은 manifest 부모가 0700 · 소유자 일치여야 한다(`offline_linux.go:149` → `safeOfflineDirectory` `:58`).
- launch 는 **디렉터리** bind 로 붙인다. 파일 하나를 bind 하면 inode 가 고정된다. 그러면 원자적 rename 으로 새 manifest 를 게시해도 컨테이너는 옛 inode 를 계속 본다.
- launch 는 ro 로 붙인다. 컨테이너가 manifest 를 바꾸지 못하게 하려는 것이다.
- httpapi 가 공식 API 를 보내는지는 호출 경로를 추적하지 않았다. 그래서 보수적으로 참여자로 셌다.

적용 명령은 `-f compose.yaml -f <release-override> -f <account-budget-override>` 셋을 **항상 함께** 준다. override 를 빼고 `up` 하면 `tossos:local` 로 떨어져 배포가 되돌아간다(기억 「운영 경로와 이미지」). `docs/operations.md` 「Docker Compose」 절의 개정안은 이렇다: 「a166 required 를 켠 뒤의 모든 `up` 은 account-budget-override 를 포함한다. 빼면 그 참여자만 OFF 가 되어 총량 계약을 벗어난다(`operator-cutover-draft.md:48`).」

## B3 — 예산 코드가 없는 바이너리와 별도 프로필

### B3-1 참여자 census (실측 `participants-census-readonly.log` · `unit-bodies-readonly.log` · `binaries-readonly.log` · `processes-readonly.log`)

| # | 실행 주체 | 실행 파일(빌드) | 예산 코드 `grep -ac TOSSOS_ACCOUNT_BUDGET_MODE` | config-dir | 주기 · 상태 | 공식 API 송신 |
|---|---|---|---|---|---|---|
| 1 | `tossos-tossos-1` 의 `console` · 자식 `engine run` · `soak run` | 이미지 a197 `e57d06cd6`(B1 수리 없음) | 있음(소스 `root_budget.go` 존재) | `/var/lib/tossos/config` | 상시 | 예 |
| 2 | `tossos-httpapi-1` 의 `httpapi` | 같음 | 있음 | 같음 | 상시 | 호출 경로 미추적 → 보수적으로 「예」 |
| 3 | `tossos-attest.timer` → `~/.local/bin/tossctl soak attest --record-renewal-status` | `488b370fd`, vcs.modified=true, 2026-10-14 설치 | **0** | `~/.config/tossctl` | 00/06/12/18시 | 명령 annotation 이 `source: local`(그 리비전 `cmd/tossctl/soak.go` attest 정의, 현재 `:223`). 호출 경로 AST 미추적. |
| 4 | `tossos-heartbeat.timer` → drop-in 으로 `~/.local/bin/tossctl-heartbeat engine heartbeat-check` | `9738a0c32`, clean | 1 | `~/.config/tossctl` | 매 분 | annotation `source: local`(`engine_heartbeat_check.go:54`) + ntfy |
| 5 | 사람 · 에이전트가 치는 `tossctl …`(PATH → `~/.local/bin/tossctl`, `type -a` 실측) | #3 과 같은 옛 바이너리 | **0** | 기본값 또는 지정값 | 비정기 | **예**(verify · quote 등) |
| 6 | `tossos-a128-kr.timer` → `run-probe.sh` → `go run ./tools/a128-jev-shadow-probe`(메인 checkout 작업 트리에서 매번 빌드) | 메인 checkout 의 그때 HEAD | 소스에 있음(`tools/a128-jev-shadow-probe/main.go:147`). env 는 unit 에 없음 → OFF | `~/.config/tossctl` | 평일 08:50, **측정 시각에 실행 중**(08:50 기동) | 예 |
| 7 | `tossos-a184-signal.timer` · `tossos-a184-evaluate.timer` → `~/.local/share/tossos-a184/bin/a184-jev-warning-signal` | a184 브랜치 `d455ce4ad`(main 에 없음), 2026-11-13 설치, vcs 미기록 | 1 — 다만 **required 에서 실행 거부** | `~/.config/tossctl-a184` | 평일 09:40 · 16:30 | 예(GET) |
| 8 | `tossos-a128-us.timer`, `tossos-a140-kr.timer` | go run(메인 checkout) | 소스에 있음(`tools/a140-breakout-signal-shadow/main.go:139`) | `~/.config/tossctl` | **disabled**(unit 파일만 있음) | 켜면 예 |
| 9 | `~/.local/bin/tossctl.{rollback,backup.*,bak-pre-de14674,p3,prev2,prev3}` 6개 | 2026-07~08 | 6개 모두 0 | — | 예약 없음. 이름을 직접 쳐야 실행된다. | 실행하면 예 |
| 10 | 메인 · 다른 worktree 의 `go run ./cmd/tossctl`, `bin/tossctl` | 그 checkout | main 기준 있음 | 지정값 | 비정기 | 예 |

- crontab 없음. 시스템 timer · `/etc/cron.d` · `/etc/systemd/system` 에 toss 항목 0.
- 다른 컨테이너 22개 중 `tossctl` · `tossos` 경로를 bind 하는 것은 TossOS 두 개뿐이다(`other-containers-mounts-readonly.log`). 다른 제품이 **자기 키로** 같은 계좌를 부르는지는 안 봤다.

### B3-2 `~/.config/tossctl-a184` 가 같은 계좌 scope 인지 (`a184-profile-readonly.log`)

- 파일은 `openapi-credentials.json`(0600, 2026-11-12) · `openapi-token.json`(0600, 2026-11-25 09:51) 두 개뿐이다.
- 두 프로필의 자격 파일 sha256 은 **다름**, 토큰 파일 sha256 도 **다름**이다. 서로 다른 객체다(dev:ino 다름).
- 코드도 키가 다름을 강제한다. a184 `refuseEngineCredentials`(브랜치 `d455ce4ad` `tools/a184-jev-warning-signal/main.go:146-163`)는 엔진 프로필과 apiKey 나 secretKey 가 같으면 실행을 거부한다.
- **같은 계좌인지: 안 봤음 — 사람 확인 필요.** 자격 파일에는 계좌 정보가 없다(최상위 키 이름만 봄). 계좌를 확인하려면 브로커를 불러야 하는데 이번 규칙이 금지한다. → 결정 기록 참조
- 계약상 이 판정은 처분을 바꾸지 않는다:
  - `operator-cutover-draft.md:7` — 「별도 키라고 별도 예산 root를 만들지 않는다. 계좌 동등성을 입증할 수 없으므로 모든 참여자가 하나의 보수적인 공통 group 총량을 나눈다」.
  - `budget-group-binding.md:7`.
  - `proposal.md:9` — 브로커 한도의 집계 단위가 계좌 · 키 · IP 중 무엇인지 미확인.

### B3-3 a166 은 참여자 아닌 호출자를 어떻게 다루나

- `design.md:8` — 「참여하지 않는 송신자의 총량은 통제하지 못한다. 배포 시 참여자 누락이 있으면 계좌 전체 보장을 선언하지 않는다」.
- `budget-group-binding.md:16` — 「bare OFF 도구나 외부 앱의 동시 호출은 이 보장의 범위 밖 … 확인되지 않으면 운영 적용을 승인하지 않는다」.
- `operator-cutover-draft.md:48` — 프로세스 하나만 OFF 로 내리는 것은 총량 계약 이탈이다.
- 위험: 예산 밖 호출이 공유 장부(state · event)를 깨뜨리지는 않는다. 장부에는 그 호출이 안 보일 뿐이다.
  1. 「rolling 1초 3건」 보장이 거짓이 된다.
  2. 그 호출 때문에 생긴 429 를 참여자들이 응답 단위 이벤트로 기록한다(`design.md:69`). 그 결과 엔진 폴링의 hint 보류가 늘어난다. 보호 가격 조회가 같은 hint 로 보류될 수 있다(`design.md:73`, `issues.md` 잔여 위험 4 「안 봤음」).
- 그래서 **알려진 예산 밖 송신자가 하나라도 남아 있으면 전환하지 않는 것**이 계약이다.

### B3-4 처분 선택지

**옛 `~/.local/bin/tossctl` (#3 · #5)**

1. **현재 main(≥ `33e4074cf`, B1 수리 포함) 빌드로 교체 [권장].**
   - `make stage-local-update` 로 `~/.local/bin/tossctl.candidate` 를 만든다(`Makefile:22-24`). 사람이 sha256 과 `go version -m` 을 확인한 뒤 `mv` 로 원자 교체한다.
   - 컨테이너 콘솔의 업데이트 화면은 쓸 수 없다. 컨테이너 실행에서는 시스템 업데이트를 끈다(`cmd/tossctl/console.go:288-290`).
   - 영향: attest · 수동 CLI 가 새 빌드로 돈다. 예산 env 가 없으면 OFF 이므로 교체 자체는 지금과 같다(토글 OFF = 기존 동작).
   - 롤백: `.candidate` 전에 옛 파일을 `tossctl.pre-a166-<UTC>` 로 보존한다.
   - heartbeat drop-in(`tossos-heartbeat.service.d/10-dedicated-binary.conf`)은 자기 주석대로 「공용 바이너리를 a134 이후 빌드로 교체하면 지운다」. 지우는 것은 선택이다. 남겨도 예산 코드가 있는 바이너리다.
2. 예산 참여자에서 명시 제외 — **불가**. 수동 CLI(#5)가 공식 API 를 보내므로 B3-3 의 계약에 걸린다. attest 만 떼어 「local 이니 제외」로 처리해도 PATH 문제는 남는다.
3. 제거 — 비권장. attest 타이머와 수동 CLI 가 깨지고 대체 경로가 없다.

**a184 프로필 · 바이너리 (#7)**

1. **a184 B 창(첫 가동일 2026-11-12, 20거래일)이 닫힌 뒤 전환하고, 전환 때 a184 타이머 둘을 끈다 [권장].**
   - 종료일은 KOSPI 달력이 정한다. 이번에 세지 않았다. → 결정 기록 참조
   - a184 바이너리는 required 모드에서 「계정 예산 manifest 모드에서는 전송 경계를 끼울 수 없어 실행하지 않음」으로 거부한다(브랜치 `d455ce4ad` `main.go:244-246`). 그래서 전환 뒤 타이머가 남아 있어도 송신은 0 이다(fail-closed). 다만 실패 기록이 매일 남으므로 끄는 것이 깨끗하다.
   - 롤백: 타이머를 다시 켠다.
2. a184 를 같은 group 에 넣는다 — **별도 코드 로트**(a184 브랜치)가 필요하다. a184 는 자체 프로세스 예산 transport 와 `official.WithHTTPClient` 를 쓴다(`main.go:247-263`). 사용자 `WithHTTPClient` 에는 권한이 없다(`design.md:23`). 그래서 `Factory.NewClient` 로 다시 배선하고 자체 예산과의 관계를 설계해야 한다. 지금 제안만 한다.
3. 지금 a184 를 멈춘다 — B 창을 잃는다. runbook 「창은 다시 시작할 수 없음」(브랜치 runbook 119행). 사람 결정이 필요하다.
4. 별도 scope 로 인정한다 — **비권장 · 계약 위반**(`operator-cutover-draft.md:7`). 계좌가 다르다는 증명이 없고(B3-2) 집계 단위도 미확인이다(`proposal.md:9`). → 결정 기록 참조

**그 밖**

- #6 a128 · #8:
  - 전환 때 unit drop-in 으로 `Environment=` 두 키를 넣는다. `run-probe.sh:32` 가 `exec go run` 이므로 env 가 그대로 넘어간다.
  - `go run` 은 메인 checkout 의 그때 작업 트리에서 빌드한다. 그래서 「게시된 정확 SHA」(`operator-cutover-draft.md:7`)를 고정하지 못한다. 사람 결정 7 로 올린다.
- #9 백업 바이너리 6개: 예산 코드 0 이다. `~/.local/bin` 밖 격리 디렉터리로 옮기는 것을 사람 결정 6 으로 올린다.

## 재실측 체크리스트 (B2 · B3 처분 뒤, 4.2 체크 조건)

모두 읽기 전용이다. `<pid>` 는 `docker inspect --format '{{.State.Pid}}' <컨테이너>`, `<B>` 는 `<HOME>/.local/share/tossos-budget`, `<M>` 은 `<B>/launch/manifest.json`.

1. **UID · user ns**: `docker inspect --format '{{.Config.User}}'`. `readlink /proc/self/ns/user` 와 `/proc/<pid>/ns/user` 가 같아야 한다.
2. **boot · time ns offset**(B1 수리 뒤 계약 `design.md:77`): `/proc/sys/kernel/random/boot_id` 가 셋에서 같아야 한다. `/proc/<pid>/timens_offsets` 는 0 · 0 이어야 한다(ns 신원은 달라도 된다).
3. **경로 문자열 동일**(B2):
   - 컨테이너: `docker inspect --format '{{range .Config.Env}}{{println .}}{{end}}' <c> | grep -E '^TOSSOS_ACCOUNT_BUDGET_(MODE|MANIFEST)='`
   - 호스트 unit: `systemctl --user show -p Environment <unit> | tr ' ' '\n' | grep -E '^TOSSOS_ACCOUNT_BUDGET_'`
   - 실행 중 프로세스: `tr '\0' '\n' </proc/<pid>/environ | grep -E '^TOSSOS_ACCOUNT_BUDGET_'`
   - 모든 참여자가 `required` 와 같은 `<M>` 이어야 한다. 다른 env 값은 출력하지 않는다.
4. **같은 객체**: `stat -c '%d:%i %u %a' <B>/root <B>/root/{queue.lock,wire.lock,event.lock} <M>` 를 호스트와 `/proc/<pid>/root<B>/…`(두 컨테이너)에서 잰다. 값이 같고 manifest identity 와 같아야 한다. identity 는 manifest 의 `identity`·`event_lock` 필드에서 숫자만 읽는다.
5. **파일시스템 · 모드**: `findmnt -T <B>/root -o FSTYPE` 가 ext4 이고, root 0700 · manifest 0600 · nlink 1 · uid 1000 이어야 한다. `namei -l <B>/root` 에 symlink 가 0 이어야 한다.
6. **오프라인 점검**: 호스트에서 `a166-budget-state check --manifest <M>`(사람 승인 · 바이너리 sha 확인 뒤, `operator-cutover-draft.md:13`).
7. **바이너리 전수**: census 의 모든 실행 파일에 `grep -ac TOSSOS_ACCOUNT_BUDGET_MODE` ≥ 1, `go version -m` 의 vcs.revision 이 `33e4074cf` 의 자손이어야 한다(`git merge-base --is-ancestor`). 이미지는 핀의 `.Id` 와 커밋을 대조한다.
8. **참여자 전수**: `systemctl --user list-timers --all`, `list-unit-files | grep -i toss`, `crontab -l`, 시스템 timer, `ps -eo pid,args | grep -E 'tossctl|a1[0-9]{2}-'`, `type -a tossctl`. 새 주체가 0 이어야 하고 a184 는 결정대로여야 한다.
9. **rollback 핀**: rollback 이미지에도 B1 수리가 있어야 한다(없으면 rollback 은 「전원 OFF」만 가능 — 사람 결정 5).
10. **전환 전 정지 확인**: 모든 참여자 종료 뒤 30초 이상(`operator-cutover-draft.md:9`). 측정 명령 원문을 영수증에 적는다.

## 전환(cutover) 순서 초안

[정지] = 엔진 정지, 손절 없음. 두 시장이 닫힌 창에서만 한다. 기억 기준 KST 05:00~09:00 이지만 KR NXT 는 08:00 에 개장하고, 11월 미국 표준시 정규장은 06:00 KST 에 마감한다. 실제 창은 그날 달력으로 사람이 확정한다. [승인] = 사람 승인 단계.

1. [승인] 사람 결정 1~8 확정(아래). — **2026-11-25 확정**(「사람 결정 기록 (2026-11-25)」)
2. [승인 · 정지 없음] main ≥ `33e4074cf` 로 `make image CHANGE=a166-account-shared-rate-limit` → release 핀을 정한다. rollback 핀도 정한다(결정 5). schema 를 대조한다(기억 「배포 전 schema 대조」).
3. [승인 · 정지 없음] 호스트 `tossctl` 를 교체한다(B3-4 권장 1). 옛 파일을 보존한다. 백업 바이너리를 격리한다(결정 6).
4. [승인 · 정지 없음] `a166-budget-state` 를 빌드하고 sha 를 기록한다. `mkdir -m 700 <B> <B>/root <B>/launch` 뒤 `provision --root <B>/root --manifest <M> --scope <opaque> --instance <opaque>` 와 `check --manifest <M>` 를 실행한다. 실패하면 부분 결과를 보존하고 중단한다(`operator-cutover-draft.md:20-28`).
5. [승인 · **정지** 창 1] account-budget-override(**volume 만**, env 없음)를 넣어 재생성한다. 이때는 OFF 와 같다. 체크리스트 4 · 5 로 컨테이너 쪽 객체를 확인한다. 엔진은 기존 절차로 다시 기동한다. 실패하면 override 를 빼고 재생성해 원상으로 돌린다.
6. [승인] a184 B 창이 닫힌 뒤 a184 타이머 둘을 끈다(B3-4 권장 1).
7. [승인 · **정지** 창 2] 활성화:
   - (a) 모든 참여자를 정지한다: 타이머 stop, 실행 중 a128 서비스 stop, 수동 CLI 0, `docker compose … stop`(자동 재시작 차단). 30초 이상 정지했음을 확인한다.
   - (b) 호스트 unit drop-in 과 셸 profile 에 env 두 키를 넣는다. 파일마다 `.bak-<UTC>` 를 만든다.
   - (c) override 에 environment 를 추가하고 `up` 한다.
   - (d) 체크리스트 1~8 을 확인한다.
   - (e) 엔진을 기동한다. required 오류가 나면 wire0 이다 — **엔진도 서지 않아 보호가 없다**. 그때는 (f)로 간다.
   - (f) 롤백은 전 참여자를 동시에 OFF 로 돌린다. override env 와 drop-in env 를 빼고, root 와 manifest 는 증거로 보존한다(`operator-cutover-draft.md:44` · `:48`).
8. [승인] 전환 영수증을 change 디렉터리에 남긴다. 영수증 내용: override · drop-in sha256, compose `config_files` label, 각 참여자 env 두 키, manifest identity 숫자, 정지 시각. 그 뒤 Manager 가 4.2 재실측 → tasks 4.2 체크를 판정한다.

## 사람 결정 목록

1. **B2 방식**
   1. (가) 홈 아래 같은 절대 경로 bind — **권장**
   2. (나) `/var/lib/tossos-budget`(sudo) 같은 경로 bind
   3. (나') 호스트 config/data 를 컨테이너 경로로 이동 — 비권장
   4. (다) 계약을 고쳐 namespace 별 manifest — 비권장(group 분열을 검출하지 못함)
2. **옛 공용 `tossctl`**
   1. 현재 main 빌드로 교체 — **권장**
   2. 제외 — 계약상 불가
   3. 제거 — 비권장
3. **a184**
   1. B 창이 닫힌 뒤 전환하고 그때 타이머를 끈다 — **권장**
   2. a184 브랜치 코드 로트로 group 에 편입(별도 로트)
   3. 지금 중지(B 창을 잃음)
   4. 별도 scope — 비권장 · 계약 위반
4. **a184 프로필이 같은 계좌인지 확인**
   1. 확인하지 않는다. 계약상 같은 group 이 기본이므로 처분이 바뀌지 않는다 — **권장**
   2. 사람이 앱에서 확인해 기록한다
5. **rollback 핀**
   1. B1 수리가 들어간 두 번째 빌드를 rollback 으로 둔다(required 상태에서 되돌릴 수 있음) — **권장**
   2. 현 a201 핀을 유지한다(rollback = 전원 OFF 뿐)
6. **백업 바이너리 6개**(예산 코드 0)
   1. `~/.local/bin` 밖 격리 디렉터리로 이동 — **권장**
   2. 그대로 둔다
7. **`go run` 참여자(a128 · a140) 의 SHA 고정**
   1. 전환 전에 고정 빌드 바이너리로 바꾼다(별도 작업) — **권장**
   2. 메인 checkout HEAD 를 전환 기간 동안 고정하는 규칙만 둔다
8. **예산 launch 모드의 audit**
   1. 영수증 · override `.bak` · compose label 로 추적한다(기존 이미지 교체 관행과 같음) — **권장(첫 전환)**
   2. 기동 때 모드와 manifest identity 를 TossOS audit 에 쓰는 별도 코드 로트

## 사람 결정 기록 (2026-11-25)

사용자 원문(2026-11-25): 「4번은 다른 계좌임. 나머지 권장대로」. 아래 선택지 문장은 위 「사람 결정 목록」에서 그대로 옮겼다. 이 기록은 결정만 적는다. 운영 변경 0, 4.2 는 여전히 미체크다(재실측 체크리스트를 통과한 뒤 Manager 가 판정).

| # | 항목 | 고른 선택지 | 근거 원문 |
|---|---|---|---|
| 1 | B2 방식 | ① (가) 홈 아래 같은 절대 경로 bind — **권장** | 「나머지 권장대로」 |
| 2 | 옛 공용 `tossctl` | ① 현재 main 빌드로 교체 — **권장** | 「나머지 권장대로」 |
| 3 | a184 | ① B 창이 닫힌 뒤 전환하고 그때 타이머를 끈다 — **권장** | 「나머지 권장대로」 |
| 4 | a184 프로필이 같은 계좌인지 확인 | ② 사람이 앱에서 확인해 기록한다 — 에 해당하는 사람 확인 결과: **엔진과 다른 계좌**. 앱으로 확인했는지 · 어떻게 확인했는지는 진술에 없다. | 「4번은 다른 계좌임」 |
| 5 | rollback 핀 | ① B1 수리가 들어간 두 번째 빌드를 rollback 으로 둔다(required 상태에서 되돌릴 수 있음) — **권장** | 「나머지 권장대로」 |
| 6 | 백업 바이너리 6개(예산 코드 0) | ① `~/.local/bin` 밖 격리 디렉터리로 이동 — **권장** | 「나머지 권장대로」 |
| 7 | `go run` 참여자(a128 · a140) 의 SHA 고정 | ① 전환 전에 고정 빌드 바이너리로 바꾼다(별도 작업) — **권장** | 「나머지 권장대로」 |
| 8 | 예산 launch 모드의 audit | ① 영수증 · override `.bak` · compose label 로 추적한다(기존 이미지 교체 관행과 같음) — **권장(첫 전환)** | 「나머지 권장대로」 |

### 결정 4 가 바꾸는 것과 안 바꾸는 것

증거 등급: **사용자 진술**이다. 브로커 조회로 확인하지 않았다(이번 규칙상 브로커 호출 금지). 계좌번호 등 식별값은 적지 않는다. 참고로 a184 runbook 도 같은 진술을 2026-11-12 사용자 결정으로 적어 두었다(브랜치 커밋 `d455ce4ad` `openspec/changes/a184-kr-jev-warning-live-lane/runbook.md:9` — 「신호 도구는 엔진과 다른 키·다른 계좌인 둘째 Open API 키만 쓰고(전용 프로필 `~/.config/tossctl-a184`)」, 같은 줄 「「호출 한도는 키별이라 엔진과 겹치지 않음」은 사용자 진술이며 저장소에 실측이 없음」). 진술이 두 번 있다는 것이지 실측이 생긴 것은 아니다.

바꾸는 것:

- B3-2 「같은 계좌인지: 안 봤음 — 사람 확인 필요」 → **사용자 진술로 다른 계좌(2026-11-25), 브로커 조회로는 미확인.**
- 「안 본 것 · 한계」의 a184 계좌 동일성 항목도 같은 상태로 바뀐다.
- B3-4 a184 선택지 4(별도 scope) 기각 근거 둘 중 첫째 「계좌가 다르다는 증명이 없고」는 약해진다. 진술은 있지만 증명(브로커 실측)은 여전히 없다.

안 바꾸는 것:

- 기각 근거 둘째 「집계 단위도 미확인」은 그대로다. `proposal.md:9` — 「현재 브로커 한도 집계 범위가 계좌·키·IP 중 무엇인지 확인하지 못했다」. 집계가 IP 단위라면 계좌가 달라도 a184 의 호출은 같은 한도를 나눠 쓴다.
- 계약 `operator-cutover-draft.md:7` — 「별도 키라고 별도 예산 root를 만들지 않는다. 계좌 동등성을 입증할 수 없으므로 모든 참여자가 하나의 보수적인 공통 group 총량을 나눈다」. 진술은 입증이 아니므로 이 규칙의 전제는 그대로다.
- 그래서 선택지 4 는 여전히 비권장 · 계약 위반이고, 결정 3 ①(창 종료 뒤 전환 + 그때 a184 타이머 둘 끄기)이 유지된다. 전환(cutover 7단계)은 a184 B 창이 닫히기 전에 하지 않는다.

### a184 B 창 종료일 추정 (읽기 전용 · 추정)

근거는 git 객체만 읽었다(`git show <커밋>:<경로>`). 실제 가동 기록(`window-B.json`, 로그, 계좌 데이터)은 열지 않았다. 그래서 아래 날짜는 **추정**이다.

세는 규칙 원문(브랜치 커밋 `d455ce4ad`):

- `design.md:93-95` — 「달력: 공식 KOSPI 일봉에 봉이 있는 날이 거래일임.」 「1일째: B 는 사람이 승인한 첫 가동일 … 그날을 포함함.」 「길이: 1일째부터 연속 20거래일. 결손일, 식별자 정지일, 실행 실패일, 신호가 없던 날, 조기 폐장·지연 개장일을 모두 산입함.」
- `design.md:96` — 「창은 다시 시작하지 않음. 정지와 재개, 누적 손실 중단, 도구 재시작으로 창이 늘거나 줄지 않음.」
- `state-format.md:142-143` — 「창은 이 날짜 이후 KOSPI 일봉 날짜(가격 유효성과 무관)가 있는 연속 20거래일이고, 오늘보다 뒤 날짜는 버리며 오늘은 16:00 KST 뒤에만 셈.」 「결손·정지·실패·신호 없음·조기 폐장·드라이런이 차지한 날을 모두 셈. 다시 시작하지 않음.」
- `runbook.md:72` — 「공휴일에는 프로브 회차가 없어 `absent_round` 로 기록됨(창 계산은 KOSPI 달력이 하므로 영향 없음).」 · `runbook.md:99` — 「그날 B 기록이 없으면 … 그날은 창에서 「기록 없음」 날로 셈.」
- 판정 시각: `runbook.md:101` 「판정 (20거래일째 장 마감 뒤, 사람이 한 번 실행)」.

정리: 첫 가동일을 1일째로 **포함**하고, 누락 · 실패일은 **빼지 않고 산입**한다. 휴장일은 KOSPI 일봉이 없으므로 세지 않는다.

**주의 — 판정 대상 창은 2026-11-12 창이 아니다.** 같은 브랜치의 `design.md:19` 「1판: … B 창 1일째 2026-11-12」, `design.md:25` 「1판 창(1일째 2026-11-12)에서 나온 제안·평가 기록은 2판 판정에 넣지 않음 … 2판은 새 기록 디렉터리와 새 창을 씀」, `design.md:26` 「1판 창에서 1판 판정(`verdict`)은 내지 않음」. 그 뒤 브랜치 끝 커밋 `2a12b2449`(2026-11-13 00:15 KST, 부모가 `d455ce4ad`) `review.md:487` 「2판 전환 기록」의 `:496` 에 「`init-window --state-dir …/state-v2 --first-day 2026-11-13 …` 종료 코드 0」이 적혀 있다. 즉 문서상 판정에 쓰이는 2판 B 창의 1일째는 **2026-11-13** 이다. 이 처분안 B3-4 의 「첫 가동일 2026-11-12」는 1판 창 날짜다.

| 창 | 1일째 | 20거래일째(주말만 뺀 추정) | 판정 |
|---|---|---|---|
| 1판(`state`) | 2026-11-12(목) | **2026-12-09(수)** | 1판 판정은 내지 않음(`design.md:26`) |
| 2판(`state-v2`) | 2026-11-13(금) | **2026-12-10(목)** 장 마감 뒤 | tasks 4.4(2판 창 기준) |

- **주말만 뺀 추정이다.** 저장소에서 KRX 휴장일 목록 출처를 찾지 못했다(`2026-12-25` · `2026-12-31` 문자열은 시험 픽스처와 OpenAPI 사본에만 있음). 창 안에 KOSPI 휴장일 · 임시 휴장이 있으면 그 수만큼 뒤로 밀린다.
- a184 전환 조건(결정 3 ①)을 판정할 때는 이 추정이 아니라 실제 KOSPI 일봉 달력 또는 a184 `verdict` 의 「남은 거래일 N」 응답(`runbook.md:108`)으로 확인한다. 그 확인은 사람 승인 뒤 별도로 한다.

## 안 본 것 · 한계

- 컨테이너 재생성 때 새 time namespace 가 생기는지, 그 offset 이 0 인지 — 재시작 승인이 필요해 못 쟀다. 체크리스트 2 로 전환 때 잰다.
- read_only rootfs 컨테이너에 `<HOME>/…` target bind 를 걸 때 Docker 가 mountpoint 를 문제없이 만드는지 — `docker run` · 재생성이 필요해 못 쟀다. 5단계(마운트만) 가 그 리허설이다.
- httpapi 가 공식 API 를 보내는지, attest · heartbeat 가 정말 송신 0 인지 — annotation 만 봤고 호출 경로 AST 는 추적하지 않았다. 보수 처분(httpapi 포함, 옛 바이너리 교체)이라 결론은 바뀌지 않는다.
- a184 프로필의 계좌 동일성 — 브로커 조회 금지로 판정 불가, 사람 확인 필요. → 결정 기록 참조(2026-11-25 사용자 진술로 다른 계좌, 브로커 조회로는 미확인)
- a184 B 창 종료일 — KOSPI 달력으로 세지 않았다. → 결정 기록 참조(주말만 뺀 달력 추정, 휴장일 미반영)
- 다른 제품(StockOS 등)이 자기 키로 같은 계좌 · 같은 IP 에서 Open API 를 부르는지 — 다른 컨테이너의 env 는 보지 않았다(mount source 만 봄).
- `docker compose config` 로 override 병합 결과를 렌더하지 않았다. override 파일을 만들지 않았기 때문이다.
- 권한 거부: 없음. 다만 첫 AST 생성 명령은 `rm` 안전 검사에 걸려 실행되지 않았다. 지우기를 빼고 다시 실행했다(파일 삭제 0).

## not-applicable

- FLM/BTM · 변이 원장 · 다중 적대 리뷰 — not-applicable. Go 함수를 편집하지 않았고(수정 함수 0), 문서 처분안이다(`docs/WORKFLOW.md:161-168` 비례 원칙). 분기를 근거로 쓴 6개 함수는 AST 를 먼저 만들었다(B2-1).
- CodeGraph — not-applicable. 호출 사슬이 아니라 단일 함수 내부 분기를 근거로 썼다. 호출자 수를 주장하지 않는다.
- 시험 실행 — not-applicable. 코드를 바꾸지 않았다. 인용한 시험은 이름과 줄만 확인했고 이번에 실행하지 않았다.
- 기억 recall:
  - ai-memory `memory_query`(「a166 manifest cutover bind mount root path tossctl-a184」) — a166 census 함정 1건 외에 B2 · B3 관련 이전 결정은 없었다.
  - Hindsight `tossos-dev-episodes` recall — a166 time namespace 로트 검증 외에 관련 결정은 없었다.
  - 둘 다 advisory 이고 현재 HEAD 로 재검증할 주장은 없었다.
