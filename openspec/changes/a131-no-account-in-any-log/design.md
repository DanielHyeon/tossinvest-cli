# a131 design — 계좌 정보는 로그에 남지 않는다

기준: base 는 `base-commit.txt`. 줄 번호는 base 의 것이다. 운영 파일 수치는 2026-10-12~13 실측(값은 읽지 않고 개수·모양만 셈).
분기 주장은 이 change 의 AST 번들(`analysis/function-logic/*/ast.json`)을 근거로 한다.

## D0. 무엇이 「계좌 정보」인가

| 종류 | 예(모양만) | 판정 |
|---|---|---|
| 계좌 번호 원문 | `99999999999`, `999-99-999999` | 계좌 정보 |
| 마스크 | `*******9999`(`attest.Mask`) | 계좌 정보 — 끝 4자리가 남는다 |
| 소금 없는 저엔트로피 유도 해시 | `pos-`·`rec-`·`mode-` + 24 hex(`journal.PositionID` · `reconcileStateID` · `operatingModeID`) | 계좌 정보 — 나머지 입력(시장·종목·작은 순번·초 단위 시각)이 같은 줄에 있거나 추측 가능, 11자리는 대입 가능 |
| 고엔트로피 유도 해시 | `obs_`/`eld_`/`exit_` + 64 hex, `strategy-attempt:v1:sha256:` | 계좌 정보 아님 — 입력에 나노초 시각·관측가·결정 신원이 들어가 대입 불가. intent/attempt id 는 원장 대조에 필요해 남김 |
| `accountSeq` | 사용자 계좌 목록 안의 작은 순번(1, 2 …) | 계좌 정보 아님 — 계좌를 식별하지 않는 서수. 치환하지 않는다(Manager 가 뒤집을 수 있음) |
| 알림 본문 | `notificationFor` 가 `Fields` 를 본문에 펼침(`account: …`) | **범위 밖 발견** — 로그가 아니라 ntfy 로 나가는 외부 전송이다. 후속 change 권고(완료 보고) |

## D1. 유출 자리 전수 (역할 기준)

census 는 `analysis/harness/account_census.go`(go/parser, 비시험 Go 전체)를 base 에서 돌린 `analysis/census/census-base.tsv` 다 — 계좌를 뜻하는
식별자·필드(`AccountRef` · `accountRef` · `account` · `AccountNo` · `AccountSeq` · 마스크 변수 · `obs.FieldAccount`)가 **싱크 호출의 인자**로 들어가는 자리를 센다.
합계: obs 싱크 22 · 오류 문자열 구성 76 · 직접 출력 12 · 계좌 json 태그 45. 여기에 `git grep` 보강(`obs.FieldAccount` 참조 전수 20, `"position_id"` 키 35)을 더했다.
존재 검사가 아니라 역할 검사다: 오류 문자열에 감싼 계좌(`fmt.Errorf(… %s …, accountRef)`)도 「어느 싱크로든 흐를 수 있는 자리」로 센다.

