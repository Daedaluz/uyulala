#!/bin/bash
docker buildx build --platform linux/amd64,linux/arm64 --push -t registry.gitlab.com/initsx/uyulala .
