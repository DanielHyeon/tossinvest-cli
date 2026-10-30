# Function Logic Map: cmd-tossctl--consolerelaunch

기준 bbe5066ec; ast.json은 수정 전 source SHA. CodeGraph 격리 worktree 인덱스 없음, 원본 손읽기.

현재 self path/argv 산출 후 바로 exec. owner.Close 성공 전 exec 금지, 오류 전파. 기존 argv·포트 유지.
