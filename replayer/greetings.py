from flask import request

def main():
    title = request.form.get('title')
    name = request.form.get('name')
    item = request.form.get('item')
    g = "Greetings, {} {}. May I take your {}?\n".format(title, name, item)
    return g