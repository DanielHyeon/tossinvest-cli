import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
# seam 의 운영 값을 바꾼다
sub(d+'/transport_probe_unix.go','''var chmodStaleSocket = os.Chmod''','''var chmodStaleSocket = func(string, os.FileMode) error { return nil }''')
