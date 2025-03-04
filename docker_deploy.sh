#!/usr/bin/sh
docker build . -t my-golang-app
docker run -d --name url-shortener --env-file .env --publish 5080:5080 my-golang-app