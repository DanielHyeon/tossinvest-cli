# a109 뮤테이션 원장 — T1 (`internal/positionpolicyrpc`, `internal/app/engine`의 세 transport)

tasks 1.6. design D1·D1a·D1b·D2·D2a·D2b의 판정을 **하나씩** 뒤집어 어떤 테스트가 죽는지
관측한 기록이다. 통과는 증거가 아니다 — 죽여 본 뒤에 근거로 쓴다.

## 실행 방법

각 뮤테이션은 치환 1건을 적용하고, **먼저 `go build ./...`가 통과하는지 확인한 뒤**
(컴파일 에러는 뮤테이션이 아니다) 아래 둘을 돌려 실패 테스트를 수집했다.

```text
go test ./internal/positionpolicyrpc/ -count=1
go test ./internal/app/engine/ -count=1 -run 'TestTheSibling|TestTheCommandEndpointSweeps|
  TestTheStagedSocketNameFits|TestEveryNameThePublishing|StartsOverALeftover|TestTheAlertSocket|
  TestTheOperatorSocket|TestRuntimeOnlyUnix|TestCommandAndRuntimeControls|TestEngineOwnsAuthenticated|
  TestPositionPolicyControlRejects|TestTheAcknowledgeRoute|TestTheAlertEndpointExposes'
```

원복은 `git checkout`이 **아니라 역편집**이다(치환된 문자열을 원래 문자열로 되돌린다).
적용 전에 작업 트리가 커밋돼 있는지 먼저 확인했다 — 커밋 전 GREEN에 checkout을 쓰면
GREEN까지 지워지고 `git diff`는 그것을 초록으로 통과시킨다. 원복 확인은 **두 가지**다.

- 파일 sha256이 적용 **전**과 정확히 일치(구동기가 매 건 검사, 불일치면 즉시 중단)
- 배치 종료 후 심볼 수 대조:
  `ValidatePrivateControlDirectory` 5 · `validateOwnerAndLinks` 11 · `privateSocketAccepts` 7 ·
  `verifyStalePrivateSocket` 3 · `verifyPrivateStagingEntry` 4 · `StagedSocketName` 12 ·
  `ListenStagedPrivateSocket` 5 · `ReclaimStalePrivateEndpoint` 7 ·
  `SweepPrivateStagingLeftovers` 5 · `MUTANT` 0
- 그리고 `git diff --quiet`(production 파일) 통과

구동기가 두 번 중단됐다(M10 1차, M13 1차) — 역편집 대상 문자열이 유일하지 않아서다.
그 두 건은 **손으로 역편집**하고 위 두 검증을 통과시킨 뒤 이어서 돌렸다. 중단이 남긴
상태를 그대로 두고 다음 뮤테이션을 적용하지 않았다.

---

## §A 죽은 뮤테이션 (1차 25건 — §B2·§D에서 7건이 더 붙어 최종 32건)

### A1. 회수의 분류 — "우리 것"과 "낯선 것"

| # | 뮤테이션 | 죽은 테스트 |
| --- | --- | --- |
| M1b | 낯선 엔트리를 `staged`에 넣어 **우리 잔재로 친다** | `TestTheSiblingEndpointsRefuseAForeignEntry` (두 endpoint 모두) |
| M2 | `case names.isStaging(name)` → `case false` (staging 분류 제거) | `TestTheSiblingEndpointsReclaimTheirStagingLeftovers` (둘) · `TestReclaimEmptiesTheDirectoryItRecovers` |
| M17 | runtime의 아는-이름에서 `.s-` 접두를 뺀다 | `TestEveryNameThePublishingPathMakesIsKnownToItsReclaim/position_policy_runtime` · `TestTheSiblingEndpointsReclaimTheirStagingLeftovers/position_policy_runtime` |
| M18 | alert의 아는-이름에서 `.endpoint-` 접두를 뺀다 | `TestEveryNameThePublishingPathMakesIsKnownToItsReclaim/alert_control` · `TestTheSiblingEndpointsReclaimTheirStagingLeftovers/alert_control` |
| M29 | 빈-접두 방어를 **두 곳 모두** 지운다(`validate` + `hasAnyPrefix`) | `TestReclaimRefusesNamesItCannotTrust/빈_staging_접두` |

