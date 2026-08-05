# a084 · Issues

## I1 — a078 I1의 설계 공백은 닫혔다

a078은 이렇게 남겼다.

> 즉 현재 코드에는 "격리를 풀고 원래 기준선을 유지하는" 경로가 없다. 이것 자체가
> 설계 공백이고, 후속 change 후보다.

a079가 사람 경로를, a084가 기계 경로를 만들었다. 저장된 진입가·최초 손절·보호선은
두 경로 모두에서 그대로 유지된다.

남은 조건: **개정이 같은 격리는 여전히 사람만 푼다.** 같은 선택기가 같은 얼어붙은
입력에 다른 답을 낼 수 없으므로 그것이 맞다.

## I2 — `ambiguous_recovery` evidence가 어느 분기에서 나왔는지 말하지 않는다

운영 원장의 세 건이 전부 접미사 없는
`exitpolicy: recovery candidate identity mismatch`다. `SelectRecoverySnapshot`에서 이
**맨** sentinel을 반환하는 경로는 두 개뿐이다 — 정체성 튜플 검사와
`compareRecoveryStage`의 rank 실패. 나머지 반환은 전부 `%w: <상세>`로 감싼다.

그 둘을 사후에 구별할 수 없다. a078은 high water가 첫 rung 문턱 바로 아래에 몰려 있다는
정황으로 특정했고, a084는 저장 snapshot을 재생해 확인했다. 둘 다 원장이 직접 말해줬어야
할 것을 추론으로 메운 것이다.

두 반환에 `%w: <어느 필드가 어긋났는가>`를 붙이면 다음 번엔 추론이 필요 없다. 격리
판정 자체는 바뀌지 않으므로 작은 change다.

## I3 — 개정 상수는 손으로 올린다

`exitpolicy.RecoverySelectorRevision`을 올릴 책임은 사람에게 있다. digest 자동 산출은
무관한 리팩터가 전 포지션을 재판정하게 만들므로 채택하지 않았다(design D3).

빠뜨린 경우의 방향은 안전하다(자동 재시도 없음 = 오늘의 동작). 다만 **선택기를 고치는
change의 tasks에 "개정 번호를 올렸는가"를 넣는 것**이 이 위험을 실질적으로 없애는
방법이고, 그것은 워크플로 문서의 몫이다.

## I4 — v27 마이그레이션 테스트가 §0.6과 충돌하고 있었다

`journalV25RowFingerprints`가 열 이름까지 해시해서, v26 이전 테이블에 nullable 열을
추가하는 것을 전부 거부했다. `schema.go`의 마이그레이션 규칙 2가 명시적으로 허용하고
`positions`·`mutation_attempts`가 이미 지난 경로인데도, v27~v29가 새 테이블만 만들어
드러나지 않았다.

fingerprint를 행만 해시하도록 좁히고 열 검사(`assertColumnsOnlyAppended`)를 분리했다.
**이 change가 없었다면 다음 additive 열 추가에서 같은 벽에 부딪혔을 것이다.**

## I5 — 독립 리뷰어 없이 배포 가능 상태에 도달했다

a083 I5와 같다. `codex`가 사용량 한도(2026-08-08 해제)로 거부됐다. 배포 전 별도 세션의
재검증이 남아 있다.

## I6 — 배포 전 `main` 대비 SchemaVersion 대조가 필요하다

`SchemaVersion` 29 → 30. 낮은 버전의 바이너리는 `ErrSchemaTooNew`로 거부하므로 롤백은
"이전 바이너리 실행"이 아니라 **DB 복원**이다. 마이그레이션 직전 자동 백업이 그 경로다.
배포 순서에서 콘솔·엔진이 같은 버전을 보게 해야 한다.

## I7 — 마이그레이션 하나가 journal 테스트 패키지를 Go 기본 timeout 밖으로 밀었다

`internal/journal`은 테스트마다 마이그레이션된 SQLite를 새로 열고 그런 테스트가 약 670개다.
그래서 **모든 새 마이그레이션 단계가 670번 곱해진다.**

```
v29 (a083 HEAD)   480.885s
v30 (a084)        795.642s      ← 둘 다 병렬 부하 아래 측정, 상대 비교용
```

v30에서 Go의 패키지당 기본 600초를 넘었고 `make test`가 timeout으로 실패했다. 매달린
것이 아니라 느린 것이다 — 넉넉한 timeout에서는 완주한다.

`make test`에 `-timeout 30m`을 붙였다. 정당한 마이그레이션이 걸리는 한계는 잘못된 실패를
보고하는 한계다.

**근본 원인은 남아 있다.** 다음 마이그레이션도 같은 곱셈을 낸다. 테스트가 스키마별
템플릿 DB를 복사해 쓰게 하면 마이그레이션 비용이 테스트 수와 무관해지지만, 그것은
별도 change다.

## I8 — `make gate`가 병행 세션의 진행 중 편집으로 막혀 있다

`make sdd-check`의 `check_agent_config_sync.py`가 drift를 보고한다.

```
 M .claude/CLAUDE.md
 M .codex/agents.md
```

두 파일 모두 **다른 세션이 지금 편집 중**이다. 한쪽에서 "에이전트 실행 순서" 절이
추가되고 advisory 관련 한 줄이 삭제된 상태이고, 두 파일의 SDD_SHARED 블록이 서로
어긋나 있다.

다른 세션의 진행 중 **안전 부트스트랩 편집을 덮어쓰지 않았다.** `--generate`는 그들의
작업을 지울 수 있다.

a084 자체의 증거는 전부 통과했다.

```
make test      EXIT=0 (go test -timeout 30m ./... 전 패키지 ok)
make vet       clean
make validate  77 passed, 0 failed
logic-map      evidence complete (11 함수)
openspec       valid --strict
```

**남은 조치**: 병행 세션이 두 파일을 커밋해 동기화한 뒤
`make gate CHANGE=a084-a-quarantine-outlives-its-cause`를 다시 돌린다.
