# a063 · Review

- 날짜: 2026-08-04
- 시점: proposal-freeze (첫 구현 task 착수 전)
- 대상: `proposal.md`, `design.md`, `specs/position-exit-policy-management/spec.md`
- 위험 등급: **High-risk** — 손절 판정을 재개시키는 운영자 mutation이고 원장에 쓴다.
  WORKFLOW가 요구하는 **적대적 Eng 관점**을 포함한다.
- 남은 절차: 구현 후 **별도 컨텍스트의 독립 검증**(WORKFLOW §9)은 아직 수행되지 않았다.

## Pre-Edit Gate

```text
- change id / task id: a063-operator-can-lift-a-quarantine / 8.1
- 대상 심볼(기존 함수 4개 — 배선 지점 3개 + 라우트 계약 테스트 1개):
    · internal/app/engine.StartPositionPolicyCommandServer   — 라우트 3개 추가
    · internal/console.(*Console).routes                     — 라우트 2개 추가
    · internal/console.(*Console).handlePositionManagement   — 행 badge/action 조립
    · internal/console.TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate
      — state-changing allowlist에 새 두 라우트 등재 (게이트가 요구한 선언)
- 기존 동작 파악 근거:
    · StartPositionPolicyCommandServer: 호출부 cmd/tossctl/engine.go:254 한 곳.
      mux 라우트 4개(/v1/health,/v1/positions,/v1/preview,/v1/apply)를 auth로 감싼다
    · NewPositionPolicyCommandService(cmd/tossctl/engine.go:250)는 **바꾸지 않는다**.
      격리 저장소는 이미 보관 중인 s.j를 지연 타입 단언으로 얻는다. 구조체에 늘어나는
      것은 zero value로 동작하는 capability 슬라이스 필드 하나뿐이라 생성자 본문이
      한 글자도 바뀌지 않는다
    · (*Console).routes: 708-820. 분기 2개(remote 로그인 등록, remote security 래핑)
    · handlePositionManagement: 203-264. 읽기 전용 렌더 경로. mutation 없음
    · Journal.ReleaseExitSnapshotQuarantine: 프로덕션 호출자 0
      (repo 전역 grep, 현재 HEAD). 유일 호출은 migration_v10_test.go:139
- 편집을 피한 지점(의도적):
    · cmd/tossctl.runConsole(36분기)과 runEngineRun(15분기)은 건드리지 않는다.
      선택적 capability는 타입 단언으로 발견한다 (D1 주석에 근거를 남긴다)
    · Journal.ReleaseExitSnapshotQuarantine 본문은 바꾸지 않는다 — 호출만 한다
    · positionPolicyLifecycleClient / positionPolicyRepository 인터페이스를 넓히지
      않는다. 넓히면 테스트 fake 전부가 깨지고, 그 fake들은 이 change와 무관하다
- upstream 상속 테스트 영향: no. 격리·정책 command plane 전체가 TossOS 신설이며
  upstream tossinvest-cli에 존재하지 않는다. 650 baseline 회귀 없음을 make test로 확인한다
- 실패 테스트 선행 작성: 계약(2장)·원장 read(3장)는 yes. 4~7장은 구현과 테스트를
  같은 루프에서 작성했고, 테스트가 실제로 물리는지는 변이 검증 6건으로 확인한다
  (아래 "구현 후 검토"). 편집 대상 기존 함수 중 판정·주문 로직은 하나도 없다
- 안전 불변식 §0 위반 여부 검토: 통과
    · §0.1 LIVE 주문 side effect: 없음. 이 경로는 주문을 만들지 않는다
    · §0.3 손절 즉시성: **회복 방향**. 현재 3개 포지션은 손절 평가가 아예 없다
    · §0.4 rate budget: 신규 공식 API 호출 없음
    · §0.5 운영 설정 audit: 해제는 주체·사유·증거와 함께 원장에 남는다
    · §0.6 원장 스키마: 변경 없음. 기존 v10 컬럼(released_at/kind/evidence)을 쓴다
    · §0.7 사람 승인: 해제는 콘솔 승인으로만 일어난다. 자동 해제 경로를 만들지 않는다
    · §0.9 단방향 안전: 판정 조건을 조금도 느슨하게 하지 않는다 (A3 참조)
```

## 적대적 Eng — 제기된 반론과 처리

### A1. "격리를 푸는 버튼은 안전장치를 끄는 버튼 아닌가" — **아니다, 구조적으로**

해제는 `released_at` 한 칼럼만 채운다. 다음 관측에서 `record()`가
`SelectRecoverySnapshot`을 **다시 돌린다**. 판정 로직은 한 글자도 바뀌지 않는다.

- 원인이 a062 결함이었다면 이제 정상 전진으로 선택되고 판정이 이어진다.
- 원인이 진짜 모호성이었다면 같은 입력이 같은 판정을 내리고 **즉시 다시 격리된다**.

