#!/bin/sh

if [ -z "$SCHEDULE_URL" ]; then
	echo "SCHEDULE_URL not defined"
	exit 1
fi

echo "Opening in browser..."
open $SCHEDULE_URL
echo "Log in to the website and paste the JavaScript code into the console."

cat scrape.js | pbcopy

printf "%s " "Press enter to continue."
read ans

go run main.go
