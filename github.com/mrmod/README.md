# Examples : A Bazel Credential Helper

A [Bazel credential helper](https://blog.engflow.com/2023/10/09/configuring-bazels-credential-helper/) allows configuring authorization headers (or any HTTP headers) for requests sent to remote urls.

Credential Helpers which `exit 0` put a JSON document of a map of header names to header values
```
{
    "Authorization": "Bearer SecureData",
    "CSRF": "SecureMark",
    "KeyId": "PublicKeyId"
}
```

A Go example [bazel credential helper](./bazelcredentialhelper/) exchanges `CLIENT_CREDENTIAL` with [a credential store](./credentialstore/) for its `SERVICE_TOKEN`.

```
BAZEL_WORKSPACE=/path/to/bazel/project

echo "Starting credential store"
pushd credentialstore ; SERVICE_TOKEN=serviceToken CLIENT_CREDENTIAL=clientCredential go run main.go &; popd

echo "Starting credentialed service Bazel will call"
pushed credentialservice ; SERVICE_TOKEN=serviceToken go run main.go & ; popd

echo "Building and installing `localhost-credential-helper` to BAZEL_WORKSPACE"
pushed bazelcredentialhelper ; go build -o localhost-credential-helper main.go ; cp localhost-credential-helper $BAZEL_WORKSPACE ; popd

```

After that's started, you can use it from the `BAZEL_WORKSPACE`:

```
# .bazelrc
common --credential_helper=/Users/bruce/Development/homewatch/localhost-credential-helper

# WORKSPACE
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "localhost_demo",
    sha256 = "InvalidChecksum",
    urls = [
        "http://localhost:8080/artifacts/localhost_demo.zip",
    ],
)
```
Then
```
CLIENT_CREDENTIAL=clientCredential bazel fetch @localhost_demo//...
```

Watch the [credential store](./credentialstore/) log for the credential exchange attempt and then the request for the `http_archive` `urls` artifact `/artifacts/localhost_demo.zip`.

# CredentialedService

An HTTP Service which reads an `Authorization` HTTP header and splits it on `Bearer ` (with a trailing whitespace) optional set with `VALID_TOKNEN`.

# CredentialStore

A HTTP Service which exchanges a received `CLIENT_CREDENTIAL` a client knows with a `SERVICE_TOKEN` it knows over the `/` HTTP route and artifacts over the `/artifacts/${Artifact}` route given a `SERVICE_TOKEN`-authorized request.

# BazelCredentialHelper

An HTTP Client which sends an env `CLIENT_CREDENTIAL` value to `localhost:8080` and exits with a JSON document of authorized headers as expected by the Bazel experimental Credential Helper interface.