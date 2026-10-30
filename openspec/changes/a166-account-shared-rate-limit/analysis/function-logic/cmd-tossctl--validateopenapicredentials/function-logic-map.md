# Function Logic Map: validateopenapicredentials

기준 5762b161b. AST는 ast.json. CodeGraph 격리 worktree 인덱스 부재; 원본 손읽기.

Accounts 성공은 OK; 인증/IP/429/서버/transport 실패는 기존 probeResult 종류로 분류. factory 오류는 실패로 반환한다.

OFF 기본 동작과 context·credential/token path·secret redaction을 유지한다.
