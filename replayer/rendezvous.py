from flask import request

def main():
    # curl URL -d "time=10&date=Sunday"
    #date = request.form.get('date')
    #time = request.form.get('time')

    #if date == None:
    #    date = request.args.get('date')     # query appended to URL
    #if time == None:
    #    time = request.args.get('time')
    time = request.values.get('time')
    date = request.values.get('date')

    r = "We'll meet at {} on {}.\n".format(time, date)
    return r
