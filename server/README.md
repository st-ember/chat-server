# Go Chat Server

This project is a robust, TCP-based chat server written in Go. It is designed with a clean, modern architecture to be both scalable and maintainable.

## Architecture

The server's design emphasizes a clear separation of concerns and safe concurrency handling, making it a great example of idiomatic Go.

*   **Concurrency Model**: The server uses a channel-based concurrency model to prevent race conditions. A central `Server` struct runs a single-threaded `cmdLoop` that processes all state-mutating commands sent from client goroutines. This serializes access to shared state (like rooms and clients), eliminating the need for complex mutexes around application logic.

*   **Separation of Concerns**: The codebase is organized into distinct layers:
    *   `domain`: Contains the core data structures of the application (`User`, `Room`, `Message`).
    *   `transport`: Manages all network-related logic, including the TCP server, client connection handling, and the mapping of the network protocol to internal commands.
    *   `infra/storage`: Provides the concrete implementation for data persistence, currently using PostgreSQL.

*   **State Management**: A hybrid approach to state is used to get the best of both worlds:
    *   **PostgreSQL Database**: The database acts as the persistent source of truth for rooms, users, and message history. This ensures data survives server restarts.
    *   **In-Memory Cache**: On startup, the server loads all rooms from the database into an in-memory map. This allows for extremely fast, low-latency operations like broadcasting messages to active users without hitting the database for every message.

## Database Schema

The server requires a PostgreSQL database. The necessary tables can be created with the following SQL:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    nickname VARCHAR(255) NOT NULL,
    remote_addr VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    room_id UUID NOT NULL REFERENCES rooms(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Getting Started

### Prerequisites

*   Go (version 1.21 or newer)
*   PostgreSQL

### 1. Installation

Clone the repository to your local machine:
```sh
git clone https://github.com/st-ember/chat-server.git
cd chat-server
```

### 2. Configuration

The server address and database connection string are currently hardcoded in `cmd/server/main.go`.

**You must edit this file to match your environment before running.**

```go
// in cmd/server/main.go

// 1. Update the database connection string
postgres, err := storage.NewDB("postgres://YOUR_USER:YOUR_PASSWORD@localhost:5432/chat?sslmode=disable")

// 2. Update the server listen address if needed
server := transport.NewServer(":8081", postgres)
```
*For a production setup, it is highly recommended to refactor this to use environment variables instead of hardcoded values.*

### 3. Running the Server

Once the configuration is set, you can run the server from the root of the project directory:

```sh
go run ./cmd/server
```

The server will start and print a message that it is listening on the configured address.

## Network Protocol

The server uses a custom binary TCP protocol for communication. The definition of this protocol is located in the `github.com/st-ember/chat-server/shared` module, which must be available.

Each message consists of a fixed-size `Header` followed by a variable-size `Content` payload.

### Client Commands

A client can interact with the server by sending messages with the appropriate type and content.

*   **List Rooms (`protocol.Rooms`)**:
    *   Sends a request to get a list of all available rooms.
    *   The server will respond with a list of room names, separated by newlines.
    *   *Content*: Empty

*   **Create Room (`protocol.CreateRoom`)**:
    *   Creates a new persistent chat room.
    *   *Content*: The desired name of the room.

*   **Join Room (`protocol.JoinRoom`)**:
    *   Joins a client to an existing chat room.
    *   *Content*: The name of the room to join.

*   **Send Message (`protocol.Chat`)**:
    *   Sends a message to the currently joined room. The server will broadcast it to all other members of the room.
    *   *Content*: The text message to send.

*   **Leave Room (`protocol.Leave`)**:
    *   Removes the client from their current room.
    *   *Content*: Empty

### Example: Connect with Netcat

You can use `netcat` or `telnet` to connect to the server, but since the protocol is binary, you will need a dedicated client to send meaningful commands.

```sh
# Connect to the server
nc localhost 8081
```
