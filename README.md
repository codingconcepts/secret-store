# secret-store
A secure way to store and share secrets.

## Installation

## Usage

### Register a client
go run main.go init --s=http://localhost:8080 --c alice.json

### Send a message
go run main.go push 2ab82eb7-cacb-40e9-befc-534f4fe5b5ce "attack at dawn" --s=http://localhost:8080