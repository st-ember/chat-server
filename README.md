# Go Chat Server

![Go version](https://img.shields.io/badge/go-1.24.2-blue.svg)

A simple chat server and client built with Go, demonstrating a custom binary network protocol over TCP, with a PostgreSQL backend for persistence. This project uses a multi-module Go workspace for a clean separation of concerns.

## Project Structure

This repository uses a [Go Workspace](https://go.dev/doc/tutorial/workspaces) to manage multiple modules within a single project.

-   `server/`: Contains the main chat server application.
    -   `cmd/server/`: The entry point for the server.
    -   `internal/`: The internal logic of the server.
        -   `domain/`: Core domain types for the application.
        -   `infra/storage/`: The PostgreSQL storage implementation.
        -   `transport/`: Handles network communication, client and room management.
-   `client/`: Contains a command-line client application for connecting to the chat server.
-   `shared/`: Contains shared code used by both the `client` and `server`.
    -   `protocol/`: Defines the custom binary network protocol.
-   `go.work`: The Go workspace file that enables the `client` and `server` to import the `shared` module seamlessly.

## Architecture

The server is built with a clean architecture in mind, separating concerns into different layers:

-   **Transport Layer**: Manages TCP connections, reads and writes messages using the custom binary protocol. It handles client connections and manages chat rooms.
-   **Domain Layer**: Defines the core data structures of the application, such as `User`, `Room`, and `Message`.
-   **Infrastructure Layer**: Implements the storage interface using PostgreSQL for data persistence.

The server uses a command-oriented approach, where client actions are translated into commands that are processed in a central loop. This ensures that access to shared resources (like rooms and clients) is synchronized.

## Database

The server uses a PostgreSQL database to persist users, rooms, and message history.

### Schema

You need to create the following tables in your PostgreSQL database:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    nickname TEXT NOT NULL,
    remote_addr TEXT NOT NULL
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    room_id UUID NOT NULL REFERENCES rooms(id),
    content TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Communication Protocol

Communication between the client and server uses a custom binary protocol. Each message is a stream of bytes with the following structure:

| Field   | Type          | Size (bytes) | Description                                                                 |
| :------ | :------------ | :----------- | :-------------------------------------------------------------------------- |
| Type    | `MessageType` | 1            | The type of message (`JoinRoom`, `CreateRoom`, `Rooms`, `Leave`, `Chat`). |
| Length  | `uint32`      | 4            | The length of the `Content` payload.                                        |
| Content | `[]byte`      | `Length`     | The variable-length message payload.                                        |

## Getting Started

### Prerequisites

-   Go version 1.24.2 or higher.
-   A running PostgreSQL database.

### Running the Server and Client

You will need two separate terminal windows.

1.  **Set up the Database:**
    Connect to your PostgreSQL instance and run the SQL commands from the [Database](#database) section to create the required tables.

2.  **Configure the Database Connection:**
    The server expects a PostgreSQL connection string. It is currently hardcoded in `server/cmd/server/main.go`. You may need to change it to match your database configuration.

    ```go
    postgres, err := storage.NewDB("postgres://postgres:Br3akD3na@localhost:5432/chat?sslmode=disable")
    ```

3.  **Start the Server:**
    In the first terminal, run the following command from the project root:
    ```bash
    go run ./server/cmd/server
    ```
    The server will start and listen on port `8081`.

4.  **Start the Client:**
    In the second terminal, run the following command from the project root:
    ```bash
    go run ./client/cmd/client
    ```
    The client will connect to the server. You can now start sending messages.

## Client Commands

The client supports the following commands:

-   `/room {name}`: Join or create a specific chat room.
-   `/rooms`: List all the current available rooms.
-   `/leave`: Leave the current chat room.
-   `/quit`: Exit the client.
-   Any other text is sent as a chat message to the current room.