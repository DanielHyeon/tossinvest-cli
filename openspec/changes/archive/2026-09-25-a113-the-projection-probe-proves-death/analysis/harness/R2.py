import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
# chmod 앞 재확인의 결과를 버린다
sub(d+'/transport_probe_unix.go','''		return true
	}
	if dead, answered := sameVerifiedSocket(socketPath, before); !answered {
		return !dead
	}
	if err := chmodStaleSocket''','''		return true
	}
	_, _ = sameVerifiedSocket(socketPath, before)
	if err := chmodStaleSocket''')
