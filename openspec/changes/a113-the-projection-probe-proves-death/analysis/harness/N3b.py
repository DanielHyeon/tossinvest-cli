import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/transport_probe_unix.go','''		return true
	}
	if dead, answered := sameVerifiedSocket(socketPath, before); !answered {
		return !dead
	}
	if err := os.Chmod''','''		return true
	}
	if err := os.Chmod''')
