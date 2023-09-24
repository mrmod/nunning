# pylint: disable=logging-fstring-interpolation,unnecessary-pass,line-too-long,too-few-public-methods
"""
Required environment:

* VIDEOS_BUCKET = some-bucket-name
* VIDEOS_ROOT = some/video/root
"""
from concurrent.futures import ThreadPoolExecutor, as_completed
from math import ceil
import os
import json
from datetime import datetime, timedelta
from typing import Any, List, Optional, Dict
import logging
import boto3

log = logging.getLogger("ListVideos")
log.setLevel(logging.DEBUG)

REQUIRED_PARAMS = ["date", "period", "step", "camera"]
DATE_FORMAT = "%Y%m%d%H"
START_TIME_FORMAT = "%Y%m%d%H%M%S"
END_TIME_FORMAT = START_TIME_FORMAT

CORS_HEADERS = {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Alllow-Methods": "GET,PUT,POST,OPTIONS,HEAD",
    "Access-COntrol-Allow-Headers": "Content-Type, X-Amz-Date, Authorization, X-Api-Key, x-requested-with",
}


def cors_response(response):
    """
    Returns:
        API Gateway integration response with CORS headers
    """
    response["headers"] = CORS_HEADERS
    return response


def integration_response(
    status_code: int,
    message: Optional[str] = None,
    error: Optional[str] = None,
    json_data: Optional[str] = None,
) -> Dict:
    """
    Returns:
        API Gateway integration response
    """
    _response = {"statusCode": status_code}
    if message:
        _response["body"] = json.dumps(
            {
                "message": message,
            }
        )
        return cors_response(_response)

    if error:
        _response["body"] = json.dumps(
            {
                "error": error,
            }
        )
        return cors_response(_response)

    if json_data:
        _response["body"] = json_data
        return cors_response(_response)

    return cors_response(_response)


def is_valid_video(s3_object_reference):
    """
    Returns:
        True when the video is mp4 and larger than 0 bytes
    """
    return (
        s3_object_reference["Key"].endswith(".mp4") and s3_object_reference["Size"] > 0
    )


def build_prefix(list_videos_request: Dict) -> str:
    """
    Returns:
        $Root/$Camera/$Y-m-d/001/dav/$LocalHour
    """
    root = os.environ["VIDEOS_ROOT"]

    date = list_videos_request.get("date", "")
    period = list_videos_request.get("period")
    camera = list_videos_request.get("camera")

    # Event date is in UTC time
    event_date = datetime.strptime(f"{date}{period:0>2}", DATE_FORMAT)

    # Videos period and event date are in local time
    log.debug(f"UTC event date {event_date}")
    event_date = event_date - timedelta(hours=7)
    video_prefix_format = f"{root}/{camera}/%Y-%m-%d/001/dav/%H/"

    prefix = datetime.strftime(event_date, video_prefix_format)
    log.debug(f"Listing videos in {prefix}")
    return prefix


def list_videos(list_videos_request: Dict):
    """
    Args:
        list_videos_request: {"date": yyyymmdd, "period": localHour, "camera": "SomeCamera"}
    Returns:
        []S3Keys matching "*.mp4" bigger than 0 bytes
    """
    s3_client = boto3.client("s3")
    prefix = build_prefix(list_videos_request)

    videos = get_videos_list(s3_client, prefix)
    log.debug(f"Found videos {videos}")
    return videos


def get_videos_list(s3_client, prefix: str) -> List[str]:
    """
    Returns:
        List of mp4 videos with more than 0 bytes
    """
    bucket = os.environ["VIDEOS_BUCKET"]
    log.debug(f"List {prefix} in {bucket}")
    res = s3_client.list_objects_v2(Bucket=bucket, Prefix=prefix)
    return [video["Key"] for video in res.get("Contents", []) if is_valid_video(video)]


class JsonError(BaseException):
    """
    JSON {"error": "string"}
    """

    def __init__(self, msg):
        super(__class__, self).__init__(msg)
        self.json = json.dumps({"error": msg})


class InvalidListVideosRequest(JsonError):
    """The request is missing features"""

    pass


class InvalidTime(JsonError):
    """The start or end time are invalid"""

    pass


APIV2_REQUIRED_PARAMS = ["camera", "start", "utcoffset"]
APIV2_OPTIONAL_PARAMS = ["end"]


def is_api_v2(query_params: Dict[str, Any]) -> bool:
    """
    Returns:
        True if all required API V2 parameters are not None
    """
    log.debug(f"IsV2? {query_params}")
    v2_request = {}
    for param in APIV2_REQUIRED_PARAMS:
        log.debug(f"Getting param: {param}")
        v2_request[param] = query_params.get(param)
        if v2_request[param] is None:
            return False
    return True


def parse_ymdhms(ymdhms: str) -> datetime:
    """
    Parse a video range timestamp of YYYYMMDDhhmmss
    """
    # pylint: disable=raise-missing-from
    try:
        return datetime.strptime(ymdhms, START_TIME_FORMAT)
    except (ValueError, TypeError):
        raise InvalidTime(f"Invalid time: {ymdhms}")