해제로 감출 수 있는 결함이 없다. 이것이 안전장치를 끄는 버튼과의 결정적 차이다.
7.2가 이 성질을 테스트로 고정한다.

### A2. "그러면 무한 루프 아닌가 — 풀고 다시 걸리고" — **맞고, 그것이 올바르다**

원인이 남아 있으면 해제는 한 주기짜리다. 그것이 정확한 피드백이다. 운영자는
"풀었는데 또 걸렸다"를 보고 원인이 자기가 생각한 것이 아니었음을 안다. 자동 재시도는
하지 않는다 — 매번 사람이 승인해야 한다.

### A3. "§0.9는 단방향 안전만 허용한다. 이건 어느 방향인가" — **더 보수적이다**

현재 상태: 격리 → `exitloop`가 refused로 분류 → **판정 없음**. 손절도 익절도 평가되지
않고, 저장된 보호선은 아무도 확인하지 않는다.

해제 후: 판정 재개. 저장된 손절선이 다시 매 관측마다 평가된다. 보호가 **없던 것에서
있는 것으로** 바뀐다. 손절선 숫자 자체는 그대로다 (D5).

### A4. "재편입으로도 되지 않나. 왜 새 경로가 필요한가" — **다른 일이다**

재편입은 새 generation + 현재가 기준 새 stop이다. 466100이면 진입가 25,700과 초기
손절 24,929가 사라진다. 손실 중인 포지션을 재편입하면 **손절선이 현재가 아래로
새로 잡혀 원래 손절폭보다 훨씬 아래로 내려간다** — §0.9 위반 방향이다. 지금 그것이
유일한 우회라는 사실 자체가 이 change의 근거다.

### A5. "정책 CAS pipeline에 Action 하나 더 넣는 게 간단하지 않나" — **아니다**

그 경로는 `Request.Validate` → `previewPositionPolicy` → `ApplyPositionPolicy`(원장
write transaction) → `prepare` → `Apply` 다섯 기존 함수를 바꾼다. 그중 하나는 정책
변경 승인의 원장 트랜잭션이다. 그리고 격리 해제는 `position_policy_lifecycles`의
version을 올리지 않으므로 그 pipeline의 CAS 대상 자체가 아니다 — 억지로 넣으면
"version이 안 바뀌는 CAS"라는 모순을 특수 분기로 처리해야 한다. D1 참조.

### A6. "capability 코드 80줄 중복은 부채 아닌가" — **의도한 값이다**

추출은 `position_policy_command.go`의 기존 함수 본문을 바꾸는 일이고 그 파일은 정책
변경 승인 경로다. 그리고 두 도메인은 만료 규칙이 이미 다르다 — 재편입의 15초
freshness는 격리 해제에 의미가 없다. 한 헬퍼에 두 정책을 담으면 다음에 한쪽을 고칠 때
다른 쪽이 조용히 따라 움직인다.

### A7. "동시에 두 콘솔이 같은 격리를 풀면" — **CAS가 막는다**

`ReleaseExitSnapshotQuarantine`의 `UPDATE … WHERE quarantine_version=? AND
released_at IS NULL` + `RowsAffected()==1` 검사가 두 번째를
`ErrExitSnapshotReleaseStale`로 거부한다. capability도 1회용이다. 4.6이 고정한다.

### A8. "해제 직후 엔진이 캐시 때문에 못 알아채면" — **캐시가 없다**

`exitloop`는 주기마다 `ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq)`를
직접 조회한다. 프로세스 내 격리 캐시가 없다. `o.refused` 맵은 알림 중복 억제용이고
판정 대상 결정에 쓰이지 않는다. 7.1이 실제 루프로 확인한다.

### A9. "타이핑 확인을 빼면 오조작 위험이 커지지 않나" — **마찰의 종류를 고른다**

체크박스 + 3초 지연 + 1회용 capability + version CAS가 이미 있다. 타이핑 확인은
사용자가 명시적으로 금지한 마찰이고(2026-07-27 지시), 그 자리에 서버가 조립한
evidence가 들어간다 — 사람이 옮겨 적는 것보다 정확하다. D3 참조.

### A10. "이 change가 3건을 실제로 푸는가" — **아니다. 도구만 만든다**

해제 실행은 §0.7 사람 판정이다. task 11.2가 그것을 명시하고, 에이전트는 실행하지
않는다.

## Manager 셀프리뷰 — 스펙 품질

- ADDED만 사용했다. 기존 "release와 re-adopt는 lifecycle을 분리한다"는 그대로 참이고,
  a063은 **lifecycle을 건드리지 않는** 세 번째 동작을 정의한다. 충돌 없음.
- Scenario 7건이 갈래를 덮는다 — 정상·재개·재격리·stale·미확인·격리없음·기준선 보존.
- "해제는 판정을 느슨하게 해서는 안 된다(MUST NOT)"가 이 요구사항의 중심 문장이고,
  재격리 시나리오가 그것을 관측 가능한 형태로 고정한다.

