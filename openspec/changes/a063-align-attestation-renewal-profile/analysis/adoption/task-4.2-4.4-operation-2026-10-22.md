# a063 task 4.2~4.4 운영 기록 — 2026-10-22

이 파일은 `task-4.2-4.4-operational-evidence-template.md` 의 완료 기록을 채우는 실제 운영 기록이다.
§1~§3 은 **변경 조작 전에** 작성했다(task 4.2 문구: 승인 범위·대상 프로필·검토된 템플릿 digest·승인된
survey 창을 해당 조작 전에 기록). §4 이후는 조작 뒤에 덧붙였다.

시크릿·토큰·계좌번호·세션 파일 내용·원시 record·원시 attestation 값은 출력하지도 남기지도 않았다.
attestation 은 `issued_at`·`expires_at`·`soak_days`·`verified_by` 메타데이터만 읽었다.

## 1. 승인 (조작 전 기록)

| 항목 | 값 |
| --- | --- |
| 승인 원문 | 사용자 "승인" (2026-10-22, Manager 설명 직후, Manager 를 거쳐 전달) |
| 승인 범위 | 4.2 — 기존 `tossos-attest` user systemd 단위 백업, 검토된 단위 정의 설치, `systemctl --user daemon-reload`. 4.3 — 설치 시각부터 console-profile survey 시작(창: 설치 시각 2026-10-22 부터 연속 3 qualifying day 기록까지, 최대 7일). 4.4 — engine profile 에 새 attestation 이 쓰이고 renewal 실패가 보이는지 확인. |
| 명시 제외 | 매매 토글·제어 변경, 엔진 재시작·정지, 실주문, `mutating: true` 명령, `~/.config/tossctl` 안에서 승인된 서비스 자신이 쓰는 것 외의 편집. 승인은 이 범위를 넘지 않는다. |
| 바이너리 교체 | **승인 범위 밖 — 하지 않는다.** 템플릿 표의 "install only the reviewed binary" 와 deployment-plan 의 `install -m 0755 "$candidate"` 줄은 이번에 실행하지 않는다(편차, §3). design.md Migration 3 의 조건 "선택한 바이너리가 `--record-renewal-status` 를 지원하는지 설치 전 확인" 은 현재 설치된 바이너리로 충족한다(§2.3). |
| timer enable | deployment-plan 활성화 블록이 정한 그대로 `disable --now` → 설치 → `verify` → `daemon-reload` → `enable --now tossos-attest.timer`. 그 밖의 단위는 건드리지 않는다. |

## 2. 조작 전 상태 (read-only, 2026-10-22T06:16:45+09:00 = 2026-10-21T21:16:45Z)

### 2.1 대상 프로필과 엔진 프로필 일치

- console 대상 프로필: `$HOME/.config/tossctl` (= `/home/daniel/.config/tossctl`, 디렉터리·비심링크).
- 실행 중 엔진: 컨테이너 `tossos-tossos-1`(이미지 `788dd6995faa`, `tossos:a157-first-leg-binding-before-sizing-e537ab23c`)
  안의 `tossctl engine run --config-dir /var/lib/tossos/config`(2026-10-19 20:26:12 KST 기동).
- `docker inspect` 마운트: `bind /home/daniel/.config/tossctl -> /var/lib/tossos/config rw=true`.
- 같은 파일 증명: 호스트 `capability-attestation.json` inode `11847558` = 컨테이너
  `/var/lib/tossos/config/capability-attestation.json` inode `11847558`.
- 판정: **엔진 프로필 = console 대상 프로필(bind mount 로 정규화하면 같은 디렉터리).** 2026-09-06 의 차단 사유
  (엔진 프로세스 0개)는 현재 해소되어 있다.

### 2.2 survey (4.3 대상)

