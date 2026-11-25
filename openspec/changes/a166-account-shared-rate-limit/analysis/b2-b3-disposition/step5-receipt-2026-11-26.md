# a166 cutover 5단계 실행 영수증 (2026-11-26)

`proposal.md` 「전환(cutover) 순서 초안」 5단계(정지 창 1)를 Manager 가 사용자 승인 뒤 직접 실행한 기록이다. 값은 모두 Manager 실측(2026-11-26 03:12~03:17 KST)이다. 이 영수증을 쓴 팀메이트는 아무것도 다시 실행하지 않았고, 원본 판정 파일 하나만 읽어 sha256 을 쟀다. `<HOME>` 은 홈 디렉터리, `<B>` 는 `<HOME>/.local/share/tossos-budget`, `<deploy>` 는 `<HOME>/.local/share/tossos-deploy` 다(`prep-steps-2-4-receipt-2026-11-25.md` 와 같은 표기). 시각은 원 측정 그대로 UTC(`Z`)로 적고, 필요한 곳에 KST 를 함께 적는다.

범위: bind mount 2개만 추가하고 새 이미지를 같이 올렸다. 예산 env 는 넣지 않았다 → 예산 OFF 그대로. 코드 변경 0. `tasks.md` 4.2 미체크.

## 창과 사용자 승인 — 양 시장이 닫힌 창이 아니었다

- 2026-11-25: 사용자 「2번 창, 새 이미지 같이 올려」. 창 = 11-26 20:00 KST 이후.
- 2026-11-26 03:12 KST: 사용자 「5단계실행」.
- Manager 가 미국 정규장 중 · 엔진 가동 중임을 보고하고 세 선택지(06:10 KST / 20:07 KST / 지금)를 제시했다.
- 사용자 「1」(06:10 예약) → 이어서 「지금 실행」.
- 실행 시각 03:15 KST = 뉴욕 2026-11-25(수) 13:15 EST, **미국 정규장 중**. `proposal.md` 의 [정지] 규칙(「두 시장이 닫힌 창에서만」)을 지키지 않은 실행이다. 사용자의 명시 지시로 진행했다.

## 원본 판정 파일

- `<deploy>/schema-compatibility-a166-f73e0917b.txt` — 모드 0600, 2814 바이트, mtime 2026-11-26 03:16:08 KST, sha256 `bd75a2a79552edb2fd6eee2239bd4cd11ca6a5c8e9c999b405d986582429eeee`(팀메이트가 2026-11-26 에 잼). 이 영수증의 근거 원본이다. 커밋하지 않았다.
- 판정 요지(파일 원문에서 옮김):
  - 새 핀 `tossos:a166-account-shared-rate-limit-f73e0917b` SchemaVersion 36 → 마이그레이션 없음. 판정은 소스 기준(`e57d06cd6..f73e0917b` 에서 `internal/journal` diff 0, 2026-11-25 확인). 운영 저널 `user_version` 은 직접 재지 않았고, 새 엔진이 기동해 첫 대사를 clean 으로 마친 것이 실행 쪽 근거다.
  - 롤백 핀(a197 이미지, `sha256:73682ed2…`) SchemaVersion 36 → 롤백은 이미지만. 태그 유지 확인(교체 직전).
  - 롤백 때는 `account-budget-override.yaml` 을 빼고 `-f compose.yaml -f rollback-override.yaml` 둘만 준다(a197 이미지에는 B1 수리가 없다).

## 사전 점검 (2026-11-26 03:14:50 KST = 18:14:50Z)

- 두 운영 컨테이너: 이미지 `sha256:73682ed243dfa2915d0decb94a4a697d6fe155437a255cf49e63a43aa5daf3ff`, healthy, RestartCount 0.
- 새 이미지 `tossos:a166-account-shared-rate-limit-f73e0917b` = `sha256:7be5c098b73eb7411bfb163600c94cb207db6e681ea3e775b00a9910b5593ca0`. a197 태그 유지.
- `<deploy>/account-budget-override.yaml` sha256 `dbb277e5b0494c1fb91cd99d0294ac96dd17e646eb9540f7b75428ff6fad9c81`.
- 교체 전 `release-override.yaml` sha256 `b8e466847757a36ab15affcf076d3993ea41a4dc82ba55b2f4d39daf722d5e08`, `rollback-override.yaml` sha256 `7a50301e37727907f49166ec4c6058d0f2c08eb5939be43674a79cd41494ca1a`.
- `tossos-attest.service` inactive.
- `compose.yaml` sha256 `ecd03108a28d4ed484165ae9f7c130eda319575aa62061ee67f22371babf4340`.
- primary 체크아웃 HEAD `f71f6f75e`, 작업 트리 깨끗.

## override 변경

