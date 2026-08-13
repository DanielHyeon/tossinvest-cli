# Tasks: a103-the-rollback-pin-is-made-not-remembered

## 1. 판단을 테스트 닿는 자리에 만든다

- [x] 1.1 [T] RED — `tools/deploy/test_image_pin.py`. 구현 전 `ModuleNotFoundError`로
  실패 확인.
- [x] 1.2 [T] GREEN — `tools/deploy/image_pin.py`. `orphan_verdict`는 이 빌드가
  롤백 대상을 없애는지 판정하고, `pin_name`은 `tossos:<change>-<commit>`을 만든다.
- [x] 1.3 [T] 통과가 증거인지 뮤테이션으로 확인 — 가드의 `MOVING_TAG` 제외 제거,
  `unknown` commit 검사 제거. 각각 다른 테스트를 실패시켰다. 새 파일이라
  `git checkout` 원복은 파일을 지우므로 사본으로 되돌리고 md5로 확인.
- [x] 1.4 [T] **리뷰 반영 — 관측 상태 셋을 구분한다.** `present`/`absent`/`unknown`.
  docker를 못 읽은 것은 "잃을 게 없다"가 아니라 거부다. 초판은 `2>/dev/null || true`로
  셋을 하나로 접었다. (리뷰 C2)
- [x] 1.5 [T] **리뷰 반영 — 핀 이름 선점을 거부한다.** `docker tag`는 이름을 옮기므로
  같은 커밋 재빌드가 직전 이미지에서 핀을 빼앗는다. 이 change가 새로 만든 유실
  경로였다. (리뷰 C3)
- [x] 1.6 [T] **리뷰 반영 — 빈 RepoTags가 "이미 무명"이다.** `{{json .RepoTags}}`는
  태그 없는 이미지에 `[]`를 낸다. 초판 테스트는 `<none>:<none>`이라는 docker가 내지
  않는 모양을 검사했고 실제 경로에서는 "이미지 없음 → 안전"으로 빠졌다.
- [x] 1.7 [T] **리뷰 반영 — docker 태그 128자 상한.** 넘으면 빌드가 끝난 뒤 거부당한다.
- [x] 1.8 [T] 18/18 통과.

## 2. 태그를 파괴가 일어나는 명령 안으로 옮긴다

- [x] 2.1 `make image CHANGE=<id>` — 빌드 전 가드와 빌드 후 자동 핀. 두 유실 형태를
  각각 막는다: 08-11·08-12는 직전 이미지가 이름을 잃은 형태, 08-13은 새 이미지가
  이름을 못 받은 형태.
- [x] 2.2 `.PHONY`에 `image` 등록. `make lint` rc=0.
- [x] 2.3 **리뷰 반영 — 빌드 후 블록을 `set -e`로 묶고 결과를 되읽어 검증한다.**
  초판은 `;` 체인의 마지막이 `echo`라 tag가 죽어도 rc=0에 "박았다"를 찍었다.
  이제 `[ -n "$id" ]` → `docker tag` → 되읽어 `[ "$got" = "$id" ]`. (리뷰 C1)
- [x] 2.4 **리뷰 반영 — CHANGE/COMMIT을 환경으로 넘긴다.** recipe 문자열에 끼우면
  셸이 먼저 해석하므로 따옴표가 든 값은 검증에 닿기 전에 탈출한다. 타겟별
  `export`로 바꿨다. (리뷰 C4)
- [x] 2.5 [T] 종단 7 시나리오를 스텁 docker로 실행 — 데몬 정지·정상·local뿐·핀 선점·
  tag 실패·첫 빌드·셸 탈출. 실제 데몬 미접촉(리뷰 전후 이미지·컨테이너 동일 확인).

## 3. 문서는 실행되는 것을 가리킨다

- [x] 3.1 `docs/operations.md` 교체 절차가 `make image CHANGE=…` 한 명령을 보여준다.
- [x] 3.2 마이그레이션이 있었던 경우 롤백은 `pre_migration_backup_path`의 백업을
  복원한 뒤 그 이미지를 띄우는 것이라는 사실을 기록. 2026-08-13 배포가 저널을
  30→31로 올렸으므로 현재 핀 `tossos:a101-2202ed9a`(schema 30)에 그대로 해당한다.
- [x] 3.3 이번 배포가 이름을 못 받은 이미지에 핀을 박았다 — `tossos:a098-ee8eca83`
  → `66553dba92d8`.
- [x] 3.4 **리뷰 반영 — 설치 절의 `docker compose build`를 최초 설치 전용으로 표시**하고
  이후 빌드는 `make image`임을 명시. 그 블록이 복붙되는 블록이다. (리뷰 C6)

## 4. 완료 조건

- [x] 4.1 Function Logic Map — **not-applicable.** 사유는 `review.md` 첫 절.
  기존 Go 함수 내부를 편집하지 않고, 분기를 근거로 삼는 주장도 없다.
- [x] 4.2 **리뷰 반영 — `tools/deploy`를 `make sdd-test`와 `compileall`에 넣는다.**
  넣기 전에는 이 change의 테스트가 아무 자동 실행에도 닿지 않았고, 그러면
  "판단을 테스트 닿는 자리에 뒀다"는 이 change의 논거가 성립하지 않는다. (리뷰 C5)
- [x] 4.3 독립 리뷰 — `review.md`. 치명 6 · 정보 6, 치명 전부 반영.
- [x] 4.4 PM 동기화 — `STORY-TOS-a103`, `_registry.yaml`, `FEAT-TOS-007` 역링크.
- [x] 4.5 `make sdd-test` OK · `make lint` rc=0
- [x] 4.6 `make sdd-sync && make sdd-check && make gate CHANGE=a103-the-rollback-pin-is-made-not-remembered`
  — sync 완료, check는 tracker 재생성 후 통과, gate 재실행 통과 (아래 커밋 직전 실측)
