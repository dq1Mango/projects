from time import sleep
import json

#with open('./data.md', 'r') as file:
    # Read all lines of the file into a list
#    rawData = file.read().split("\n")

file = open("./data.md", "r")
rawData = file.read()[1:-2]
rawData = rawData.split(",\n")
for line in range(len(rawData)):
    rawData[line] = json.loads(rawData[line])

file2 = open("./manualIntervention.md", "r")
events = json.loads(file2.read())

def pareseFluff(sporkee):
    sporkee = sporkee.replace(" got ", "")
    sporkee = sporkee.replace(".", "")
    sporkee = sporkee.replace(" has ", "")
    sporkee = sporkee.strip()
    return sporkee

def formatDate(date):
    date = date.split(" ")
    month = date[0]
    if month == "Sep":
        month = 0
    elif month == "Oct":
        month = 1
    elif month == "Nov":
        month = 2
    elif month == "Dec":
        month = 3
    elif month == "Jan":
        month = 4
    elif month == "Feb":
        month = 5
    elif month == "Mar":
        month = 6
    else:
        month = 7 #hopefully i will have stopped caring by then
    return [month, date[1]]

def compareDates(date1, date2):
    date1 = formatDate(date1)
    date2 = formatDate(date2)
    if date1[0] != date2[0]:
        return date1[0] > date2[0]
    else:
        return date1[1] > date2[1]

def formatName(name):
    idk = name.split(" ")
    if len(idk) != 2:
       return name 
    return idk[0] + " " + idk[1][0]

alive = []
killers = {}
recentEntry = {}

for event in events:
    for person in event["content"]:
        person = person.lower()
        person = person.strip()
        person = formatName(person)
        if person not in recentEntry:
            recentEntry[person] = event["date"]
        elif compareDates(event["date"], recentEntry[person]):
            recentEntry[person] = event["date"]

#print(recentEntry)

for data in rawData:
    message = data["content"].lower()
    if "sporked" in message: #maybe add like "got out" here too if possible or u care
        if " by " not in message:
            message = message.split("sporked")
            
            sporkee = pareseFluff(message[0]) #UNREADABLE plz change, nice job :+1
            sporkee = formatName(sporkee)
            if sporkee not in alive:
                alive += [sporkee]
                killers[sporkee] = 1
            else:
                killers[sporkee] += 1
        else:
            message = message.split("sporked") 
            message[1] = message[1].split("by") #more readable, less bad +10 aura
            sporker = message[1][1]
            sporker = sporker.strip()
            sporker = sporker.replace(".", "")
            sporker = formatName(sporker)
            if sporker not in alive:
                alive += [sporker]
                killers[sporker] = 1
            else:
                killers[sporker] += 1

killers = {k: v for k, v in sorted(killers.items(), key=lambda item: item[1], reverse=True)}

for data in rawData:
    message = data["content"].lower()
    if "sporked" in message: #maybe add like "got out" here too if possible or u care
        
        message = message.split("sporked")
        if " by " not in message[1]:
            sporkee =sporkee = message[1]
            sporkee =sporkee.replace(".", "")
            sporkee =sporkee.strip()
            
        else:
            sporkee = message[0]
            sporkee =sporkee.replace(" got ", "")
            sporkee =sporkee.replace(".", "")
            sporkee =sporkee.strip()
        
        sporkee = formatName(sporkee)
        #print(sporkee) #edge case finder
        if sporkee in recentEntry:
            if compareDates(data["date"], recentEntry[sporkee]):
                if sporkee in alive:
                    alive.remove(sporkee)


        elif sporkee in alive:
            alive.remove(sporkee)

print(killers)
print(sorted(alive))
print(killers["joey h"])


    

