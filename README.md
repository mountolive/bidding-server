# Example of a bidding server in Golang

The bidding server is suppose to be able to lookup, according to a publisherId and a position
the campaign with the greatest price possible.

## Test data
In order to load the test data you must have redis-server and cli installed in your machine.
Then execute the following:

```cat script.redis | redis-cli —pipe```

## Run
To run the server, you can simply run the entry point, main.go

```go run main.go```

This starts the server in port 8080

