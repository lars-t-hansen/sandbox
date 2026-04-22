default:
	@echo "Select a target, see Makefile"

# NOTE! cpu_time.go hardcodes some decisions.

cpu_time.csv: cpu_time.txt cpu_time.go
	go run cpu_time.go cpu_time.txt > cpu_time.csv

# NOTE!  Time ranges are hardcoded in the sql script
# NOTE!  The output file must be edited to remove junk at the beginning and end

cpu_time.txt: cpu_time.sql
	psql -f cpu_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > cpu_time.txt
