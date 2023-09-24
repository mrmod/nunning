"""
DAV 1D video transcoder lambda
"""
# pylint: disable=logging-fstring-interpolation
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
import subprocess
import os
import json
import logging
from urllib.parse import unquote
from glob import glob
from typing import Any, List, Optional, Dict, Union
import base64

import boto3
import botocore

FFMPEG = "/opt/bin/ffmpeg"
log = logging.getLogger("DavTranscoder")
log.setLevel(logging.DEBUG)


class JsonError(BaseException):
    """
    Error renderable with JSON
    """

    def __init__(self, msg):
        super().__init__()
        self.msg = msg

    def to_json(self):
        """Return json string"""
        return json.dumps({"error": self.msg})


class TranscodingError(JsonError):
    """
    Error thrown when transcoding fails
    """


class FailedToSaveMetadata(JsonError):
    """
    Error thrown when transcoding metadata fails to save
    """


class FailedToWriteMetric(JsonError):
    """
    Error thrown when writing the CloudWatch metric fails
    """

# BackyardSanctuary authorization component
def check_cookie_is_authorized(signed_jwt: str) -> bool:
    authorizer_function = os.environ.get("AUTHORIZER_FUNCTION_NAME", "CookieAuthorizer")
    lc = boto3.client("lambda")
    try:
        result = lc.invoke(
            FunctionName=authorizer_function,
            InvocationType="RequestResponse",
            Payload=json.dumps({"SignedCookie": signed_jwt}).encode("utf-8"),
        )
        log.debug(f"Cookie authorization result {result}")
        body = json.load(result.get("Payload"))
        if body is None or body.get("statusCode", 900) == 200:
            return True
        log.debug(f"Decoded authorization body: {body}")
        return False
    except Exception as err:
        log.error(f"Failed to authorizer jwt: {err}")
    return False


# BackyardSanctuary authorization component
def is_authorized(cookies: List[str]) -> bool:
    for cookie in cookies:
        log.debug(f"Trying to authorize {cookie}")
        if cookie and cookie.startswith("byst="):
            try:
                log.debug(f"Checking cookie authorization for {cookie}")
                return check_cookie_is_authorized(cookie.split("=", 1)[1])
            except Exception as err:
                log.debug(f"Cookie authorization failed: {err}")
                return False
    return False



def response(
    status_code: int,
    message: Optional[str] = None,
    error: Optional[str] = None,
    json_data: Optional[str] = None,
) -> Dict:
    """
    Create a Lambda response
    Returns:
        Dict[statusCode, body: str]
    """
    _response = {"statusCode": status_code}
    if message:
        _response["body"] = json.dumps(
            {
                "message": message,
            }
        )
        return _response

    if error:
        _response["body"] = json.dumps(
            {
                "error": error,
            }
        )
        return _response

    if json_data:
        _response["body"] = json_data
        return _response

    return _response


FRAMES_PATH = "/tmp/frames"

keyframes_cmd = [
    "-vf",
    # "select='gt(scene,0.04)'",  # > 4% of pixels changed, not good at night
    "select='eq(pict_type,I)'",  # Select only I/Intraprediction frames (Iframes)
    "-vsync",
    "0",
    # only capture one frame
    # "-vframes",
    # "1",
    # Seek 2 seconds in to avoid gray/missing Iframe videos
    "-ss",
    "2.0",
    # scale frame to 960x540 from 3840x2160
    "-s",
    "960x540",
    "-f",
    "image2",  # JPEG
]
oneframe_only_cmd = ["-vframes", "1", "-"]
transcode_to_mp4_cmd = [
    # 1080p from 4K (3840x2160)
    # "-vf",
    # "scale=1920x1080",
    "-s",
    "896x414",
    "-format:v",
    "fps=12",
    # Seekable MP4 container
    "-f",
    "ismv",
    # "-preset",
    # "ultrafast",
    "-",
]


def decode_keyframes(
    base_command: List[str], uniq_id=""
) -> subprocess.CompletedProcess:
    """
    Decode keframes from a video where keyframes is defined as frames with a 4% difference in
    pixels or greater
    """
    try:
        os.makedirs(FRAMES_PATH)
    except FileExistsError:
        pass
    local_path = os.path.join(FRAMES_PATH, f"{uniq_id}_%02d.jpeg")
    log.debug(f"Decoding keyframes to {local_path}")
    command = base_command + keyframes_cmd + [local_path]
    log.debug(f"Decoding with command {' '.join(command)}")
    return ffmpeg_transcode(command)