- 같은 컨테이너 안에서 `tossctl soak run --config-dir /var/lib/tossos/config` 1개가 2026-10-19 20:26:19 KST 부터 실행 중.
- 같은 프로필 record `capability-soak.jsonl` mtime 2026-10-22 06:02:52 KST (계속 추가 중).
- 템플릿 절차대로 **이미 실행 중인 survey 는 중단·중복하지 않는다.** 셸에서 `soak run` 을 새로 띄우지 않는다.
- `soak status`(로컬 record 만 읽음, 네트워크 0) 결과: cycles 5959 (2026-08-03T00:26:00Z .. 2026-10-21T21:01:33Z),
  credential streak 6 of 3, 최근 일자 2026-10-15 FAIL(auth failures 27) 이후 10-16~10-21 pass, `READY`.
  이 과거 일자들은 4.3 증거로 쓰지 않는다(소급 금지).

### 2.3 설치된 바이너리

- `$HOME/.local/bin/tossctl`: 정규 파일·비심링크, SHA-256 `5287fe3cb0bda6ed91ef2aa5c55df7e8deab84e07e7f644cb280acbd60873807`.
- `soak attest --help` 에 `--record-renewal-status` 있음(grep exit 0). 이 바이너리를 그대로 쓴다.
- 참고: 2026-09-06 검토 후보 `/tmp/a063-reviewed-20260906/tossctl`(`d9ef7f5e…`)는 이번 승인 범위가 아니다.
- 컨테이너 console 바이너리도 `--record-renewal-status` 를 가진다(a063 구현 `676bd4b46` 는 main 의 조상) — renewal-status
  reader 가 운영 console 에 이미 있다.

### 2.4 설치된 단위 (교체 대상)

| 파일 | SHA-256 | 상태 |
| --- | --- | --- |
| `~/.config/systemd/user/tossos-attest.service` | `1524b07ae9765b883c232fe855aead90ef6320286dfce22a8b67a66a88fc055b` | static, inactive/dead, Result=success, ExecMainStatus=0 (2026-10-22 02:34:18 KST) |
| `~/.config/systemd/user/tossos-attest.timer` | `983cac7dfea15f23e7aa13001d191ce086d63e9069786eac03c6cb3778cc8d23` | enabled, active/waiting, `OnBootSec=10min`·`OnUnitActiveSec=6h`·`Persistent=true` |

- 두 단위 모두 FragmentPath 가 정확히 위 경로이고 DropInPaths 비어 있음.
- 설치된 service 는 이미 `--config-dir %h/.config/tossctl` 을 쓰지만 `/bin/sh -c '… || printf "not earned yet"'` 로
  **거절 exit 를 성공으로 가린다**(design 2026-09-05 계약의 바로 그 결함).
- `systemctl --user list-timers` (발췌): `tossos-attest.timer` NEXT 2026-10-22 08:34:18 KST, LAST 02:34:18 KST.
  나머지 `tossos-a128-kr`·`tossos-a128-us`·`tossos-a140-kr` timer 는 이번 범위 밖이며 건드리지 않는다.
- persistent stamp: `~/.local/share/systemd/timers/stamp-tossos-attest.timer` mtime 2026-10-22 02:34:18 KST.

### 2.5 현재 attestation (메타데이터만)

- 경로 `~/.config/tossctl/capability-attestation.json`, mode 0600, mtime 2026-10-22 02:34:18 KST.
- `issued_at` 2026-10-21T17:34:18Z, `expires_at` **2026-11-20T17:34:18Z**, `soak_days` 6, `verified_by` `tossos-attest.timer`.
- `<attestation>.renewal-status.json` 없음(이전 단위는 `--record-renewal-status` 를 쓰지 않음).

## 3. 설치할 정의와 검토 digest (조작 전 기록)

| 파일 | 검토된 digest (deployment-plan·template) | 현재 저장소 파일 digest | 일치 |
| --- | --- | --- | --- |
| `deploy/systemd/tossos-attest.service` | `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96` | `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96` | 예 |
| `deploy/systemd/tossos-attest.timer` | `0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde` | `0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde` | 예 |

