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
