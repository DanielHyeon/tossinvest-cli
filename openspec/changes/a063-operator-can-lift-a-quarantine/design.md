# a063 · 설계

## D1 — 격리 해제는 정책 CAS pipeline이 아니라 자기 seam을 갖는다

**결정**: `positionpolicy.Action`에 다섯 번째 값을 추가하지 않는다. 격리 해제는
`internal/exitquarantine` 계약과 전용 command service·RPC 라우트·콘솔 핸들러를 갖는다.

**근거 1 — 도메인이 다르다.** positionpolicy가 답하는 질문은 "이 포지션을 어떤 exit
정책이 지배하며 자동관리 대상인가"다. 격리가 답하는 질문은 "저장된 exit snapshot을
믿을 수 있는가"다. 격리 해제는 `position_policy_lifecycles`의 status도
`desired_policy_id`도 version도 바꾸지 않는다 — 그 pipeline의 CAS 대상이 아예 아니다.

**근거 2 — 원장 쓰기 경로를 새로 만들지 않는다.** `ReleaseExitSnapshotQuarantine`는
이미 존재하고, `WHERE … AND quarantine_version=? AND released_at IS NULL` +
`RowsAffected()==1` 형태의 CAS와 kind·evidence 검증을 이미 갖췄다. 이 change는 그
함수를 **호출만** 한다.

**근거 3 — 기존 High-risk 함수 편집을 최소화한다.** Action을 추가하는 설계는
`positionpolicy.Request.Validate`, `previewPositionPolicy`,
`Journal.ApplyPositionPolicy`(원장 트랜잭션), `PositionPolicyCommandService.prepare`,
`.Apply` 다섯 개를 모두 바꿔야 하고 그중 하나는 원장 write transaction이다. 전용
seam은 대부분 새 파일이고, 기존 편집은 배선 지점 여섯 곳에 그친다.

## D2 — capability 모양은 정책 pipeline과 같되 코드는 공유하지 않는다

격리 해제 capability는 정책 것과 같은 성질을 갖는다.

| 성질 | 값 | 정책 pipeline과 같은가 |
|---|---|---|
| 토큰 | 32바이트 난수, SHA-256 digest만 서버 보관, 1회용 | 같다 |
| 위험 지연 | 3초 (`notBefore`) | 같다 |
| TTL | 5분 | 같다 |
| 엔진 인스턴스 바인딩 | instanceID + domain 상수 | 같다 (domain 문자열은 다르다) |
| 승인 확인 | `Confirmed` 체크박스 | 같다 |
| 소비 시점 | 인가된 Apply 진입 직후, 저장소 접근 전 | 같다 |

**공통 헬퍼로 추출하지 않는다.** 추출은 `position_policy_command.go`의 기존 함수
본문을 바꾸는 일이고, 그 파일은 정책 변경 승인 경로다. 약 80줄 중복이 그 파일을
건드리는 것보다 싸다. 두 도메인이 각자의 만료 정책을 독립적으로 갖는 것도 이 쪽이
낫다 — 재편입의 15초 freshness 같은 규칙이 한 곳에 섞이지 않는다.

## D3 — evidence는 서버가 만든다

`ReleaseExitSnapshotQuarantine`는 비어 있지 않은 `evidence`를 요구한다. 그 문자열은
**서버가 조립한다**.

```
LOCAL_OPERATOR released quarantine v1 (reason=ambiguous_recovery) from the console at 2026-08-04T12:00:00Z
```

운영자에게 문구를 타이핑시키지 않는다. 콘솔 UI에 타이핑 확인·추가 승인 마찰을 넣지
않는 것은 사용자 지시(2026-07-27)다. 위험 확인은 이미 체크박스와 3초 지연이 담당하고,
"누가·언제·무엇을" 은 서버가 아는 사실이므로 사람이 옮겨 적을 이유가 없다.

## D4 — 해제는 억제가 아니라 재시도다

이것이 §0.3·§0.9를 통과하는 핵심 논증이다.

해제는 `exit_snapshot_quarantines.released_at`만 채운다. 다음 관측 주기에서
`exitloop`의 `ActiveExitSnapshotQuarantine` 조회가 비고, 포지션은 working set에
돌아오며, `record()`가 `SelectRecoverySnapshot`을 **다시 돌린다**.

- 원인이 a062 결함이었다면(현재 3건) 이제 정상 전진으로 선택되고 판정이 이어진다.
- 원인이 진짜 모호성이었다면 같은 관측이 같은 판정을 내리고 **즉시 다시 격리된다**.

즉 해제로 감출 수 있는 결함이 없다. 해제가 하는 일은 판정을 한 번 더 시도하게 하는
것뿐이고, 판정 자체는 조금도 느슨해지지 않는다. 지금 상태(판정 없음)보다 어느
방향으로도 덜 안전해지지 않는다.

## D5 — 기준선은 유지된다. 재편입과 무엇이 다른가

| | 판정 격리 해제 (a063) | 자동관리 해제 → 재편입 (기존) |
|---|---|---|
| adoption generation | 그대로 | **새로 만든다** |
| 진입가 · 초기 손절 | 그대로 | **현재가에서 다시 만든다** |
| high-water · rung | 그대로 | **버린다** |
| `exit_states` write | **없다** | 있다 (`resetExitStateForReadoptTx`) |
| 쓰는 테이블 | `exit_snapshot_quarantines` 한 행 | lifecycle + audit + exit state |

466100으로 말하면, 해제는 진입가 25,700과 초기 손절 24,929를 그대로 둔 채 판정만
재개한다. 재편입은 둘 다 오늘 가격 기준으로 갈아치운다.

## D6 — 무엇을 보여주고 무엇을 묻는가

preview는 운영자가 판단에 쓸 사실만 담는다.

- 종목 · generation · quarantine version
- 격리 사유(`reason`)와 저장된 원문 증거(`evidence`)
- 격리 시각과 경과
- 해제 후 유지되는 보호선 — 저장된 snapshot의 `CurrentProtection`
- 명시 경고: **원인이 남아 있으면 다음 관측에서 다시 격리된다**

묻는 것은 체크박스 하나다. 3초 지연 뒤에만 적용된다.

## D7 — 정지선

- `ReleaseExitSnapshotQuarantine` 본문을 바꾸지 않는다.
- `quarantineExitSnapshotTx`·`QuarantineExitSnapshot`를 바꾸지 않는다. 격리 생성
  조건은 이 change의 범위 밖이다.
- `AUTHORITATIVE_RECONCILE`을 배선하지 않는다. 사람이 아닌 무엇도 격리를 풀지 않는다.
- `exit_states`·`positions`·`position_policy_*`에 쓰지 않는다.
- 평가기(`EvaluateLadderSnapshot`·`EvaluateRatchetSnapshot`)와 손절·익절 계산식을
  건드리지 않는다.
- 새 공식 API 호출이 없다. rate budget(§0.4) 계상 대상이 아니다.
