# Function Logic Map: runconsole

기준 5762b161b. AST는 ast.json. CodeGraph 격리 worktree 인덱스 부재; 원본 손읽기.

console onboarding 기본 deps에 root를 전달하는 호출만 변경한다. signal/HTTP/worker lifecycle은 이 로트 범위 밖이다.

OFF 기본 동작과 context·credential/token path·secret redaction을 유지한다.
