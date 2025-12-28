package storage

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	dm "github.com/st-ember/chat-server/internal/domain"
)

type DB struct {
	conn *sql.DB
}

func NewDB(connString string) (*DB, error) {
	conn, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("error opening postgres connection: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging postgres database: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) SaveUser(user *dm.User) error {
	_, err := db.conn.Exec("INSERT INTO users (id, nickname, remote_addr) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET nickname = EXCLUDED.nickname, remote_addr = EXCLUDED.remote_addr", user.ID, user.Nickname, user.RemoteAddr)
	if err != nil {
		return fmt.Errorf("error saving user: %w", err)
	}

	return nil
}

func (db *DB) GetUserByID(id uuid.UUID) (*dm.User, error) {
	row := db.conn.QueryRow("SELECT id, nickname, remote_addr FROM users WHERE id = $1", id)

	user := &dm.User{}
	if err := row.Scan(&user.ID, &user.Nickname, &user.RemoteAddr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}

	return user, nil
}

func (db *DB) GetUserByRemoteAddr(remoteAddr string) (*dm.User, error) {
	row := db.conn.QueryRow("SELECT id, nickname, remote_addr FROM users WHERE remote_addr = $1", remoteAddr)

	user := &dm.User{}
	if err := row.Scan(&user.ID, &user.Nickname, &user.RemoteAddr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by remote address: %w", err)
	}

	return user, nil
}

func (db *DB) SaveRoom(room *dm.Room) error {
	_, err := db.conn.Exec("INSERT INTO rooms (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", room.ID, room.Name)
	if err != nil {
		return fmt.Errorf("error saving room: %w", err)
	}

	return nil
}

func (db *DB) ListRooms() ([]*dm.Room, error) {
	rows, err := db.conn.Query("SELECT id, name FROM rooms")
	if err != nil {
		return nil, fmt.Errorf("error listing rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*dm.Room
	for rows.Next() {
		room := &dm.Room{}
		if err := rows.Scan(&room.ID, &room.Name); err != nil {
			return nil, fmt.Errorf("error scanning room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (db *DB) SaveMessage(message *dm.Message) error {
	_, err := db.conn.Exec("INSERT INTO messages (id, user_id, room_id, content) VALUES ($1, $2, $3, $4)", message.ID, message.UserID, message.RoomID, message.Content)
	if err != nil {
		return fmt.Errorf("error saving message: %w", err)
	}

	return nil
}

func (db *DB) ListMessagesByRoom(roomID uuid.UUID, limit int) ([]*dm.Message, error) {
	rows, err := db.conn.Query("SELECT id, user_id, room_id, content FROM messages WHERE room_id = $1 ORDER BY id DESC LIMIT $2", roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing messages by room: %w", err)
	}
	defer rows.Close()

	var messages []*dm.Message
	for rows.Next() {
		message := &dm.Message{}
		if err := rows.Scan(&message.ID, &message.UserID, &message.RoomID, &message.Content); err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}