서비스가 실행하는 명령(유일한 `ExecStart`, 셸 없음):

```text
%h/.local/bin/tossctl --config-dir %h/.config/tossctl soak attest --record-renewal-status --verified-by tossos-attest.timer
```

판정: **non-mutating.** `runSoakAttest`(cmd/tossctl/soak.go:476) 는 로컬 soak record·로컬 verify record 를 읽고,
기준을 채우면 attestation 파일 하나와 `<attestation>.renewal-status.json` 하나만 쓴다. 주문·취소·엔진 재시작·토글
변경 경로가 없고 네트워크 호출이 없다. `--record-renewal-status` 는 `--record`/`--out` 을 거절하므로 경로는 프로필에서만
유도된다.

timer 는 `OnCalendar=*-*-* 00/6:00:00`, `Persistent=true` — 로컬 시각 00·06·12·18 시. 기존 stamp(02:34)가 06:00 슬롯보다
앞이므로 `enable --now` 직후 한 번 즉시 실행될 수 있다(Persistent 의 정상 동작, 수동 `start` 는 하지 않는다).

### 편차 (검토 문서 대비)

1. 바이너리 설치 줄 생략(§1). 대신 설치된 바이너리가 플래그를 지원함을 §2.3 에서 확인했다.
2. 백업 디렉터리에는 deployment-plan 대로 바이너리·service·timer 셋을 다 복사한다(롤백 블록이 셋을 복원하므로).
   바이너리를 교체하지 않으니 복원은 같은 내용이다.

### 승인된 survey 창

- 시작: 설치(§4 의 실제 `enable --now` 시각) — 2026-10-22.
- qualifying day 의 단위는 `soak status` 의 일자(UTC 날짜)다. 설치 시각(2026-10-21T21:xxZ) 이후 **처음으로 온전히
  시작되는 일자 2026-10-22(UTC)** 부터 센다. 가장 이른 완료: 2026-10-22·10-23·10-24(UTC) 모두 pass →
  2026-10-25 09:00 KST 이후 판정 가능.
- 상한: 7일 — 2026-10-29(UTC) 일자까지. 그 안에 연속 3일을 못 채우면 4.3 은 미완으로 보고한다. 일자 하나라도
  FAIL 이면 연속 계수를 다시 시작한다.

## 4. 조작 (task 4.2) — 2026-10-22 06:18 KST

§1~§3 은 커밋 `ae7000458` 로 조작 전에 고정했다. 실행은 deployment-plan 활성화 블록에서 바이너리 설치 줄만 뺀
스크립트 한 번(exit 0)이다. 순서와 결과:

| 단계 | 결과 |
| --- | --- |
| 검토 digest 재확인(`sha256sum -c`) | service OK · timer OK |
| 설치된 바이너리 `--record-renewal-status` 지원 | 있음 |
| 전제 재검증(정규 파일·비심링크·FragmentPath 정확·drop-in 없음·timer enabled+active·service inactive) | 모두 통과 |
| 백업 디렉터리 | `/home/daniel/.config/systemd/user/a063-backup-20261021T211830Z` |
| 백업 검증(`cmp -s` 원본=사본) | 셋 다 일치 — `tossctl` `5287fe3c…3807`, service `1524b07a…055b`, timer `983cac7d…cc8d23` |
| `systemctl --user disable --now tossos-attest.timer` | 06:18:30 KST, 이후 service inactive 확인 |
| `install -m 0644` service·timer | 설치본 digest `0bf50972…de96` · `0db4e0cb…1edde` = 검토 digest, mode 644 |
| `systemd-analyze --user verify` (설치본 둘) | exit 0, 경고 없음 |
| `systemctl --user daemon-reload` | exit 0 |
| `systemctl --user enable --now tossos-attest.timer` | 06:18:31 KST, enabled · active |

