# SneakerNet
#
# Usage:
#   sn up filename    // Filename can have path delimiters
#   sn down filename [dest] // Filename can have no path delimiters but dest can
#   sn ls pattern     // Pattern can have no path delimiters, and is prefix-only?

# (DONE) Task 1: make it work
# (DONE) Task 2: can we use the API and not go via the CLI?
# Task 3: can we detect authorization failure and provide a sensible error message or ask for authority?
#     The credentials can be passed explicitly when we create the s3 resource, we just have to
#     obtain them from somewhere
# Task 4: can we use standard command line parsing?

import boto3
import os
import re
import sys

s3 = boto3.resource('s3')

def up():
    if len(sys.argv) != 3:
        usage()
    fn = sys.argv[2]
    assert_valid_filename(fn)
    assert_file_exists(fn)
    # TODO: Should allow path name in the file, but should be reduced to base name in the
    # destination
    #
    # TODO: Could allow destination to be specified separately
    s3.Bucket(AWS_SN_BUCKET).upload_file(fn, 'TRANSIT/' + fn)

def down():
    if len(sys.argv) < 3 or len(sys.argv) > 4:
        usage()
    fn = sys.argv[2]
    dest = sys.argv[3] fn if len(sys.argv) == 4 else fn
    assert_valid_filename(fn)
    assert_file_not_exists(dest)
    # TODO: Should allow path name in the file, but should be reduced to the base name in
    # the source?  ie, sn down a/b/c gets "c" from the server and stores it locally at a/b/
    s3.Bucket(AWS_SN_BUCKET).download_file('TRANSIT/' + fn, fn)

def ls():
    if len(sys.argv) > 3:
        usage()
    pattern = sys.argv[2] if len(sys.argv) == 3 else ""
    for x in s3.Bucket(AWS_SN_BUCKET).objects.filter(Prefix="TRANSIT/"):
        print(x.key)

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

AWS_SN_BUCKET = os.environ['AWS_SN_BUCKET']

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