M17·M18이 죽는 방식이 freeze P2-5의 요구 그대로다: 집합에서 하나를 빼면 **완전성 테스트와
잔재 회수 핀이 함께** 죽는다. 완전성 테스트 혼자 죽으면 그것은 집합에 대한 주장일 뿐이고,
회수 핀이 함께 죽어야 그 주장이 동작과 묶인다.

### A2. 사망 증명(probe)의 갈래 — 1차 네 갈래, §1-fix 이후 세 갈래

⚠️ **M7의 대상 절은 §1-fix F1이 삭제했다.** owner 쓰기 비트로 생사를 **추정**하던 절이
그것이고, 수락 중인 socket을 지우는 값이었다(A1 P1-A). 아래 M7 행은 1차 측정의 기록으로
남기고, 지금 그 자리를 지키는 것은 §D의 F1-N1이다.

| # | 뮤테이션 | 죽은 테스트 |
| --- | --- | --- |
| M3 | `if privateSocketAccepts(...)` → `if false` (probe 자체를 건너뛴다) | `TestTheSiblingEndpointsRefuseToTakeALiveOwnersSocket` (둘) |
| M4 | 수락 성공을 `return false`로 (산 주인을 죽었다고 읽는다) | 위 + `TestThePrivateSocketProbeReadsOnlyThreeThingsAsDead/수락_중이면_살아_있다` |
| M5 | ECONNREFUSED를 사망으로 읽지 않는다 | `.../연결_거부는_사망` · `TestTheSiblingEndpointsStartOverAPreChmodSocket`(둘) · **a108 관용 핀 4건**(`…StartsOverALeftover/socket만_남았다`·`둘_다_남았다`) · `TestReclaimEmptiesTheDirectoryItRecovers` |
| M6 | 파일 부재를 사망으로 읽지 않는다 | `.../파일_부재는_사망` |
| M7 | owner 쓰기 비트 판정을 뒤집는다 | `.../owner_쓰기_비트가_없으면_사망` |

M5가 죽이는 목록이 이 change의 요지를 그대로 보여 준다: 사망 판정 하나를 잃으면 **모든
정상 잔재가 산 주인으로 오독되어** 기동이 매번 거부된다.

### A3. 완화의 폭과 위치

| # | 뮤테이션 | 죽은 테스트 |
| --- | --- | --- |
| M8 | 잔재 socket의 perm 검사를 통째로 없앤다 | `TestTheSiblingEndpointsRejectAPermissiveSocketLeftover` (둘) |
| M9 | 잔재 socket perm을 **정확-0600으로 되돌린다**(병의 복원) | `TestTheSiblingEndpointsStartOverAPreChmodSocket` (둘) · `TestReclaimEmptiesTheDirectoryItRecovers` |
| M19 | 클라이언트 검증 `statPrivateSocket`을 `perm&0o077`로 완화 | `TestTheClientSocketChecksStayExactlyZeroSixHundred/ValidatePrivateSocket` |
| M20 | 클라이언트 검증 `ValidateRuntimeSocket`을 같은 방식으로 완화 | `.../ValidateRuntimeSocket` |

M8과 M9가 완화의 **양 끝**이다: 넓히면 M8이 죽고, 좁히면 M9가 죽는다. M19·M20은 freeze
P1-3이 지목한 누출 경로이고, 손쉬운 구현(공유 helper 완화)이 정확히 이 둘을 건드린다.

### A4. 완화하지 않은 검증들