| # | 싱크(파일) | 자리 | 필드/경로 | 수 | 현재 유출(실측) | a131 처리 |
|---|---|---|---|---|---|---|
| L1 | engine.log JSON | `obs.FieldAccount` 를 싣는 로그 인자 — `adoption.go` 6 · `exitloop.go:895`·`:1898`(`logErr`) · `exitwiring.go:112`·`:150`(알림 Fields)·`:424`·`:429` · `reconcileloop.go:538`·`:543`·`:565` · `runtime.go:371`·`:451`·`:523`(모든 줄 앞) · `interlock.go:441`(마스크) · `obs/record_only.go:184` | `account` | 20 | **예** — 80,195 줄 | R1 핸들러가 필드를 버림 |
| L2 | engine.log JSON | 알림 선점 줄 `obs/notifier.go:716`·`:748`·`:765` | `claimed_by` 값 안의 계좌 | 3 | **예** — 13 줄 | R2 값 치환(등록 원문) |
| L3 | engine.log JSON | 오류 원문 — `obs.Logger.Error` 의 `error`, `detail`, 알림 `Body` 가 로그로 갈 때. 계좌를 문자열에 넣는 구성 76(census `errtext`: `journal` 60 · `risk` 4 · `reconcile` 4 · `filldetect` 2 · `app/engine` 2 · `soak`·`reversallane`·`execgw`·`attest` 각 1) | `error`·`detail`·메시지 | 76 | **예** — 1 줄(`exit.proposal_refused`) | R2 값 치환 — 감싼 오류도 문자열이므로 같은 규칙 |
| L4 | engine.log JSON | `position_id` 키 35 + 오류 문구 안 `pos-…` | `position_id` | 35 | **예** — 3,405 줄 | R2 `pos-[redacted]` |
| L5 | engine.log 비JSON | `runEngineRun` `engine.go:368` `account %s` | stdout | 1 | **예** — 69 줄 | R3 쓰기 대상 감쌈 |
| L6 | engine.log 비JSON | `main` `main.go:14` 마지막 오류 줄(엔진 거절 문구 — 인터록 `ErrAccountMismatch` 는 마스크 둘) | stderr | 1 | 잠재 | R3 |
| L7 | capability-soak.jsonl | `soak.Recorder.Append` ← `Cycle.AccountRef`(`record.go:142`) | `account_ref`(원문) | 1 | **예** — 5,082 줄 | R4 봉인 |
| L8 | capability-verify(-us).jsonl | `verifylive.Recorder.Append` ← `Entry.AccountRef`(`record.go:274`) · 대사 줄 `encodeReconcileLine` → `lockedRecord.appendLine` | `account_ref`(마스크) | 2 | **예** — 25 줄(+ 잔재 US 34) | R4 봉인 |
| L9 | audit.log | 게이트 수락 `interlock.go:418` `account=%s …` → `audit.Log.Record` | `detail` | 1 | **예** — 78 줄 | R5 Record 가 치환 |
| L10 | attest-autorun.log · journald | `runSoakAttest` `soak.go:549` `account %s`(마스크) | stdout | 1 | **예** — 147 줄 | 줄 삭제 |
| L11 | attest-autorun.log · journald | `soak attest` 실패 문구(`evaluateIssues` 계좌 모호 · `acceptSupervised` 계좌 다름 — 마스크) | stderr(L6 경로) | 2 | 잠재 | R3(main) + R2 마스크 꼴 |
| L12 | soak.log(soak run 의 출력) | 계좌 출력 자리 0(census `print` 에 `runSoakRun` 없음) | — | 0 | 아니오 — 0 줄 | L6 경로(main)만 |

**범위 밖(로그가 아님).** `capability-attestation.json`(인터록이 대조하는 상태 파일) · `journal.db`(정본 원장) · 설정 · 갱신 상태 파일(`RenewalStatus` 는
사유 코드만) · 대화형 명령의 터미널 출력(`account list` · `verify report` · `soak status` · flatten 확인 문구 — 사람이 요청한 표시이며 파일로 남지 않음) ·
census `record` 의 나머지 43 태그(원장·API·서명 다이제스트 입력). 알림 본문은 D0 의 범위 밖 발견.

## D2. 규칙 하나 — `internal/redact`

모든 싱크는 이 패키지의 함수만 부른다. 규칙을 자리마다 베끼지 않는다.

- `Register(refs ...string)` — 프로세스 전역 집합(뮤텍스). 원문(공백 제거)과, 숫자가 8자리 이상이면 숫자열 정규식(자리 사이 `-`·공백 허용, 앞뒤 숫자 아님)을 넣는다.
  짧은 참조(시험의 `123-45`)는 원문 그대로만 — 숫자 정규식은 가격·수량을 지울 수 있어 8자리 미만엔 쓰지 않는다.
- `String(s)` — 순서대로: 등록 원문(긴 것부터) → 등록 숫자열 → 마스크 꼴 `\*{3,}[0-9]{1,8}` → `\b(pos|rec|mode)-[0-9a-f]{24}\b` 를 각각
  `[account]` · `[account]` · `[account]` · `$1-[redacted]` 로 바꾼다.
