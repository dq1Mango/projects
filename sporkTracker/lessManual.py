import json

dontAccidentlyRun = True
if dontAccidentlyRun:
    import sys
    sys.exit()
file = open("./manualIntervention.md", "r")

events = json.loads(file.read())

#put the names of the people who got back in here
#names = "Michael Sutter, Behbod, Jayson D, Eric D, Grayson Smith, Colin Doran, Casper, Avery D, Kelly O, Hadley W, Thomas R, Chloe M, Katie V, Ben S, Phoebe M, Emma L, Ava Warren, Elsie Ely, Sabrina Cox"
names = "Cece Morrow, Emma Lefebvre, Annabella Phelps, Franny Sgourakis, Greg Lazarro, Bryce Beamon, Noah Patrick, Isaiah Irwin Evans, AJ Hyland, Mackenna Pozza, John Peters, Kaden Lane, Andrew Hartung, Yousuf Khan, Ava Pessacreta, and Liam Dwyer"
#put the uh date they got back in here
date = "Sep 20"

names = names.split(",")

events.append({"date": date, "content": names})

file2 = open("./manualIntervention.md", "w")
file2.write(str(events).replace("\'", "\""))


