# proposal-freeze 리뷰 2회차(마지막) — codex (원문)

- 보이스: codex-cli 0.165.0, `codex exec -s read-only -C <worktree> --ephemeral --color never -o <파일> "<프롬프트>" </dev/null`, rc 0, 2026-11-23.
- 기준: HEAD `119d7f171` + Manager 결정 반영 문서(미커밋). 원문 = 마지막 메시지(`-o`) 그대로, 전체 도구 로그는 커밋하지 않음. 리뷰 뒤 change 디렉터리 밖 변경 0.

## 프롬프트

```text
TossOS change a196 의 proposal-freeze 리뷰 2회차(마지막)다. 저장소 루트는 현재 디렉터리(~/tossos-work/a196-impl). 파일을 수정하지 말 것(읽기 전용).

대상: openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/proposal.md, design.md, specs/engine-safety/spec.md(정본 openspec/specs/engine-safety/spec.md 의 같은 Requirement 를 MODIFIED), tasks.md, analysis/spec-edit-intent.md(정본 대조 diff 와 의도한 편집 16 목록). 1회차 원문: openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/analysis/review/proposal-freeze-r1-codex.md. 1단계 증거: openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/analysis/cited-ast/README.md, analysis/code-context/evidence-reconciliation.md, analysis/function-logic/*/function-logic-map.md.
관련 코드: internal/app/engine/attestation_watch.go, internal/app/engine/a135_attestation_watch_internal_test.go, internal/soak/attest.go, cmd/tossctl/soak.go, internal/verifylive/endpoints.go, internal/obs/notifier.go, internal/journal/outbox.go, internal/obs/ntfy.go.

1회차 발견과 Manager 처분(2026-11-23):
- R1(P1, 실린 증거가 endpoint 당 첫 유효 증거 하나라 두 시장 기록이면 기한이 이르다, 정각 경계): Q1 — 이른 경보를 받아들이고 발급자는 고치지 않음. 기한을 「파일에 실린 증거」로 정의, spec 문장을 「그 시각부터 갱신은 그 증거를 싣지 못하며, 다른 유효한 감독 증거가 없으면 POST endpoint 가 빠진다」로, D4 를 「실린 증거가 같은 한 바뀌지 않는다, 바뀌면 새로 센다」로, design §6 잔여에 기록. 정각 = 이미 지남, D2 시험에 V−1ns·V·V+1ns 세 행.
- R2(P1, 기한 경과 뒤 갱신 실패 시 파일이 만료까지 통과, daysLeftCeil 재사용 시 약 48시간 침묵): Q4 — 남은 시간 ≤ 0 은 새 단계 proof_overdue, 단계+기한 시각+KST 날짜 key 로 하루 한 번 critical, daysLeftCeil 은 양수에만. spec 의 「D-2 이하」를 「D-2 와 D-1」로 되돌리고 하루 한 번 critical 문장 + Scenario 추가.
- R3(P2, 기존 Scenario 「만료 8일 전」 등이 새 본문과 모순): Q2 — 4개 Scenario 의 WHEN/THEN 을 「기한」 기준으로 고침. 이름은 strict validator 가 정본 Scenario 이름 누락을 거절해서 그대로 둠. 중복 「감독 증거의 기한 2일 전」은 「만료 2일 전과 1일 전」 THEN 으로 합침.
- R4(P2, 재시작 뒤 같은 단계 재알림): Q6 — 정본에 「보냄 상태는 엔진 프로세스 수명 안에서만 유지되므로, 재시작 뒤에는 현재 단계를 한 번 더 알릴 수 있다」 추가.
- R5(P2, §5-11 영향 예상 오류·변이 누락): Q5 — 시험 함수 TestA135StageBoundaries(표)·TestA135SameStageOnANewAttestationIsAnnouncedAgain(재작성)만 고치고 헬퍼는 무편집, 「통과 이유가 바뀐 시험」 목록을 §5-12 에, 변이 6 → 12.
- 별도 Q3: 결과 문장을 제목과 본문 첫 문장 둘 다에(기기 제목 표시 미실측).

할 일: (1) 위 처분이 틀렸거나 문서 반영이 처분과 어긋나면 근거(파일:줄)와 함께 반박하라. (2) 반영된 문서에서 새 결함을 P0/P1/P2 로 — 특히 proof_overdue 와 다른 단계의 전이·key 충돌, proof_overdue 와 critical() 판정·알림 꺼진 엔진, 「이름은 만료, 내용은 기한」인 Scenario 의 모순, spec-edit-intent 의 목록이 실제 diff 와 일치하는지, 변이 12개가 각각 design §5 의 어느 시험에 죽는지. 현재 소스로 직접 확인하고, 추측과 확인한 사실을 구분하라. 새 기능 제안은 하지 말고 결함만.
```