- `IsAccountKey(k)` — 소문자·밑줄 제거 뒤 `account` · `accountref` · `accountrefs` · `accountno` · `accountnumber` · `accountseq` · `maskedaccount`.
- `Handler(h slog.Handler)` — R1·R2: 계좌 키 속성을 버리고(그룹 안까지) 문자열·오류·Stringer 값과 메시지를 `String` 으로 바꾼다. 그 밖의 `Any` 값은 JSON 으로
  만들어 치환이 생기면 치환된 문자열로 바꾼다(안 생기면 원래 값). 미리 묶은 속성(`With`)은 **Handle 시점에** 치환한다 — `With` 가 등록보다 먼저 불려도 새지 않음.
- `Writer(w io.Writer)` — R3: 쓰기마다 `String` 을 적용한다(줄 단위 쓰기 전제 — fmt · slog 모두 한 줄 한 번 쓰기).
- `SealRecordLine(line, dir)` · `OpenAccount(sealed, dir)` — R4(D4).

**등록 지점 하나(R0).** 계좌 원문이 프로세스에 들어오는 입구는 `official.Client.accountsLocked`(공개 `Accounts` 와 지연 계좌 범위 결정이 같이 쓰는 한 경로) 하나다
(엔진 `resolveAccountRef` · soak `probeAccounts` · verify 가 모두 이것을 부름). 받은 모든 `accountNo` 를 등록한다. WTS 클라이언트는 엔진·soak·verify 경로에 없다
(엔진은 official-only — engine-safety). 등록 전 줄은 계좌를 모르므로 계좌를 실을 수 없다 — 단 계좌 조회 **실패** 문구(브로커 응답 본문)는 등록 전이라 원문 치환을
못 받는다; 숫자열 규칙은 등록 값에서 오므로 이것도 못 받는다(잔재 ①).

## D3. 싱크 배선

| 싱크 | 편집 | 동작 변화 |
|---|---|---|
| R1·R2 obs 로거 | `obs.NewLogger`: 내부 핸들러를 `redact.Handler` 로 감싼 **뒤** `Base` 를 붙임 | 줄의 키 집합에서 `account` 가 빠지고 문자열 값이 치환됨. 레벨·이벤트·시각 불변 |
| R3 엔진 출력 | `runEngineRun`: `out, errOut` 을 `redact.Writer` 로 감쌈(로거도 `errOut` 을 받으므로 이중 적용 — 무해) | 비JSON 줄의 계좌 → `[account]` |
| R3 마지막 오류 | `main`: `fmt.Fprintln(redact.Writer(os.Stderr), err)` | 모든 명령의 마지막 오류 줄 — 대화형 사용자에게도 계좌가 `[account]` 로 보임(수용: 사용자 결정) |
| R5 audit | `audit.Log.Record`: `Setting` · `Old` · `New` · `Detail` 을 `redact.Redact` 로 치환(같은 함수를 정리 명령이 씀 — `audit.RedactEntry`) | 행위·설정 이름·이전/이후 값·시각·주체 보존. `account=…` 는 `account=[account]` |
| L10 | `runSoakAttest`: `account` 줄 삭제 | 출력 한 줄 감소 |

**audit 의미 보존(D3-a).** 감사 기록의 질문은 「누가(Subject) 언제(At) 무엇을(Action·Setting) 무엇에서 무엇으로(Old→New) 바꿨나」이다. 실측 audit.log 603 줄에서
계좌는 `detail` 에만 있고(78 줄, 전부 `automation_gate.accepted`) `Old`/`New`/`Setting`/`Subject` 에는 0 이다. 한 프로필은 한 계좌만 다루므로(엔진은 첫 계좌만 씀 —
`resolveAccountRef`) 감사 줄의 계좌는 추적성에 더하는 것이 없다. `RecordChange` 의 중복 판정(`previous.New == newValue`)은 `New` 에 계좌가 없으므로 불변 — 시험이 고정.

## D4. soak · 검증 기록 — 봉인(로컬 키), 읽는 쪽은 종전과 같은 값

**읽는 쪽이 있다(과제 4).** 두 기록은 로그이면서 인터록 증거의 입력이다.

