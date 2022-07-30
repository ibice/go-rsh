# Remote shell over gRPC

This project demonstrates how a remote shell session can be spawned using gRPC as transport.

No authorization is made and all session requests are accepted.

## Usage

You'll need to have go > 1.18 installed.

1. Create the server

    ```bash
    go run ./server
    ```

2. Run the client

    ```bash
    go run ./client
    ```

Server and client use `localhost:2222` for the connections by default.

