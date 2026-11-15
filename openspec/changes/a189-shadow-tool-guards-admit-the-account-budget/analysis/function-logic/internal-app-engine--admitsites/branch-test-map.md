# Branch Test Map: `admitSites` (internal/app/engine)

변이 원문: `../../mutation-2/mutation2-run5.txt`. RED = 편집 전 판본으로 `TestExactlyOneProductionSiteAdmitsIntoTheSeam` 이 `type 블록:Admit` 로 실패(경계 실측 `../../boundary/boundary.txt`, OLD 대조군).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 802 import 순회 | 전부 | yes | yes |
| B2 | 806 import 갈래 | 전부 | yes | yes |
| B3 | 807 dot import 표시 | `TestAdmitCensusSeesEverySpellingItClaims`(dottedCall) — 변이 g3 FAIL | yes(g3) | yes |
| B4 | 809 경계 경로(해독 뒤 끝 원소) | `TestAdmitCensusSeesEverySpellingItClaims`(example.com 경로) — 변이 g5 FAIL; 원시 문자열·이스케이프 철자는 `TestA189AdmitCensusDecodesTheImportPathLiteral` — 변이 g8 FAIL, a7·a8 FAIL | yes(g5) | yes |
| B5 | 811 별칭 | `TestAdmitCensusSeesEverySpellingItClaims`(sh) — 변이 g1 FAIL; 기본 이름 `TestExactlyOne…` — 변이 g7 FAIL | yes(g1·g7) | yes |
| B6 | 818 선언 순회 | 전부 | yes | yes |
| B7 | 822 type-switch | 전부 | yes | yes |
| B8 | 823 선택자 Sel 기록 | `TestA189AdmitCensusIsConservativeInADotImportedFile` — 변이 g4 FAIL | yes(g4) | yes |
| B9 | 825 경계 선택자 참조 | `TestExactlyOneProductionSiteAdmitsIntoTheSeam` — 변이 g6 FAIL, a1·a2·a4·a5·a6 FAIL(둘째 자리), b1·b2 PASS(무해) | yes | yes |
| B10 | 828 맨 식별자 갈래 | 전부(아래 B11 의 입구) | yes | yes |
| B11 | 829 dot import 파일의 맨 이름 | `TestExactlyOne…` — 변이 a3 FAIL; `TestA189AdmitCensusIsConservative…` | yes(a3) | yes |
| 좁힘 자체 | 앞 판본으로 되돌림 | `TestA189AdmitCensusIgnoresSameNamedMembersOutsideTheSeam` — 변이 g2 FAIL | yes(g2) | yes |
| 잃음 | l1(문 이름 필드 빼돌리기)·l3(래퍼 패키지) PASS — 앞 판본은 잡음; l3 은 `TestOnlyTheEngineImportsThisSeam` FAIL; l2 는 두 판본 다 못 잡음 | design R4 | — | — |
