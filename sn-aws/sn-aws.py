# Task 1: make it work
# Task 2: can we use the API and not go via the CLI?
# Task 3: can we detect authorization failure and provide a sensible error message or ask for authority?

# Usage:
#   sn up filename    // Filename can have path delimiters
#   sn down filename [dest] // Filename can have no path delimiters but dest can
#   sn ls pattern     // Pattern can have no path delimiters, and is prefix-only?

import os
import re
import sys

def doit(cmd):
    # sys.stdout.write(cmd + "\n")
    os.system(cmd)
    
def up():
    global AWS_PREFIX
    if len(sys.argv) < 3:
        usage()
    fn = sys.argv[2]
    assert_valid_filename(fn)
    assert_file_exists(fn)
    # FIXME: Should allow path name in the file, but should be reduced to base name in the
    # destination
    #
    # FIXME: Could allow destination to be specified separately
    doit(f"aws s3 cp {fn} {AWS_PREFIX}{fn}")

def down():
    global AWS_PREFIX
    if len(sys.argv) < 3:
        usage()
    fn = sys.argv[2]
    dest = fn if len(sys.argv) < 4 else sys.argv[3]
    assert_valid_filename(fn)
    assert_file_not_exists(dest)
    # FIXME: Should allow path name in the file, but should be reduced to the base name in
    # the source?  ie, sn down a/b/c gets "c" from the server and stores it locally at a/b/
    doit(f"aws s3 cp {AWS_PREFIX}{fn} {dest}")

def ls():
    global AWS_PREFIX
    pattern = sys.argv[2] if len(sys.argv) >= 3 else ""
    doit(f"aws s3 ls {AWS_PREFIX}{pattern}")

def assert_valid_filename(fn):
    if not re.compile(r"^[-a-zA-Z0-9._]+$").match(fn):
        fail(f"Bad file name {fn}, must match /^[-a-zA-Z0-9._]+$/")

def assert_file_exists(fn):
    if not os.access(fn, os.R_OK):
        fail(f"File not found: `{fn}`")

def assert_file_not_exists(fn):
    if os.access(fn, os.F_OK):
        fail(f"File already exists: `{fn}`")

def usage():
    sys.stderr.write(
"""Usage: sn up filename
       sn down filename [dest]
       sn ls [pattern]
""")
    sys.exit(1)

def fail(msg):
    sys.stderr.write(msg + "\n")
    sys.exit(1)

if not "AWS_SN_BUCKET" in os.environ:
    fail("No value for AWS_SN_BUCKET in environment")

AWS_PREFIX = f"s3://{os.environ['AWS_SN_BUCKET']}/TRANSIT/"

if len(sys.argv) < 2:
    usage()

if sys.argv[1] == "up":
    up()
elif sys.argv[1] == "ls":
    ls()
elif sys.argv[1] == "down":
    down()
else:
    usage()
