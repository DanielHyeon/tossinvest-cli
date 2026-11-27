# a063 D6 — gstack review r1 (2026-11-28)

- 대상: `8b5ca6edd`(feat/a063-renewal-scope) vs origin/main `baeddfe31`. 프로덕션 Go 변경 19줄 → small diff, specialist 생략.
- 보이스: gstack `review` 핵심 패스(작성자 Opus) + 별도 컨텍스트 적대 패스 1(general-purpose 서브에이전트, 읽기 전용).
  같은 모델 계열이다 — 다른 계열 보이스(codex)는 쓰지 않았다(경량 리뷰, Low risk).
- gate scan: `GATE_SUMMARY listed=4 eligible=0` — 완화된 시험·게이트 없음. banned terms: 목록 없음.

| # | 발견(적대 패스) | 등급 | 처분 |
|---|---|---|---|
| F1 | 옛 콘솔이 새 파일(issued+코드)을 `unknown` 으로 본다, format_version 1 이라 이유를 모른다 | Medium, 비차단 | **거절(문서화)** — design D6.5 가 의도한 방향(성공으로 안 읽음). 버전을 올려도 옛 콘솔은 v2 를 똑같이 `unknown` 으로 본다. attestation 줄의 「미검증 endpoint」는 따로 보인다. 배포 순서 메모로 남김. |
| F2 | 콘솔 문장이 게이트 OFF 에서 거짓(인터록이 `!gate.Enabled` 면 attestation 을 안 읽음, `internal/app/engine/interlock.go` `runInterlock`) | Low | **수용·수정** — 「자동매매 게이트가 켜져 있으면」을 붙임(`renewalReasonMessage` 문자열만). |
| F3 | `refused`/`failed` 에 새 코드가 붙어도 Validate 통과(작성자는 만들지 않음, 0600 손편집만 도달) | Low | **거절** — 새 가드 금지 원칙. 도달 경로가 소유자 손편집뿐이고 표시는 고정 문장이다. |
| F4 | 판정 집합이 `soak.LiveOnlyEndpoints()` 이고 엔진은 `engine.RequiredEndpoints()` | Low | **거절(범위 밖)** — 발급 안내(B20)와 같은 집합을 일부러 썼다(D6.1). 두 목록의 결속은 기존 구조이며 이 change 가 바꾸지 않는다. |
| F5 | 계좌 사유 코드 등이 일반 문장으로 떨어진다(기존) | Low | **거절(기존 부채)** — 이 change 이전부터. |

- 적대 패스가 확인한 무결함: 종료 코드 불변, attestation 바이트 불변(`attest.Save`·`BuildAttestation` 무편집), 상태 파일의 다른 소비자 없음, 새 콘솔이 옛 파일을 그대로 받음.
- 수정 뒤: `go test -count=1 ./internal/console -run 'A063|Renewal|Attestation'` ok, 번들(`renewalReasonMessage`·`Console.readAttestation`) 해시 재추출.

## r2 (같은 날)

- r1 수정은 문자열 한 줄이다. r2 는 수정 확인만 했다(편집 0): 선택 시험 재실행 초록, 새 발견 없음. 별도 보이스는 다시 돌리지 않았다 — 상한 2회 안에서 수렴.
