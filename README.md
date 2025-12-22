# Go Chat Server

![Go version](https://img.shields.io/badge/go-1.24-blue.svg)

A simple chat server and client built with Go, demonstrating a custom binary network protocol over TCP. This project uses a multi-module Go workspace for a clean separation of concerns.

## Project Structure

This repository uses a [Go Workspace](https://go.dev/doc/tutorial/workspaces) to manage multiple modules within a single project.

-   `server/`: Contains the main chat server application. It's responsible for accepting client connections, processing messages, and broadcasting them to other clients.
-   `client/`: Contains a command-line client application for connecting to the chat server.
-   `shared/`: Contains shared code used by both the `client` and `server`. This primarily includes the data structures and encoding/decoding logic for the communication protocol.
-   `go.work`: The Go workspace file that enables the `client` and `server` to import the `shared` module seamlessly.

## Communication Protocol

Communication between the client and server uses a custom binary protocol. Each message is a stream of bytes with the following structure:

| Field   | Type          | Size (bytes) | Description                                        |
| :------ | :------------ | :----------- | :------------------------------------------------- |
| Type    | `MessageType` | 1            | The type of message (e.g., `Join`, `Leave`, `Chat`). |
| Length  | `uint32`      | 4            | The length of the `Content` payload.               |
| Content | `[]byte`      | `Length`     | The variable-length message payload.               |

## Getting Started

### Prerequisites

-   Go version 1.22 or higher.

### Running the Server and Client

You will need two separate terminal windows.

1.  **Start the Server:**
    In the first terminal, run the following command from the project root:
    ```bash
    go run ./server/cmd/server
    ```
    The server will start and listen on port `8081`.

2.  **Start the Client:**
    In the second terminal, run the following command from the project root:
    ```bash
    go run ./client/cmd/client
    ```
    The client will connect to the server. You can now start sending messages.

## Commands (Planned)

The following commands are planned for future implementation. Currently, all input is sent as a standard `Chat` message.

-   `/room {name}`: Join or create a specific chat room.
-   `/rooms`: List all the current available rooms.
-   `/leave`: Leave the current chat room.