from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.common.actions.pointer_actions import PointerActions
from bs4 import BeautifulSoup
import re
from time import sleep
from random import uniform

browser = webdriver.Firefox()
browser.get('https://monkeytype.com/')

ActionChains(browser).move_to_element(browser.find_element(By.CLASS_NAME, 'rejectAll')).perform()
ActionChains(browser).click().perform()

ActionChains(browser).move_to_element(browser.find_elements(By.CLASS_NAME, 'textButton')[11]).perform()
ActionChains(browser).click().perform()

#wait = input("waiting to begin...")
#wpm = int(input('how fast?'))
sleep(3)
wpm = 178

soup = BeautifulSoup(browser.page_source, features="html.parser")
words = str(soup.find(id='words'))
words = words.split('<div class="word"')

delay = ((1 / (wpm * 6)) * 60)
textField = browser.find_element(By.ID, 'wordsInput')
for word in words:
    word = letters = re.findall('<letter>(.+?)</letter>', word)
    
    for letter in word:
        textField.send_keys(letter)
        #sleep(delay + uniform(-delay / 1.2, delay / 1.2))
        sleep(delay)
    
    textField.send_keys(Keys.SPACE)
    #sleep(delay + uniform(-delay / 1.2, delay / 1.2))
    sleep(delay)

#browser.quit()
