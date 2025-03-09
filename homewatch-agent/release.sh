#!/bin/bash

LOCAL_TAG=homewatch:agent-${RELEASE_VERSION}
REGISTRY=588487667149.dkr.ecr.us-west-2.amazonaws.com
REGISTRY_URL=${REGISTRY}/homewatch
REMOTE_TAG=${REGISTRY_URL}:agent-${RELEASE_VERSION}

# SERVICE_RESTART : yes | no
SERVICE_RESTART=${SERVICE_RESTART:-no}
# PUSH_CONTAINER : yes | no
# Pushes to ECR when 'yes'
PUSH_CONTAINER=${PUSH_CONTAINER:-no}

# BUILD_CONTAINER : yes | no
# Builds the container when 'yes'
BUILD_CONTAINER=${BUILD_CONTAINER:-no}

# Build the homewatch agent binary
make build-linux Version=${RELEASE_VERSION} Output=_dist

if [[ "$BUILD_CONTAINER" == "yes" ]]; then
    # Stage the Homewatch Agent container components
    pushd _deploy
    ansible-playbook \
        --inventory inventory.yaml \
        --extra-vars @pre_build.vars.yaml \
        --extra-var release_version=${RELEASE_VERSION} \
        pre_build.playbook.yaml
    popd
    # Build the Homewatch Agent container
    docker build -t ${LOCAL_TAG} .
    echo Built Container: ${LOCAL_TAG}
    exit 0
fi

if [[ "$PUSH_CONTAINER" == "yes" ]]; then
    aws --region us-west-2 ecr \
        get-login-password |
        docker login \
            --username AWS \
            --password-stdin ${REGISTRY}
    # Push the Homewatch Agent container to the registry
    docker tag ${LOCAL_TAG} ${REMOTE_TAG}
    docker push ${REMOTE_TAG}
    echo "Pushed to ${REMOTE_TAG}"
    exit 0
fi