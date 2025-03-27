package database

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
    var err error
    DB, err = sql.Open("sqlite3", "./sqlite.db")
    if err != nil {
        return err
    }

    err = createTables()
    if err != nil {
        return err
    }

    return nil
}

func createTables() error {
    userTable := `
    CREATE TABLE IF NOT EXISTS user (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        contact_type TEXT NOT NULL,
        contact TEXT NOT NULL
    );`

    requestTable := `
    CREATE TABLE IF NOT EXISTS user_requests (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        request_method TEXT NOT NULL,
        tokens_in INTEGER,
        tokens_out INTEGER,
        date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(user_id) REFERENCES user(id)
    );`

    _, err := DB.Exec(userTable)
    if err != nil {
        return err
    }

    _, err = DB.Exec(requestTable)
    return err
}
