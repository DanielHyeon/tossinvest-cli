import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/transport_probe_unix.go','''		return !errors.Is(err, os.ErrNotExist)
	}
	if dead, answered := sameVerifiedSocket(socketPath, before); !answered {
		return !dead
	}
	return projectionSocketAccepts(socketPath)''','''		return !errors.Is(err, os.ErrNotExist)
	}
	return projectionSocketAccepts(socketPath)''')
