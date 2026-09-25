import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/transport_unix.go','''	if errors.Is(err, unix.ECONNREFUSED) || errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}''','''	if errors.Is(err, unix.ECONNREFUSED) || errors.Is(err, os.ErrNotExist) {
		return false
	}
	if info, statErr := os.Lstat(socketPath); statErr == nil && info.Mode().Perm()&0o200 == 0 {
		return false
	}
	return true
}''')
