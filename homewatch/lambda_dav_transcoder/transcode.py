
import argparse
from concurrent.futures import ThreadPoolExecutor, as_completed
import json
from datetime import datetime
from argparse import ArgumentParser
import boto3


FUNCTION_NAME = "devDavTranscoder"

def transcode(lambda_client, video):
    start_time = datetime.now()
    result = lambda_client.invoke(
        FunctionName=FUNCTION_NAME,
        Payload=json.dumps({"video": video}),
    )
    end_time = datetime.now()
    print(f"Transcoded {video} in {end_time-start_time}seconds")
    return end_time-start_time, result

def display_result(time_taken, result, video):
    body = json.loads(
        json.loads(
            result['Payload'].read().decode(encoding="utf-8"),
        ).get("body"),
    )
    try:
        print(f"{time_taken}: {body.get('output_key')}")
    except KeyError as key_name:
        print(f"Error: {key_name} missing transcoding {video}")

def main(s3_listing=None, single_file=None):
    # print(f"S3Listing: {s3_listing}")
    # print(f"SingleFile? {single_file}")
    client = boto3.client("lambda")
    if single_file:
        display_result(*transcode(client, s3_listing), single_file)
        return
    
    try:
        videos = []
        with open(s3_listing, 'r') as fp:
            videos = [line.strip("\n").split(" ")[-1] for line in fp.readlines()]
        davs = [video for video in videos if video.endswith(".dav")]
        print(f"Transcoding {len(videos)} videos")
        results = []
    
        with ThreadPoolExecutor(max_workers=16) as executor:
            for video in davs:
                results.append(executor.submit(transcode, client, video))
            completed = 0
            for execution in as_completed(results):
                execution.result()
                completed += 1
                # duration, result = execution.result()
                if completed%10 == 0:
                    print(f"{completed}/{len(videos)} transcoded")
    except FileNotFoundError:
        print(f"S3 Listing {s3_listing} doesn't exist")


if __name__ == "__main__":
    cli = ArgumentParser("transcode", usage="Pass an s3 listing to transcode")
    cli.add_argument("s3_listing", help="File path with an S3 listing output")
    cli.add_argument("--single-file", action="store_true", help="Input is a single prefix and not an s3 listing")
    user_input = cli.parse_args()
    print(user_input)
    main(**user_input.__dict__)
