from flask import Flask, jsonify, request, make_response, redirect
from flask import Response
from flask_cors import CORS
import requests
import json
import os

api = Flask("mock-api")
CORS(api)
if os.environ.get("AUTHORIZATION_BYPASS_CODE"):
    print("Bypassing authorization with faked bypass code")

AUTHORIZATION_BYPASS_CODE = os.environ.get("AUTHORIZATION_BYPASS_CODE")
GET_DATAPOINTS_URL=os.environ.get("GET_DATAPOINTS_URL")
GET_CAMERAS_URL=os.environ.get("GET_CAMERAS_URL")

def remote_integration(url, query={}) -> Response:
    print(f"Remote Integration: {url}")
    cookies = {"byst": AUTHORIZATION_BYPASS_CODE }
    api_response = requests.get(GET_DATAPOINTS_URL, cookies=cookies, params=query)
    return make_response(api_response.json(), api_response.status_code)

@api.route("/api/datapoints")
def get_datapoints():
    if GET_DATAPOINTS_URL and AUTHORIZATION_BYPASS_CODE:
        return remote_integration(GET_DATAPOINTS_URL, query=request.args)

    camera = request.args.get("camera")
    try:
        with open(f"api.get.datapoints.{camera}.json", "r") as data:
            return jsonify({"datapoints": json.load(data)})
    except Exception:
        return 500

@api.route("/api/cameras", methods=["GET"])
def get_cameras():
    if GET_CAMERAS_URL and AUTHORIZATION_BYPASS_CODE:
        return remote_integration(GET_CAMERAS_URL, query=request.args)

    camera = request.args.get("camera")
    try:
        with open(f"api.get.cameras.{camera}.json", "r") as data:
            return jsonify(json.load(data))
    except Exception:
        return 500

@api.route("/api/cameras", methods=["PUT"])
def update_camera_enabled():
    return 200

@api.route("/Events/<path:event_path>")
def get_events(event_path):
    camera = event_path.split("/")[0]
    print(f"Camera: {camera} Rest: {event_path}")
    try:
        with open(f"api.get.events.{camera}.json", "r") as data:
            return jsonify(json.load(data))
    except Exception:
        return 500

api.run()