- `soak attest`(`BuildAttestation`)는 soak 요약의 `AccountRef`(기록 `account_ref` 원문에서 옴)를 **attestation 의 계좌로 그대로 쓴다**(`soak/attest.go:263`).
  엔진 인터록은 그 값을 해석된 계좌와 대조한다(`attest.Verify` — `interlock.go:518`).
- 계좌 모호 판정(`summary.go:137` `distinctAccounts`)과 계좌 없음 판정(`evaluateIssues`)이 기록의 계좌를 쓴다.
- 감독 증거(`acceptSupervised`)는 검증 기록의 마스크를 soak 원문의 마스크와 비교한다(`attest.SameAccountMasked`).
- 검증 대사(`bindReconcileAccount`)는 대상 줄들의 마스크가 하나이고 현재 자격의 마스크와 같은지 본다.
- 서명·해시: 기록 줄 자체를 서명하는 곳은 없다. 대사의 `sha256(raw)` 지문(`reconcile.go:189`·`:287`)은 **한 번의 대사 안에서** 읽기와 다시 읽기를 대조하는
  동시성 검사이고 저장되지 않는다 — 정리 명령이 대사와 겹치지만 않으면 무관(정리 명령은 검증 실행 잠금이 신선하면 거절).

그래서 필드를 **지우면** attestation 갱신이 깨진다(계좌 없음 → `ReasonAccountMissing`, 감독 증거의 계좌 다름 → 오류). 선택지:

| 안 | 기록에 남는 것 | attestation 계좌의 출처 | 읽는 쪽 변경 | 판정 |
|---|---|---|---|---|
| 지우기 | 없음 | 없음 | 결속 소멸 — 인터록 증거 약화 | 기각(안전 방향 위반) |
| 소금 없는 해시 | `sha256(계좌)` | 없음 | — | 기각 — 11자리 대입 가능 |
| 키드 HMAC | `HMAC(k, 계좌)` | attest 가 브로커에서 다시 읽거나 attestation 도 태그로 | attest 에 브로커 호출 추가(§0.4 예산) 또는 인터록 `Verify` 변경 | 기각 — 인터록 변경·호출 추가 |
| **로컬 키 봉인** | `account_ref_sealed` = AES-256-GCM(k, 원래 값) | 종전과 같음(열린 값) | 읽을 때 연다 — 결속 판정 코드 무변경 | **채택** |

**채택 근거.** 키 파일(`<기록 디렉터리>/account-seal.key`, 32바이트 hex, 0600)은 계좌 원문을 이미 가진 attestation(0600)·원장과 **같은 디렉터리**에 있다 —
키를 가진 사람은 이미 계좌를 가진다. 그래서 「되돌릴 수 있음」이 노출을 더하지 않고, 로그만 새면(복사·공유) 계좌는 나가지 않는다. 봉인은 무작위 nonce 라
같은 계좌도 줄마다 다른 값 — 줄 사이 상관도 남지 않는다. 결속이 꼭 필요함은 위 다섯 읽는 자리가 증명한다(YAGNI 시험 통과).

- 쓰기: 구조체를 종전대로 직렬화한 줄을 `redact.SealRecordLine(line, dir)` 에 넣는다 — `account_ref` 를 떼어 봉인해 `account_ref_sealed` 로 두고, 나머지 모든 문자열 값을
  `String` 으로 치환한다(줄 안 오류 원문 등). 숫자는 `json.Number` 로 원문 보존, HTML 이스케이프 끔. 키가 없으면 만든다(임시 파일 + `link` — 경합 시 먼저 만든 쪽 키를 읽음).
- 읽기: 해독 뒤 `AccountRefSealed` 가 있으면 `redact.OpenAccount` 로 열어 `AccountRef` 를 채운다. 키가 없거나 열기 실패면 빈 값 — 읽는 쪽 판정이 「계좌 없음」으로
  fail-closed(attestation 발급 거절 · 감독 증거 거절 · 대사 거절). 평문 `account_ref` 만 있는 옛 줄은 그대로 읽는다(전환기 호환).
