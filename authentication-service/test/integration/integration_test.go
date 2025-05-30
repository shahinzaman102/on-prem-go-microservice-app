package integration

import (
	"authentication/api"
	"authentication/data"
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v4/stdlib" // Ensure this is imported to set up the Postgres driver
)

var mockDB *sql.DB
var mock sqlmock.Sqlmock
var app api.Config

func TestMain(m *testing.M) {
	var err error
	// Set up the mock database
	mockDB, mock, err = sqlmock.New(sqlmock.MonitorPingsOption(true)) // Enable ping monitoring
	if err != nil {
		panic("failed to create mock database: " + err.Error())
	}

	// Initialize the app with the mock database
	app = api.Config{
		DB:     mockDB,
		Models: data.New(mockDB),
	}

	code := m.Run()
	mockDB.Close() // Close the mock database
	os.Exit(code)
}

func TestDatabaseConnectivity(t *testing.T) {
	// Mocking a successful database connection using mock.Expectations
	mock.ExpectPing().WillReturnError(nil) // Mock a successful Ping

	// Now we test the actual pinging method
	err := mockDB.Ping()
	if err != nil {
		t.Fatalf("failed to ping the database: %v", err)
	}

	// Ensure all expectations are met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}
