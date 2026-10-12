# a131 — 시스템이 쓰는 어떤 로그·기록 파일에도 계좌 정보가 남지 않는다 (no-account-in-any-log)

## Why

사용자 결정(2026-10-12): **계좌 정보는 로그에 남기지 않는다.** 안전 불변식 8(시크릿·세션·계좌 개인정보를 로그에 저장 금지)의 집행이다.

a129 리뷰와 Manager 실측(2026-10-12, 운영 파일을 값 없이 개수만 셈)이 찾은 현재 유출:

| 파일 | 줄 수 | 계좌가 실린 줄 | 실린 자리 |
|---|---|---|---|
| `~/.config/tossctl/engine.log` | 85,394 | 원문 80,127 줄 · 마스크 149 줄 | `account` 필드 80,195 · `claimed_by` 13 · `error` 1 · 비JSON `account …` 줄 69 |
| `~/.config/tossctl/capability-soak.jsonl` | 5,139 | 5,082 | `account_ref`(원문) |
| `~/.config/tossctl/capability-verify.jsonl` | 25 | 25 | `account_ref`(마스크 `*******1234` 꼴) |
| `~/.local/share/tossos/audit.log` | 603 | 78 | `detail` 의 `account=<원문>`(게이트 수락 줄) |
| `~/.local/share/tossos/attest-autorun.log` | 5,383 | 147 | `soak attest` 출력의 `account <마스크>` 줄 |

그리고 `position_id`(`journal.PositionID`)는 계좌·시장·종목·작은 순번을 소금 없이 sha256 한 앞 24 hex 이고, 같은 줄에 시장·종목이 있으므로
11자리 계좌는 오프라인 대입으로 되돌릴 수 있다 — 계좌 유도 식별자도 계좌 정보로 본다(engine.log 3,405 줄).

## What Changes

1. **규칙 하나(`internal/redact`)** — 계좌 필드 이름 판정, 문자열 안의 계좌(등록된 원문·숫자열 꼴·마스크 꼴)와 계좌 유도 저엔트로피
   식별자(`pos-`·`rec-`·`mode-` + 24 hex) 치환, 기록 줄 봉인(로컬 키 AES-256-GCM). 모든 싱크가 이 패키지의 함수만 부른다.
2. **등록 지점 하나** — 계좌 원문을 아는 유일한 입구 `official.Client.accountsLocked` 가 받은 계좌 번호를 프로세스 규칙에 등록한다.
3. **싱크 배선**:
   - `obs.NewLogger` 의 slog 핸들러를 규칙으로 감싼다(engine.log 의 모든 구조화 줄 · flatten 로거) — 계좌 필드는 빠지고 모든 문자열 값·오류 원문·메시지가 치환된다.
   - `tossctl engine run` 의 stdout/stderr 쓰기와 `main` 의 마지막 오류 줄이 같은 규칙을 지난다(engine.log 비JSON 줄).
   - `audit.Log.Record` 가 같은 규칙으로 문자열 칸을 치환한다 — 행위·설정·이전/이후 값·시각·주체는 그대로.
   - soak · 검증 기록 쓰기(`soak.Recorder.Append` · `verifylive.Recorder.Append` · 대사 줄 `lockedRecord.appendLine`)는 줄을
     `redact.SealRecordLine` 에 통과시켜 `account_ref` 를 `account_ref_sealed`(로컬 키 봉인)로 바꾸고 나머지 문자열을 치환한다.
     읽는 쪽(`soak.LoadCycles` · `verifylive.LoadEntries` · 대사 엄격 해독)은 봉인을 열어 메모리의 `AccountRef` 를 종전과 같게 채운다 —
     attestation 의 계좌 결속·마스크 비교·대사 계좌 결속은 바뀌지 않는다.
   - `soak attest` 출력의 계좌 줄을 지운다(attest-autorun.log · journald).
4. **일회성 정리 명령** `tossctl scrub-account-records` — 기존 soak · 검증(KR·US) · audit 파일을 **같은 함수**로 다시 쓴다.
   기본은 미리보기(파일별 바뀔 줄 수만 출력, 쓰기 0), `--apply` 는 사람 실행(원자 교체, 백업 없음 — 사용자 결정). `mutating: true`.
5. 회귀 방지 — 계좌 센티널을 계좌 참조·오류·브로커 응답에 심어 모든 싱크 출력에 없음을 단언하는 시험 + 새 append 쓰기·새 로거 구성이
   규칙을 우회하면 실패하는 구조 시험(go/parser).

## What Does NOT Change

- 주문·청산·대사·체결 감지·Guardian·인터록의 동작과 순서. 브로커 호출 수(0 추가). 손절 즉시성.
- attestation 파일(`capability-attestation.json`, 상태 파일 — 로그가 아님)의 형식과 엔진의 계좌 대조.
- 원장(journal.db)·설정 파일 — 계좌는 시스템의 정본 상태로 남는다(로그가 아님).
- 과거 engine.log 삭제 — Manager 가 사람 승인으로 따로 처리(범위 밖).

## Impact

- Affected specs: `engine-safety`(요구 1개 추가).
- Affected code: 새 `internal/redact/`, `internal/obs/log.go`(`NewLogger`), `internal/official/reads.go`(`accountsLocked`),
  `cmd/tossctl/engine.go`(`runEngineRun`), `cmd/tossctl/main.go`(`main`), `internal/audit/audit.go`(`Record`),
  `internal/soak/record.go`(`Append` · `LoadCycles` · `Cycle` 필드 하나), `internal/verifylive/record.go`(`Append` · `appendLine` · `LoadEntries` ·
  `Entry` 필드 하나), `internal/verifylive/reconcile_record.go`(`readRecordStrictNoTail`), `cmd/tossctl/soak.go`(`runSoakAttest`),
  새 `cmd/tossctl/scrub_account_records.go`.
- High-risk: 인터록 증거(soak·검증 기록)의 계좌 결속 경로와 엔진 기동 로깅을 만진다 — 결속의 **값**은 불변(봉인/열기 왕복).
  Pre-Edit 선언 + 편집하는 기존 함수마다 Function Logic Map.
- 배포 순서(design 「배포」): 읽는 쪽(호스트 attest 바이너리)과 쓰는 쪽(컨테이너 soak·엔진)을 함께 교체한 **뒤에** 정리 명령을 사람이 돌린다.
