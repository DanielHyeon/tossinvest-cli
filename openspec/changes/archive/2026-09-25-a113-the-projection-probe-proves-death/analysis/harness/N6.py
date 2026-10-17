import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/transport_unix.go','''		if staleProjectionSocketAccepts(socketPath, socketInfo) {''','''		fresh, _ := os.Lstat(socketPath)
		_ = socketInfo
		if staleProjectionSocketAccepts(socketPath, fresh) {''')
