# a180 리뷰 기록

- 날짜: 2026-11-04
- 위험 등급: Small(개발 도구·시험). High-risk 경로 영향 없음. 게이트 판정 변경 없음.

## 절차 이탈 기록

- **역순.** 구현과 RED/GREEN 이 2026-11-04 `/doctor` 후속으로 Story·계약·base 고정·증거 조정·Pre-Edit 보다 먼저 일어났다.
  READY 를 거치지 않았다. `docs/WORKFLOW.md` 기준 이 등급은 계약이 필수는 아니며, 사용자 선택으로 사후에 맞췄다.
- **위임.** 구현자가 Manager 세션 자신이다. Terra 위임 없음(이 세션에 paseo 도구 없음). 검증은 분리된 리뷰 패스에 맡긴다.
- **첫 리뷰.** 수리 직후 리뷰는 gstack 이 아닌 Claude 계열 서브에이전트 1개였다. 아래 1차 기록이다.

## proposal-freeze

사후 작성이라 동결 시점이 없다. 도구 change 의 경량 리뷰(validate + Manager 셀프리뷰 + 기록)에 해당한다.
분리된 gstack `review` 결과는 아래 「gstack 리뷰」에 기록한다.

## 생략과 면제

- Function Logic Map: not-applicable — base→작업 트리에 수정된 Go 함수가 0개다(`.sh`·`.json`·`.py` 만 변경).
- Pre-Edit Gate: not-applicable — 비자명 production 코드가 아니고 High-risk 경로가 아니다. 사후 기록임을 밝힌다.
- 변이 원장·다중 적대 리뷰: not-applicable — 비례 원칙(High-risk·게이트 판정 변경이 아님). 변이 하네스는 별도로 돌렸다.
- CodeGraphContext·GBrain: 갱신 실패로 advisory 미갱신(`analysis/code-context/evidence-reconciliation.md`).

## 1차 리뷰 (서브에이전트, 2026-11-04)

| 지적 | 판정 | 처리 |
|---|---|---|
| 수동 실행이 잠금에 막히면 쓰지 않고 exit 0 | 수용 | 120초 대기 후 exit 1 |
| 5분 간격이 마지막 상태를 버림 | 수용 | 턴 종료 `Stop` 훅 추가 |
| 수동 메모가 저장마다 지워짐 | 수용 | 다음 요약으로 이전 |
| 시험이 `-print` 누락을 못 잡음 | 수용 | 단언 강화, 변이로 확인 |
| 같은 초 저장 시 백업 덮어쓰기 | 거절 | 드묾. `design.md` 에 `unsupported` 로 기록 |
| prune 대상 확대 | 거절 | 측정상 효과 없음(1.66초 대 1.39~1.53초) |
| `LATEST_JSONL` 작은따옴표 | 거절 | 이 저장소 경로에 없음. 기존 동작 |
| `async` 키가 무시되는지 | 미해결 | 새 세션 측정(tasks 4.1). 기존 `codegraph sync` 훅도 같은 키를 쓴다 |

## 검증 결과

실행한 명령과 결과는 `analysis/verification.md` 에 있다.

## gstack 리뷰

### 1차 (2026-11-04): CHANGES REQUESTED

- **보이스:** 작성자와 분리된 Claude 리뷰어, Claude 적대 서브에이전트 1개, Codex CLI(`codex exec -s read-only`, 정적 검토만). 세 보이스 합의: 1·2·5번.
- **완결도:** 스킬 기준 INCOMPLETE. Testing·Maintainability 전문가 미실행, `gstack-review-log`·`gstack-gate-diff`·`git fetch` 미실행.

| # | 지적 | 판정 | 처리 |
|---|---|---|---|
| 1 | 비동기 저장의 git 조회가 `index.lock` 을 잡아 동시 git 쓰기 실패 (이 change 가 만든 회귀, 재현 2/256 등) | 수용 | `GIT_OPTIONAL_LOCKS=0`. 리뷰 제안만으로는 부족했다: 새 시험이 `git diff --stat` 의 index 쓰기를 잡아 `diff.autoRefreshIndex=false` 를 추가함 |
| 2 | 비원자 쓰기로 메모 손실, 쓰기 실패가 exit 0 | 수용 | 임시 파일 + `mv`, 실패 시 exit 1, 백업 성공 시에만 회전 |
| 3 | 시험이 시나리오를 못 박지 못함. 추가 변이 10개 중 9개 생존. `no status cap` 변이 이름이 범위를 과장 | 수용 | 시험 17개로 보강, 변이 31개, 이름 수정 |
| 4 | 자동 경로 실패가 조용히 사라짐 | 수용 | `save-failures.log` |
| 5 | 다른 세션 대화가 요약에 섞임 | 부분 수용 | 기존 동작. `design.md` 위험에 기록. 코드 수정 안 함 |
| 6 | 자식이 잠금 FD 상속, 시간 제한 없음 | 부분 수용 | shim 재현뿐이라 `unsupported` 로 `design.md` 에 기록 |
| 7 | 메모 안의 비슷한 줄에서 잘림 | 수용 | 제목 전체 일치 |
| - | 문서 사실 정정 3건(5분 상한, 요약 크기, 호출자 수) | 수용 | 반영 |
| - | Codex: 메모 편집 경쟁, Stop 대기열 | 리뷰어가 기각 | 조용한 실패 부분만 4번에 반영 |

