# Remote shell over gRPC

This project demonstrates how a remote shell session can be spawned using gRPC as transport.
It contains a library, client and a server CLIs.

No authorization is made and all session requests are accepted.

## Usage

You'll need to have go > 1.18 installed.

1. Create the server

    ```bash
    go run ./cmd/server
    ```

2. Run the client

    ```bash
    go run ./cmd/client
    ```

Server and client use `127.0.0.1:22222` for the connections by default.