| # | 뮤테이션 | 죽은 테스트 |
| --- | --- | --- |
| M10a | 잔재 socket의 hard link 요구를 지운다(`true`→`false`) | `TestReclaimRefusesASocketWithASecondHardLink` |
| M10b | 잔재 socket의 소유·링크 검증을 통째로 지운다 | 같음 |
| M11 | staging 엔트리의 모양 검증을 지운다 | `TestReclaimRefusesAStagingEntryOfTheWrongShape` · `TestTheStagingSweepLeavesEverythingElseAlone` |
| M13 | 회수의 control 디렉터리 검증을 지운다 | `TestReclaimRefusesAnUnsafeControlDirectory` |
| M14 | 잔재 descriptor의 형식 검증을 지운다 | `TestReclaimRefusesADescriptorOfTheWrongShape` |

⛔ **M10a·M10b·M13·M14는 1차 측정에서 살아남았다.** 그 넷을 죽인 세 핀
(`…ASocketWithASecondHardLink` · `…AnUnsafeControlDirectory` · `…ADescriptorOfTheWrongShape`)은
**뮤테이션이 열어서 추가한 것**이지 원래 있던 것이 아니다. 즉 §1.5의 경계 테스트만으로는
"검증을 완화하지 않았다"를 증명할 수 없었다 — 검증이 거기 있다는 사실과 그 검증이 무언가를
막는다는 사실은 다른 주장이다.

### A5. 발행·제거·위생

| # | 뮤테이션 | 죽은 테스트 |
| --- | --- | --- |
| M15 | 제거의 ErrNotExist 관용을 지운다 | a108 관용 핀 6건 + `TestTheSiblingEndpoints…StagingLeftovers`(둘) · `…StartOverAPreChmodSocket`(둘) |
| M16 | 회수가 디렉터리를 비우지 않는다(rmdir 생략) | 위 + `TestReclaimEmptiesTheDirectoryItRecovers` (총 17건) |
| M21 | staged listen의 chmod 0600을 지운다 | **23건** — 두 endpoint의 기동·라우트 테스트 전부 |
| M24 | command endpoint의 staging 위생을 지운다 | `TestTheCommandEndpointSweepsItsStagingAndLeavesStrangers` |
| M25 | 위생이 접두를 보지 않는다(전부 지운다) | 위 + `TestTheStagingSweepLeavesEverythingElseAlone` |
| M28 | hex 절단을 되돌린다(staging 12자) | `TestTheStagedSocketNameIsElevenCharacters` · `TestTheStagedSocketNameFitsInsideEverySiblingsFinalName` |

M28이 죽는 것이 D1a의 핵심이다: 12자 staging은 **`runtime.sock`에서는 통과하고
`alerts.sock`에서만 깨진다.** 길이를 상수에서 계산하는 테스트였다면 두 뮤테이션 모두
살아남았을 것이다 — 생성기의 산출물을 세기 때문에 죽는다.

---

## §B 생존 뮤테이션 (3건) — 사유와 함께 선언

⚠️ **1차 §B는 6건이었다.** 그중 셋(M1a·M22·M23)은 A1 적대 리뷰가 반증했다 — 죽일 수
없었던 것이 아니라 **방법을 안 썼던 것**이다. §B2에 재분류를 적는다.

| # | 뮤테이션 | 왜 죽일 수 없었나 |
| --- | --- | --- |
| M12 | staging 엔트리의 **소유 uid** 검증을 지운다 | 비root 테스트는 파일 소유자를 바꿀 수 없어 이 절만 실패시키는 디스크 상태를 만들 수 없다(a108 `ownedByEffectiveUser` 주석의 같은 사유). 모양 검증(M11)과 socket 쪽 소유 검증(M10b)은 죽는다. |
| M26 | `AlertControlServer.Close`의 listener 직접 닫기를 지운다 | 경합이다. Close 계약(세 경로 부재)은 그대로 통과한다 — 그것을 재는 테스트는 있고, 늦은 unlink를 재는 테스트는 없다. ⚠️ M22가 반증된 이상 이 사유도 **잠정**이다: listener를 직접 닫지 않아도 `Shutdown`이 대개 닫아 주므로 "누가 닫았는가"를 파일 관측으로 가르는 방법을 아직 찾지 못했다는 뜻이지, 없다는 뜻이 아니다. issues.md I3에 후속 후보로 등록했다. |
| M27 | 이름 집합의 빈-접두 방어(`validate`)를 **단독으로** 지운다 | **등가다.** `hasAnyPrefix`가 독립적으로 빈 접두를 건너뛰므로 단독 제거는 동작을 바꾸지 않는다. 두 곳을 함께 지우면(M29) 죽는다 — 방어가 이중임이 그 대비로 측정됐다. |

