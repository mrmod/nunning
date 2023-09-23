Homewatch proxies SFTP uploads from Lorex cameras to a web-hosting service.

This is just a toy.

## SFTP Proxy Feature

When an SFTP upload arrives at the Homewatch host (aka: CamerasManagerHost)
And the upload has a `.dav` extension
And the file has finished writing or an renamed to its permantent filename
Then the DAV encoded video is uploaded to S3
And then an S3 Event outside of Homewatch does transcoding and publishing

## Syslog Listener Feature

Given we want to upload DAV-encoded H.265 Videos to S3 so they can be transcoded
And we can use syslog to learn `.dav` encoded files are ready for upload
Then we should listen for syslog messages from SFTP to trigger the SFTP Proxy Feature

# Building

## For Linux
```
make build-linux
```

## For Raspberry Pi
```
make build-pi
```

# Deploying

Deployment pushes a built binary to a host managing cameras SFTP and SYLOG data.

## Via SCP
```
make deploy-scp CamerasManagerUser=cameras CamerasManagerHost=cameras-manager
```

## Via SFTP
```
make deploy-sftp CamerasManagerUser=cameras CamerasManagerHost=cameras-manager
```