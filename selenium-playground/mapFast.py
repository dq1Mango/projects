from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.common.actions.pointer_actions import PointerActions
from bs4 import BeautifulSoup
import re
from time import sleep

totalTime = 0

browser = webdriver.Firefox()
browser.get('https://www.geoguessr.com/seterra/en/vgp/3007')

#wait = input('waiting to begin ... ')
sleep(5)

def correctName(name):
    if name == 'United Kingdom': return 'AREA_THEUNITEDKINGDOM'
    elif name == 'Vatican City': return 'CITY_VATICAN'
    elif name == 'San Marino': return 'CITY_SANMARINO'
    elif name == 'Luxembourg': return 'AREA_LUXEMBURG'
    elif name == 'Norway': return 'rect12_32_'
    elif name == 'Sweden': return 'rect12_10_'
    elif name == 'Finland': return 'rect12_31_'
    elif name == 'Bosnia and Herzegovina': return 'AREA_BOSNIAANDHERZEGOVINA'
    elif name == 'Serbia': return 'AREA_SERBIA'
    elif name == 'Andorra': return 'CITY_ANDORRA'
    elif name == 'Denmark': return '_x3C_Path_x3E__7_'
    elif name == 'Liechtenstein': return 'C5'
    elif name == 'Greece': return 'AREA_GREECE'
    elif name == 'Italy': return 'rect12_17_'
    elif name == 'Czech Republic (Czechia)': return 'CHECH_REPUBLIC'
    elif name == 'Croatia': return 'C4'
    elif name == 'Monaco': return 'CITY_MONACO'
    elif name == 'North Macedonia': return 'MACEDONIA'
    elif name == 'Estonia': return 'AREA_ESTONIA'
    elif name == 'France': return 'AREA_FRANCE'
    elif name == 'Kosovo': return 'AREA_KOSOVO'
    else: return name.upper()


delay = 0
for i in range(46):
    soup = str(BeautifulSoup(browser.page_source, features="html.parser"))
    country = re.findall('Click on (.+?)<button', soup, 2)[0]
    country = correctName(country)
    ActionChains(browser).move_to_element(browser.find_element(By.ID, country)).perform()
    ActionChains(browser).click().perform()
    sleep(delay)
