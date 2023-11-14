package main

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "gchat"
	dbname   = "postgres"
)

type Storage interface {
  CreateUser(UserJSON) (*UserJSON, error)
  GetUserById(int) (*UserJSON, error)
  GetUserByUsername(string) (*UserJSON, error)
  UpdateUser(UserJSON) error
  DeleteUserById(int) error
  DeleteUserByUsername(string) error

  CreateChat(ChatJSON) (*ChatJSON, error)
  GetChatById(int) (*ChatJSON, error)
  UpdateChat(ChatJSON) error
  DeleteChatById(int) error

  CreateMessages([]MessageJSON) ([]MessageJSON, error)
  GetMessagesByChatId(int) ([]MessageJSON, error)
  DeleteMessagesByChatId(int) error
  DeleteMessagesByAuthorName(string) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		err = fmt.Errorf("sql open failed: %w", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		err = fmt.Errorf("ping failed: %w", err)
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	if err := s.CreateTableUsers(); err != nil {
		err = fmt.Errorf("create table users failed: %w", err)
		return err
	}

	if err := s.CreateTableChats(); err != nil {
		err = fmt.Errorf("create table chats failed: %w", err)
		return err
	}

	if err := s.CreateTableMessages(); err != nil {
		err = fmt.Errorf("create table messages failed: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) CreateTableUsers() error {
	query := `create table if not exists users(
    id serial primary key,
    username text,
    passwordHashed text,
    chatIds integer[]
  )`

	if _, err := s.db.Exec(query); err != nil {
		err = fmt.Errorf("create table users query failed: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) CreateTableChats() error {
	query := `create table if not exists chats(
    id serial primary key,
    usernames text[]
  )`

	if _, err := s.db.Exec(query); err != nil {
		err = fmt.Errorf("create table chats query failed: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) CreateTableMessages() error {
	query := `create table if not exists messages(
    id serial primary key,
    chatId integer,
    text text,
    authorName text,
    timestamp integer
  )`

	if _, err := s.db.Exec(query); err != nil {
		err = fmt.Errorf("create table messages query failed: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) CreateUser(user UserJSON) (*UserJSON, error) {
	query := `insert into users(username, passwordHashed, chatIds)
  values ($1, $2, $3) returning id`

	nullArray := []sql.NullInt64{}
	row := s.db.QueryRow(query, user.Username, user.PasswordHashed, pq.Array(nullArray))

	err := row.Scan(&user.Id)
	if err != nil {
		err = fmt.Errorf("scan error: %w", err)
		return nil, err
	}

	return &user, nil
}

func (s *PostgresStore) GetUserById(id int) (*UserJSON, error) {
	query := `select * from users where id = $1 limit 1`
	row := s.db.QueryRow(query, id)

	user := &UserJSON{}
	nullArray := []sql.NullInt64{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHashed, pq.Array(&nullArray))
	if err != nil {
		err = fmt.Errorf("scan error: %w", err)
		return nil, err
	}

	for _, v := range nullArray {
		if v.Valid {
			user.ChatIds = append(user.ChatIds, int(v.Int64))
		}
	}

	return user, nil
}

func (s *PostgresStore) GetUserByUsername(username string) (*UserJSON, error) {
	query := `select * from users where username = $1 limit 1`
	row := s.db.QueryRow(query, username)

	user := &UserJSON{}
	nullArray := []sql.NullInt64{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHashed, pq.Array(&nullArray))
	if err != nil {
		err = fmt.Errorf("scan error: %w", err)
		return nil, err
	}

	for _, v := range nullArray {
		if v.Valid {
			user.ChatIds = append(user.ChatIds, int(v.Int64))
		}
	}

	return user, nil
}

func (s *PostgresStore) UpdateUser(user UserJSON) error {
	query := `update users set username=$1, passwordHashed=$2, chatIds=$3 where id=$4`
	_, err := s.db.Exec(query, user.Username, user.PasswordHashed, pq.Array(user.ChatIds), user.Id)
	if err != nil {
		err = fmt.Errorf("update user query failed: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) DeleteUserById(id int) error {
	query := `delete from users where id=$1`
	_, err := s.db.Exec(query, id)
	if err != nil {
		err = fmt.Errorf("delete user query failed: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) DeleteUserByUsername(username string) error {
	query := `delete from users where username=$1`
	_, err := s.db.Exec(query, username)
	if err != nil {
		err = fmt.Errorf("delete user query failed: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) CreateChat(chat ChatJSON) (*ChatJSON, error) {
	query := `insert into chats(usernames) values($1) returning id`
	row := s.db.QueryRow(query, pq.Array(chat.Usernames))

	err := row.Scan(&chat.Id)
	if err != nil {
		err = fmt.Errorf("chats scan error: %w", err)
		return nil, err
	}

	return &chat, nil
}

func (s *PostgresStore) GetChatById(id int) (*ChatJSON, error) {
	query := `select usernames from chats where id=$1 limit 1`
	row := s.db.QueryRow(query, id)

	chat := &ChatJSON{Id: id}
	err := row.Scan(pq.Array(&chat.Usernames))
	if err != nil {
		err = fmt.Errorf("chats scan error: %w", err)
		return nil, err
	}

	return chat, nil
}

func (s *PostgresStore) UpdateChat(chat ChatJSON) error {
	query := `update chats set usernames=$1 where id=$2`
	_, err := s.db.Exec(query, pq.Array(chat.Usernames), chat.Id)
	if err != nil {
		err = fmt.Errorf("update chats query error: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) DeleteChatById(id int) error {
	query := `delete from chats id = $1`
	_, err := s.db.Exec(query, id)
	if err != nil {
		err = fmt.Errorf("delete chats query error: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) CreateMessages(messages []MessageJSON) ([]MessageJSON, error) {
  values := ""
	for i, msg := range messages {
		messageValues := fmt.Sprintf("(%v, %v, %v, %v)", msg.ChatId, msg.Text, msg.AuthorName, msg.Timestamp)

    if i != len(messages) - 1 {
      values = values + messageValues + ", "
    } else {
      values = values + messageValues 
    }
	}

	query := `insert into messages(chatId, text, authorName, timestamp) values($1) returning id`
  rows, err := s.db.Query(query, values)
  if err != nil {
    err = fmt.Errorf("insert messages query error: %w", err)
    return nil, err
  }
  defer rows.Close()

  msgIndex := 0
  for rows.Next() {
    err = rows.Scan(&messages[msgIndex].Id)
    if err != nil {
      err = fmt.Errorf("messages scan error: %w", err)
      return nil, err
    }

    msgIndex += 1
  }

  return messages, nil
}

func (s *PostgresStore) GetMessagesByChatId(id int) ([]MessageJSON, error) {
  query := `select * from messages where chatId=$1 order by timestamp desc`
	rows, err := s.db.Query(query, id)
	if err != nil {
		err = fmt.Errorf("messages query error: %w", err)
		return nil, err
	}
	defer rows.Close()

  messages := []MessageJSON{}
	for rows.Next() {
		message := MessageJSON{}
		err = rows.Scan(&message.Id, &message.ChatId, &message.Text, &message.AuthorName, &message.Timestamp)
		if err != nil {
			err = fmt.Errorf("messages scan error: %w", err)
			return nil, err
		}

    messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("messsages rows error: %w", err)
		return nil, err
	}

  return messages, nil
}

func (s *PostgresStore) DeleteMessagesByChatId(id int) error {
  query := `delete from messages where chatId=$1`
  _, err := s.db.Exec(query, id)
  if err != nil {
    err = fmt.Errorf("messages delete query error: %w", err)
    return err
  }

  return nil
}

func (s *PostgresStore) DeleteMessagesByAuthorName(name string) error {
  query := `delete from messages where authorName=$1`
  _, err := s.db.Exec(query, name)
  if err != nil {
    err = fmt.Errorf("messages delete query error: %w", err)
    return err
  }

  return nil
}
