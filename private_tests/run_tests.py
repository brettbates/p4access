#!/usr/bin/python3.9
import csv
import re
import sys
import subprocess

P4PORT="localhost:1998" # Set to your local p4broker address
P4USER="perforce"

# run_test will run the command against the broker and capture the output
def run_test(user, reqAccess, path, expected):
    cmd = f"p4 -p localhost:1998 -u {user} access {reqAccess} {path}" 
    print(cmd)
    try:
        cres = subprocess.check_output(
            cmd, stderr=subprocess.STDOUT, shell=True)
    except subprocess.CalledProcessError as e:
        if expected != "ERROR":
            print(f"FAIL, expected ERROR from {cmd}")
            return False
        else:
            print(f"PASS, error captured as expected")
            return True
    # Split result into lines, ignore blank lines
    res = [x for x in cres.decode('utf-8').split('\n') if x != '']
    res = parse_result(res)
    checked = check_result(res, reqAccess, path, expected)
    if checked:
        print(f"PASS, correct result received")
        return True
    else:
        print(f"FAIL, unexpected result \n{cres.decode('utf-8')}")
        return False


# parser_result finds all the groups/access level's returned
# result order is preserved
def parse_result(res):
    parsed = []
    reg = re.compile(r'.*Group (.*) grants (.*) access to the path.*')
    for line in res:
        ms = reg.match(line)
        if ms:
            group = ms.group(1)
            perm = ms.group(2)
            parsed.append((group, perm))

    return parsed


def check_result(res, reqAccess, path, expected):
    if '&&' in expected:
        exp = expected.split('&&')
    else:
        exp = [expected]

    if expected == 'NONE':
        exp = []

    # Check we have the correct number of groups
    if len(exp) != len(res):
        print(f"FAIL, expected result of length {len(exp)}, got {len(res)}")
        print(f"Expected: {exp}")
        print(f"Result: {res}")
        return False

    for i, r in enumerate(res):
        # Check the group is correct and in correct order
        if r[0] != exp[i]:
            print(f"FAIL, expected r[0] {r[0]} == exp[{i}] {exp[i]}")
            return False
        # Check we are getting the reqAccess
        if r[1] != reqAccess:
            print(f"FAIL res {res}, expected {expected} to path {path}")
            print(f"reqAccess {reqAccess} != result {r[1]} received")
            return False
    return True


# For each line in the csv, run a test and check the output
if __name__ == '__main__':
    with open('private.csv') as test_file:
        test_reader = csv.reader(test_file, delimiter=',')
        for test in test_reader:
            print("-" * 80)
            print('')
            print(' | '.join(test))
            res = run_test(test[0], test[1], test[2], test[3])
            print('')
            print("-" * 80)
            if not res:
                sys.exit(1)