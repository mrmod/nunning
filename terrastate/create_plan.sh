#!/bin/bash

D=$(date -u +%s)
terraform plan -out self.dev.$D.plan && \
    terraform show -json self.dev.$D.plan > self.dev.$D.plan.json && \
    cp self.dev.$D.plan.json self.dev.latest.plan.json && \
    cp self.dev.$D.plan self.dev.latest.plan


