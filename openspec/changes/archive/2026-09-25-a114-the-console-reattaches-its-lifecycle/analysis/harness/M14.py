import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/console_lifecycle_attach.go','''	client, seat, wanted := a.current()
	if wanted {
		a.wake()
	}
	return client, seat''','''	client, seat, _ := a.current()
	return client, seat''')
