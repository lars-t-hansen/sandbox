default:
	@echo "Select a target, see Makefile"

# NOTE! The programs hardcode some decisions.

cpu_time.csv: cpu_time.txt cpu_time.go util.go
	go run cpu_time.go util.go cpu_time.txt > cpu_time.csv

gpu_time.csv: gpu_time.txt gpu_time.go util.go
	go run gpu_time.go util.go gpu_time.txt > gpu_time.csv

card_time.csv: card_time.txt card_time.go util.go
	go run card_time.go util.go card_time.txt > card_time.csv

# NOTE!  Time ranges are hardcoded in the sql scripts

cpu_time.txt: cpu_time.sql
	psql -f cpu_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > cpu_time.txt

gpu_time.txt: gpu_time.sql
	psql -f gpu_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > gpu_time.txt

card_time.txt: card_time.sql
	psql -f card_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > card_time.txt