하지 않은 것: 바이너리 교체, 다른 단위(`tossos-a128-*`·`tossos-a140-kr`) 변경, `systemctl start tossos-attest.service`
수동 실행, 엔진·console·survey 재시작/정지, 토글·설정 변경, 주문 명령, `make` 명령.

롤백: deployment-plan 「Proposed rollback」 블록의 `backup=` 에 위 디렉터리를 넣어 실행한다.

## 5. 첫 renewal 주기 관측 (read-only)

`enable --now` 직후 `Persistent=true` 가 놓친 06:00 KST 슬롯을 즉시 실행했다(기존 stamp 02:34 < 06:00).
수동 트리거는 쓰지 않았다.

| 관측 | 값 |
| --- | --- |
| `tossos-attest.service` | 시작·종료 2026-10-22 06:18:31 KST, `Result=success`, `ExecMainStatus=0` |
| `tossos-attest.timer` | LAST 06:18:31 KST, NEXT **2026-10-22 12:00:00 KST** (이후 18·00·06 시) |
| attestation (engine profile) | mtime 2026-10-22 06:18:31 KST, `issued_at` 2026-10-21T21:18:31Z, `expires_at` **2026-11-20T21:18:31Z**, `soak_days` 6, `verified_by` `tossos-attest.timer` — 설치 시각 이후 새로 쓰임 |
| renewal-status | `~/.config/tossctl/capability-attestation.json.renewal-status.json`, mode 0600, 154 B: `format_version` 1, `attempted_at` 2026-10-21T21:18:31Z, `outcome` `issued`, `reason_codes` 없음, `expires_at` 2026-11-20T21:18:31Z |
| 엔진 컨테이너에서 보임 | `/var/lib/tossos/config/` 아래 두 파일 모두 같은 크기·시각으로 보임(bind mount) |
| 운영 console 표면 | `GET https://127.0.0.1:37085/verify-console` 200 — `renewal 진단: issued · 마지막 시도 2026-10-21 21:18Z (1시간 미만 전)`, `attestation 만료까지 약 720시간 남았다.`, 상태 `유효`, 발급/만료 2026-10-21 → 2026-11-20 |
| 엔진·survey·console 프로세스 | 조작 전과 같은 PID·기동 시각(engine 2026-10-19 20:26:12, soak run 20:26:19) — 재시작 없음 |

journal 출력에는 엔드포인트 목록·만료·"gate will still refuse (clause 9)" 안내만 있고 이 기록에는 위 메타데이터만 옮겼다.

## 6. 남은 것

- **4.2: 완료.**
- **4.3: 진행 중.** survey 는 같은 프로필에서 이미 돌고 있어 중단·중복하지 않았다. 셀 일자는 2026-10-22(UTC) 부터,
  가장 이른 판정은 2026-10-25 09:00 KST 이후, 상한 2026-10-29(UTC). 각 일자는 `tossctl --config-dir ~/.config/tossctl
  soak status`(로컬 읽기) 의 일자 행으로 기록한다.
- **4.4: 부분.** (1) 프로필 일치 ✔ (2) 설치 이후 새 attestation ✔ (3) renewal-status 가 같은 프로필 sibling 경로에 있고
  마지막 renewal 과 일치 ✔ (5) 엔진 재시작·주문 변경 0 ✔. 남은 것은 (4) **실패 가시성의 운영 관측** — 새 단위는
  셸 `||` 가림이 없어 거절 시 systemd `Result=exit-code` 로 남고 renewal-status 가 `refused`/`failed` 가 되지만,
  이번 주기는 `issued` 라 실제 실패는 관측되지 않았다. 실패를 일부러 만들려면 승인 범위 밖 조작(record·프로필 변경)이
  필요하므로 하지 않았다. 템플릿 4.4 는 3일 창 뒤에 확인하도록 되어 있으므로 창이 끝날 때 다시 판정한다.
- 4.4.1 base 재고정, 4.5 독립 리뷰·PM 동기화·`make gate`·archive 는 그 뒤.