---

## §B2 재분류 — "죽일 수 없다"가 아니라 "방법을 안 썼다" (2026-08-15, §1-fix)

A1 적대 리뷰(별도 컨텍스트, Opus)가 §B의 세 사유를 반증했다. 셋 다 **같은 오류**를
공유한다: 죽이려면 그 뮤테이션이 되살리는 **사고 시나리오를 재현해야 한다**고 가정했다는
것. 재현할 필요가 없었다 — 그 판정이 남긴 **관측 가능한 흔적**을 재면 된다.

| # | 1차 선언 | A1의 반증 | 재적용 결과 (§1-fix) |
| --- | --- | --- | --- |
| M23 | "crash 창이라 죽일 수 없다 — listen과 chmod 사이에서 프로세스를 죽일 수 없다" | **프로세스를 죽일 필요가 없다.** bind한 이름이 무엇인지는 `listener.Addr()`이 그대로 들고 있다 | **사망.** `TestTheStagedListenBindsAStagingNameNotTheFinalName` · `TestTheStagedListenDoesNotUnlinkTheNameItRemembers` |
| M22 | "경합이라 관측 가능한 피해가 없다 — a108도 300라운드 중 3회" | **경합을 재현할 필요가 없다.** listener가 **기억하는 이름**에 파일을 두고 Close하면, unlink가 켜져 있으면 사라지고 꺼져 있으면 남는다 | **사망.** `TestTheStagedListenDoesNotUnlinkTheNameItRemembers` |
| M1a | "등가에 가깝다 — 무시해도 rmdir이 ENOTEMPTY로 실패해 기동은 거부되고 파일도 남는다" | **등가가 아니다.** 이물과 **우리 잔재**가 공존하면, 무시한 판은 rmdir 전에 우리 socket·descriptor를 **이미 지운다.** 1차 핀은 이물만 있는 디렉터리를 만들어 그 차이를 볼 수 없었다 | **사망.** `TestTheSiblingEndpointsRefuseAForeignEntryAndKeepTheirOwnLeftovers`(둘) · `TestTheSiblingEndpointsNameTheEntryTheyRefused`(둘) · `TestReclaimSaysWhichEntryItRefused/낯선_엔트리` |

M1a의 오류가 특히 값진 반증이다. 1차 사유는 **관측한 것이 아니라 추론한 것**이었고
(rmdir이 실패하니 결과가 같을 것이다), 그 추론은 디렉터리에 우리 잔재가 하나도 없는
상태만 상정했다. 사고의 실제 모양은 잔재와 이물이 **함께** 있는 것이다.

정산: 생존 6 → **3**(등가 1 · 경합 1(잠정) · 측정 불가 1). 사망 25 → **28**.

---

## §D §1-fix 라운드 — A1 판결 구현의 방어를 잰다 (2026-08-15)

구동은 §A와 같다(적용 → `go build` → 두 패키지 → 역편집 → sha256 대조). 구동기가 두 번
중단됐다(M1a 역편집 대상 비유일, F5-N1 역편집원이 빈 문자열) — 두 건 다 **손으로 역편집**한
뒤 `git status --porcelain`과 `git diff --stat`이 **빈 출력**임을 확인하고 이어서 돌렸다.

