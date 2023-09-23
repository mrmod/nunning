# v0.2
# bucket=my-bucket-name
bucket=$S3UploadBucket

# environment=dev
environment=$Stage

# TrimPrefix=/homewatch/working/directory/
# File=/homewatch/working/directory/file.dav
# onTrimPrefix($File) => file.dav
# NOTE: The exact literal prefix is trimmed
TrimPrefix=/home/cameras/

nohup ./homewatch --cleanup-all-files \
        --debug \
        --vvv \
        --enable-event-upload \
        --s3-index-bucket-url="s3://${bucket}/${environment}/IndexEvents" \
        --index-trim-prefix="$TrimPrefix" \
        --enable-video-upload \
        --s3-video-bucket-url="s3://${bucket}/${environment}/Videos" \
        --video-trim-prefix="$TrimPrefix" 2>&1 