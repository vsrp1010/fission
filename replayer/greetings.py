from flask import request

def main():
    '''
    title = request.form.get('title')
    if title == None:
        title = request.args.get('title')
    name = request.form.get('name')
    if name == None:
        name = request.args.get('name')
    item = request.form.get('item')
    if item == None:
        item = request.args.get('item')
    '''
    content = request.get_json(force=True)
    title = content['title']
    name = content['name']
    item = content['item']
    g = "Greetings, {} {}. May I take your {}?\n".format(title, name, item)
    return g