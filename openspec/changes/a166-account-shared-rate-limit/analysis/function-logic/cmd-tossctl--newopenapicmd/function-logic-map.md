# Function Logic Map: newopenapicmd

기준 5762b161b. AST는 ast.json. CodeGraph 격리 worktree 인덱스 부재; 원본 손읽기.

test와 status의 검증 호출만 root factory로 연결한다. login 저장·logout·status 구조는 유지한다.

OFF 기본 동작과 context·credential/token path·secret redaction을 유지한다.
