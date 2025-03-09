Homewatch agent listens for traditional security camera DAV uploads, re-encodes them to H264, then uploads them to a cloud-based app for viewing over the web.

# Release Process

Releases can be deployed to a aystem for testing

```
./release.sh
```

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

# Packaging

## Build the Package

`RELEASE_VERSION=2.0.0-pre1`

```
make build-linux VERSION=$RELEASE_VERSION
```

## Build the Container Artifacts

```
cd _deploy
ansible-playbook \
    --inventory inventory.yaml  \
    --extra-vars @pre_build.vars.yaml \
    --extra-var release_version=${RELEASE_VERSION} \
    --extra-var "prometheus.homewatch_agent_address=localhost:9112" \
    pre_build.playbook.yaml
```

## Build the Container

```
docker build \
    -t homewatch:${RELEASE_VERSION} \
    .
docker tag homewatch:${RELEASE_VERSION} ${HOMEWATCH_ECR_REPO}:${RELEASE_VERSION}
```

# Deploy

```
docker push $REPO
```

# Release and Run

```
docker run \
    --name homewatch \
    --publish 2112:2112 \
    -v "${UploadsPath}:/upload" \
    -u "${UploadsUserUID}:${UploadsUserGID}" \
    homewatch:${RELEASE_VERSION}
```

# Building

## For Linux
```
make build-linux
```

## For Raspberry Pi
```
make build-pi
```

# Infrastructucture

Found [in _infra](_infra/README.md). Deployed with terraform.

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
