bebe-sched: main.go ratings.csv scrape.js
	go build -ldflags="-s -w -X 'main.scheduleURL=$(SCHEDULE_URL)' -X 'main.exportURL=$(EXPORT_URL)'"

install:
	go install -ldflags="-s -w -X 'main.scheduleURL=$(SCHEDULE_URL)' -X 'main.exportURL=$(EXPORT_URL)'" .
