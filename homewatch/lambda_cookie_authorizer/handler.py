import base64
import json
import os
from typing import Any, Dict, Optional

import boto3

RedemptionCode = str
SignedJwt = str

CORS_HEADERS = {
    "Access-Control-Allow-Credentials": "true",
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Alllow-Methods": "GET,PUT,POST,OPTIONS,HEAD",
    "Access-Control-Allow-Headers": "Content-Type, Credentials, X-Amz-Date, Authorization, X-Api-Key, x-requested-with, Set-Cookie, set-cookie",
}
def cors_response(response):
    response["headers"] = CORS_HEADERS
    return response


class Unauthorized(BaseException):
    """403: Forbidden"""


def authorization(arn, allow=False):
    effect = "Deny"
    if allow:
        effect = "Allow"

    return {
        "PrincipalId": "user",
        "PolicyDocument": {
            "Version": "2012-10-17",
            "Statement": [
                {"Effect": effect, "Action": ["execute-api:Invoke"], "Resource": [arn]}
            ],
        },
    }


class DownstreamError(BaseException):
    def __init__(self, msg):
        super(BaseException).__init__()
        if isinstance(msg, DownstreamError):
            self.msg = msg.msg
        else:
            self.msg = msg


class RedemptionCreationError(DownstreamError):
    pass


class RedemptionVerificationError(DownstreamError):
    pass


class CookieCreationError(DownstreamError):
    pass


class CookieVerificationError(DownstreamError):
    pass


def invoke(lambda_client, fn: str, payload: bytes) -> Dict[Any, Any]:
    """
    Raises:
        When "errorType" is "errorString", raises an exception
    """
    print(f"Calling Function {fn}")
    result = lambda_client.invoke(
        FunctionName=fn, InvocationType="RequestResponse", Payload=payload
    )

    body = json.load(result.get("Payload"))
    if body is None:
        return
    if isinstance(body, str):
        return body
    if body.get("errorType") == "errorString":
        raise DownstreamError(body.get("errorMessage"))
    return body


def try_get_jwt(c=RedemptionCode) -> Optional[SignedJwt]:
    """
    Raises:
        RedemptionVerificationError: When the code can't be verified
        CookieCreationError: When cookie can't be created
    """
    verifier = os.environ.get("REDEMPTION_VERIFIER_NAME", "RedemptionVerifier")
    # Function name of the Lambda for verifying redemption codes
    creator = os.environ.get("COOKIE_CREATOR_NAME", "CookieCreator")
    lc = boto3.client("lambda")

    try:
        request = json.dumps({"QueryStringParameters": {"r": c}}).encode("utf-8")
        invoke(lc, verifier, request)
    except DownstreamError as err:
        raise RedemptionVerificationError(err)

    try:
        request = json.dumps(
            {
                "RedemptionCode": c,
            }
        ).encode("utf-8")
        cookie = invoke(lc, creator, request)
        return cookie.get("SignedCookie")
    except DownstreamError as err:
        raise CookieCreationError(err)


def try_verify_jwt(signed_cookie: Dict[str, str]):
    """
    Raises:
        When JWT is not authorized or verifiable
    """
    verifier = os.environ.get("COOKIE_VERIFIER_NAME", "CookieVerifier")
    lc = boto3.client("lambda")
    print(f"Verifing cookie {signed_cookie}")
    try:
        request = json.dumps(signed_cookie).encode("utf-8")
        invoke(lc, verifier, request)
    except DownstreamError as err:
        raise CookieVerificationError(err)


def try_send_authorization(email_address: str) -> bool:
    """
    Send a redemption url to the given emaill address
    Raises:
        DownstreamError when unable to contact a RedemptionCreator
    """
    creator = os.environ.get("REDEMPTION_CREATOR_NAME", "RedemptionCreator")

    # email_address = base64.decodebytes(email_address.encode("utf-8")).decode()
    print(f"Trying to notify {email_address}")

    if not "@" in email_address:
        print(f"Invalid email address {email_address}")
        return False
    lambda_client = boto3.client("lambda")

    response = None
    try:
        response = invoke(
            lambda_client,
            creator, # redemptionCreator
            json.dumps(
                {
                    "HomeId": "Web",
                    "HomeOwner": email_address,
                }
            ),
        )
        
    except DownstreamError as err:
        raise RedemptionCreationError(err)

    return response


