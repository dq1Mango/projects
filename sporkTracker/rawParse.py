with open('./badData.md', 'r') as file:
    # Read all lines of the file into a list
    badlines = file.read().split("·")

messages = []
for line in badlines:
    messages.append(line.strip())

rawData = []
for message in messages:
    message = message.split("\n")
    try:
        rawData.append({
            "date": message[0],
            "content": message[1]
        })
    except:
        pass

file = open("./data.md", "w")
file.write(str(rawData).replace("},", "},\n").replace("\'", "\""))
