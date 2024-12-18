#!/bin/bash
BOOTSTRAP=$(docker compose -f compose/docker-compose.yml ps -a | wc -l)

PROFILE=${PROFILE:-default}

cmd=$1
shift 1

case $cmd in
  up)
      if [ -n "$PROFILE" ]; then
        docker compose -f compose/docker-compose.yml --profile "$PROFILE" up -d --build "$@"
      else
        docker compose -f compose/docker-compose.yml up -d --build "$@"
      fi
      if [ "$BOOTSTRAP" -ge 3 ]; then
        echo 'Already created'
      else
        docker exec -ti uyulala uyulala available -w
        docker exec -ti uyulala uyulala create key
        docker exec -ti uyulala uyulala create app --demo demo
        echo '------------------------'
        echo 'Done!'
        echo '* Head to https://localhost:8080/demo (certificate should be a self-signed one)'
        echo '* Create a new user'
        if [ -n "$PROFILE" ]; then
          echo '* Head to http://localhost:3000 to login to grafana with your new user'
        fi
        echo ''
        echo 'The QR code will not work due to the fact that its hosted on localhost.'
      fi
    ;;
  down)
      if [ -n "$PROFILE" ]; then
        docker compose -f compose/docker-compose.yml --profile "$PROFILE" down
      else
        docker compose -f compose/docker-compose.yml down  "$@"
      fi
    ;;
  *)
    echo "Usage: $0 {up|down}"
    exit 1
esac