## 결론

**진행 승인.** 완료 보고에 다음을 반드시 포함한다.

1. 격리 3건의 실제 해제는 이 change가 하지 않는다. 운영자 결정이다.
2. 격리 **생성** 시점의 관측과 critical 알림 전달은 a064가 맡는다.
3. `AUTHORITATIVE_RECONCILE`은 배선하지 않은 채로 남는다 (9.4가 확인).

---

## 구현 후 검토 (2026-08-04, 별도 패스)

**한계 명시**: 구현을 만든 컨텍스트가 diff를 다시 읽은 것이며 WORKFLOW §9의 **별도
세션 독립 검증이 아니다**.

**TDD 방식 고지**: 계약(2장)과 원장 read(3장)는 테스트 선행으로 썼다. engine service·
transport·console(4~6장)은 구현과 테스트를 같은 루프에서 썼다. WORKFLOW의 검증
기준은 요구사항↔테스트 추적성이며, 이 change에서 테스트가 실제로 물리는지는 아래
변이 검증 6건이 증명한다.

### 변이 검증 (tasks 9.1~9.6)

전부 백업 파일로 복구했고 복구 후 `diff`로 바이트 동일함을 확인했다.
(§ a061의 교훈: 미커밋 작업에는 `git checkout --`를 쓰지 않는다.)

| 변이 | 대상 테스트 | 결과 |
|---|---|---|
| `findQuarantine`의 version 비교 제거 | 4.6 | FAIL ✓ |
| `Confirmed` 검사 제거 | 4.3 | FAIL ✓ |
| 원장 read의 세대 술어 제거 | 3.2 | FAIL ✓ |
| 콘솔 token purpose 검사 제거 | 6.7 | FAIL ✓ |
| 위험 지연 3초 → 0초 | 4.4 | FAIL ✓ |
| 해제가 exit state를 함께 쓰도록 | 4.8 | FAIL ✓ |

**변이 검증이 실제로 값을 한 건**: `ReleaseQuarantine` 안의 두 번째 version 비교를
지웠을 때 **아무 테스트도 실패하지 않았다**. 조사해 보니 그 가드는 도달 불가능한
죽은 코드였다(`issues.md` I2). 제거하고 왜 하나로 충분한지를 주석에 남겼다.

### 구현 중 잡은 결함 (`issues.md` I3)

콘솔 행 조립의 첫 초안이 격리를 `State.AdoptionGeneration`으로 매칭했다. 그 값은
`position_policy_lifecycles`에서 오고 lifecycle 명령이 없던 포지션에서는 기본값 1인
반면, 격리의 generation은 `positions.instance_seq`다. 운영 원장에서 IONQ는
`instance_seq=3`이므로 **초안 그대로였으면 IONQ와 042660의 격리가 화면에서 조용히
사라졌다.** 세대 규칙을 원장 읽기 한 곳에만 두도록 고쳤다.

### 게이트가 실제로 막은 것

`TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`가 새 mutation 라우트 두 개를
state-changing 목록에 등재하지 않은 상태에서 실패했다. 라우트 계약이 문서가 아니라
게이트라는 증거다.

### 실행 결과

- `go build ./...` 성공, `go vet ./...` 무결함.
- `go test ./...` — 79 패키지 **6094건 green** (a063 신규 44건 포함).
- `AUTHORITATIVE_RECONCILE` 프로덕션 호출자 0 유지 (task 9.7).
- `check_analysis.py` — 편집한 기존 함수 4개
  (`StartPositionPolicyCommandServer`, `Console.routes`,
  `Console.handlePositionManagement`, 그리고 라우트 계약 테스트
  `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`) 증거 완성. 구현 후 AST를
  재생성해 분기 번호를 갱신했다. 마지막 하나는 계획에 없던 것으로, 검사기가 테스트
  함수도 기존 함수로 계산하기 때문에 필요했다 — 그리고 그 map이 이 편집이 검사 완화가
  아니라 게이트가 요구한 선언임을 보인다.

### 계획 대비 축소한 것

- `NewPositionPolicyCommandService` 편집을 없앴다. 격리 저장소는 이미 보관 중인
  `s.j`를 지연 타입 단언으로 얻으므로 생성자 본문이 바뀌지 않는다.
- `cmd/tossctl.runConsole`(36분기)과 `runEngineRun`(15분기)은 처음부터 손대지 않았다.
  선택적 capability를 타입 단언으로 발견하는 방식이 그 둘을 모두 피한다.

### 남은 것

- 활성 격리 3건은 그대로다 (`issues.md` I1). 해제 실행은 운영자 결정이다.
- task 11.1(배포 후 실측)은 배포 전까지 열려 있다. a061의 6.7과 같은 상태다.
- 격리 **생성** 시점의 관측과 critical 알림 전달은 a064.
- 해제 이력 화면 부재(I6), a062 테스트가 실제로 무엇을 덮고 있었는지 재확인(I4)은
  후속 후보다.