def handle(event, context):
    # Where to redirect when authz succeeds
    redirect_url = os.environ.get("REDIRECT_URL", "/redirect.html")

    _ = context

    headers = event.get("multiValueHeaders", {})

    cookies = headers.get("Cookie", [])

    query = event.get("queryStringParameters", {})

    if query == None:
        query = {}
    redemption_code = query.get("r", None)

    if redemption_code:
        print(f"Verifying redemption code {redemption_code}")
        signed_jwt = None
        try:
            signed_jwt = try_get_jwt(redemption_code)
        except RedemptionVerificationError as err:
            print(f"Failed to verify redemption code {redemption_code}: {err}")
            return cors_response({
                "statusCode": 301,
                "headers": {
                    "Location": "/loginfailed.html",
                },
            })
        except CookieCreationError as err:
            print(f"Failed to create cookie for {redemption_code}: {err}")
            return cors_response({
                "statusCode": 301,
                "headers": {
                    "Location": "/loginfailed.html",
                },
            })
        print("Created Cookie")
        if signed_jwt:
            return cors_response({
                "statusCode": 200,
                # "cookies": [f"byst={signed_jwt}; Path=/; Secure; HttpOnly; SameSite=LAX;"],
                "body": "OK",
                "multiValueHeaders": {
                    "Set-Cookie": [f"byst={signed_jwt}; Path=/; Secure; HttpOnly; SameSite=LAX;"],
                },
                "headers": {
                    # "Location": "/home.html",
                    "content-type": ["text/html"],
                    # "set-cookie": [
                    #     f"byst={signed_jwt}; Path=/; Secure; HttpOnly; SameSite=LAX;",
                    #     "simple=cookie",
                    # ],
                },
            })
    email_address = query.get("email", None)
    print(f"Query: {query}")
    
    if email_address:
        email_address = email_address.replace("%40", "@")
        print(f"Trying to create login for {email_address}")
        ok = False
        try:
            redemption_url = try_send_authorization(email_address)
        except RedemptionCreationError as err:
            print(f"Failed to send authorization: {err}")
        except Exception as err:
            print(f"Failed to create login token: {err}")
            return cors_response({
                "statusCode": 301,
                "headers": {
                    "Location": "/loginfailed.html",
                },
            })
        if redemption_url:
            print(f"Sending redemption url {redemption_url}")
            return cors_response({
                "statusCode": 200,
                "body": json.dumps({"url": redemption_url}),
            })
            # return cors_response({
            #     "statusCode": 301,
            #     "headers": {
            #         "Location": redemption_url,
            #     }
            # })

    do_redirect = False
    for cookie in cookies:
        print(f"Inspecting cooke {cookie}")
        # TODO: Authenticate token
        if cookie == "Token=token=secureToken":
            return cors_response({"statusCode": 200})
        # TODO: Manage secret
        if cookie == "SignedNonce=nonce=abc123":
            do_redirect = True

    # TODO: Send out login URL
    if do_redirect:
        print(f"Redirecting to {redirect_url}")
        return cors_response({
            "statusCode": 301,
            "headers": {
                "Location": redirect_url,
            },
        })
    print("No nonce or token found")

    try:
        signed_cookie = event["SignedCookie"]
        try_verify_jwt(event)
        print(f"Verified JWT")
        return cors_response({"statusCode": 200})
    except CookieVerificationError:
        print(f"Unable to verify JWT")
    except Exception:
        print(f"No Signed JWT present in event: {event}")
    # No signed nonce or token are present
    return cors_response({"statusCode": 403})
