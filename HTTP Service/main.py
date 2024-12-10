#libraries imported
from flask import Flask
from flask import request, jsonify

app = Flask(__name__)

#/hello route
@app.route('/hello', methods=['GET', 'POST'])
def hello():
    #GET request (with no parameter) 
    if request.method == 'GET':
        data = {'message': 'world'}
        #returns the JSON response body{"message":"world"} and status code 200
        return jsonify(data), 200
    #POST request
    else:
        data = 'Method Not Allowed'
        #returns the JSON response 'Method Not Allowed' and status code 405
        return jsonify(data), 405

#/hello/<name> route
@app.route('/hello/<name>', methods=['GET', 'POST'])
def hello_name(name):
    #POST request with the path-parameter "name"
    if request.method == 'POST':
        data = {'message': f'Hi, {name}.'}
        #returns the JSON response body {"message":"Hi, <name>."} and status code 200
        return jsonify(data), 200
    #GET request
    else:
        data = 'Method Not Allowed'
        #returns the JSON response 'Method Not Allowed' and status code 405
        return jsonify(data), 405

#/test route
@app.route('/test', methods=['GET', 'POST'])
def test():
    #GET request with no query parameters.
    if request.method == 'GET':
        data = {"message":"test is successful"}
        #returns the JSON body {"message":"test is successful"} and status code 200
        return jsonify(data), 200
    
    #POST method
    elif request.method == 'POST':

        #<msg> is the string passed to the msg query parameter
        msg = request.args.get('msg')

        #POST request with a msg query parameter. 
        if msg:
            data = {'message': msg}
            #returns the JSON body {"message":"<msg>"} and status code 200
            return jsonify(data), 200
        
        #POST request with no msg query parameter
        else:
            data = "Bad Request"
            #returns the JSON body {"Bad Request"} and status code 400
            return jsonify(data), 400