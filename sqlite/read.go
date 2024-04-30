package sqlite

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hoshinonyaruko/gensokyo-dashboard/config"
	"github.com/hoshinonyaruko/gensokyo-dashboard/structs"
)

// RobotStatus represents the structure corresponding to the robot_status table
type RobotStatus struct {
	SelfID          int64  `json:"self_id"`
	MessageReceived int    `json:"message_received"`
	MessageSent     int    `json:"message_sent"`
	LastMessageTime int64  `json:"last_message_time"`
	InvitesReceived int    `json:"invites_received"`
	KicksReceived   int    `json:"kicks_received"`
	DailyDAU        int    `json:"daily_dau"`
	Nickname        string `json:"nickname"`
	ImgHead         string `json:"imgHead"`
	IsOnline        bool   `json:"isOnline"`
}

// FetchOnlineRobots returns a JSON array of all online robots' statuses for the current day
// 返回机器人信息，会返回当日所有机器人，包括不在线的
func FetchOnlineRobots(db *sql.DB, cfg *config.Config) ([]byte, error) {
	currentDate := time.Now().Format("2006-01-02") // Get the current date in YYYY-MM-DD format

	// Adjust the query to select only today's entries and directly use the 'online' column.
	query := `SELECT self_id, message_received, message_sent, last_message_time, invites_received, kicks_received, daily_dau, online
              FROM robot_status
              WHERE date = ?` // Only fetch entries for the current date
	rows, err := db.Query(query, currentDate)
	if err != nil {
		return nil, fmt.Errorf("error querying robots for today: %w", err)
	}
	defer rows.Close()

	var robots []RobotStatus
	onlineBots := make(map[string]bool) // Map to track which bots are online
	for rows.Next() {
		var robot RobotStatus
		err = rows.Scan(&robot.SelfID, &robot.MessageReceived, &robot.MessageSent, &robot.LastMessageTime,
			&robot.InvitesReceived, &robot.KicksReceived, &robot.DailyDAU, &robot.IsOnline)
		if err != nil {
			return nil, fmt.Errorf("error reading robot status rows: %w", err)
		}

		// Map robot info from config and handle the image
		botFound := false
		for _, botInfo := range cfg.BotInfos {
			if fmt.Sprintf("%d", robot.SelfID) == botInfo.BotID {
				robot.Nickname = botInfo.BotNickname
				imgData, imgErr := os.ReadFile(botInfo.BotHead)
				if imgErr == nil {
					robot.ImgHead = base64.StdEncoding.EncodeToString(imgData)
				} else {
					fmt.Printf("Error reading image file: %v\n", imgErr)
					robot.ImgHead = "" // Use an empty string if the image cannot be loaded
				}
				botFound = true
				break
			}
		}

		// If bot not found in config, initialize with default settings
		if !botFound {
			newBotInfo := config.BotInfo{
				BotID:       fmt.Sprintf("%d", robot.SelfID),
				BotNickname: "NewBot-" + fmt.Sprintf("%d", robot.SelfID),
				BotHead:     "images/head.gif", // Default image path
			}
			cfg.BotInfos = append(cfg.BotInfos, newBotInfo)
			config.WriteConfigToFile(*cfg)
		}

		onlineBots[fmt.Sprintf("%d", robot.SelfID)] = true // Mark this bot as online
		robots = append(robots, robot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over robot status rows for today: %w", err)
	}

	// Append offline robots
	for _, botInfo := range cfg.BotInfos {
		if !onlineBots[botInfo.BotID] {
			imgData, _ := os.ReadFile(botInfo.BotHead)
			botid64, _ := strconv.ParseInt(botInfo.BotID, 10, 64)
			robots = append(robots, RobotStatus{
				SelfID:          botid64,
				MessageReceived: 0,
				MessageSent:     0,
				LastMessageTime: 0,
				InvitesReceived: 0,
				KicksReceived:   0,
				DailyDAU:        0,
				Nickname:        botInfo.BotNickname,
				ImgHead:         base64.StdEncoding.EncodeToString(imgData),
				IsOnline:        false,
			})
		}
	}

	jsonData, err := json.Marshal(robots)
	if err != nil {
		return nil, fmt.Errorf("error marshaling today's robot statuses to JSON: %w", err)
	}

	return jsonData, nil
}

// FetchFieldValuesForRobot queries the robot_status table for a specified number of past days for a given field type.
// 根据机器人id 需要的数据天数 数据类型，获取数据 数据类型=表的列名
func FetchFieldValuesForRobot(db *sql.DB, selfID int64, days int, fieldType string) ([]string, error) {
	// Calculate the start date for the query.
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	query := fmt.Sprintf(`SELECT %s FROM robot_status WHERE self_id = ? AND date BETWEEN ? AND ? ORDER BY date DESC`, fieldType)
	rows, err := db.Query(query, selfID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("Error querying robot_status: %v", err)
		return nil, fmt.Errorf("error querying robot_status: %w", err)
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		if err != nil {
			log.Printf("Error reading rows: %v", err)
			return nil, fmt.Errorf("error reading rows: %w", err)
		}
		values = append(values, value)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return values, nil
}

func FetchAllFieldsForRobot(db *sql.DB, selfID int64, days int) ([]structs.RobotStatus, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	query := `SELECT self_id, date, online, message_received, message_sent, last_message_time,
              invites_received, kicks_received, daily_dau 
              FROM robot_status 
              WHERE self_id = ? AND date BETWEEN ? AND ? ORDER BY date DESC`
	rows, err := db.Query(query, selfID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("Error querying robot_status: %v", err)
		return nil, fmt.Errorf("error querying robot_status: %w", err)
	}
	defer rows.Close()

	var robots []structs.RobotStatus
	for rows.Next() {
		var robot structs.RobotStatus
		var date time.Time // Use time.Time for proper date handling
		err = rows.Scan(&robot.SelfID, &date, &robot.Online, &robot.MessageReceived, &robot.MessageSent,
			&robot.LastMessageTime, &robot.InvitesReceived, &robot.KicksReceived, &robot.DailyDAU)
		if err != nil {
			log.Printf("Error reading robot status rows: %v", err)
			return nil, fmt.Errorf("error reading robot status rows: %w", err)
		}
		robot.Date = date.Format("2006-01-02") // Format the date as "YYYY-MM-DD"
		robots = append(robots, robot)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return robots, nil
}

type CommandStat struct {
	CommandName       string `json:"command_name"`
	SelfID            int64  `json:"self_id"`
	TotalCalls        int    `json:"total_calls"`
	LastCallTimestamp int64  `json:"last_call_timestamp"`
}

func FetchTopCommands(db *sql.DB, selfId int64, rank int) ([]CommandStat, error) {
	query := `SELECT command_name, self_id, total_calls, last_call_timestamp 
              FROM command_stats 
              WHERE self_id = ? 
              ORDER BY total_calls DESC 
              LIMIT ?`
	rows, err := db.Query(query, selfId, rank)
	if err != nil {
		log.Printf("Error querying top commands for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error querying top commands for selfId %d: %w", selfId, err)
	}
	defer rows.Close()

	var results []CommandStat
	for rows.Next() {
		var stat CommandStat
		if err := rows.Scan(&stat.CommandName, &stat.SelfID, &stat.TotalCalls, &stat.LastCallTimestamp); err != nil {
			log.Printf("Error reading command stats for selfId %d: %v", selfId, err)
			return nil, fmt.Errorf("error reading command stats for selfId %d: %w", selfId, err)
		}
		results = append(results, stat)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error during rows iteration for selfId %d: %w", selfId, err)
	}

	return results, nil
}

func FetchTopDailyCommands(db *sql.DB, selfId int64, date time.Time, rank int) ([]CommandStat, error) {
	query := `SELECT command_name, self_id, calls, last_call_timestamp 
              FROM daily_command_stats 
              WHERE self_id = ? AND date = ? 
              ORDER BY calls DESC 
              LIMIT ?`
	rows, err := db.Query(query, selfId, date.Format("2006-01-02"), rank)
	if err != nil {
		log.Printf("Error querying daily top commands for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error querying daily top commands for selfId %d: %w", selfId, err)
	}
	defer rows.Close()

	var results []CommandStat
	for rows.Next() {
		var stat CommandStat
		if err := rows.Scan(&stat.CommandName, &stat.SelfID, &stat.TotalCalls, &stat.LastCallTimestamp); err != nil {
			log.Printf("Error reading daily command stats for selfId %d: %v", selfId, err)
			return nil, fmt.Errorf("error reading daily command stats for selfId %d: %w", selfId, err)
		}
		results = append(results, stat)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error during rows iteration for selfId %d: %w", selfId, err)
	}

	return results, nil
}

type GroupStat struct {
	GroupID                int64  `json:"group_id"`
	SelfID                 int64  `json:"self_id"`
	TotalMessagesSent      int    `json:"total_messages_sent,omitempty"`
	LastMessageTimestamp   int64  `json:"last_message_timestamp,omitempty"`
	ConsecutiveMessageDays int    `json:"consecutive_message_days,omitempty"`
	MessagesSent           int    `json:"messages_sent,omitempty"`  // For daily stats
	ActiveMembers          int    `json:"active_members,omitempty"` // For daily stats
	Date                   string `json:"date,omitempty"`           // Only for daily stats
}

func FetchTopGroups(db *sql.DB, selfId int64, rank int) ([]GroupStat, error) {
	query := `SELECT group_id, self_id, total_messages_sent, last_message_timestamp, consecutive_message_days 
              FROM group_stats 
              WHERE self_id = ? 
              ORDER BY total_messages_sent DESC 
              LIMIT ?`
	rows, err := db.Query(query, selfId, rank)
	if err != nil {
		log.Printf("Error querying top groups for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error querying top groups for selfId %d: %w", selfId, err)
	}
	defer rows.Close()

	var results []GroupStat
	for rows.Next() {
		var stat GroupStat
		if err := rows.Scan(&stat.GroupID, &stat.SelfID, &stat.TotalMessagesSent, &stat.LastMessageTimestamp, &stat.ConsecutiveMessageDays); err != nil {
			log.Printf("Error reading group stats for selfId %d: %v", selfId, err)
			return nil, fmt.Errorf("error reading group stats for selfId %d: %w", selfId, err)
		}
		results = append(results, stat)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error during rows iteration for selfId %d: %w", selfId, err)
	}

	return results, nil
}

func FetchTopDailyGroups(db *sql.DB, selfId int64, date time.Time, rank int) ([]GroupStat, error) {
	// Updated SQL query to include selfId in the WHERE clause
	query := `SELECT group_id, self_id, messages_sent, active_members, date 
              FROM daily_group_stats 
              WHERE self_id = ? AND date = ? 
              ORDER BY messages_sent DESC 
              LIMIT ?`
	// Pass selfId along with date and rank to the query
	rows, err := db.Query(query, selfId, date.Format("2006-01-02"), rank)
	if err != nil {
		log.Printf("Error querying daily top groups for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error querying daily top groups for selfId %d: %w", selfId, err)
	}
	defer rows.Close()

	var results []GroupStat
	for rows.Next() {
		var stat GroupStat
		if err := rows.Scan(&stat.GroupID, &stat.SelfID, &stat.MessagesSent, &stat.ActiveMembers, &stat.Date); err != nil {
			log.Printf("Error reading daily group stats for selfId %d: %v", selfId, err)
			return nil, fmt.Errorf("error reading daily group stats for selfId %d: %w", selfId, err)
		}
		results = append(results, stat)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error during rows iteration for selfId %d: %w", selfId, err)
	}

	return results, nil
}

type UserStat struct {
	UserID                 int64  `json:"user_id"`
	SelfID                 int64  `json:"self_id"`
	Nickname               string `json:"nickname"`
	Role                   string `json:"role"`
	TotalMessagesSent      int    `json:"total_messages_sent,omitempty"`
	LastMessageTimestamp   int64  `json:"last_message_timestamp,omitempty"`
	ConsecutiveMessageDays int    `json:"consecutive_message_days,omitempty"`
	MessagesSent           int    `json:"messages_sent,omitempty"`           // For daily stats
	IncludedInGroupCount   bool   `json:"included_in_group_count,omitempty"` // For daily stats
	Date                   string `json:"date,omitempty"`                    // Only for daily stats
}

func FetchTopUsers(db *sql.DB, selfId int64, rank int) ([]UserStat, error) {
	query := `SELECT user_id, self_id, nickname, role, total_messages_sent, last_message_timestamp, consecutive_message_days 
              FROM user_stats 
              WHERE self_id = ? 
              ORDER BY total_messages_sent DESC 
              LIMIT ?`
	rows, err := db.Query(query, selfId, rank)
	if err != nil {
		log.Printf("Error querying top users for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error querying top users for selfId %d: %w", selfId, err)
	}
	defer rows.Close()

	var results []UserStat
	for rows.Next() {
		var stat UserStat
		if err := rows.Scan(&stat.UserID, &stat.SelfID, &stat.Nickname, &stat.Role, &stat.TotalMessagesSent, &stat.LastMessageTimestamp, &stat.ConsecutiveMessageDays); err != nil {
			log.Printf("Error reading user stats for selfId %d: %v", selfId, err)
			return nil, fmt.Errorf("error reading user stats for selfId %d: %w", selfId, err)
		}
		results = append(results, stat)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error during rows iteration for selfId %d: %w", selfId, err)
	}

	return results, nil
}

func FetchTopDailyUsers(db *sql.DB, selfId int64, date time.Time, rank int) ([]UserStat, error) {
	query := `SELECT user_id, self_id, nickname, role, messages_sent, last_message_timestamp, included_in_group_count, date 
              FROM daily_user_stats 
              WHERE self_id = ? AND date = ? 
              ORDER BY messages_sent DESC 
              LIMIT ?`
	rows, err := db.Query(query, selfId, date.Format("2006-01-02"), rank)
	if err != nil {
		log.Printf("Error querying daily top users for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error querying daily top users for selfId %d: %w", selfId, err)
	}
	defer rows.Close()

	var results []UserStat
	for rows.Next() {
		var stat UserStat
		if err := rows.Scan(&stat.UserID, &stat.SelfID, &stat.Nickname, &stat.Role, &stat.MessagesSent, &stat.LastMessageTimestamp, &stat.IncludedInGroupCount, &stat.Date); err != nil {
			log.Printf("Error reading daily user stats for selfId %d: %v", selfId, err)
			return nil, fmt.Errorf("error reading daily user stats for selfId %d: %w", selfId, err)
		}
		results = append(results, stat)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration for selfId %d: %v", selfId, err)
		return nil, fmt.Errorf("error during rows iteration for selfId %d: %w", selfId, err)
	}

	return results, nil
}
