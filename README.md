# Bebe Schedule

- Scrape schedule as JSON (deep inside several layers of `iframe`s and `frame`s)
- Convert JSON shifts to iCal events
- Import iCal file to Google Drive

## Requirements

- File `ratings.csv` - include rating for each unit (keeping mine a secret hehe)
- Environment variable `SCHEDULE_URL` - URL to online schedule

## Building

```bash
# or run `make`
go build -ldflags="-X 'main.scheduleURL=$SCHEDULE_URL'"
```
