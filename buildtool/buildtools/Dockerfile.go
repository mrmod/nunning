FROM ubuntu:jammy as base

RUN apt -qq update && apt install -y golang ca-certificates openssl

ENTRYPOINT ["go"]
# Default action
CMD  ["version"]