## 응답

**P0 0건, P1 0건, P2 3건. 현 문서 그대로 freeze하기에는 계약 모순과 검증 공백이 남습니다.**

기준: HEAD `119d7f171211685645d2390a1f08278e31d43570` + 현재 미커밋 문서. 아래는 원문 대조와 정적 반례이며, a196 구현·시험 실행 결과가 아닙니다.

1. **[P2] 같은 날 재발급 Scenario가 정상적인 날수 단계 전이까지 금지합니다.**

   [delta:65–67](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:65)은 같은 증거로 재발급되면 “그날 알림은 더 나가지 않는다”고 요구합니다. [design §5-2:40](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/design.md:40)도 같은 표현입니다. 그러나 [delta:11](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:11)의 억제 단위는 **같은 기한의 같은 날수 단계**입니다.

   **정적 반례:** 00:00 KST에 기한까지 `5일 30분`이면 올림 결과는 D-6입니다. 06:00에 같은 증거·같은 유효 기간으로 재발급되면 기한은 그대로지만 잔여 시간은 `4일 18시간 30분`, 즉 D-5입니다. D4와 본문은 새 D-5 알림을 요구하지만 Scenario는 그날 추가 알림을 금지합니다. 계산 근거는 [daysLeftCeil:125](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:125)입니다.

   재발급에 따른 중복 억제와 시간 경과에 따른 새 단계 알림이 계약에서 구분되지 않았습니다. **실제 구현의 누락을 확인한 것은 아닙니다.**

2. **[P2] Q1의 대체 증거 조건이 D7 알림 문구에는 반영되지 않았습니다.**

   [design D7:13](~/tossos-work/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/design.md:13)의 경로가 아니라, 현재 대상인 [design D7:13](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/design.md:13)은 다음을 여전히 단정합니다.

   - 날수 통지: “그 뒤 첫 갱신부터 엔진 재시작이 거부된다”
   - `proof_overdue`: “다음 갱신부터 엔진 재시작이 거부된다”

   반면 수정된 [delta:7](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:7)은 **다른 유효한 감독 증거가 없을 때** POST가 빠진다고 정확히 한정합니다.

   **확인 사실:** 발급자는 기간을 벗어난 증거를 제외한 다음, endpoint별 첫 유효 증거를 선택합니다([acceptSupervised:331–340](~/tossos-work/a196-impl/internal/soak/attest.go:331)). 따라서 실린 A가 만료해도 기록에 유효한 B가 있으면 다음 성공한 갱신은 POST를 유지합니다. 갱신 전 `proof_overdue` 상태에서도 이 반례가 성립합니다.

   **이른 경보를 수용한 Q1 자체를 반박하지 않습니다.** 다만 확정된 증거 기한과 조건부인 재시작 거부를 D7이 다시 동일시합니다. Q1 반영이 문서 내부에서 일관되지 않습니다.

