#!/usr/bin/env bash

cd "$(dirname "$0")"

echo "Starting mysql docker container...">&2
docker run \
  --platform linux/x86_64 \
  -d \
  --rm \
  --name jurassic-park-sql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_USER=admin \
  -e MYSQL_PASSWORD=password \
  -e MYSQL_DATABASE=jurassicpark \
  -v $(pwd):/resources \
  mysql:5.7.36

until docker exec -it jurassic-park-sql mysql -uadmin -ppassword --execute "SHOW DATABASES;" >/dev/null 2>&1; do
  echo "Waiting for mysql to be ready..." >&2
  sleep 1
done

echo "Creating database in local mysql docker container..." >&2
docker exec -it jurassic-park-sql mysql -uadmin -ppassword --execute "CREATE DATABASE jurassicpark;">/dev/null 2>&1 
echo "Creating the tables in local mysql docker container..." >&2
docker exec -it jurassic-park-sql mysql -uadmin -ppassword --execute "use jurassicpark;SOURCE /resources/jurassic-park-db.sql;" >/dev/null 2>&1;