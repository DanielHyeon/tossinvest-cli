# TossOS SDD Control Graph

관측 전용 TypeDB schema다. 로컬 `stockos-sdd-typedb` 서비스를 공유할 수 있지만
database는 `tossos_sdd`를 사용한다. 연결 실패는 개발·테스트·커밋을 차단하지 않는다.

```bash
.sdd/.venv/bin/python tools/sdd-history/ingest_typedb.py
```

기본 endpoint는 `localhost:1729`다. 자격증명은 `TYPEDB_USER`,
`TYPEDB_PASSWORD` 환경변수로만 전달한다.
