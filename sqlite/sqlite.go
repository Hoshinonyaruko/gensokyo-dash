package sqlite

import (
	"database/sql"
	"fmt"
	"log"
)

// 网页登入cookie
func EnsureCookieTablesExist(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS cookies (
        cookie_id VARCHAR(36) PRIMARY KEY,
        expiration BIGINT NOT NULL
    );`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating cookies table: %w", err)
	}
	return nil
}

// 消息表
func EnsureMessagesTableExists(db *sql.DB) error {
	// Create the table with the appropriate data types and primary key settings
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS messages (
        message_id INTEGER PRIMARY KEY AUTOINCREMENT,
        message_type TEXT,
        time INTEGER,
        self_id INTEGER,
        raw_message TEXT,
        user_id INTEGER,
        group_id INTEGER,
        message_date DATE
    );`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Printf("Error creating messages table: %v", err)
		return fmt.Errorf("error creating messages table: %w", err)
	}

	// Create indexes separately
	indexesSQL := []string{
		"CREATE INDEX IF NOT EXISTS idx_message_type ON messages(message_type);",
		"CREATE INDEX IF NOT EXISTS idx_self_id ON messages(self_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_id ON messages(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_group_id ON messages(group_id);",
		"CREATE INDEX IF NOT EXISTS idx_message_date ON messages(message_date);",
	}
	for _, sql := range indexesSQL {
		if _, err := db.Exec(sql); err != nil {
			log.Printf("Error creating index: %v", err)
			return fmt.Errorf("error creating index: %w", err)
		}
	}

	log.Println("Ensured that messages table and indexes exist")
	return nil
}

