import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
# chmod 뒤 재확인의 결과를 버린다(post-review P1-1)
sub(d+'/transport_probe_unix.go','''		return !errors.Is(err, os.ErrNotExist)
	}
	if dead, answered := sameVerifiedSocket(socketPath, before); !answered {
		return !dead
	}''','''		return !errors.Is(err, os.ErrNotExist)
	}
	_, _ = sameVerifiedSocket(socketPath, before)''')
