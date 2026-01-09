package data

func Connect() {
	dsn := "host=localhost user=myuser password=mysecretpassword dbname=mydatabase port=5432 sslmode=disable"
	_ = dsn
}
func Disconnect() {

}
