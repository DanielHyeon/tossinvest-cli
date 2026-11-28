# a210 — 엔진 시험이 운영 audit 파일에 쓰지 않는다

## 왜

2026-11-28 a209 배포 뒤 확인에서, 운영 audit(`~/.local/share/tossos/audit.log` — 컨테이너 `/var/lib/tossos/data/audit.log`
로 붙는 그 파일)에 **실제로 일어나지 않은 설정 변경**이 기록돼 있음을 발견했다.

- 2026-11-28T02:27:23Z, `subject=daniel`(호스트 사용자 — 운영 엔진은 컨테이너 안이라 `unknown`)로 기동 설정 행 묶음이
  한꺼번에 기록됨: `engine.automation_gate.enabled true → false`, `engine.adoption.enabled true → false`,
  `engine.notifications.* → 빈 값`, `engine.cancel_external_opposite_buy true → false`, 그리고
  `engine.adoption.floor_only_exit '' → false`. 그 시각 운영 이미지는 a207 이고, a207 바이너리에는
  `engine.adoption.floor_only_exit` 문자열이 없다(`grep -ac` 0) — a208 개발 중인 코드를 돌린 호스트 프로세스가 쓴 것.
- 같은 모양(`subject=daniel`, 켜짐 → 꺼짐 묶음)이 11-16 이후 최소 9 번 있다(시각 불규칙 — 타이머가 아니라 개발 중 시험 실행 시각).
  운영 엔진이 다음 기동 때 실제 값으로 되돌려 써서(`subject=unknown`, 꺼짐 → 켜짐) 운영 동작은 바뀐 적 없다.
- 결과: audit 에 「사람이 자동매매 게이트를 껐다」는 거짓 기록이 반복해서 쌓였다. 프로젝트 불변식 「runtime config 변경은
  audit 로 추적 가능해야 한다」를 거짓 행이 깨뜨린다.

원인(Manager 가 코드로 확인, 구현 전에 CodeGraph·실측으로 재확인할 것):

- `internal/app/engine/engine_test.go:110` `engineOptions` 는 `ConfigDir` 만 임시 경로로 주고 `AuditFile` 을 비운다.
- `internal/app/engine/interlock.go:860` `openAuditLog` 는 `AuditFile` 이 비면 `journal.DataDir()` 로 간다.
  `journal.DataDir()`(`internal/journal/paths.go:64`)는 `TOSSOS_DATA_DIR` → `XDG_DATA_HOME` → `$HOME/.local/share/tossos` 순.
- `internal/app/engine` 시험 패키지에는 `TestMain` 이 없고 데이터 디렉터리 env 를 돌리는 곳이 없다. `engineOptions`·`startEngine` 호출 자리 20 곳.
- `audit.RecordChange` 가 같은 값을 다시 쓰지 않으므로, 운영 엔진이 기동한 뒤 **처음 도는 시험 한 번**만 흔적을 남긴다 —
  그래서 늘 일어나는데 가끔만 보였다.

## 무엇을

- 엔진 시험 패키지가 실제 사용자 데이터 디렉터리(그리고 설정 디렉터리)에 아무것도 쓰지 않게 한다 — 시험 쪽에서만 고친다.
- 같은 누수가 다른 시험 패키지에도 있는지 **실측으로** 센다(가짜 HOME 상자에서 시험을 돌려 생긴 파일을 센다).
- 이미 쌓인 거짓 행은 지우거나 고치지 않는다(audit 는 append-only 증거). 식별 규칙과 시각 목록만 change 에 기록한다.

## 비목표

- 생산 코드(`openAuditLog` 의 기본 경로 등) 변경 없음 — 운영 엔진은 그 기본 경로를 정상으로 쓴다. 생산 diff 0 이 목표이며
  그래서 spec delta 가 없다(`skip_specs: true`, a202 선례).
- 운영 audit 파일 편집 · 삭제 · 주석 행 추가 없음.
- 전체 시험(`make test`)을 측정 수단으로 쓰지 않는다 — 대상 패키지만.

## 영향

- 위험 등급: 시험 전용(High-risk 경로 생산 코드 변경 없음). 단 사용자 지시(2026-11-28)에 따라 구현자와 별도의 적대적 리뷰 + gstack 리뷰를 한다.
- 이 수리 뒤에도 남는 것: 다른 패키지가 새로 같은 누수를 만들 수 있다(가드 없음) — design 「남는 위험」.
