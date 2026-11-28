# a210 D5 — 운영 audit 의 거짓 행 묶음 (읽기 전용 분석)

- 대상: 운영 audit `~/.local/share/tossos/audit.log`(컨테이너 `/var/lib/tossos/data/audit.log`). **열어서 읽기만 함** — 편집 · 삭제 · 주석 행 추가 없음.
- 도구: `contamination_scan.py`(같은 디렉터리), 출력 원문 `contamination-scan.txt`. 실행 2026-11-28, 그때 파일 1,247 행
  (호스트 subject 587 · `unknown` 660).
- 적는 것: 시각 · 행 수 · 설정 이름 · 불리언 토글의 old→new 만. 한도 · 금액 · 토픽 · URL · 계좌 값은 출력도 기록도 안 함.

## 식별 규칙

1. subject 가 호스트 사용자인 행만 본다(운영 엔진은 컨테이너 안이라 `unknown`).
2. 그 행들을 시각 간격 1 초 이하로 이어 붙여 묶음을 만든다(다른 subject 행이 끼면 끊음).
3. 묶음 안 `engine.*` 설정 이름이 **2 개 이상**이고 불리언 토글 하나 이상이 **true → false** 인 묶음을 후보로 친다
   (`recordGateSettings` 기동 행 묶음의 모양 — 사람의 CLI 조작은 설정 하나씩 쓴다).
4. 후보 중 운영 컨테이너가 처음 쓴 시각(**2026-07-30T19:31:11Z**, 첫 `unknown` 행) 이후의 것을 거짓 행 묶음으로 판정한다.
   그 전 묶음은 호스트 엔진이 실제로 기동했을 수 있어 판정 불가로 가른다.
5. 보조 증거 `restored=k/n`: 묶음에서 true→false 된 토글 n 개 중, 그 뒤 첫 `unknown` 행이 false→true 로 되돌린 수 k.
   운영 엔진이 다음 기동 때 실제 값을 다시 쓴 것 — 운영 동작이 바뀐 적이 없다는 증거.

## 결과

| 분류 | 묶음 수 |
|---|---|
| 호스트 subject 묶음 전체 | 67 |
| 규칙 3 후보 | 36 |
| 컨테이너 이후(거짓 행 판정) | **35** — `restored` 완전 34, 부분 1(2026-08-04T13:03:26Z, 3/4) |
| 컨테이너 이전(판정 불가) | 1 — 2026-07-30T12:27:08Z(1/2) |

- 거짓 행 묶음 시각 목록 · 설정 이름 · 토글은 `contamination-scan.txt` 의 `era=container` 줄 35 개.
- 2026-11-16 이후만 세면 10 묶음(proposal 의 「최소 9 번」과 맞음). 마지막은 2026-11-28T02:27:23Z(17 행,
  `engine.adoption.floor_only_exit ∅→false` 포함 — a208 개발 코드).
- 묶음마다 바뀐 토글은 늘 같은 모양: `engine.automation_gate.enabled` · `engine.adoption.enabled` · `engine.notifications.enabled` ·
  `engine.notifications.topic_configured` true→false, 2026-11-25 이후 `engine.cancel_external_opposite_buy` true→false 추가.
  시험 config 의 기본값(꺼짐)이 운영 값(켜짐) 위에 기록된 것.

## 읽는 사람에게

- 이 35 묶음의 「게이트 · 편입 · 알림을 껐다」 행은 실제 설정 변경이 아니다. 같은 시각 운영 엔진의 설정은 그대로였다(`restored`).
- 2026-08-04T13:03:26Z 묶음의 되돌려지지 않은 토글 1 개는 이 스캔이 원인을 가르지 않는다(값 비교는 범위 밖). 거짓 행 판정 자체는 규칙 1~4 로 성립.
- 운영 audit 에 이미 쌓인 행은 append-only 증거라 그대로 둔다(design 「남는 위험」).
