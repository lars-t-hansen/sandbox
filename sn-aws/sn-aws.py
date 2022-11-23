# Task 1: make it work
# Task 2: can we use the API and not go via the CLI?
# Task 3: can we detect authorization failure and provide a sensible error message or ask for authority?

# Usage:
#   sn up filename    // Filename can have path delimiters
#   sn down filename  // Filename can have no path delimiters
#   sn ls pattern     // Pattern can have no path delimiters, and is prefix-only?

import os
import re
import sys

def doit(cmd):
    sys.stdout.write(cmd + "\n")

def up():
    global AWS_PREFIX
    if len(sys.argv) < 3:
        usage()
    fn = sys.argv[2]
    assert_valid_filename(fn)
    assert_file_exists(fn)
    # FIXME: Should allow path name in the file, but should be reduced to base name in the
    # destination
    doit(f"aws s3 cp {fn} {AWS_PREFIX}{fn}")

def down():
    global AWS_PREFIX
    if len(sys.argv) < 3:
        usage()
    fn = sys.argv[2]
    assert_valid_filename(fn)
    assert_file_not_exists(fn)
    # FIXME: Should allow path name in the file, but should be reduced to the base name in
    # the source?  ie, sn down a/b/c gets "c" from the server and stores it locally at a/b/
    doit(f"aws s3 cp {AWS_PREFIX}{fn} .")

def ls():
    global AWS_PREFIX
    pattern = sys.argv[2] if len(sys.argv) >= 3 else ""
    doit(f"aws s3 ls {AWS_PREFIX}{pattern}")

def assert_valid_filename(fn):
    if not re.compile(r"^[-a-zA-Z0-9._]+$").match(fn):
        sys.stderr.write(f"Bad file name {fn}, must match /^[-a-zA-Z0-9._]+$/")
        sys.exit(1)

def assert_file_exists(fn):
    if not os.access(fn, os.R_OK):
        sys.stderr.write(f"File not found: `{fn}`\n")
        sys.exit(1)

def assert_file_not_exists(fn):
    if os.access(fn, os.F_OK):
        sys.stderr.write(f"File already exists: `{fn}`\n")
        sys.exit(1)

def usage():
    sys.stderr.write(
"""Usage: sn up filename
       sn down pattern
       sn ls [pattern]
""")
    sys.exit(1)

if not "AWS_SN_BUCKET" in os.environ:
    sys.stderr.write("No value for AWS_SN_BUCKET in environment")
    sys.exit(1)

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
