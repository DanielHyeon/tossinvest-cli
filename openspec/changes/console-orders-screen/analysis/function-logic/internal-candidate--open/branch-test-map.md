# Branch Test Map: `Open`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

B1·B2·B3·B5·B7은 직접 테스트가 없다. B1·B3은 프로덕션 전용 경로이고(`internal/testenv`의 `TestFixedFSProberIsTestOnly`가 프로덕션이 테스트용 prober를 주입하지 못하게 막으므로, nil 경로가 곧 프로덕션 경로다), 나머지는 OS·드라이버 실패 경로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 경로 없이 열면 데이터 디렉터리 규칙으로 푼다 | 없음 — `cmd/tossctl`의 `openCandidateStore` 프로덕션 경로 | 아니오(사후 증거) | 미커버 |
| B2 | 데이터 디렉터리를 못 정함 | 없음 | 아니오(사후 증거) | 미커버 |
| B3 | prober 없이 열면 시스템 prober | 없음 — 프로덕션 전용 | 아니오(사후 증거) | 미커버 |
| B4 | 허용목록 밖 마운트는 거부하고 아무것도 만들지 않는다 | `TestTheStoreRefusesAFilesystemThatCannotPromiseAWrite` | 아니오(사후 증거) | yes |
| B5 | 디렉터리 생성 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B6 | busy timeout 미설정은 5초 | `openStore`를 쓰는 모든 저장소 테스트 | 아니오(사후 증거) | yes |
| B7 | 드라이버 open 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B8 | clock 미설정은 시스템 시계 | `openStore` | 아니오(사후 증거) | yes |
| B9 | cooling TTL 미설정은 기본값 | `TestAReEntryWithinTheCoolingTTLKeepsTheOriginalFirstSeenAt` | 아니오(사후 증거) | yes |
| B10 | staleness TTL 미설정은 기본값 | `TestACandidateNobodyCooledDoesNotStayActiveForever` | 기록됨(§1 2-a 실행 재현) | yes |
| B11 | 마이그레이션 실패는 핸들을 닫고 올린다 | `TestAStoreFromANewerBuildIsRefused` | 아니오(사후 증거) | yes |
