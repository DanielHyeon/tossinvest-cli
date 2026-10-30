# a169 재개 초안 독립 문서 검토

대상 `/home/daniel/tossos-work/a166-runtime/openspec/changes/a166-account-shared-rate-limit/analysis/a169-restart-contract-draft.md` 2026-10-31 읽은 원문, source local origin/main 6e0968222993756a41065044996f674b35e20a27. 손읽기만; AST·시험·운영 미실행. 결론: 초안은 대체로 승인 결정과 일치하지만 구현 freeze 전 아래 세 경계가 필요하다.

## P1: 정책 해석과 최초 cursor
초안 :7–11의 strict 공통 3/s·보호대기·별도 프로필·top·25초/U2 late·같은 날 abort 금지는 기존 결정을 보존한다. :18·24의 그룹 전체 429 즉시 abort는 기존 엔진 로그 감시보다 범위가 넓다. 도구 429는 원래 측정 제외였지만 모든 같은 group 타 프로세스 429까지 즉시 abort할지는 기존 문구만으로 자동 확정하지 말고 더 보수적인 상세안으로 명시해 Manager가 기존 사용자 결정과 맞는지 판정해야 한다. 누락/gap/오류에 대한 fail-closed는 안전한 상세 방향이나 과거 손실을 baseline에서 무시할지와 abort 지속성 선택은 남는다.

`ObservationCursor{}`는 event/gap 0부터 읽고 오래된 ring이 이미 잘렸으면 overrun을 보여준다(`internal/accountbudget/observation_linux.go:625-642,675-695`). 그대로 현재 cursor만 받아 과거 gap/overrun을 버리는 것은 완전 이력 확인이 아니다. 최소한 시작 전 사건 제외 정책과, 과거 gap/overrun·Busy/Orphan 중 어떤 조건을 허용하는지, cursor 기록/같은 날 abort 원장의 atomic 실패·동시실행·restart를 별도로 명세하라. Busy가 현재 wire owner가 있다는 관측이지 엔진 정체성이나 건강 증명은 아니다.

## P1: 현재 API만으로 fence 완료가 아님
`accountbudgetruntime/runtime_linux.go:43-49,85-102`은 validated observed budget을 소유하나 소비자에게 동일 store Snapshot 경계를 제공하지 않는다. `transport_linux.go:215-244`는 Acquire→Begin→poll hint만 하며 a169 event fence는 없다. 외부 watcher가 Snapshot한 다음 요청을 내는 조합은 다른 process Finish와 경합하므로 그것만으로 선형화되지 않는다.

최소 설계 후보: factory가 소유한 같은 observed store에 대한 read-only Snapshot을 제공하고, a169 전용 context가 지정한 기준 cursor를 private transport가 **같은 wire lease를 보유한 상태에서** 검증한다. Begin은 이전 orphan을 gap으로 전환하므로(`observation_linux.go:440-470`) Begin 이후 gap을 포함한 사건/overrun과 동일 instance를 확인한 다음 첫 Write 전에 거절한다. 이 시점 자기 Begin pending은 Busy이며 sequence가 자기 것임을 확인해 정상 처리해야 한다. Snapshot/Begin 오류 또는 자신 외 불일치·과거 누락은 보수적으로 전송하지 않고 Begin 성공에 한해 Finish 및 Release 오류를 보존한다. 모든 토큰·401 시도에도 이 fence context가 이어져야 한다. 기존 Snapshot은 자신의 wire lock을 Busy로 나타내므로 외부에서 Busy만 보고 구별할 수 없고, transport가 보유한 sequence와 실제 lease를 결속하는 좁은 내부 계약이 필요하다. 이는 기존 부품으로 설계 가능한 후보이지 현재 API가 이미 제공하는 기능이나 증명된 구현이 아니다.

위 주장 전제는 모든 관련 writer가 동일 pinned store와 wire lock을 지키고 Finish 기록 또는 실패의 pending/dirty를 유지한다는 것이다. 다른 비협조 writer·임의 조작은 제외하고, 다른 writer의 Finish 이후 다음 a169 Write 0, 자기 Begin의 orphan→gap, own Finish 오류, 401/token context, 동시 cancel과 실제 첫 Write를 fixture로 증명해야 한다. 이미 시작된 wire를 소급 취소하지 못하며 429 발생 이후 관측·기록 전까지도 절대 wire0을 보장하지 않는다. watcher 자체 실패와 prewire fence의 역할도 따로 고정해야 한다.

## P1: 활성 엔진/프로필 증명은 별도
`internal/enginelock/enginelock.go:5,27,80-82`은 flock이 배타 수단이며 소유자의 신원을 읽는 신호가 아님을 설명한다. `cmd/tossctl/engineproc.go:196-199,261-264`의 marker는 advisory 확인이다. 디렉터리의 engine.lock/engine.log/credential 존재, marker 신선도, 동일 budget group 중 어느 하나도 실제 실행 엔진이 그 특정 profile 및 v2 manifest를 사용한다고 증명하지 않는다. marker와 lock 점유를 결합해도 PID/instance·lock/경로 재사용·실제 launch manifest 결속 및 죽은 뒤 전환 경합에 대한 신뢰 경계가 필요하다. 현재 source만으로 신원/동일 예산 결속을 완료했다고 결론낼 수 없다.

최소 후보는 운영자가 확인한 단일 profile/launch 배포 영수증과 실행 중 engine이 생성한 검증 가능한 instance·예산 신원 연결을 대조하고, 수집 중 신선한 liveness·instance 유지 실패를 abort하는 것이다. 단순 자기 주장 marker는 authority가 아니며 현재 그런 결속의 구현·신뢰성은 미확인이다. 필요한 소스 변경 여부와 liveness 조건은 설계에서 확정해야 하고, 증거 부재 시 측정 보류가 기존 정책에 부합한다. 실 운영 확인/리허설은 별도 사람 단계다.

## 추가
25초 72/70 및 250ms+350ms 조건부 41/39는 낙관적 상한일 뿐이고, token/경쟁·거절·응답/저장 시간으로 성공 보장 0인 설명은 적절하다. U2 실패 전체 late와 선정 시각을 그대로 보존한다. 재개 승인·운영 허가는 여기서 발생하지 않는다.
