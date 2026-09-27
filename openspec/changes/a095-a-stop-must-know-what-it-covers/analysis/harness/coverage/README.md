# a095 3판 커버리지 프로파일 (대상 파일만 남김)

- 측정 HEAD: obs · journal · app/engine = `4fd400ea`, reconcile = `02716357` (두 HEAD 사이에 대상 소스 무변화 —
  각 번들의 `source_sha256` 이 base `02716357` 의 파일과 일치)
- 명령: `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=<pkg>.out` (패키지별 한 번)
- 원본 프로파일에서 FLM 대상 소스 파일의 줄만 남겼음(`mode:` 머리줄 유지). 「진입 실측」 재현:
  `python3 ../render_bundles.py <ast 디렉터리> obs.out journal.out app-engine.out reconcile.out`
