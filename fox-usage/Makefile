default:
	@echo "Select a target, see Makefile"

# NOTE! cpu_time.go snd gpu_time.go hardcode some decisions.

cpu_time.csv: cpu_time.txt cpu_time.go
	go run cpu_time.go cpu_time.txt > cpu_time.csv

gpu_time.csv: gpu_time.txt gpu_time.go
	go run gpu_time.go gpu_time.txt > gpu_time.csv

# NOTE!  Time ranges are hardcoded in the sql scripts
# NOTE!  The output files must be edited to remove junk at the beginning and end

cpu_time.txt: cpu_time.sql
	psql -f cpu_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > cpu_time.txt

gpu_time.txt: gpu_time.sql
	psql -f gpu_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > gpu_time.txt
