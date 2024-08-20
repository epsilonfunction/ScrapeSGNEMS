import os
import time
from urllib.parse import urlencode
import requests
import json


from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

# Not Needed for now; generally a good idea to keep around for error handling
# from selenium.webdriver.support import expected_conditions as EC
# from selenium.webdriver.common.keys import Keys
# from selenium.webdriver.common.action_chains import ActionChains
# from selenium.common.exceptions import TimeoutException
# from selenium.webdriver.common.desired_capabilities import DesiredCapabilities

nems_url = "https://www.nems.emcsg.com/nems-prices"

# Default Options; running in headless mode
driveroptions = Options()
driveroptions.add_argument("--headless")

driver = webdriver.Edge(options=driveroptions)
driver.get(nems_url) 
selenium_cookies = driver.get_cookies()
cookies = {cookie['name']: cookie['value'] for cookie in selenium_cookies}

# driver.wait(10) TODO: Determine if wait time or WebDriverWait is better
#TODO: Consider removing and hardcoding the form action url since it is porbably fixed.
form_elem = driver.find_element(By.XPATH, "//form[@action='/api/sitecore/DataSync/DataDownload']")

# Query Parameters
params = json.load(open("Parameters.json"))["Parameters"]
filepath = json.load(open("Parameters.json"))["filepath"]

form_action_url = form_elem.get_attribute('action')
full_url = f"{form_action_url}?{urlencode(params)}"

user_agent = driver.execute_script("return navigator.userAgent;")
headers = {
    'User-Agent': user_agent,
    'Referer': full_url,
}

response = requests.get(full_url, cookies=cookies, headers=headers)
# TODO: add exception to handle non 200 response

#Overwrites current data. Becareful
with open(filepath, "wb") as file:
    file.write(response.content)

driver.quit()