# ScrapeSGNEMS
Scraping NEMS Prices


This repo is created to scrape SG NEMS data for reporting date. Sometimes, UIPath fails so we go Old School.

Steps (GoLang):
1) Change Date in Parameters.JSON
2) in powershell, input the following command

    ./main.exe


Steps (Python):
1) $pip install -r requirements.txt
2) Alter Parameters.json to desired date. Set the following:
    1) "fromDate"
    2) "toDate"
    3) "filepath"
3) $python TPCScrape.py

Dependencies:
1) Python 3.11. Probably works for earlier version. Not tested for anything outside of 3.11

TODO:
1) Containerisation
2) Implement test cases (PRIORITY !!!)
3) CI/CD

Version History:
v0.2.1: Added ChromeDriverUpdate.go to update Chromedriver
v0.1.1: TPC Scraping with GoLang. ChromeDriver preinstalled.