def transcode_dav_to_mp4(
    base_command: List[str], uniq_id=""
) -> subprocess.CompletedProcess:
    """
    Transcode a DAV to an MP4
    """
    command = base_command + transcode_to_mp4_cmd
    log.debug("Transcoding DAV to h264")
    return ffmpeg_transcode(command)


def ffmpeg_transcode(command: List[str]) -> subprocess.CompletedProcess:
    """
    Invoke FFMPEG to run a command against a video behind a presigned URL
    """

    try:
        return subprocess.run(
            command, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
    except subprocess.CalledProcessError as error:
        # pylint: disable=raise-missing-from
        log.error(f"Error calling ffmpeg: {error}")
        raise TranscodingError(f"Failed to process video: {error}") from error


def upload_frame(s3_client, bucket, key, jpg):
    log.debug(f"Uploading frame {jpg}")
    with open(jpg, "rb") as fp:
        s3_client.put_object(
            Body=fp,
            Bucket=bucket,
            ContentType="image/jpeg",
            Key=key,
            StorageClass="STANDARD_IA",
        )
    log.debug(f"Uploaded {jpg} to {key}")
    try:
        os.remove(jpg)
    except Exception as err:
        log.warning(f"Unable to delete {jpg}: {err}")
    return jpg


def upload_frames(
    s3_client, ffmpeg: subprocess.CompletedProcess, root: str, uniq_id=""
) -> List[str]:
    """
    Upload keyframes to S3
    """
    bucket = os.environ["VIDEOS_BUCKET"]
    frames_manifest = []
    keyframes = glob(f"{FRAMES_PATH}/{uniq_id}_*.jpeg")
    log.debug(f"Uploading {len(keyframes)} keyframes from {FRAMES_PATH} to {root}")
    for jpg in keyframes:
        key = f"{root}/{os.path.basename(jpg)}"
        frames_manifest.append(key)
        futures = []
        with ThreadPoolExecutor(max_workers=4) as executor:
            futures.append(executor.submit(upload_frame, s3_client, bucket, key, jpg))
        for _future in as_completed(futures):
            result = _future.result()
            log.debug(f"Result: {result}")
    return frames_manifest


def upload_video(
    s3_client, ffmpeg: subprocess.CompletedProcess, key: str, uniq_id=""
) -> List[str]:
    """
    Upload a transcoded video to S3
    """
    bucket = os.environ["VIDEOS_BUCKET"]
    log.debug(f"Uploading video to {key}")
    s3_client.put_object(
        Body=ffmpeg.stdout,
        Bucket=bucket,
        ContentType="video/mp4",
        Key=key,
    )
    return [key]


S3_DAV_VIDEO_URL = "s3DavVideoUrl"
IS_IMAGE_ONLY_TRANSCODE = "isImageOnlyTranscode"
TranscodeMetadata = Dict[str, Union[List[str], str]]


def get_frames_path(dav_video_key: str):
    """
    Removes camera mfr-specific cruft from the end of the DAV video S3 Key
    Args:
        dav_video_key: $prefix/$video.[abcd].dav
    Returns:
        $prefix/$video
    """
    source, _ = os.path.splitext(dav_video_key)

    return source.split("[")[0]


# pylint: disable=too-many-locals
def transcode(video: str, only_keyframes=False, uniq_id="") -> TranscodeMetadata:
    """
    Transcode a video. Removes bracketed nonsense from the video name
    Args:
        video: Prefix within an S3 bucket
    """
    s3_client = boto3.client("s3")
    log.debug(f"Transcoding {video}")

    bucket = os.environ["VIDEOS_BUCKET"]

    url = s3_client.generate_presigned_url(
        "get_object",
        Params={
            "Bucket": bucket,
            "Key": video,
        },
        ExpiresIn=300,
    )
    base_command = [
        FFMPEG,
        "-y",
        "-i",
        url,
    ]

    key = f"{get_frames_path(video)}.mp4"
    decode_fun = transcode_dav_to_mp4
    upload_fun = upload_video

    if only_keyframes:
        decode_fun = decode_keyframes
        upload_fun = upload_frames
        key = f"{get_frames_path(video)}/keyframes"

    log.debug(f"Transcoding from {url} to {key}")
    try:
        ffmpeg = decode_fun(base_command, uniq_id)
    except TranscodingError as error:
        raise TranscodingError(f"Error processing {key}: {error}") from error
    if ffmpeg.returncode != 0:
        log.error(f"Error: {ffmpeg.stderr}")
    log.debug("Transcoded video successfully")
    if not only_keyframes and len(ffmpeg.stdout) == 0:
        log.info(f"No frames decoded in {key}")

    output = upload_fun(s3_client, ffmpeg, key, uniq_id)
    log.debug(f"Created {output} in S3")
    return {
        S3_DAV_VIDEO_URL: video,
        "urls": output,
        IS_IMAGE_ONLY_TRANSCODE: only_keyframes,
    }


def write_transcode_metadata(transcode_metadata: TranscodeMetadata) -> None:
    """
    Writes transcode metadata to $FramesPath.json
    Raises:
        FailedToSaveMetadata when S3 doesn't 200 on writing the JSON object
    """
    bucket = os.environ["VIDEOS_BUCKET"]
    prefix = get_frames_path(transcode_metadata.get(S3_DAV_VIDEO_URL))

    s3_client = boto3.client("s3")
    key = f"{prefix}.json".replace("Videos", "Events")
    log.debug(f"Creating metadata {key}")
    res = s3_client.put_object(
        Bucket=bucket,
        Key=key,
        Body=json.dumps(transcode_metadata),
    )
    if res.get("ResponseMetadata", {}).get("HTTPStatusCode", 999) != 200:
        raise FailedToSaveMetadata(f"Failed to save metadata for {key}")
    log.debug(f"Saved metadata to {key}")


VideoEventItem = Dict[Any, Any]
VIDEO_EVENT_TIME_FORMAT = "%Y%m%d%H%M%S"
VIDEO_EVENT_TYPE = {
    0: "MP4Transcode",
    1: "Keyframes",
}


def build_video_event_item(transcode_metadata: TranscodeMetadata) -> VideoEventItem:
    """
    Creates a DynamoDB table Item
    Returns
        VideoEventItem
    """
    prefix = get_frames_path(transcode_metadata.get(S3_DAV_VIDEO_URL))
    transcode_data = f"{prefix}.json".replace("Videos", "Events")
    camera_name = get_camera_name(transcode_metadata.get(S3_DAV_VIDEO_URL))
    event_datetime = datetime.utcnow().strftime(VIDEO_EVENT_TIME_FORMAT)
    return {
        "Src": camera_name,
        "DateTime": event_datetime,
        # Key where the DAV encoded video was uploaded to
        "DavKey": transcode_metadata.get(S3_DAV_VIDEO_URL),
        # Key for JSON Transcode Metadata
        "TranscodeDataKey": transcode_data,
        "EventType": VIDEO_EVENT_TYPE[transcode_metadata.get(IS_IMAGE_ONLY_TRANSCODE)],
    }


def write_event(transcode_metadata: TranscodeMetadata) -> None:
    """
    Writes transcode metadata to Homewatch Events table
    Raises:
        KeyError when the environment variable HOMEWATCH_TABLE is not defined
    """
    table = os.environ["HOMEWATCH_TABLE"]

    video_event_item = build_video_event_item(transcode_metadata)
    homewatch = boto3.resource("dynamodb").Table(table)

    homewatch.put_item(
        TableName=table,
        Item=video_event_item,
    )


def write_iframe_metric(camera: str, iframe_count: int) -> None:
    """
    Writes the count of iframes decoded from a DAV upload to the $EnvironmentHomewatch Namespace
    Raises:
        KeyError if the environment is missing critical configuration
        FailedToWriteMetric if a metric fails to write for any reason
    """
    cw_metrics = boto3.client("cloudwatch")
    log.debug(f"Writing iframe count metric for {camera}: {iframe_count}")
    try:
        cw_metrics.put_metric_data(
            Namespace=os.environ["IFRAMECOUNT_NAMESPACE"],
            MetricData=[
                {
                    "MetricName": "IFrameCount",
                    "Value": iframe_count,
                    "Dimensions": [
                        {
                            "Name": "Camera",
                            "Value": camera,
                        }
                    ],
                }
            ],
        )
    except botocore.exceptions.ClientError as error:
        raise FailedToWriteMetric(
            f"Failed to write IFrameCount metric: {error}"
        ) from error


def unwrap_event(event: Dict):
    """
    Unwrap S3 event to get records or unwrap API GW integration event
    Returns:
        video: Path excluding bucket to the s3 resource
    """
    records = event.get("Records", [])
    log.debug(f"Unwrapping {len(records)} records")
    if records:
        video = records[0]["s3"]["object"]["key"]
        uniq_id = records[0]["responseElements"]["x-amz-id-2"][0:8]
        log.debug(f"Transcoding video from record {video}")
        # Safen the amazon key just in case
        uniq_id = base64.urlsafe_b64encode(uniq_id.encode("utf-8")).decode()[0:7]
        return (uniq_id, video)
    log.info("Handling test event")
    video = event["video"]
    uniq_id = base64.urlsafe_b64encode(video.split("/")[7].encode())[0:7].decode()
    return (uniq_id, video)


def get_camera_name(video: str) -> str:
    """
    Returns
        camera name from the S3 key for the video
    """
    try:
        return video.split("/")[2]
    except IndexError:
        return "UnknownCamera"


def is_transcoding_disabled(camera_name: str) -> bool:
    """
    Returns:
        By default, False
    """
    try:
        table = os.environ["CAMERAS_TABLE"]
    except KeyError:
        log.error("Missing CAMERAS_TABLE configuration ")
        return False
    ddb = boto3.client("dynamodb")

    response = ddb.get_item(TableName=table, Key={"Name": {"S": camera_name.lower()}})
    try:
        camera_state = response["Item"]
    except KeyError:
        log.debug(f"No state available for {camera_name}. Defaulting to Enabled")
        return False

    # By default transcoding is enabled
    state = camera_state.get("State", {}).get("S", "enabled")
    log.debug(f"Checking camera state is 'enabled' {camera_state}")

    return state.lower() == "disabled"


# pylint: disable=broad-except,too-many-return-statements
def handle(event, context):
    """
    Handle an S3 event or API Gateway event integration
    """
    _ = context
    log.debug(f"Transcoding event: {event}")
    uniq_id = "nani"

    is_s3_event = event.get("Records") != None
    is_test_event = event.get("video") != None

    is_api_call = not is_test_event and not is_s3_event

    if is_api_call:
        cookies = event.get("multiValueHeaders", {}).get("Cookie", [])
        authorized = is_authorized(cookies)
        if not authorized:
            log.debug("Authorization is required for API calls")
            return response(403, json_data=json.dumps({"error": "Missing authorization"}))
    try:
        os.stat(FFMPEG)
    except FileNotFoundError:
        log.error(glob("/opt/bin/*", recursive=True))
        return response(500, error="FFMpeg is missing")

    try:
        uniq_id, video_key = unwrap_event(event)
        video = unquote(video_key)
    except KeyError as key_name:
        return response(400, error=f"Missing required input {key_name}")
    except Exception:
        return response(400, error="Failed to decode video name from event")

    if video is None:
        return response(400, error="'video' is required")

    camera_name = None
    try:
        camera_name = video.split("/")[2]
    except IndexError:
        log.warning(f"Unexpected video path {video}")
    only_keyframes = os.environ.get("ONLY_KEYFRAMES_OUTPUT") != None
    transcode_metadata = None
    try:
        if camera_name and is_transcoding_disabled(camera_name):
            log.info(f"Transcoding is disabled for {camera_name}")
            transcode_metadata = {
                S3_DAV_VIDEO_URL: video,
                "urls": [],
                IS_IMAGE_ONLY_TRANSCODE: None,
            }
        else:
            log.info(f"Transcoding {camera_name} keyframes")
            transcode_metadata = transcode(video, only_keyframes=only_keyframes, uniq_id=uniq_id)
        
        write_transcode_metadata(transcode_metadata)
        camera_name = get_camera_name(transcode_metadata.get(S3_DAV_VIDEO_URL))
        write_iframe_metric(camera_name, len(transcode_metadata.get("urls")))
        write_event(transcode_metadata)
        log.info(f"Transcoded {len(transcode_metadata.get('urls', []))} keyframes from {camera_name}")
    except KeyError as error:
        log.error(f"Missing environment var for {error}")
        return response(500, error="Environment error")
    except TranscodingError as error:
        return response(500, json_data=error.to_json())
    except FailedToSaveMetadata as error:
        log.warning(f"Failed to save metadata {error}")
        return response(200, json_data=error.to_json())
    except FailedToWriteMetric as error:
        log.warning(f"Failed to write metric {error}")
        return response(200, json_data=error.to_json())

    return response(200, json_data=json.dumps(transcode_metadata))
