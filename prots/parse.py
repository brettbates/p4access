#!/usr/bin/env python2

# This is to see what the unmarshalled dict looks like
# p4 -G info | ./parse.py

import sys, marshal

try:
    num=0
    while 1:
        num=num+1
        print '\n--%d--' % num
        dict =  marshal.load(sys.stdin)
        for key in dict.keys():
            print "%s: %s" % (key,dict[key])

except EOFError:
    pass