// 机器人状态表
// EnsureRobotStatusTableExists creates or alters the robot_status table as necessary.
func EnsureRobotStatusTableExists(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS robot_status (
        self_id INTEGER,
        date DATE NOT NULL,
        online BOOLEAN NOT NULL,
        message_received INTEGER NOT NULL,
        message_sent INTEGER NOT NULL,
        last_message_time INTEGER,
        invites_received INTEGER DEFAULT 0,
        kicks_received INTEGER DEFAULT 0,
        daily_dau INTEGER DEFAULT 0,
        PRIMARY KEY (self_id, date)
    );`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Error creating robot_status table: %v", err)
		return fmt.Errorf("error creating robot_status table: %w", err)
	}
	log.Println("Ensured that robot_status table exists")
	return nil
}

// 用户表
// EnsureUserStatsTableExists creates or alters the user_stats and daily_user_stats tables as necessary.
func EnsureUserStatsTableExists(db *sql.DB) error {
	// Update user_stats table to store cumulative data only
	createCumulativeTableSQL := `
    CREATE TABLE IF NOT EXISTS user_stats (
        user_id INTEGER PRIMARY KEY,
        self_id BIGINT,
        nickname TEXT,
        role TEXT,
        total_messages_sent INTEGER DEFAULT 0,
        last_message_timestamp INTEGER,
		consecutive_message_days INTEGER DEFAULT 0
    );`
	_, err := db.Exec(createCumulativeTableSQL)
	if err != nil {
		log.Printf("Error creating cumulative user_stats table: %v", err)
		return fmt.Errorf("error creating cumulative user_stats table: %w", err)
	}

	// Create a new table for daily statistics
	createDailyTableSQL := `
    CREATE TABLE IF NOT EXISTS daily_user_stats (
        user_id INTEGER,
        self_id BIGINT,
        date DATE NOT NULL,
        nickname TEXT,
        role TEXT,
        messages_sent INTEGER DEFAULT 0,
		last_message_timestamp INTEGER,
        included_in_group_count BOOLEAN DEFAULT FALSE,
        PRIMARY KEY (user_id, date)
    );`
	_, err = db.Exec(createDailyTableSQL)
	if err != nil {
		log.Printf("Error creating daily user_stats table: %v", err)
		return fmt.Errorf("error creating daily user_stats table: %w", err)
	}

	// Create an index on the self_id field in the cumulative table
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_user_self_id ON user_stats (self_id);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		log.Printf("Error creating index on user_stats: %v", err)
		return fmt.Errorf("error creating index on user_stats: %w", err)
	}

	log.Println("Ensured that user_stats and daily_user_stats tables and index on self_id exist")
	return nil
}

// 群表
// EnsureGroupStatsTableExists creates or alters the group_stats and daily_group_stats tables as necessary.
func EnsureGroupStatsTableExists(db *sql.DB) error {
	// Update group_stats table to store cumulative data only
	createCumulativeTableSQL := `
    CREATE TABLE IF NOT EXISTS group_stats (
        group_id INTEGER PRIMARY KEY,
        self_id BIGINT,
        total_messages_sent INTEGER DEFAULT 0,
        last_message_timestamp INTEGER,
        consecutive_message_days INTEGER DEFAULT 0
    );`
	_, err := db.Exec(createCumulativeTableSQL)
	if err != nil {
		log.Printf("Error creating cumulative group_stats table: %v", err)
		return fmt.Errorf("error creating cumulative group_stats table: %w", err)
	}

	// Create a new table for daily statistics
	createDailyTableSQL := `
    CREATE TABLE IF NOT EXISTS daily_group_stats (
        group_id INTEGER,
        self_id BIGINT,
        date DATE NOT NULL,
        messages_sent INTEGER DEFAULT 0,
        active_members INTEGER DEFAULT 0,
        PRIMARY KEY (group_id, date)
    );`
	_, err = db.Exec(createDailyTableSQL)
	if err != nil {
		log.Printf("Error creating daily group_stats table: %v", err)
		return fmt.Errorf("error creating daily group_stats table: %w", err)
	}

	// Create an index on the self_id field in the cumulative table
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_group_self_id ON group_stats (self_id);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		log.Printf("Error creating index on group_stats: %v", err)
		return fmt.Errorf("error creating index on group_stats: %w", err)
	}

	log.Println("Ensured that group_stats and daily_group_stats tables and index on self_id exist")
	return nil
}

// 指令表
func EnsureCommandStatsTables(db *sql.DB) error {
	// Create the command_stats table
	createCommandStatsTableSQL := `
    CREATE TABLE IF NOT EXISTS command_stats (
        command_name TEXT,
        self_id BIGINT,
        total_calls INTEGER DEFAULT 0,
		last_call_timestamp INTEGER,
        PRIMARY KEY (command_name, self_id)
    );`
	if _, err := db.Exec(createCommandStatsTableSQL); err != nil {
		log.Printf("Error creating command_stats table: %v", err)
		return err
	}

	// Create index on the self_id field in the command_stats table
	createCommandStatsIndexSQL := `CREATE INDEX IF NOT EXISTS idx_command_self_id ON command_stats (self_id);`
	if _, err := db.Exec(createCommandStatsIndexSQL); err != nil {
		log.Printf("Error creating index on command_stats: %v", err)
		return err
	}

	// Create the daily_command_stats table
	createDailyCommandStatsTableSQL := `
    CREATE TABLE IF NOT EXISTS daily_command_stats (
        command_name TEXT,
        self_id BIGINT,
        date DATE NOT NULL,
        calls INTEGER DEFAULT 0,
        last_call_timestamp INTEGER,
        PRIMARY KEY (command_name, self_id, date)
    );`
	if _, err := db.Exec(createDailyCommandStatsTableSQL); err != nil {
		log.Printf("Error creating daily_command_stats table: %v", err)
		return err
	}

	// Create index on the date field in the daily_command_stats table
	createDailyCommandStatsIndexSQL := `CREATE INDEX IF NOT EXISTS idx_daily_command_date ON daily_command_stats (date);`
	if _, err := db.Exec(createDailyCommandStatsIndexSQL); err != nil {
		log.Printf("Error creating index on daily_command_stats: %v", err)
		return err
	}

	log.Println("Ensured that command_stats and daily_command_stats tables and their indexes exist")
	return nil
}

// api状态表
func EnsureAPITableExists(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS api_status (
        api_url TEXT NOT NULL,
        date DATE NOT NULL,
        online BOOLEAN NOT NULL,
        response_time INTEGER,
        checks_performed INTEGER DEFAULT 0,
        checks_failed INTEGER DEFAULT 0,
        PRIMARY KEY (api_url, date)
    );`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Error creating api_status table: %v", err)
		return fmt.Errorf("error creating api_status table: %w", err)
	}
	log.Println("Ensured that api_status table exists")
	return nil
}
