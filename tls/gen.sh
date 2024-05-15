#!/bin/bash
mkcert -cert-file server.crt -key-file server.key \
	localhost https://localhost:8080 https://localhost:5173
