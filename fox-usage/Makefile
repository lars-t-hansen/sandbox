default:
	@echo "Select a target, see Makefile"

# NOTE! process.go hardcodes some decisions.

cpu_time.csv: cpu_time.txt process.go
	go run process.go cpu_time.txt > cpu_time.csv

# NOTE!  Time ranges are hardcoded in the sql script

cpu_time.txt: cpu_time.sql
	psql -f cpu_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > cpu_time.txt
