import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/console_lifecycle_attach.go','''	a.client, a.failed = client, false
	a.seat++
	announce := !a.attached''','''	a.failed = false
	a.seat++
	announce := !a.attached''')