# pylint: disable=too-many-instance-attributes
class ListVideosRequest:
    """
    A request to list videos. The init times are assumed to be local
    and the utcoffset, their distance from UTC

    Args:
        camera: Camera name
        utcoffset: Requestors local time's +/- distance from UTC
        start: YYYYMMDDhhmmss precision time string in the Requestors local time
        [end]: Optional. Same protocol as the `start` argument
    """

    def __init__(
        self,
        camera: str = None,
        utcoffset: str = None,
        start: str = None,
        end: Optional[str] = None,
    ):
        """
        Args:
            camera: Camera name
            utcoffset: signed string of integer
            start: YYYYMMDDHHMMSS
            end: Optionally the same format as start
        """

        self.camera = camera
        # pylint: disable=raise-missing-from
        try:
            self.utcoffset = int(utcoffset)
        except TypeError:
            raise InvalidListVideosRequest("Invalid UTC offset")

        self.start = start
        self.end = end

        self.__validate()
        self.__set_utc_time()
        log.debug(f"Start: {self.utc_start} End: {self.utc_end}")

    def __validate(self) -> None:
        """
        Validate request input
        Raises: InvalidListVideosRequest
        Returns: None
        """

        missing_keys = [
            k for k in APIV2_REQUIRED_PARAMS if self.__getattribute__(k) is None
        ]
        if len(missing_keys) > 0:
            raise InvalidListVideosRequest(f"Missing the required keys: {missing_keys}")

    def __set_utc_time(self) -> None:
        """
        Set the UTC start and end boundaries
        Raises:
            InvalidTime
            InvalidListVideosRequest
        Returns: None
        """
        # pylint: disable=raise-missing-from
        try:
            self.local_start = parse_ymdhms(self.start)
            self.utc_start = self.local_start + timedelta(hours=self.utcoffset)
        except InvalidTime:
            raise InvalidListVideosRequest(f"Invalid start time: {self.start}")

        try:
            self.local_end = parse_ymdhms(self.end)
            self.utc_end = self.local_end + timedelta(hours=self.utcoffset)
        except InvalidTime:
            self.local_end = datetime.utcnow() - timedelta(hours=self.utcoffset)
            self.utc_end = datetime.utcnow()

        if self.utc_end < self.utc_start:
            raise InvalidListVideosRequest(
                f"End time is before start {self.end}, {self.start}"
            )

    def hours(self) -> int:
        """
        Returns:
            hours between start and end
        """
        return ceil((self.local_end - self.local_start).seconds / 60 / 60)

    def __str__(self) -> str:
        return f"'Camera: {self.camera} Start: {self.start} End: {self.end}'"


def s3_video_prefix(req: ListVideosRequest, video_datetime: datetime) -> str:
    """
    Create the s3 prefix where videos for a req:video_datetime coordinate
    """
    root = os.environ["VIDEOS_ROOT"]
    return datetime.strftime(
        video_datetime, f"{root}/{req.camera}/%Y-%m-%d/001/dav/%H/"
    )


def v2_list_videos(req: ListVideosRequest):
    """
    APIV2
    List videos
    Args:
        req: Request to list camera videos in a certain timerange
    """
    videos = []
    video_prefixes = []

    s3_client = boto3.client("s3")
    for hours in range(req.hours() + 1):
        video_prefixes.append(
            s3_video_prefix(req, req.local_start + timedelta(hours=hours))
        )

    with ThreadPoolExecutor(max_workers=8) as executor:
        futures = [
            executor.submit(get_videos_list, s3_client, pfx) for pfx in video_prefixes
        ]
        for future in as_completed(futures):
            videos += future.result()
    log.debug(f"Found {len(videos)} matching request {req}")
    return videos


def is_authorized(authorization: str) -> bool:
    """
    Args:
        authorization: Bearer token: "Bearer Token"
    Returns:
        True if authorized
    """
    try:
        psk = os.environ["PSK"]
    except KeyError:
        log.error("Missing PSK environment configuration")
        return False
    try:
        token = authorization.split(" ")[1]
    except IndexError:
        return False

    return psk == token


# pylint: disable=too-many-return-statements
def handle(event, context):
    """
    ListVideosQueryParameters:
        camera=Cameraname
        period=7
        step=2
        date=YYYYMMDD
        utcoffset=HH
        start=YYYYMMDDHHMMSS
        [end=YYYYMMDDHHMMSS] | Now
    Returns:
        API Gateway integration response with CORS headers
    """
    _ = context
    params = event.get("queryStringParameters", {})
    authorization = event.get("headers", {}).get("Authorization")
    if authorization is None:
        return integration_response(403, error="Missing authorization")
    if not is_authorized(authorization):
        return integration_response(403, error="Unauthorized")

    list_videos_request = {}
    for param in REQUIRED_PARAMS:
        list_videos_request[param] = params.get(param)

    # pylint: disable=consider-iterating-dictionary
    if None in list_videos_request.keys() and not is_api_v2(params):
        missing_keys = [key for key in list_videos_request.keys() if key is None]
        return integration_response(
            400, error=f"Required keys {missing_keys} are missing"
        )

    if is_api_v2(params):
        req = ListVideosRequest(
            camera=params.get("camera"),
            utcoffset=params.get("utcoffset"),
            start=params.get("start"),
            end=params.get("end"),
        )
        videos = v2_list_videos(req)
        return integration_response(
            200,
            json_data=json.dumps(
                {
                    "videos": videos,
                    "start": f"{req.local_start}",
                    "utc_start": f"{req.utc_start}",
                }
            ),
        )

    try:
        videos = list_videos(list_videos_request)
    except KeyError:
        return integration_response(500, error="Missing environment configuration")
    except TypeError:
        return integration_response(400, error="Invalid request payload")

    return integration_response(200, json_data=json.dumps(videos))
