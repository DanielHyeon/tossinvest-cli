import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
# chmod 의 어떤 실패도 사망으로 읽는다(위험 방향)
sub(d+'/transport_probe_unix.go','''		return !errors.Is(err, os.ErrNotExist)''','''		return false''')
