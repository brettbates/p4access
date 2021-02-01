#!/usr/bin/python3.9
import csv

if __name__ == '__main__':
    with open('private.csv') as test_file:
        test_reader = csv.reader(test_file, delimiter=',')
        for test in test_reader:
            print(' | '.join(test))