**이 change 의 첫 수리가 회귀를 만들었다.** 동기 훅을 비동기로 바꾼 것이 원인이고, 수정 전까지 이 세션의 작업 트리에서 훅이 켜져 있었다.
리뷰 시점에 실제 `.git/index.lock` 잔류는 없었다. 실제 충돌 발생 여부는 재지 않았다.

### 2차 (2026-11-04): CHANGES REQUESTED

- **보이스:** 1차의 리뷰어 1명. Codex·서브에이전트는 돌리지 않았다.
- **(a) 동시 git 쓰기:** `/tmp` 픽스처(미추적 3만, 추적 3천)에서 양성 대조군 대 수리본. `git add` stat-clean 12/231 대 0/246, stat-dirty 15/293 대 0/384, `index.lock` 폴링 관측 1,371/68,091 대 0/123,565. 실제 저장소 경합은 재지 않았다(tasks 4.1).
- **(b) 원자 쓰기:** 새 결함 없음. write 오류·SIGTERM·SIGKILL·로그 쓰기 불가 경로를 재현했다.
- **git 2.43 주장:** `git diff --stat` 이 `GIT_OPTIONAL_LOCKS=0` 과 `--no-optional-locks` 를 무시함을 리뷰어가 재현했다. 1차 제안 「한 줄 수정」은 리뷰어가 정정했다.
- **5·6번 무수정 판단:** 리뷰어가 맞다고 판정했다.

| # | 지적 | 판정 | 처리 |
|---|---|---|---|
| 1 | `age_summary(299)` 가 초 경계에서 흔들림(2/400, 부하 3/150, 290초는 0/400) | 수용 | 290초로 변경 |
| 2 | 스펙 「Every failure」 가 거짓. 백업·git 실패는 rc 0, 로그 없음 | 수용 | 스펙을 네 실패로 좁히고 `design.md` 위험에 기록 |
| 3 | 변이 수 32 는 틀림. 31 이다(대조군까지 센 줄 수) | 수용 | 정정 |
| 4 | 「`.git/index` mtime 불변」 은 수리 전에도 나오는 값이라 증거가 아님 | 수용 | `verification.md` 에서 삭제하고 픽스처 수치로 대체 |
| 5 | `\| 0` 줄 주장이 재현 안 됨 | 수용 | 삭제 |
| 권고 | 요약 안내문이 믿으면 안 되는 파일로 이어받으라고 함 | 수용 | 안내문에 경고 한 줄 추가 |

**최종 판정 상태:** 고친 뒤 리뷰어의 APPROVE 는 받지 않았다. 리뷰어는 코드 수정이 시험 한 줄뿐이면 전체 재리뷰 없이 결과 기록으로 충분하다고 봤다.
그 뒤 안내문 한 줄이 더 추가됐다. 결과는 `analysis/verification.md` 에 있다.

## base 재고정 — Manager 승인 2026-11-25

- **승인:** Manager(Fable) 2026-11-25. 조건: 옛 base `232d6b86c` 이후 이 change 디렉터리를 만지며 `.go` 를 고친 비병합 커밋이 0 이면 base 를 현재 main HEAD 쪽으로 재고정한다(`docs/WORKFLOW.md` 「사람 승인 base 재고정」).
- **사유:** 이 change 는 고정 번들이 0 이라 `--record-landing` 을 받을 수 없고, gate ⑤ 창의 끝이 워킹트리라 base 뒤에 착지한 형제 change 의 Go 함수가 required 로 들어온다(옛 base 에서 required 55, 창 안 커밋 274).
- **귀속 실측(요건 1):** `analysis/base-refix/attribution.txt` — 자기 Go 커밋 0(이 디렉터리를 만진 커밋은 `ca7a3f4f7` 하나, `.go` 0개, renumber 없음). required 55 함수의 34 파일은 전부 형제 착지 몫(a187 · a189 · a191 · a192 · a193 · a194 · a195 · a196 · a198 · a199 · a200 · a201 · `1ea261100`).
- **요건 3:** `base-commit.txt` 한 파일만 단독 커밋, 메시지에 영수증.
