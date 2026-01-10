package database

import ( 
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	dsn := "host=localhost user=myuser password=mysecretpassword dbname=mydatabase port=5432 sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Database connection established")
	return db, nil
}
