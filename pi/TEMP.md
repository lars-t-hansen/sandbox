We want

- a cronjob that runs temperature every n minutes (probably about every 5)
- to capture that output in a file along with a timestamp
- the file must be suitable for gnuplot or other plotting sw, or for importing into spreadsheet
- probably time along x and temperature along y
- what can google sheets do?
  - it can take a timestamp 10:05 and a value, and plot it
  - probably we don't care about the date, so long as we can pin the start of a plot to a particular date
