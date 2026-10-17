import sys,os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d=sys.argv[1]
exec(open(os.path.join(os.path.dirname(__file__),'N1.py')).read().split('d=sys.argv[1]')[1])
exec(open(os.path.join(os.path.dirname(__file__),'N2.py')).read().split('d=sys.argv[1]')[1])