| # | 뮤테이션 | 결과 | 죽인 테스트 |
| --- | --- | --- | --- |
| A1-N1 | 발행 접두를 리터럴 `.endpoint2-`로 갈라놓는다(A1 원본 재적용) | **사망** | `TestEveryPublishedStagingPrefixIsTheSharedConstant/alert_control_transport_unix.go` |
| F1-N1 | probe를 chmod 없는 owner-write 추정으로 되돌린다(병의 복원) | **사망** | `…ProbeReadsOnlyTwoThingsAsDead/쓰기_비트가_깎여도_수락_중이면_생존` · `TestReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped` · `TestTheSiblingEndpointsRefuseALiveOwnerWithoutTheWriteBit`(둘) |
| F5-N1 | staging socket의 사망 증명을 지운다(이름표만 보고 지운다) | **사망** | `TestReclaimRefusesALiveStagingSocket` · `TestReclaimSaysWhichEntryItRefused/수락_중인_staging_socket` · `TestTheSiblingEndpointsRefuseALiveStagingSocket`(둘) |
| F6-N1 | 거부 오류에서 엔트리 이름을 뺀다(세 자리 전부) | **사망** | `TestReclaimSaysWhichEntryItRefused`(3 하위) · `TestTheSiblingEndpointsNameTheEntryTheyRefused`(둘) |

A1-N1이 이 라운드의 요지다. 1차에서 그것이 살아남은 것은 이름-집합 완전성 테스트가
**상수를 가져다 자기가 `os.CreateTemp`를 돌렸기** 때문이다 — 발행 경로가 리터럴로
갈라져도 그 테스트는 통과한다. "발행이 이 상수를 쓴다"는 **동적으로 잴 수 없는 성질**이고,
그래서 소스를 파는 정적 핀으로 갈랐다.

F1-N1은 M9(완화를 정확-0600으로 되돌린다)와 짝이다. M9는 회수의 **폭**을 되돌리고,
F1-N1은 회수의 **판정 방법**(묻기 → 추정)을 되돌린다. 둘 다 죽는다.

---

## §C 컴파일 실패로 폐기한 시도 (뮤테이션 아님)

| 시도 | 왜 뮤테이션이 아닌가 |
| --- | --- |
| M5 1차 (`unix.ECONNREFUSED` 절 삭제) | `unix` import가 미사용이 되어 빌드 실패. `!errors.Is(...) &&`로 재구성해 다시 측정했다. |
| M7 1차 (`statErr == nil && false`) | `info`가 미사용이 되어 빌드 실패. 판정을 뒤집는 형태로 재구성했다. |
| M27 1차 (`if false`) | `prefix`가 미사용이 되어 빌드 실패. 비교 대상을 바꾸는 형태로 재구성했다. |

---

## 정산

- 1차 적용 31건 + §1-fix 4건(신규) + 재적용 3건(M1a·M22·M23) = **35종 적용**
- 최종 = **사망 32 · 생존 3**(등가 1 · 경합 1(잠정) · 측정 불가 1)
- 뮤테이션이 **새 핀 3개를 열었다**(hard link · 디렉터리 위생 · descriptor 모양).
  §1.5 경계 테스트만으로는 A4의 네 뮤테이션이 전부 살아남았다.
- 회수의 모든 분기(분류 5 · probe 3 · 완화 4 · 검증 5 · 제거·위생 6 · 발행 의례 2 ·
  거부의 지목 1)에 죽는 뮤테이션이 하나 이상 있다. 예외는 §B에 사유와 함께 선언했다.
  (probe가 4갈래에서 3갈래가 된 것이 §1-fix F1이다 — owner-write 추정 절을 지웠다.)

### 이 원장이 1차에 틀린 방식

§B2의 셋은 전부 "이 뮤테이션이 되살리는 **사고를 재현해야** 죽일 수 있다"고 적었다.
crash 창을 열 수 없어서, 경합을 재현할 수 없어서, 결과가 같아 보여서. 셋 다 **사고를
재현할 필요가 없었다** — 판정이 남긴 흔적(bind한 이름, 기억하는 이름, 지워진 우리 파일)을
재면 됐다.

생존 선언은 그래서 **"이 뮤테이션은 죽지 않는다"가 아니라 "내가 쓴 방법으로는 안 죽었다"**로
읽어야 한다. 남은 §B 셋도 그 지위다.
