# Purpose

Backend of the app which is an API only receiving calls from a VueJS frontend.

# Dev

1. `cd ~/backend`
1. `export GOPATH=~/backend`
1. `go run src/go_project/*.go`

# Deployment

`docker run --net my_network --ip 172.50.0.10 -p 8000:8000 -e "CORS_ALLOWED_ORIGIN=http://api.example.com:9000" -e "REMOTE_DB_HOST=10.10.10.10" -e "LOCAL_DB_HOST=172.50.0.1" -e "LOG_FILE_PATH=/var/log/backend/errors.log" -e "USER_EMAIL=me@example.com" -v /var/log/backend:/var/log/backend -d --name backend_v1_container myaccount/myrepo:backend_v1`


