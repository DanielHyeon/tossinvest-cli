# a166 cutover 준비 2~4단계 실행 영수증 (2026-11-25)

`proposal.md` 「전환(cutover) 순서 초안」의 정지 없는 준비 단계 2 · 3 · 4 를 Manager 가 사용자 승인(「2단계 승인」「3단계 승인」「4단계 승인」) 뒤 직접 실행한 기록이다. 값은 Manager 실측이다. 이 영수증을 쓴 팀메이트는 아무것도 다시 실행하지 않았고, 각 단계 끝의 「재확인」 줄만 읽기 전용으로 다시 쟀다(재확인 명령 원문은 맨 아래). `<HOME>` 은 홈 디렉터리, `<B>` 는 `<HOME>/.local/share/tossos-budget`, `<M>` 은 `<B>/launch/manifest.json` 이다(`proposal.md` 「재실측 체크리스트」와 같은 표기).

운영 컨테이너 변경 0. 엔진 정지 0. `tasks.md` 4.2 미체크.

## 2단계 — 이미지 빌드 (컨테이너 교체 없음)

- 명령: `make image CHANGE=a166-account-shared-rate-limit`, rc 0, 2026-11-25 18:41:03~18:41:34 KST. 빌드 전후 HEAD 모두 `f73e0917b32ba02072cbf6fdf8380d3dbd0f6889`. 원문 로그: `make-image.log`(Manager 원본과 바이트 동일, sha256 `9970fb4eb33b64c48e9b7b65da5933c7bc6afd203b6f523bbb1d7244db4fe91a`).
- 결과: `tossos:a166-account-shared-rate-limit-f73e0917b` → `sha256:7be5c098b73eb7411bfb163600c94cb207db6e681ea3e775b00a9910b5593ca0`. `tossos:local` 태그도 이 이미지로 옮겨졌다.
- 빌드 전 확인:
  - `make image` 는 override 파일을 고치지 않는다(`Makefile:206-230`).
  - 운영 이미지 커밋 `e57d06cd6` 과 HEAD 의 `internal/journal` diff 0. `SchemaVersion` 둘 다 36(`internal/journal/schema.go:6`). 운영 저널 파일은 열지 않았다.
  - `e57d06cd6`..HEAD 의 생산 Go diff: `internal/clock/bootstamp.go`, `internal/official/reconcile_detail_reads.go` · `reconcile_reads.go` · `reconcile_response.go`.
- 빌드 후 확인:
  - 실행 중 두 컨테이너는 그대로 `tossos@sha256:73682ed243dfa2915d0decb94a4a697d6fe155437a255cf49e63a43aa5daf3ff`(시작 2026-11-25T00:10Z).
  - override 두 파일 mtime 불변: release 2026-11-25 09:10:24, rollback 2026-11-24 17:50:36 KST.
  - a197 · a201 이미지 이름이 남아 있다.
  - 두 이미지에서 `/usr/local/bin/tossctl` 을 꺼내 `grep -ac reconcileUnambiguous`: 새 이미지 1, 운영 이미지 0.
- 한계:
  - B1 수리는 조건 삭제라 문자열로 구분하지 못한다. 새 이미지에 B1 수리가 들어 있다는 것은 HEAD 조상 관계로만 추정한다.
  - 이미지 속 실행 파일에 vcs 정보가 없다(`-trimpath`, Dockerfile 빌드).
- 재확인(2026-11-25, 팀메이트): `docker image inspect … --format '{{.Id}}'` = `sha256:7be5c098b73eb7411bfb163600c94cb207db6e681ea3e775b00a9910b5593ca0` — **일치**. `docker ps` 의 `tossos-tossos-1` · `tossos-httpapi-1` 이미지 = `73682ed243df` — 운영 이미지 id 앞 12자리와 **일치**(태그가 옮겨져 이름 대신 id 로 표시됨, 전체 id 는 이 명령으로 다시 재지 않음).

## 3단계 — 호스트 `tossctl` 교체 · 백업 격리

- 명령: `make stage-local-update`, rc 0, HEAD `f73e0917b` 불변. 원문 로그: `stage-local-update.log`(Manager 원본 sha256 `860b919f1e68f3eaf4fd7bcb51d286a9377469664813e3fa193852deac0901f6`. 커밋본은 4 · 5행의 홈 경로만 `<HOME>` 으로 바꿨다 — 이 디렉터리의 다른 로그와 같은 표기. 나머지 바이트는 같다).
- 새 `<HOME>/.local/bin/tossctl`: sha256 `ed904db0e64644b341ba4af0c756479838495c40e735b4f0f31f513ff18b398a`, `vcs.revision=f73e0917b32ba02072cbf6fdf8380d3dbd0f6889`, `vcs.modified=false`, go1.26.5, `TOSSOS_ACCOUNT_BUDGET_MODE` grep 1건.
- 옛 파일: sha256 `5287fe3cb0bda6ed91ef2aa5c55df7e8deab84e07e7f644cb280acbd60873807`(2026-10-14 빌드, 예산 코드 0).
  - 격리 디렉터리 `<HOME>/.local/share/tossos-quarantine-bin/20261125T094623Z/`(700)에 `tossctl.pre-a166-20261125T094623Z` 로 `cp -p` 보존, 해시 일치 확인 뒤 `mv -f tossctl.candidate tossctl`.
- 같은 격리 디렉터리로 옮긴 백업 6개(결정 6 ①). 삭제 없음.
  - `tossctl.backup.b0e805f35b94c289da25864c129c79bc5066f950b3575357e01b743fdbda96f9`
  - `tossctl.bak-pre-de14674`
  - `tossctl.p3`
  - `tossctl.prev2`
  - `tossctl.prev3`
  - `tossctl.rollback`
