#!/bin/sh

set -o pipefail -x

ENVIRONMENT=dev
BUCKET=${S3_BUCKET}
TRIM_PREFIX=/uploads
WATCH_PATHS=/uploads/source1,/uploads/source2

echo Starting homewatch
# Exposes Prometheus metrics endpoint on default port
/app/homewatch \
        --v2 \
        --v2-enable-metrics \
        --v2-watch-paths $WATCH_PATHS \
        --v2-enable-watch-reaper \
        --enable-video-upload \
        --s3-video-bucket-url="s3://$BUCKET/$ENVIRONMENT/Videos" \
        --video-trim-prefix="$TRIM_PREFIX"
