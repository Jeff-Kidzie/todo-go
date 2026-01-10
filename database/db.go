package database

import ( 
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func Connect() *sql.DB {
	dsn := "host=localhost user=myuser password=mysecretpassword dbname=mydatabase port=5432 sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic("failed to connect database")
	}

	if err = db.Ping(); err != nil {
		panic("failed to ping database")
	}

	fmt.Println("Database connection established")
	return db
}
