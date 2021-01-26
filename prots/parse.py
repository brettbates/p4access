#!/usr/bin/env python2

"""
parse.py takes '-G' output from p4 and prints the dictionary
p4 -G info | ./parse.py
"""

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

