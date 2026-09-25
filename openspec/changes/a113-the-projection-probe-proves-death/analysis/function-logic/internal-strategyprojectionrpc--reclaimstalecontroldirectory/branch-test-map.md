# Branch Test Map: `reclaimStaleControlDirectory`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 디렉터리 모드·symlink 불안전 | `TestStartRefusesUnsafeLeftoverShapes` 디렉터리 두 행 (a108) | no(회귀 핀) | yes |
| B2 | 소유 uid 불일치 | not-applicable: 비root 가 파일 소유자를 못 바꾼다 (a108 A1 F6 기록) | no(회귀 핀) | yes |
| B3 | ReadDir 실패 | not-applicable: 0700 우리 소유 디렉터리의 ReadDir 실패를 결정적으로 만들 수단 없음 | no(회귀 핀) | yes |
| B4 | 엔트리 순회 | `TestStartRecoversFromUnpublishedStagingLeftover` (a108) | no(회귀 핀) | yes |
| B5 | 이름 분류 switch | `TestStartRecoversFromUnpublishedStagingLeftover` · `TestStartRefusesControlDirectoryWithUnknownEntry` (a108) | no(회귀 핀) | yes |
| B6 | 최종 이름 → seen | `TestStartReclaimsExactDeadProjectionEndpoint` (a108) | no(회귀 핀) | yes |
| B7 | `.s-` staging 후보 | `TestStartRecoversFromUnpublishedStagingLeftover` (a108) | no(회귀 핀) | yes |
| B8 | staging 모양 이상 | `TestStartRefusesStagingLeftoverOfAnUnexpectedShape` (a108) | no(회귀 핀) | yes |
| B9 | 낯선 엔트리 | `TestStartRefusesControlDirectoryWithUnknownEntry` (a108) | no(회귀 핀) | yes |
| B10 | socket 이 있다 | `TestStartRecoversFromDeadSocketOnlyLeftover` (a108) | no(회귀 핀) | yes |
| B11 | socket 모양 불안전 | `TestStartRefusesUnsafeLeftoverShapes` socket 세 행 (a108) | no(회귀 핀) | yes |
| B12 | 산 주인 거부 — 쓰기 비트가 깎인 산 socket 포함, 거부 사유 "still alive"·보존·0600 복원 | `TestStartRefusesLiveProjectionOwnerWithoutRemovingIt` (a108) · `TestTheReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped` (a113) · 배선 핀 `TestTheReclaimHandsTheProbeTheInodeItVerified` (a113) | yes — base 에서 "치우면 안 되는 상태에서 기동이 받아들여졌다" | yes |
| B13 | descriptor 가 있다 | `TestStartRecoversFromDescriptorOnlyLeftover` (a108) | no(회귀 핀) | yes |
| B14 | descriptor 형식 불안전 | `TestStartRefusesUnsafeLeftoverShapes/descriptor_권한이_0600이_아니다` (a108) | no(회귀 핀) | yes |
| B15 | 제거 루프 | `TestStartRecoversFromUnpublishedStagingLeftover` (a108 — 회수 후 잔재 0) | no(회귀 핀) | yes |
| B16 | 파일 제거 실패 | not-applicable: 비root 로 결정적 재현 수단 없음 | no(회귀 핀) | yes |
| B17 | rmdir 실패 | not-applicable: 비root 로 결정적 재현 수단 없음 | no(회귀 핀) | yes |