- 자리: `soak.Recorder.Append` · `verifylive.Recorder.Append` · `lockedRecord.appendLine`(대사 줄 — 쓴 뒤 다시 읽어 대조하는 바이트는 봉인된 줄) /
  `soak.LoadCycles` · `verifylive.LoadEntries` · `readRecordStrictNoTail`(대사 엄격 해독). `reconcile.go:284` 의 `decodeRecordStrict(again)` 는 해독 가능성만 보므로 무편집.
- 구조체 필드 추가: `Cycle.AccountRefSealed` · `Entry.AccountRefSealed`(`json:"account_ref_sealed,omitempty"`). 쓰기 때는 비어 있고 줄 변환이 채운다.

## D5. 일회성 정리 명령 `tossctl scrub-account-records`

- 대상(기본): 프로필 디렉터리의 `capability-soak.jsonl` · `capability-verify.jsonl` · `capability-verify-us.jsonl`(있으면), 데이터 디렉터리의 `audit.log`
  (엔진과 같은 해석 — `journal.DataDir()`). `--file` 로 덧붙이거나 바꿀 수 있다(시험·잔재 디렉터리용). 없는 파일은 건너뛴다.
- 같은 함수: 기록 줄은 `redact.SealRecordLine`(쓰기 경로와 같은 함수), audit 줄은 `audit.RedactEntry`(= `Record` 가 부르는 함수). 규칙 사본 없음.
- audit 의 원문 치환에는 등록 값이 필요하다 — 명령은 먼저 soak·검증 기록의 평문 `account_ref` 와 attestation 의 `account_ref` 를 읽어 `redact.Register` 한다
  (읽기만). 마스크 꼴은 등록 없이 치환된다.
- 미리보기(기본): 파일별 「줄 수 · 바뀔 줄 수」만 출력. 쓰기 0(키 파일도 만들지 않음 — 미리보기는 봉인하지 않고 변환 필요 여부만 셈).
- `--apply`(사람): 같은 디렉터리 임시 파일(0600)에 쓰고 fsync → 원본의 크기·수정 시각이 읽은 때와 같은지 다시 보고 다르면 중단(원본 무변경) →
  `rename` 으로 원자 교체 → 디렉터리 fsync. 백업 없음(사용자 결정). 손상(해독 불가) 줄은 그대로 두고 수를 보고(마지막 줄 찢김 등).
- 거절: 검증 실행 잠금이 신선하거나 이 기록에 붙은 soak 프로세스가 보이면 `--apply` 거절 — `soak run` 은 기록 fd 를 쥐고 있어 교체 뒤의 쓰기가 사라진 inode 로 간다.
- 주석 `mutating: true`, `source: local` — 대화형 에이전트는 자동 실행하지 않는다. 시험은 임시 디렉터리에서만 돈다.

## D6. 동작 불변 · 손절 즉시성

- 바뀌는 것은 로그·기록의 **바이트**뿐이다. 주문·청산·대사·체결 감지 경로의 판정 입력은 로그를 읽지 않는다(`readLogTail` 은 기동 실패 꼬리 표시만).
- 핸들러 치환은 로그 쓰기 안에서 일어나며 비용은 줄당 정규식 몇 번이다(엔진은 5초 주기 · 줄 수십). a129 D6 과 같이 로그 쓰기는 주기 끝이거나 판정 밖이다.
- 기록 봉인 실패(키 파일을 못 만듦)는 그 기록 쓰기의 오류다 — soak·검증 측정 도구의 기록 실패이며 엔진 경로가 아니다. audit `Record` 의 치환은 실패하지 않는다(순수 함수).
- 등록은 `accountsLocked` 안의 메모리 append 하나다 — 브로커 호출·오류·반환값 불변.

## D7. 시험 전략

1. 센티널 `987-65-SENTINEL` · 숫자형 `98765432101` · 그 마스크 · `pos-` id 를 계좌 참조 · 오류 원문(감싼 오류 포함) · 브로커 응답(httptest `/api/v1/accounts`) ·
   알림 Fields · `With` 속성에 심고: obs JSON 줄 · 엔진 Writer · audit 파일 · soak/검증 기록 파일 · 정리 명령 결과 어디에도 없음을 단언.
