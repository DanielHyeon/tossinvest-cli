# 영수증 — 6.1~6.3 실기록 대사 확인 (2026-10-28)

읽기 전용. 계좌번호는 `****`, 브로커 requestId 는 `<가림>` 으로 바꿈. 조건주문 id 는 tasks 0.6 영수증에 이미 기록된 값이라 그대로 둠.

## 6.1 배포

- a164 커밋 6개 `1e5a848a9` · `35f4e952a` · `4785e41ad` · `ffeb4fb79` · `ba07b6c08` · `cf6ced797` 가 모두 `eaf713ff9` 의 조상임
  (`git merge-base --is-ancestor <c> eaf713ff9`, 2026-10-28 종결 에이전트 재실행 — 6/6 ok).
- 운영 컨테이너 `tossos-tossos-1` 이미지 `sha256:e6d61dd03684a911872d6afdf0a3e58287b8b0cb66dadad06bb2c1c29411e0f4`, StartedAt `2026-10-27T09:36:38Z`
  (`docker inspect`, 2026-10-28 재확인). 바이너리에 `reconciled_expired` · `expiry-evidence-absent` · `expiry-sources-disagree` 문자열이 있음 — Manager 확인(2026-10-28).
  교체는 사람이 2026-10-27 에 함. 이 종결은 이미지·컨테이너를 건드리지 않음.

## 6.2 실기록 대사 (사람 실행)

- 기록 `~/.config/tossctl/capability-verify.jsonl`(컨테이너 안 `/var/lib/tossos/config/capability-verify.jsonl`) 28번째 줄:
  `kind=reconcile` · `step_id=reconcile-expired` · `title=reconciled expired` · `format_version=2` ·
  `started_at=2026-10-23T00:24:47.495639008Z` · `finished_at=2026-10-23T00:24:48.219426609Z`(09:24 KST) ·
  artifact `kind=conditional-order` · `symbol=333430` · `reconciled_expired=true`.
  필드 이름·값만 뽑아 읽음(계좌 봉인값·관측 원문은 출력하지 않음).
- 사실 기록: 그 실행 때 엔진이 약 12초 멈췄음 — 장중(09:24 KST)이고 엔진 정지 창 관례(KST 05:00~09:00) 밖이었음(Manager 확인).

## 6.3 읽기 전용 확인

명령(2026-10-28, 종결 에이전트 재실행 — Manager 가 같은 날 먼저 실행): `docker exec tossos-tossos-1 tossctl --config-dir /var/lib/tossos/config verify status --market KR`
— `verify status` 는 `source: local`, 「Reads only the local record. It makes no network call.」(`cmd/tossctl/verify.go:222-230`), lock 없음. rc 0.
PATH 의 호스트 `tossctl`(10-13 dirty 옛 빌드)은 쓰지 않음.

```text
기록               /var/lib/tossos/config/capability-verify.jsonl
계좌               ****

단계
  read-fixtures          pass (통과)
  sellable-baseline      pass (통과)
  idempotency            pass (통과)
  idempotency-ttl-edge   skipped (생략)             요청하지 않았다 — 실행하려면 --include-ttl-edge를 붙여라. 경고: 이 …
  order-cancel           pass (통과)
  order-amend            pass (통과)
  sell-boundary          fail (실패)                official: API error 422: {"error":{"requestId":"<가림>","code":"market-not-suppo…
  conditional-register   pass (통과)
  sellable-reserved      pass (통과)
  conditional-persist    pass (통과)
  conditional-trigger    deferred (보류)            별도 세션 — 시장 조건 필요: 체결될 의도의 주문 없이는 발동을 임…
  conditional-modify     fail (실패)                official: API error 404: {"error":{"requestId":"<가림>","code":"conditional-orde…
  conditional-cancel     fail (실패)                official: API error 404: {"error":{"requestId":"<가림>","code":"conditional-orde…
  costs                  pass (통과)

대사로 종결된 객체 1건 (공식 목록·상세 읽기 근거 — 성공한 요청·증명이 아니다):
  conditional-order DJwYn8P_dD9lQMEeYc-5l5_yDTzu2fo55FEJkhU8WVg (333430) reconciled expired
```

판정(tasks 6.3 의 넷):

| 항목 | 결과 |
|---|---|
| outstanding 0 | 충족 — 미해결(정리 대상) 목록이 출력에 없음 |
| reconciled expired 1 | 충족 — 「대사로 종결된 객체 1건 … reconciled expired」 |
| 기록 줄 수 +1 | 충족 — 기록 28줄(`wc -l`), 마지막 28번째 줄이 6.2 의 대사 줄이고 27번째 줄은 그보다 앞선 `step order-cancel`(finished `2026-10-19T05:16:29Z`). 실행 전 줄 수의 직접 기준선은 없어 27→28 은 이 순서에서 나온 판단 |
| 이전 줄 바이트 불변 | **직접 증명 불가** — 실행 전 줄 해시 기준선이 없음. 간접 근거: 파일 mtime `2026-10-23 09:24:48.218 +0900` 이 28번째 줄 `finished_at` 과 같음(그 뒤 쓰기 없음) · 기록은 append 계약(시험 2.5 「기존 바이트 불변」, `analysis/red-green/`) |

단계 판정 메모: sell-boundary · conditional-modify · conditional-cancel = fail, conditional-trigger = deferred 는 a164 범위 밖(conditional-trigger 는 a100 몫).
a164 의 수락 조건은 대상 조건주문 하나가 「reconciled expired」 로 종결되어 HOLD·정리 대상에서 빠지는 것이고, 위 출력이 그것을 보임.