3. **[P2] 변이 12개를 선언했지만, 실제 outbox key 결함과 일부 경계 변이를 죽일 단언이 빠져 있습니다.**

   [design §5:39–55](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/design.md:39)의 시험 계획과 변이를 연결하면 다음과 같습니다. 모두 **계획의 검출 가능성 평가**이며 CAUGHT 실행 판정은 아닙니다.

   | 변이 | 연결된 시험 | 정적 검토 |
   |---|---|---|
   | ① 만료만 계산 | §5-1 | 감독 기한 D-7인데 만료 D-30인 입력으로 구별 가능 |
   | ② min→max | §5-1 | 같은 입력으로 구별 가능 |
   | ③ 날수 신원에 발급 시각 | §5-2 | 같은 기한·같은 단계에서 재발급하면 구별 가능 |
   | ④ 거절 key에 발급 신원 | §5-6 | **outbox key만 변형하면 검출 불명확.** 메모리 억제가 유지되면 새 이벤트 자체가 없어 건수 시험은 통과 |
   | ⑤ 누락 집합 제거 | §5-6 | 메모리 key 변형은 검출 가능. **outbox key만 변형하면 서로 다른 Key라는 단언이 필요하지만 계획에 없음** |
   | ⑥ D-2 경계 이동 | §5-3·11 | 수정된 경계 표까지 포함하면 구별 가능 |
   | ⑦ `>=`→`>` | §5-10 | 정확히 V인 행으로 구별 가능 |
   | ⑧ 가장 늦은 증거 사용 | §5-10 또는 기존 binding 시험 | 기존 [binding 시험:518](~/tossos-work/a196-impl/internal/app/engine/a135_attestation_watch_internal_test.go:518)은 binding 자체 변형을 검출. **새 기한 계산에서 binding을 우회하는 변형을 죽일 서로 다른 endpoint 시각 입력은 §5에 명시되지 않음** |
   | ⑨ 0 이하를 imminent로 | §5-3 | 단계 자체와 D-1→경과 전이를 단언하면 검출 가능 |
   | ⑩ KST→UTC | §5-3·6 | **KST 날짜만 바뀌는 시각 쌍이 없음.** 기존 시험도 25시간 전진하여 양쪽 모두 날짜가 바뀜 |
   | ⑪ 단계 제거 | 대응 시험 없음 | **scope_lost→expired 전이·Key 비교가 없음.** 누락 집합이 남으면 단계만 제거해도 두 key가 여전히 다를 수 있어, 변이 설명 자체도 충돌을 보장하지 않음 |
   | ⑫ 누락 집합 정렬 삭제 | 대응 시험 없음 | **동일 집합의 순서 차이를 검증하지 않음.** 현재 `MissingEndpoints`가 이미 정렬하므로 추가 정렬 삭제는 동작상 동등할 수도 있음 |

   특히 [a135Notifier:149](~/tossos-work/a196-impl/internal/app/engine/a135_attestation_watch_internal_test.go:149)는 실제 outbox 중복 억제를 구현하지 않습니다. 실제 notifier는 공급된 `e.Key`를 그대로 사용하고([notifier:991](~/tossos-work/a196-impl/internal/obs/notifier.go:991)), 원장은 그 key로 기존 행을 찾습니다([outbox:289](~/tossos-work/a196-impl/internal/journal/outbox.go:289)). 따라서 **가짜 notifier의 이벤트 건수 증가만으로 서로 다른 상태가 실제 전달되는지 증명할 수 없습니다.**

처분 대조에서 그 밖에 확인한 결과는 다음과 같습니다.

- **Q4:** `remaining ≤ 0`을 별도 단계로 분리하고 양수에만 `daysLeftCeil`을 적용하는 설계는 R1의 약 48시간 침묵 반례를 해소합니다. `critical()`은 healthy·notice 외에는 true이고, `eventFor`는 `w.critical`도 확인하므로 `proof_overdue` 역시 알림 OFF에서 일반 등급으로 강등됩니다([critical:74](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:74), [eventFor:315](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:315)).
- **전이·재시도:** D5대로 단계·기한·KST 날짜를 key에 보존하면 `imminent → proof_overdue → scope_lost/expired`의 서로 다른 단계는 구분됩니다. 기록 실패 뒤에는 `sent`를 설정하지 않습니다([checkOnce:248](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:248)). **이 설계 자체가 영구 또는 하루 초과 침묵을 만든다는 새 반례는 확인하지 못했습니다.**
- **Q2·Q6:** 기존 Scenario의 WHEN/THEN 수정과 재시작 예외는 반영됐습니다. 이름에 “만료”가 남았다는 이유만으로 새 실행 조건 모순을 판정하지 않았습니다. Q6의 “알릴 수 있다”도 실제 outbox의 재알림 창과 충돌하지 않습니다.
- **정본 diff:** `spec-edit-intent.md`에 적힌 `awk` 방식으로 Requirement를 추출해 대조했습니다. **저장된 diff와 완전 일치**, 본문 6·Scenario 수정 4·추가 6, 총 16개이며 Scenario 11→17도 맞습니다.
- **Q5:** 기존 시험 두 함수의 수정 계획과 통과 이유 변경 목록은 반영됐습니다. 다만 변이 수를 늘린 것만으로 위 검증 공백이 해소되지는 않았습니다.

파일 수정·시험·validator·운영 명령 실행은 하지 않았습니다. 인용 AST 22개와 FLM AST 6개의 소스 SHA는 현재 파일과 모두 일치했습니다. AST·CodeGraph 재생성은 하지 않았습니다. ai-memory CLI 조회는 샌드박스 연결 제한으로 실패했습니다. `impeccable: N/A — 화면 변경 없는 읽기 전용 계약·소스 리뷰`.