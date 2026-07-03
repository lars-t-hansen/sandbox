This is just to check the size of a pipe.  Per doc, on linux this is 64K, but ulimit -p reports 4KB
if it is read sensibly (8 items of 512 bytes each).  Looks like the doc is right.  And the -p thing
probably reports the largest atomic read/write.

Anyway, try `./writer 100000 1000` for a start.  It should stop printing after 65000.

