#!/bin/bash
docker buildx build -f docker/Dockerfile --platform linux/amd64,linux/arm64 --push -t ghcr.io/daedaluz/uyulala:latest .
