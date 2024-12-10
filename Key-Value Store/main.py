# Libraries imported
from flask import Flask, request, jsonify
import requests
import os
from requests.exceptions import ConnectionError

app = Flask(__name__)

# In-memory key-value store
key_value_store = {}

# Get the forwarding address from environment variable
FORWARDING_ADDRESS = os.getenv('FORWARDING_ADDRESS', '')

# PUT request at /kvs/<key> with JSON body {"value": <value>}
def kvs_put(key):
    if FORWARDING_ADDRESS:
        # Forward the PUT request if this is a forwarding instance
        url = f'http://{FORWARDING_ADDRESS}/kvs/{key}'
        try:
            response = requests.put(url, json=request.json)
            return response.content, response.status_code
        except ConnectionError:
            data = {'error': 'Cannot forward request'}
            return jsonify(data), 503

    # If the length of the key <key> is more than 50 characters, then return an error
    if len(key) > 50:
        data = {'error': 'Key is too long'}
        return jsonify(data), 400

    # If the request body is not a JSON object with key "value", then return an error
    if not request.json or 'value' not in request.json:
        data = {'error': 'PUT request does not specify a value'}
        return jsonify(data), 400

    # Extract the value from the JSON body
    value = request.json['value']

    # If the key <key> does not exist in the store, add a new mapping from <key> to <value> into the store
    if key not in key_value_store:
        key_value_store[key] = value
        data = {'result': 'created'}
        return jsonify(data), 201
    # Otherwise, if the <key> already exists in the store, update the mapping to point to the new <value>
    else:
        key_value_store[key] = value
        data = {'result': 'replaced'}
        return jsonify(data), 200

# GET request at /kvs/<key>
def kvs_get(key):
    if FORWARDING_ADDRESS:
        # Forward the GET request if this is a forwarding instance
        url = f'http://{FORWARDING_ADDRESS}/kvs/{key}'
        try:
            response = requests.get(url)
            return response.content, response.status_code
        except ConnectionError:
            data = {'error': 'Cannot forward request'}
            return jsonify(data), 503

    # If the key <key> exists in the store, return the mapped value in the response
    if key in key_value_store:
        data = {'result': 'found', 'value': key_value_store[key]}
        return jsonify(data), 200
    # If the key does not exist in the store, return an error
    else:
        data = {'error': 'Key does not exist'}
        return jsonify(data), 404

# DELETE request at /kvs/<key>
def kvs_delete(key):
    if FORWARDING_ADDRESS:
        try:
            # Forward the DELETE request if this is a forwarding instance
            url = f'http://{FORWARDING_ADDRESS}/kvs/{key}'
            response = requests.delete(url)
            return jsonify(response.json()), response.status_code
        except ConnectionError:
            data = {'error': 'Cannot forward request'}
            return jsonify(data), 503

    # If the key <key> exists in the store, remove it
    if key in key_value_store:
        del key_value_store[key]
        data = {'result': 'deleted'}
        return jsonify(data), 200
    # If the key <key> does not exist in the store, return an error
    else:
        data = {'error': 'Key does not exist'}
        return jsonify(data), 404

# Main route for each method
@app.route('/kvs/<key>', methods=['PUT', 'GET', 'DELETE'])
def kvs_requests(key):
    if request.method == 'PUT':
        return kvs_put(key)
    elif request.method == 'GET':
        return kvs_get(key)
    elif request.method == 'DELETE':
        return kvs_delete(key)
    