2. 왕복: 봉인한 soak 기록으로 `Summarize` → `BuildAttestation` 의 계좌가 원문과 같음 · 감독 증거의 마스크 비교 통과 · 대사 계좌 결속 통과 · 키 없음이면 fail-closed.
3. 호환: 평문 옛 줄과 봉인 줄이 섞인 기록이 같은 결과.
4. 구조(go/parser): ① 비시험 코드에서 `slog.New*Handler`·`slog.New(` 는 `internal/obs/log.go` 와 `internal/redact` 밖에 없음 ② `O_APPEND` 로 여는 비시험 자리는
   허용 목록(soak · verify · audit · engineproc · soakproc)뿐이고, 앞 셋의 같은 파일이 `redact.` 를 부름 ③ `NewLogger` 가 `redact.Handler` 를 부름
   ④ `accountsLocked` 가 `redact.Register` 를 부름 ⑤ `runEngineRun` 의 `out, errOut` 이 `redact.Writer` 로 감싸짐.
5. 정리 명령: 미리보기는 파일 바이트·mtime 불변 · 키 파일 없음 · 바뀔 줄 수 정확 / apply 는 원자 교체·권한 0600·센티널 0·attestation 왕복 성립 · 도중 변경이면 중단.
6. 기존 시험 전부 green(`internal/obs` · `internal/audit` · `internal/soak` · `internal/verifylive` · `internal/official` · `cmd/tossctl` · `internal/app/engine`).

변이(사본·무변이 대조군 GREEN 먼저): 핸들러 우회(NewLogger 가 감싸지 않음) · 계좌 키 버림 삭제 · 마스크 규칙 삭제 · 등록 삭제 · Writer 감쌈 삭제 · audit 치환 삭제 ·
봉인 대신 평문 · 열기 삭제 · `pos-` 규칙 삭제 · `With` 속성 치환을 WithAttrs 시점으로(등록 전 묶음 유출) · 미리보기가 씀.

## 배포 (범위 밖 — 사람 절차)

1. **읽는 쪽과 쓰는 쪽을 함께 교체.** 새 쓰기(봉인)를 옛 읽기가 만나면 계좌 없음으로 attestation 갱신이 거절된다(fail-closed, 갱신 창 30일 중 하루치).
   호스트 `~/.local/bin/tossctl`(tossos-attest.timer)과 컨테이너 이미지(`make image CHANGE=a131-…`)를 같은 KST 05:00~09:00 창에 바꾼다.
2. 교체 뒤 engine.log 새 줄에 `account` 키 0 · `pos-` 0 을 사람이 확인(`grep -c`).
3. 정리 — 사람이 `tossctl --config-dir ~/.config/tossctl scrub-account-records`(미리보기) → soak 정지 → `… --apply` → soak 재시작.
   다음 `tossos-attest.timer` 실행이 발급(또는 종전과 같은 사유로 거절)하는지 확인.
4. 과거 engine.log 는 Manager 가 사람 승인으로 따로 처리.

롤백: 이미지·호스트 바이너리 핀을 되돌리면 옛 읽기가 봉인 줄의 계좌를 못 본다 — 롤백 시 soak 기록을 새로 시작해야 attestation 이 다시 발급된다(정리 뒤라면 같음).
원장·설정·attestation 형식 변경 없음.

## 잔재(명명)

① 계좌 조회 실패 문구에 브로커가 계좌를 되돌려 주면 등록 전이라 치환되지 않는다. ② 런타임 패닉의 스택(fd 2 직접)은 Writer 를 지나지 않는다.
③ 알림 본문(ntfy)은 범위 밖. ④ 대화형 터미널 출력은 범위 밖. ⑤ `accountSeq` 는 치환하지 않음(D0). ⑥ 콘솔 프로세스는 계좌를 등록하지 않는다 —
콘솔이 쓰는 audit 줄에 계좌가 0 임은 실측(D3-a)이고, 그 판정이 바뀌면 다시 볼 것.
