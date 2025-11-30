package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() error {
    var err error
    db, err = sql.Open("sqlite3", "./database.db")
    if err != nil {
        return err
    }

    // 
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS worklogs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            date TEXT NOT NULL,
            description TEXT,
            hours REAL NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users(id)
        )
    `)
    
    return err
}
