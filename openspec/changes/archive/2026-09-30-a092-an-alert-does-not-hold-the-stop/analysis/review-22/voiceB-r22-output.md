<!-- 22라운드 보이스 B(산출물 무기화 렌즈) — Claude code-inspector-tester 서브에이전트(독립 컨텍스트, 같은 하네스). 프롬프트 prompt-common.md + 호출 렌즈.
     대상 7cf80832(HEAD 09465697 와 a092·인용 코드 diff 0). 원문 그대로. -->

**판정: BLOCK**

(voice B · 산출물 공격 관점 · 대상 `7cf80832`. HEAD `09465697`과 `git diff 7cf80832 HEAD -- internal cmd openspec/specs <a092>` 결과는 비어 있음)

재현한 것: `analysis/head-ast-21/extract.py`를 `/tmp/claude-1000/a092-r22/voiceB-copy`(archive 7cf80832)에서 `git rev-parse` 줄만 바꾼 사본으로 돌림 (GOCACHE=/tmp/claude-1000/a092-gocache). 결과는 JSON 39개가 커밋본과 **바이트 동일**하고, MANIFEST 본문 39줄도 동일함. 13개 파일의 sha는 c1e34dc4와 7cf80832에서 같음.

## 발견 표

| # | 등급 | 주장 | 증거(파일:줄) | 권고 |
|---|---|---|---|---|
| 1 | **P0** | 모드 전이 통지 key에 에피소드 신원이 없음. 설계 D0.3g 1이 key를 「같아야 한다」고 고정해서 a092 ADDED 「완화도 운영자에게 통지되어야 한다(SHALL)」와 정본 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만 전송한다(SHALL NOT)」가 **같은 key 위에서 충돌**함. a092가 완화 경로를 처음 만들기 때문에 「완화 → 1시간 안 재강화 → 재완화」가 생산에서 도달 가능해짐. 그러면 재강화 통지와 두 번째 완화 통지가 조용히 흡수됨(ClaimSettled → nil 반환). 게다가 `mode-release`는 「통지됨」으로 보고함 | key `"operating_mode:"+acct+":"+mode` `internal/obs/mode.go:62` · 설계 `design.md:621-622` · `claimOwed` `outbox.go:407-408` · `ClaimSettled` `notifier.go:285-293` → `notifyCritical` `:223` 불통과 → nil · `DefaultRemindAfter = time.Hour` `notifier.go:59` · 정본 `engine-safety/spec.md:708-710` · risk-management `:137` 「알림이 발송된다」 · 델타 `:306` | key에 전이 rowid(22판이 `OperatingModeRecord`에 이미 싣는 값)를 넣음. 통지는 `changed==true`일 때만 나가므로 폭주하지 않음. a094 D−5.2와 같은 형태 |
| 2 | P1 | Scenario 「승인 뒤 늦게 끝난 동기 전송 실패는 다시 잠그지 않는다」가 무조건임. 그런데 설계는 해제 세대를 근거가 **확정된 뒤**(결과가 돌아온 순간) 읽음. 그래서 「확정~읽기」 사이의 해제는 다시 잠김. a124 정본은 이 창을 명시적으로 허용하고 Scenario에 한정어를 달았는데, 델타는 한정어를 뺌 → 지킬 수 없는 THEN | 델타 `:64` · `:101-104` · `design.md:651-653` · 정본 `:1471` 「확정 순간과 해제 세대를 읽는 순간 사이의 해제를 「앞」으로 보는 것은 허용된다」 · `:1523-1524` · `:1534-1535` · `retry.go:557` 주석 「판정 근거를 얻은 직후 읽는 값」 | a124처럼 「해제가 세대 읽기 뒤일 때; 확정과 읽기 사이의 해제는 보수적으로 다시 잠글 수 있다」를 붙임 |
| 3 | P1 | Q4 핀의 양성 대조군(「오늘의 `parkAlert`는 통과」)이 핀 자신의 금지 형태 ③(「`entry == nil` 갈래」)과 모순됨. `Block`은 `if g.entry != nil` 안에만 있고 `EnqueueAlert`는 무조건 실행됨. 또 설계의 AST 인용 「B2 `:548` 안에서 `EnqueueAlert`(`:551`)」가 틀림 — B2는 `if err != nil { payload = nil }`(548-550)이고, 삽입은 함수 최상위(열 9)에 있음. 델타 SHALL(「잠금을 건너뛴 채 삽입에 닿는 경로」 금지)은 게이트 미배선일 때 HEAD가 위반함. 정본 `:840-844`의 「배선된 것까지」 면제가 델타에는 없음 | `replay.go:535-537` · `:548-551` · `head-ast-21/…gateway.parkalert.json`(B1 :535 · B2 :548 · call :551 col 9) · `design.md:658-666` · `tasks.md:31-32` · 델타 `:62` | ③에서 미배선 게이트를 명시적으로 면제하거나, `parkAlert`를 `entry==nil`이면 적재하지 않는 구조로 바꿈. 인용 정정 |
| 4 | P1 | 새 공개 원장 표면 `RecordAlert(ctx,a,remindAfter)`는 정착 행을 **재무장**함(PENDING을 만듦). 그런데 Q4 핀은 `EnqueueAlert` 호출자만 봄. `n.mu` 밖에서 `RecordAlert`를 부르는 호출자가 생기면 `Acknowledge`가 막고 있는 a096 경합(셈과 해제 사이 재무장)이 그대로 다시 열리는데, 핀이 이를 못 잡음 | `design.md:636` · `:646` · `:666` · `notifier.go:834-839`(n.mu가 막는 경합) · `outbox.go:294-345`(rearm) | 「`RecordAlert` 비시험 호출자 = `n.mu` 아래의 어댑터 하나」 구조 핀 추가(또는 obs 전용 seam) |
| 5 | P1 | 교차 change 충돌. (a) a094 D−5.3/D−4.6은 exit 관측 루프에서 critical을 **직접 `EnqueueAlert`**하고 먼저 잠그지 않음. 이는 a092 Q4 SHALL과 핀 ③에 정면으로 걸림. (b) a092 델타 `:38` 「exit goroutine의 critical 기록은 재무장 동작 그대로(SHALL)」는 a094의 remindAfter 0 기록이 문자 그대로 위반함. (c) a066의 **미커밋 작업 트리**(peer 세션, HEAD 아님)도 먼저 잠그지 않는 critical 직접 적재를 더함. D0.3g 4의 「오늘은 `parkAlert` 하나」는 HEAD에서는 참이지만 곧 거짓이 됨. 또 알림 성격의 critical(완화 통지 · 청소 연속 실패)에 「자기 진입 차단 사유」를 세우라는 SHALL은 과잉 일반화임 | a094 `design.md:55-64` · `specs/exit-policy/spec.md:39` · 작업 트리 `internal/app/engine/risk_relaxation_command.go:158-166`(untracked) · 델타 `:38` · `:62` · `design.md:666` | SHALL에 「알림기의 기록 전용 입구를 거치는 것」을 직접 기록자의 정식 경로로 명시함. `:38`은 「알림기 입구를 거친 기록」으로 한정. a094와 조정 |
| 6 | P1 (T) | 정본을 통째로 교체하는 MODIFIED 「배달 실행자의 정지가 …」가 거짓·낡은 서술을 a092 소유로 다시 게시함. `:227-229` · `:240` 「잃은 것 하나는 a092가 진다 — a092 Scenario(`:126-130`)는 모드 승격을 SHALL로」는 거짓임: 20판이 11-1을 적용했고, D0.3g 7은 잔여로 수용했으며, `:126-130`은 지금 「일반 등급도…」임. 같은 요구의 규범 `:183`(「함께 하지도 않는다」)과도 모순됨. 메모 `:155-156` 「바뀐 것은 근거 ① 한 곳」은 전수가 불완전함 | 델타 `:155-156` · `:183` · `:227-229` · `:240` · `tasks.md:92` | 해당 인용 블록을 정정하거나 「a098 기록」으로 표지 |
| 7 | P2 (T) | 같은 MODIFIED 안의 낡은 좌표. 이제 근거 ②가 **유일한** 근거인데, 그 인용 `risk-management/spec.md:102-108`은 「한도 완화 시도」 Scenario를 가리킴(모드 완화 승인은 `:133`). `retry.go:498-505`: `Block`은 지금 `:532-539`. `engine.go:377-398`: 루프 집합은 지금 `cmd/tossctl/engine.go:666-704` | 델타 `:166` · `:189` · `:204` · `:253` · `risk-management/spec.md:101-104` · `:133` | 정정 |
| 8 | P2 (T) | 설계 좌표 오류. `cmd/tossctl/engine.go:636`은 MANIFEST HEAD `bce793a7`과 `7cf80832`에서 `:639`임(c1e34dc4에서도 `:637`). 「인용한 운영 파일의 sha는 c1e34dc4 이후 바뀌지 않았다」도 거짓: `cmd/tossctl/engine.go`는 `bce793a7`(a066 5.5)에서 바뀌었고 MANIFEST에 없음 | `design.md:600-601` · `:610` · `git log c1e34dc4..7cf80832 -- cmd/tossctl/engine.go` = `bce793a7` | 정정. sha 주장 범위를 MANIFEST 13개 파일로 한정 |
| 9 | P2 (T) | MODIFIED 「등급화된 알림」 `:42` 「**이 요구**의 범위는 exit 관측 goroutine이다(SHALL)」는 archive 뒤 정본 요구 전체(`:32`의 전역 critical/outbox/heartbeat 규범)를 좁히는 문장으로 읽힘. 같은 요구의 `:64`(범위 밖 발송자 의무)와도 충돌함 | 델타 `:32` · `:42` · `:64` | 「위 원격 전송 비대기 문단들의 범위」로 한정 |
| 10 | P2 (T) | 델타 `:72` 「(승격이 상태를 바꾸면 그 통지도 … 기록만 한다)」와 `:36` 「통지를 없애서는 안 된다」는 설계가 유지하는 기록 실패 경로와 어긋남. 그 경로는 `n.escalate`인데 announcer가 의도적으로 nil이라 통지 자체가 없음 | `design.md:422` · `notifier.go:219` · `:360-368` · `:382-383` | 통지 없음을 명시적으로 예외 처리(구조화 로그가 기록) |
| 11 | P2 (T) | exit-policy MODIFIED `:28` 「캡 발생은 알림된다(SHALL)」와 Scenario `:54-56` 「알림이 발송된다」는 무조건인데, 같은 요구 `:30`과 `:64`, D0.3g 6은 버퍼가 차면 버림(C8). 정본 편입 요구 `exit-policy/spec.md:89` · `:101`(무관리 알림 「발송된다」)도 같은 모양임 | 델타 exit-policy `:28` · `:54-56` · `:64` · `design.md:693` · `obs/event.go:317-337`(두 이벤트 모두 일반 등급) | 「버퍼가 비어 있는 동안; 버려지면 기록」 한정어 추가 |
| 12 | P2 | 「현재 모드 = 커밋 순서 하나」의 열거에서 `OperatingModeHistory`(`ORDER BY created_at, rowid`)가 빠짐. rowid = 커밋 순서의 전제에 VACUUM 관련 단서가 없음(추측: SQLite 문서는 INTEGER PRIMARY KEY가 없는 테이블의 rowid를 VACUUM이 바꿀 수 있다고 함. 복원 경로가 VACUUM INTO 사본이고, 테이블을 재구성하는 마이그레이션도 같음. 상대 순서는 보존될 가능성이 높음) | `operating_mode.go:593-596` · `core_domain.go:183-193` · `backup.go:27` · `:76` | 이력 순서도 rowid로 바꾸거나 제외 사유를 적음. 복원·마이그레이션 뒤 순서 핀 추가 |
| 13 | P2 (T) | AC2 「보장 시점 = 전이 호출의 반환」 대 risk-management `:133` 「journal 영속과 **동시에** … 투영」: 「동시에 = 같은 전이 호출 안」으로 읽으면 모순은 아님. 다만 ADDED가 그 해석을 명시하지 않음 | 델타 `:295` · risk-management `:133` | 해석 문장 한 줄 추가 |
| 14 | P2 | D0.3g 3은 「a124 늦은 적용과 같이」라고 하지만, a124 `:1463`(승격 쓰기 실패 시 무조건 차단)을 옮기지 못함. `n.escalate`는 반환값이 없어 발송자가 실패를 알 수 없음 → 해제 + 원장 오류가 겹치면 차단과 모드를 둘 다 잃음 | `notifier.go:378-399` · 정본 `:1463-1464` · `design.md:654` | 승격 성공 여부를 발송자에게 돌려주거나 잔여로 명시 |
| 15 | P2 | tasks §22·§21의 커버리지 빈칸. ADDED 「자동 강화…」의 「AND 청산은 영향 없음」 · 완화 통지 **성공** 경로(C13은 실패 보고만 봄) · 입구 도달 경로 전수 구조 핀 · 「진입 차단이 늦게 걸린다」의 「청산 무영향」 · 선점 기록과 「남은 행 계속」 · 옛 §6(⛔)에 매달린 Scenario 넷(상한을 읽지 않는 transport · 예산을 줄여도 · 다시 올릴 주기 · 사이클 총 체류). 22.2 FLM 목록에 `RestoreOperatingModeProjection`(rowid를 실음) · `Notifier.AnnounceOperatingMode`(사건 구성 추출, `design.md:621-622`) · `currentModeFromRow`가 없음. C15 복원 실패 처분을 델타가 언급하지 않음 | `tasks.md:17-41` · `:59-79` · 델타 `:289` · `:306` · `:310-313` | task 추가 |
| 16 | P2 | `Flush`가 `n.mu`를 쥔 채 `Publish`함. 오늘 생산 호출자는 0이고 델타가 「엔진이 조립해 실행하는」으로 범위를 좁혀 둠. 그러나 생산 호출자가 생기는 것을 막는 핀이 없음 | `notifier.go:734-735` · `:779` · 델타 `:52` | 「Flush 비시험 호출자 0」 핀 |
| 17 | P2 (T) | `check_values.py`는 지연 수치만 검사함. 22판은 새 수치가 없어 검사 대상이 0임. FAIL 60건은 전부 22판 이전 절(design `:777`~`:2209`, proposal `:158` · `:454`, tasks `:349`~`:2295`). 좌표 · AST 포함 관계 · MANIFEST sha 주장 · 정본 줄 인용을 보지 못함(#3·#7·#8이 통과됨) | 실행 rc=1, 60건. `tools/check_values.py:70-252` | 좌표 대조 검사 추가(extract JSON 대 문서의 `file:line`) |

## 필수 판정 1~8

1. **제목 달성 — 설계대로면 원격 동기 대기 도달 경로는 0 (조건부).**
   - 도달 경로 전수: Notify 입구 하나(`o.alert` 7자리 `exitloop.go:830` · `:1536` · `:1606` · `:1632` · `:1656` · `:1686` · `exit_quarantine_announce.go:71` → `:1710`)와 Announce 입구 셋(`checkOutage` `:846`, `observe` `:747` → `retry.go:357` · `:365` · `:413`, `ConfirmedFloor` `exitwiring.go:207` · `:231`).
   - 그 밖을 배제한 근거: Gateway Submit에는 notifier가 없고 `parkAlert`는 적재만 함(`replay.go:551`). Guardian은 `escalateFor`를 `IssueEntry`(`riskguardian.go:436`)에서만 부르고 `IssueReduction`에는 없음. 원장에서 통지하는 곳은 `operating_mode.go:479` 하나뿐. 기록 실패 시 승격은 announcer nil(`notifier.go:382-383`). 편입은 대사 루프 소속(`adoption.go`의 `ReconcileDriver`).
   - D0.3g 1의 측정은 참: `Retrier`는 설정값만 가진 구조체이고 메서드는 읽기만 함(`retry.go:305-328` · `:335-446`). 생성은 한 곳(`exitwiring.go:47` ← `gateway.go:324`). floor는 exit만 씀(`engine.go:589` → `exitwiring.go:345`).
   - 조건: C16(`Journal==nil`일 때 동기 publish 금지)과 C8. 도달 경로 전수를 기계로 강제하는 장치는 없음(#15).
2. **RecordAlert 안 나 — 조건부 유지.**
   - 어댑터가 `n.mu` 아래에서 부르므로 셈~해제 배제가 유지됨. 중복 제거와 재무장은 `recordAlertTx` 그대로임.
   - 정착 행은 임차를 들고 있지 않음(승인은 `outbox.go:494-499`, 재무장은 `:334-342`에서 `alertClaimCleared`로 지움) → 재무장이 산 임차를 건드리지 않음.
   - 위험: 공개 `RecordAlert`에 핀이 없음(#4).
3. **원칙 E — 해제 세대 읽기 뒤의 해제에 대해서는 닫힘.**
   - 확정~읽기 창은 남음. a124는 허용하지만 델타 Scenario가 무조건이라 지킬 수 없음(#2).
   - a092의 세 자리에서 게이트를 잘못 여는 새 경로는 찾지 못함. 해제에는 미전달 0이 필요하고, 그 조건이면 그 행이 이미 승인됐거나 사라진 상태라 제때 적용한 결과와 같음.
   - 이중 결함 시 보수 조항이 누락됨(#14).
   - 교차: a094는 세대를 적재 **전**에 읽음(a094 `design.md:62`). 적재 실패 앞의 해제로 차단이 건너뛰어짐 → a124 `:1465` 위반(a094 몫).
4. **커밋 순서 — 설계상 방향 판정·복원·울타리가 한 순서로 모임.**
   - rowid 근거 셋은 참: rowid 테이블(`core_domain.go:183-193`) · 생산 `DELETE`/`UPDATE` 0(grep; INSERT는 `operating_mode.go:445` 하나) · BEGIN IMMEDIATE(`journal.go:225` `_txlock=immediate`, `:391`).
   - 빠진 것: 이력 순서(#12).
   - AC2는 모순이 아님. 해석을 명시할 것(#13).
5. **archive 정합성 — 불가.**
   - 정본 모순: #1.
   - 거짓 자기 인용과 낡은 좌표: #6 · #7.
   - 범위 오독: #9. 지킬 수 없는 무조건 문장: #2 · #10 · #11.
   - 새 MODIFIED 「배달 실행자의 정지가 …」의 규범 문장은 정본과 근거 ①만 다름(raw diff 확인: `:36-38` 교체와 메모 `:155-156` 추가). 그러나 같은 블록 안의 다른 서술이 거짓이 됨.
   - 21판→22판 삭제(보조 실행자 문단 · 실행자 정지 문단 · Scenario 「배달 실행자가 죽는다」 · HEAD 좌표)는 전부 정본 문장이 소유함: `:203` · `:1005` · `:1022-1045` · `:1134-1135`. 해당 없음을 확인함.
6. **Q4 핀은 적힌 대로는 서지 않음(#3 · #4 · #5). C8은 섬.**
   - C8: 정본 `:203`의 보조 실행자 의무 셋을 지고, `AuxiliaryExecutor.OnStop`이 실행자별이라(`auxiliary.go:71-76`) 「래치 없음」이 가능함. 문서 결함은 #11.
7. **a094 교차 — 같은 지반을 밟음.**
   - (a) Q4 SHALL·핀 대 a094의 직접 적재(선행 잠금 없음)
   - (b) 델타 `:38`의 재무장 SHALL 대 a094의 remindAfter 0
   - (c) 세대 읽기 시점 규약이 둘로 갈림
   - (d) a092 자신의 통지 key에 에피소드 신원이 없음(#1). a094의 처방(key에 에피소드, outbox 계약 무변경)이 a092에도 맞음.
   - C8은 일반 등급이라 outbox가 없어 겹치지 않음. a066 작업 트리(미커밋)도 (a)와 같은 모양.
8. **좌표·AST 인용 전수.**
   - 맞음: `exitloop.go:747` · `:846` · `:1513` · `:1710` · `retry.go:305-328` · `:357` · `:362` · `:365` · `:413` · `:572` · `:574` · `:577` · `exitwiring.go:207` · `:231` · `:333` · `gateway.go:269-271` · `:324` · `:350-353` · `reconcileloop.go:360` · `position_policy_command.go:94`(7cf80832) · `adoption.go:257-258` · `mode.go:49-74` · `:57` · `outbox.go:131-154` · `:231-267` · `:276-364` · `:372-418` · `notifier.go:223` · `:228` · `:484` · `:520` · `:571` · `:851` · `:855` · `:864` · `:875` · `operating_mode.go:388` · `:391` · `:478` · `:479-481` · `:618-626` · `core_domain.go:183-193` · `engine_risk_relaxation.go:136` · `:194` · risk-management `:133` · `:137` · MANIFEST 개수 39.
   - 틀림: #3(B2 포함 관계) · #7 · #8.

`git status --short -- openspec/changes/a092-an-alert-does-not-hold-the-stop` 결과: `?? openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/review-22/` 한 줄뿐(건드리지 않음). 이 리뷰가 만든 파일은 전부 `/tmp/claude-1000/a092-r22/voiceB-*` 아래에 있음.

Recommendation: BLOCK because 모드 통지 key에 에피소드 신원이 없어서 a092 ADDED의 통지 SHALL과 정본 중복 억제 SHALL이 a092가 새로 여는 완화→재강화 주기에서 정면 충돌하고(P0), Q4 핀·원칙 E Scenario·MODIFIED 인용 블록이 스스로 모순되거나 지킬 수 없는 문장을 정본에 넣기 때문(P1 다섯) — key에 전이 rowid를 넣고, Scenario에 a124 한정어를 복원하고, 핀 범위를 RecordAlert와 미배선 게이트까지 바로잡고, 「a092가 진다」 블록과 낡은 좌표를 정정한 뒤 재검토할 것.
