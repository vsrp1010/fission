from flask import request

def main():
    date = request.form.get('date')
    time = request.form.get('time')
    r = "We'll meet at {} on {}.\n".format(time, date)
    return r