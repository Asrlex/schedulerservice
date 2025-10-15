package db

import (
		"database/sql"
		"encoding/json"
		"fmt"
		"os"
		"time"

		"github.com/Asrlex/schedulerservice/internal/metrics"

		_ "github.com/mattn/go-sqlite3"
)

var (
	db  *sql.DB
	err error
)

const dbPath = "jobs.db"
const dbTablesPath = "internal/db/db-tables.json"

// InitDBConnection initializes the database connection
func InitDBConnection() error {
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		
		CreateDBTables()
		ScheduleDBPing()
		return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
		if db == nil {
			InitDBConnection()
		}
		return db
}

// CloseDB closes the database connection
func CloseDB() error {
		if db != nil {
				return db.Close()
		}
		return nil
}

// PingDB pings the database to check its availability
func PingDB() error {
		if db != nil {
				return db.Ping()
		}
		return fmt.Errorf("database not initialized")
}

// ScheduleDBPing periodically checks the database connection
func ScheduleDBPing() error {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				if pingErr := PingDB(); pingErr != nil {
					fmt.Printf("PingDB error: %v\n", pingErr)
				} else {
					fmt.Println("PingDB successful")
				}
			}
		}()
		return nil
}

// CreateDBTables creates the necessary tables in the database
func CreateDBTables() error {
		// Read the db-tables.json file
    data, err := os.ReadFile(dbTablesPath)
    if err != nil {
        return fmt.Errorf("failed to read db-tables.json: %w", err)
    }

    // Parse the JSON data
    var tables struct {
        Tables []struct {
            Name string `json:"name"`
            SQL  string `json:"sql"`
        } `json:"tables"`
    }
    if err := json.Unmarshal(data, &tables); err != nil {
        return fmt.Errorf("failed to parse db-tables.json: %w", err)
    }

    // Execute each table creation query
    for _, table := range tables.Tables {
        if _, err := db.Exec(table.SQL); err != nil {
            return fmt.Errorf("failed to create table %q: %w", table.Name, err)
        }
        fmt.Printf("Table %q created successfully\n", table.Name)
    }

    return nil
}

func UpdateMetric(name metrics.MetricName, value float64) error {
		if db == nil {
        fmt.Printf("[WARN] Database connection not set for metrics")
        return fmt.Errorf("database connection not set")
    }

    _, err := db.Exec(`
        UPDATE metrics SET
        metric_value = metric_value + ?,
        recorded_at = CURRENT_TIMESTAMP
        WHERE metric_name = ?
    `, value, string(name))
		return err
}

// SeedDB seeds the database with initial data
func SeedDB() error {
		return nil
}
