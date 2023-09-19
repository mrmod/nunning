# Terrastate (aka Landscape)

Terrastate is a tool for viewing the changes a `terraform plan` output would make to your environment.

## How it works

A `terraform plan` is run to create some output. That output is converted to JSON and posted to the **Terrastate Backend**.

The **Backend** hashes the plan and stores it in s3 and the date-ordered hash in DynamoDB.

# Usage

## 1. Create a plan

```
cat << CONTAINER > Dockerfile

FROM $TerraformContainer

EXPORT GIT_REF=GIT_SHORT_REF

COPY /module /module
ENTRYPOINT ["terraform"]

CONTAINER

docker build -t $Repo/Terrastate:$ChangeSetId
docker push
```

This creates an executable change.

## 2. Generate a plan

```
docker run -v /terraformModule:/module Terrastate:$ChangeSetId plan -out /module/plan
docker run -v /terraformModule:/module Terrastate:$ChangeSetId show -json /module/plan > /module/plan.json

changeSetId=$(curl -v -d @/module/plan.json $apiUrl/plans | jq '.changeSetId')
```

A saved plan is referencable by its changeSetId.

## 3. Link a change set to a git sha

```
cat << EXECUTION > execute.json
{
    "changeSetId": $changeSetId,
    "sha": "$(git rev-parse --verify --short HEAD)"
}
EXECUTION

curl -v -d @execution.json $apiUrl/execution
```

A submitted plan can be executed using the container of its commit

## 4. Run the plan

```
curl -v $apiUrl/run/$changeSetId?env=$env
```

`env` can be one of `prod|dev|test`. When the environment is test, the change is applied in a disposable environment.

# Developing

## Start a Backend

```
bazel run //:gazelle
bazel run //backend:gazelle-update-repos
bazel run //backend
```

## Start a UI

```
bazel fetch @npm//...
bazel run //webui:start
```