- 백업: `<deploy>/release-override.yaml.bak-20261125T181450Z`(`b8e46684…`) · `<deploy>/rollback-override.yaml.bak-20261125T181450Z`(`7a50301e…`).
- 새 `release-override.yaml`: sha256 `fc00afa3e3a99d6e56e53ba31ee3873b1c2633424e09ff404ea82c13149fc3f4`. 두 서비스 `tossos@sha256:7be5c098…`.
- 새 `rollback-override.yaml`: sha256 `b8e46684…`(= 옛 release). 두 서비스 `tossos@sha256:73682ed2…`.
- 병합 렌더링(`docker compose -p tossos -f compose.yaml -f release-override.yaml -f account-budget-override.yaml config`):
  - 두 서비스 image `tossos@sha256:7be5c098…`.
  - bind: `/var/lib/tossos/config` rw · `/var/lib/tossos/data` rw · `<B>/root` rw · `<B>/launch` ro.
  - `ACCOUNT_BUDGET` env 키 0.

## 교체

명령(저장소 primary 디렉터리에서, 서비스마다 한 번):

```text
docker compose -p tossos -f compose.yaml -f <deploy>/release-override.yaml -f <deploy>/account-budget-override.yaml up -d --no-deps --no-build --wait --wait-timeout 90 <service>
```

| 서비스 | 시작 | 종료 | rc | 출력 순서 | StartedAt |
|---|---|---|---|---|---|
| httpapi | 2026-11-25T18:15:09.793Z | 18:15:15.898Z | 0 | Recreate → Recreated → Starting → Started → Waiting → Healthy | 18:15:10.250Z |
| tossos | 18:15:27.090Z | 18:15:33.168Z | 0 | 같은 순서 | 18:15:27.531Z |

## 엔진

- 교체 전 마지막 대사: 18:14:54.07Z `reconcile.clean`.
- 18:15:27.31Z `engine.heartbeat` — 옛 엔진의 마지막 줄로 **추정**한다. 어느 프로세스가 썼는지는 재지 않았다(원본 판정 파일은 이것을 추정 표시 없이 「옛 엔진 마지막 로그」로 적었다).
- 새 `engine.operating_mode` 18:15:33.71Z(`gate_enabled` true · `gate_verified` true).
- 부재 약 6.4초. 이전 교체 실측 2~6초보다 길다.
- `reconcile.clean` blocked 0: 18:15:38.64Z · 18:16:41.13Z 두 번.
- 18:15:27Z 이후 로그 4줄 전부 INFO(WARN · ERROR 0).
- 18:16:53Z 두 컨테이너 healthy, RestartCount 0.
- 실행 digest: 두 컨테이너 모두 `sha256:7be5c098…`.

## B2 실측 — 컨테이너 쪽 객체

`/proc/<컨테이너 pid>/root<경로>` 를 `stat` 했다. 두 컨테이너 모두 같은 값이다.

| 객체 | `%d:%i` |
|---|---|
| `<B>/root` | `66309:11853039` |
| `<B>/root/queue.lock` | `66309:11853046` |
| `<B>/root/wire.lock` | `66309:11853047` |
| `<B>/root/event.lock` | `66309:11853049` |
| `<B>/launch/manifest.json` | `66309:11853051` |

- 호스트 `stat` 및 `prep-steps-2-4-receipt-2026-11-25.md` 4단계의 manifest v2 identity 와 일치.
- `docker inspect` Mounts: launch rw=false, root rw=true.
- 그 밖에 `/run/tossos/session.json` · `tls.crt` · `tls.key` bind(rw=false)가 있다. 이번에 추가한 것이 아니라 `compose.yaml` 쪽이다. 4단계 이전과 같은지는 따로 대조하지 않았다.

## 예산 상태 — OFF 그대로

- 두 컨테이너 `Config.Env` 에 `ACCOUNT_BUDGET` 을 포함한 키 0건.
- time namespace 는 호스트(`time:[4026531834]`)와 다르다: tossos `4026533806`, httpapi `4026533867`.
- `timens_offsets` 는 둘 다 monotonic 0 0 · boottime 0 0.

## 안 한 것 · 못 본 것

- 운영 저널 `user_version` 직접 측정 안 함. schema 36 판정은 소스 diff 0 + 새 엔진 기동 · clean 대사.
- 03:12 KST 에 본 직전 400줄 속 `reconcile.mismatch` 3건 · `engine.manual_buy_refresh_failed` 2건의 내용은 열어 보지 않았다.
- 열린 포지션 · 미체결 수는 재지 않았다.
- 콘솔 화면은 띄워 보지 않았다.
- 예산 required 모드 동작은 이 단계의 범위가 아니다(활성화는 7단계).
- 4.2 재실측 체크리스트 10항은 돌리지 않았다. `proposal.md` 5단계가 적은 「체크리스트 4 · 5」 중 4(같은 객체)는 위 B2 실측으로 했고, 5(컨테이너 쪽 파일시스템 · 모드)는 미확인이다.

## 이후 규칙

- 모든 `up` 은 `-f compose.yaml -f release-override.yaml -f account-budget-override.yaml` 셋을 함께 준다.
- 되돌릴 때는 `-f compose.yaml -f rollback-override.yaml` 둘만 준다(account-budget-override 제외).
- 남은 단계: 6(a184 B 창 종료 뒤 타이머 끔, 추정 2026-12-10) → 7(정지 창 2, 활성화) → 8.

## not-applicable

- 시험 · `make gate` · `make sdd-sync` · gstack 리뷰 — not-applicable. 문서 전용 기록이고 코드 변경 0 이다.
