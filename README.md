# ScrapeSGNEMS
Scraping NEMS Prices


This repo is created to scrape SG NEMS data for reporting date. Sometimes, UIPath fails so we go Old School.

Steps:
1) $pip install -r requirements.txt
2) Alter Parameters.json to desired date. Set the following:
    1) "fromDate"
    2) "toDate"
    3) "filepath"
3) $python TPCScrape.py

Dependencies:
1) Python 3.11. Probably works for earlier version. Not tested for anything outside of 3.11

Future plans:
1) Rewrite in golang for ease of use
2) Containerisation
3) Implement test cases (PRIORITY !!!)
