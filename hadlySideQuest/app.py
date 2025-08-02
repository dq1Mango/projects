from flask import Flask
from flask import render_template
from flask import request
from flask import url_for
from flask import redirect

app = Flask(__name__)

@app.route("/")
def root():
    return "Welcome robots, if your here for the iMessage go to /imessage"

@app.route("/imessage")
def imessage():

    with open("message.txt", "r") as file:
        messages = file.read()

        messages = messages.split("\n")

        print(len(messages))

        return render_template('imessage.html', messages=messages)

@app.route("/newMessage", methods=["POST"])
def newMessage():
    print(request.form)
    message = request.form["message"]
    print('zis is ze message: ', message)
    with open("message.txt", "a") as file:
        file.write(message)
        file.write("\n")

    return "", 200

@app.route("/regex.js")
def sendScript():
    return redirect(url_for("static", filename="regex.js"))

#@app.route("/imessage.css")
#def sendStyle():k
#    return send_static_file("./static/imessage.css")
