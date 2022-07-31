# Remote shell over gRPC

> Like `ssh` or `kubectl exec`, but over gRPC.

This library contains code to run remote commands using the [gRPC framework].
Apart from the library, [client](cmd/client/main.go) and [server](cmd/server/main.go) CLIs are included.

Features:

- Execute arbitrary processes other than the server shell.
- Interactive PTY sessions used to run the commands.
- Client is able to exit using the exit code of the remote command.

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


## Building

Run:

```bash
make
```

[gRPC framework]: https://grpc.io
