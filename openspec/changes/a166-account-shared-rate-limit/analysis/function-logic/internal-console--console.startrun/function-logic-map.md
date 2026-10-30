# Function Logic Map: internal-console--console.startrun

기준 bbe5066ec; ast.json은 수정 전 source SHA. CodeGraph 격리 worktree 인덱스 없음, 원본 손읽기.

spent면 거절, 진행 중이면 기존 run 반환; finishedFlag가 먼저 세워지는 동안 포인터를 교체할 수 있다. 이전 done 닫힘 전 교체 금지.
