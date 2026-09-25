import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
# chmod 의 부재 오류도 생존으로 읽는다
sub(d+'/transport_probe_unix.go','''		return !errors.Is(err, os.ErrNotExist)''','''		return true''')
