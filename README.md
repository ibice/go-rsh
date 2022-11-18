# Remote shell over gRPC

> Like `ssh` or `kubectl exec`, but over gRPC.

This library contains code to run remote commands using the [gRPC framework].
Apart from the library, [client](cmd/client/main.go) and [server](cmd/server/main.go) CLIs are included.

Features:

- Execute shell commands or spawn an interactive shell on the server.
- Interactive PTY sessions are used to run the commands.
- Client is able to exit using the exit code of the remote command.

## Usage

You'll need to have go > 1.18 installed.

1. Create the server

    ```bash
    go run ./cmd/server
    ```

2. Run the client

    ```bash
    # Spawn interactive shell
    go run ./cmd/client

    # Run command
    go run ./cmd/client -- ping 1.1.1.1 -c 3
    ```

Server and client use `127.0.0.1:22222` for the connections by default.


## Building

To compile server and client binaries, run:

```bash
make
```

[gRPC framework]: https://grpc.io