- 교체 전 확인: 공용 바이너리를 쓰는 호스트 프로세스 0. 새 빌드 `soak attest --help` 에 `--record-renewal-status` · `--verified-by` 있음. `engine heartbeat-check` 있음.
- 계획과 다른 점:
  - 옛 파일을 `<HOME>/.local/bin` 이 아니라 격리 디렉터리에 보존했다. 예산 코드 없는 실행 파일을 PATH 디렉터리에 다시 두지 않으려고.
  - `tossctl.rollback` 을 격리해서 호스트 콘솔 시스템 업데이트의 되돌리기 슬롯이 비었다(`docs/operations.md:926-927`).
- 미확인:
  - attest 타이머가 새 파일로 도는지는 다음 실행(2026-11-26 00:00 KST) 뒤에 확인한다.
  - `tossctl version` 이 최신 버전 조회로 외부에 한 번 접속했다(브로커 아님).
- 되돌리기: 격리 디렉터리의 `tossctl.pre-a166-20261125T094623Z` 를 `<HOME>/.local/bin/tossctl` 로 `mv`.
- 재확인(2026-11-25, 팀메이트): `sha256sum <HOME>/.local/bin/tossctl` = `ed904db0…b398a` — **일치**. 격리 디렉터리 `ls` = 7개(`tossctl.pre-a166-20261125T094623Z` + 위 6개, 이름 모두 같음) — **일치**. 격리 디렉터리 안 파일의 해시와 디렉터리 모드는 다시 재지 않았다.

## 4단계 — `a166-budget-state` 빌드 · provision · check

- 도구 빌드: `go build -o <B>/bin/a166-budget-state ./tools/a166-budget-state`, sha256 `0f66d829f7c59c9a2d00dd7f01ab556c8292acd430ae1fc7c2a9e2f6603a5616`, `vcs.revision=f73e0917b32ba02072cbf6fdf8380d3dbd0f6889`, `vcs.modified=false`.
- 디렉터리: `mkdir -m 700` 으로 `<B>` · `<B>/root` · `<B>/launch` · `<B>/bin`. 경로 조상에 symlink 0, ext4, uid 1000.
- `provision --root <B>/root --manifest <M> --scope tossos-shared-1 --instance epoch-20261125-1` → rc 0 「provisioned」(2026-11-25 19:05 KST).
- `check --manifest <M>` → rc 0 「verified」.
- stdout · stderr 는 `<B>/provision.out` · `<B>/provision.err` 에 있다(커밋하지 않음).
- scope · instance 이름은 Manager 가 고른 불투명 이름이다. 계좌 · 키 정보 없음.
- manifest(v2) identity 와 실측 `stat -c '%d:%i'` 대조(Manager, 모두 일치):

  | 객체 | `%d:%i` | 모드 | nlink | uid |
  |---|---|---|---|---|
  | `<B>/root` | `66309:11853039` | 0700 | — | 1000 |
  | `<B>/root/queue.lock` | `66309:11853046` | 0600 | 1 | 1000 |
  | `<B>/root/wire.lock` | `66309:11853047` | 0600 | 1 | 1000 |
  | `<B>/root/event.lock` | `66309:11853049` | 0600 | 1 | 1000 |
  | `<M>`(manifest 자신) | `66309:11853051` | 0600 | 1 | 1000 |

- 아직 읽는 참여자 없음: 두 컨테이너 env 의 `TOSSOS_ACCOUNT_BUDGET_*` 0건, attest · heartbeat · a128-kr unit `Environment` 0건, 셸 env 0건. 모든 참여자 OFF 그대로.
- 재확인(2026-11-25, 팀메이트): `sha256sum <B>/bin/a166-budget-state` = `0f66d829…5616` — **일치**. `stat -c '%d:%i %u %a %h %n'` 다섯 객체 — 위 표의 `%d:%i` · uid · 모드 · nlink 모두 **일치**(root 의 nlink 는 2, 디렉터리라 표에는 적지 않음). manifest 파일 내용은 다시 읽지 않았다(identity 대조는 Manager 실측).

## 재확인 명령 원문 (읽기 전용, 팀메이트)

```text
sha256sum ~/.local/bin/tossctl ~/.local/share/tossos-budget/bin/a166-budget-state
stat -c '%d:%i %u %a %h %n' <B>/root <B>/root/queue.lock <B>/root/wire.lock <B>/root/event.lock <M>
ls -la ~/.local/share/tossos-quarantine-bin/20261125T094623Z/   # 7개
docker image inspect tossos:a166-account-shared-rate-limit-f73e0917b --format '{{.Id}}'
docker ps --format '{{.Names}} {{.Image}}'   # tossos 두 컨테이너
```

불일치 0건.

## 남은 단계

- cutover 순서 기준 5(정지 창 1, mount 리허설) · 6 · 7 · 8 은 미실시다.
- 운영 컨테이너 변경 0. `tasks.md` 4.2 미체크.
- 새 이미지(`tossos:a166-account-shared-rate-limit-f73e0917b`)에는 a168 과 B1 수리가 들어 있으나, 운영 반영은 별도 승인 사항이다. release · rollback 핀(결정 5 ①)은 아직 바꾸지 않았다.

## not-applicable

- 시험 · `make gate` · `make sdd-sync` · gstack 리뷰 — not-applicable. 문서 전용 기록이고 코드 변경 0 이다.
