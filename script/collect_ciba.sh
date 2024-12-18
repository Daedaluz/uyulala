#!/bin/bash
curl -Ss -X POST https://localhost:5173/api/v1/collect \
	-H "Content-Type: application/x-www-form-urlencoded" \
	-d grant_type=urn:openid:params:grant-type:ciba \
	-d client_id=demo \
	-d client_secret=demo \
	-d auth_req_id=$1
