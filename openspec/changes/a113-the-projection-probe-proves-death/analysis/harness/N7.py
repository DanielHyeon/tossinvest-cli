import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
# 회수가 새 함수를 버리고 순수 probe 로 되돌아간다(배선 뮤테이션)
sub(d+'/transport_unix.go','''		if staleProjectionSocketAccepts(socketPath, socketInfo) {''','''		_ = socketInfo
		if projectionSocketAccepts(socketPath) {''')
