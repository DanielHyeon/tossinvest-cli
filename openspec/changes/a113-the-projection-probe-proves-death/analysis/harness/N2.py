import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/transport_probe_unix.go','''	if err := chmodStaleSocket(socketPath, 0o600); err != nil {
		return !errors.Is(err, os.ErrNotExist)
	}
''','')
