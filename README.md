# jurassic-park
Welcome to the Jurassic Park cage management system! This service manages the inventory of all cages in the park, their power status, and it tracks the dinosaurs that are kept within each cage. Special care has been taken to ensure that only compatible species are kept within each cage. Power for each cage is also managed and protected by this system. This system ensures that we don't cut power to any cage that is housing any dinosaurs as per our learnings from recent events. We've even made an extra effort to ensure that all dinosaurs in the park are female. This time, life will not find a way.

## Prerequisites
You will need go 1.21.1 or later installed on your developer machine.

## Setup
Pull the code from github by running the following command in your go path.
```
git clone https://github.com/EdgarH78/jurassic-park.git
```

Run go get to pull down any dependencies.
```
go get -u
```

## Running Integration tests
All automated tests for the jurassic-park management system are integration tests, so you'll need to run a test database. To install and run the test database run the following from the project directory:
```
scripts/run-tests-db.sh
```
That will install a test mysql database on your machine running in docker with the database schema initialized. This runs on port 3307 in order to isolate the test code from any production systems. Use the following command to run the tests:
```
go test ./...
```

## Running Locally
The simplest way to run the jurassic-park management system locally is use the built in script to run the mysql server on your local machine. Run the following command:
```
scripts/run-db.sh
```
No run the server with the following command:
```
go run main.go
```
The server is now running and listening on port 8080.

## Using a remote database
You can use the following environment variables to configure the jurassic-park management server to use a different MySQL Server.
```
SQL_HOST
SQL_USER
SQL_PASSWORD
SQL_DATABASE_NAME
```

## Using the API
The jurassic-park management system uses  REST API. It is focused on creating cages, adding dinosaurs to the park, adding dinosaurs to different cages and managing the power status of each cage. Detailed documentation for the API can be found in the swagger.yaml file.

## Future Improvements
We need to use transactions when changing the power status of a cage or adding a dinosaur to it. There currently is the potential for race conditions until that is resolved. Filtering support for dinosaurs is fairly robust. However we can only filter on cages based on their power status. We should add the ability to filter on cages that can house a dinosaur, so park managers can more quickly find the right cage for a dinosaur.

