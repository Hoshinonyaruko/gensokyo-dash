package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hoshinonyaruko/gensokyo-dashboard/structs"
)

// lastActivityMap stores the last activity time for each self_id.
var lastActivityMap sync.Map

func init() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			checkAndUpdateRobotStatus()
		}
	}()
}

func checkAndUpdateRobotStatus() {
	now := time.Now()
	lastActivityMap.Range(func(key, value interface{}) bool {
		selfID := key.(int64)
		lastActivity := value.(time.Time)
		if now.Sub(lastActivity) > time.Minute { // Check if last activity was more than a minute ago
			updateRobotOffline(selfID) // Update online status to false
		}
		return true
	})
}

func updateRobotOffline(selfID int64) {
	db, err := sql.Open("sqlite3", "file:mydb.sqlite?cache=shared&mode=rwc")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE robot_status SET online = ? WHERE self_id = ?", false, selfID)
	if err != nil {
		log.Printf("Error setting robot status offline: %v", err)
	}
}

// 用于解析RawMessage以提取指令名
func parseCommandName(rawMessage string) string {
	parts := strings.SplitN(rawMessage, " ", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return rawMessage
}

// 处理消息事件
func ProcessMessageEvent(db *sql.DB, event structs.MessageEvent) error {
	// 当前时间戳和日期，提前计算
	currentDate := time.Unix(event.Time, 0).Format("2006-01-02")

	// 更新或插入每日用户统计
	dailyUserSQL := `
	INSERT INTO daily_user_stats 
		(user_id, self_id, date, nickname, role, messages_sent, last_message_timestamp, included_in_group_count)
	VALUES 
		(?, ?, ?, ?, ?, 1, ?, TRUE)
	ON CONFLICT(user_id, date) DO UPDATE SET
		messages_sent = daily_user_stats.messages_sent + 1,
		last_message_timestamp = excluded.last_message_timestamp,
		included_in_group_count = CASE
			WHEN strftime('%Y-%m-%d', daily_user_stats.last_message_timestamp, 'unixepoch') != strftime('%Y-%m-%d', excluded.last_message_timestamp, 'unixepoch') THEN TRUE
			ELSE FALSE
		END
	`
	if _, err := db.Exec(dailyUserSQL, event.UserID, event.SelfID, currentDate, event.Sender.Nickname, event.Sender.Role, event.Time, event.Time); err != nil {
		log.Printf("Error updating daily user stats: %v", err)
		return err
	}

	// 更新总用户统计
	userSQL := `
	 INSERT INTO user_stats (user_id, self_id, nickname, role, total_messages_sent, last_message_timestamp, consecutive_message_days)
	 VALUES (?, ?, ?, ?, 1, ?, 1)
	 ON CONFLICT(user_id) DO UPDATE SET
		 nickname = excluded.nickname,
		 role = excluded.role,
		 total_messages_sent = user_stats.total_messages_sent + 1,
		 last_message_timestamp = excluded.last_message_timestamp,
		 consecutive_message_days = CASE 
			 WHEN strftime('%Y-%m-%d', user_stats.last_message_timestamp, 'unixepoch', '+1 day') = strftime('%Y-%m-%d', ?, 'unixepoch') THEN user_stats.consecutive_message_days + 1 
			 ELSE 1 
		 END
	 `
	if _, err := db.Exec(userSQL, event.UserID, event.SelfID, event.Sender.Nickname, event.Sender.Role, event.Time, currentDate); err != nil {
		log.Printf("Error updating user stats: %v", err)
		return err
	}

	// 获取用户的 included_in_group_count 状态
	var includedInGroupCount bool
	err := db.QueryRow("SELECT included_in_group_count FROM daily_user_stats WHERE user_id = ?", event.UserID).Scan(&includedInGroupCount)
	if err != nil {
		log.Printf("Error fetching included_in_group_count: %v", err)
		return fmt.Errorf("error fetching included_in_group_count: %v", err)
	}

	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	// 尝试在事务中完成所有数据库操作
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()

	// 插入或更新消息到 messages 表
	messageSQL := `
    INSERT INTO messages (message_id, message_type, time, self_id, raw_message, user_id, group_id, message_date)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(message_id) DO UPDATE SET
        message_type = excluded.message_type,
        time = excluded.time,
        raw_message = excluded.raw_message,
        user_id = excluded.user_id,
        group_id = excluded.group_id,
        message_date = excluded.message_date;`
	if _, err = tx.Exec(messageSQL, event.MessageID, event.MessageType, event.Time, event.SelfID, event.RawMessage, event.UserID, event.GroupID, currentDate); err != nil {
		return fmt.Errorf("error inserting message: %v", err)
	}

	// 处理指令统计
	commandName := parseCommandName(event.RawMessage)

	// 更新总指令统计
	commandTotalSQL := `
    INSERT INTO command_stats (command_name, self_id, total_calls, last_call_timestamp)
    VALUES (?, ?, 1, ?)
    ON CONFLICT(command_name, self_id) DO UPDATE SET
        total_calls = command_stats.total_calls + 1,
        last_call_timestamp = ?;`
	if _, err := tx.Exec(commandTotalSQL, commandName, event.SelfID, event.Time, event.Time); err != nil {
		return fmt.Errorf("error updating command total stats: %v", err)
	}

	// 更新每日指令统计
	commandDailySQL := `
    INSERT INTO daily_command_stats (command_name, self_id, date, calls, last_call_timestamp)
    VALUES (?, ?, ?, 1, ?)
    ON CONFLICT(command_name, self_id, date) DO UPDATE SET
        calls = daily_command_stats.calls + 1,
        last_call_timestamp = ?;`
	if _, err := tx.Exec(commandDailySQL, commandName, event.SelfID, currentDate, event.Time, event.Time); err != nil {
		return fmt.Errorf("error updating daily command stats: %v", err)
	}

	// 更新 群发信息条数 每日
	updateMessagesSQL := `
	INSERT INTO daily_group_stats (group_id, self_id, date, messages_sent)
	VALUES (?, ?, ?, 1)
	ON CONFLICT(group_id, date) DO UPDATE SET
		messages_sent = daily_group_stats.messages_sent + 1;
	`
	if _, err := tx.Exec(updateMessagesSQL, event.GroupID, event.SelfID, currentDate); err != nil {
		return fmt.Errorf("error updating messages sent in daily group stats: %v", err)
	}

	// 更新 群发信息条数 总

	updateTotalMessagesSQL := `
	INSERT INTO group_stats (group_id, self_id, total_messages_sent, last_message_timestamp)
	VALUES (?, ?, 1, ?)
	ON CONFLICT(group_id) DO UPDATE SET
		total_messages_sent = group_stats.total_messages_sent + 1,
		last_message_timestamp = excluded.last_message_timestamp;
	`
	if _, err := tx.Exec(updateTotalMessagesSQL, event.GroupID, event.SelfID, event.Time); err != nil {
		return fmt.Errorf("error updating total messages sent in group stats: %v", err)
	}

	// 下方分支 每个用户每天仅第一次调用会统计
	if includedInGroupCount {

		// 更新每日群组日活统计
		updateActiveMembersSQL := `
		INSERT INTO daily_group_stats (group_id, self_id, date, active_members)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(group_id, date) DO UPDATE SET
			active_members = daily_group_stats.active_members + 1;
		`
		if _, err := tx.Exec(updateActiveMembersSQL, event.GroupID, event.SelfID, currentDate); err != nil {
			return fmt.Errorf("error updating active members in daily group stats: %v", err)
		}

		// 更新累积群组连续活跃天数统计
		updateConsecutiveDaysSQL := `
		UPDATE group_stats
		SET consecutive_message_days = CASE
			WHEN date(julianday(?, 'unixepoch')) = date(julianday(last_message_timestamp, 'unixepoch') + 1) THEN consecutive_message_days + 1
			WHEN date(julianday(?, 'unixepoch')) > date(julianday(last_message_timestamp, 'unixepoch') + 1) THEN 1
			ELSE consecutive_message_days
		END
		WHERE group_id = ?;`

		if _, err := tx.Exec(updateConsecutiveDaysSQL, event.Time, event.Time, event.GroupID); err != nil {
			return fmt.Errorf("error updating consecutive message days in group stats: %v", err)
		}

		// 更新机器人状态表，每接收到一个当日新用户，活跃度（daily_dau）加1
		updateRobotStatsSQL := `
		UPDATE robot_status
		SET
			daily_dau = daily_dau + 1,
			last_message_time = ?
		WHERE self_id = ? AND date = ?;`

		if result, err := tx.Exec(updateRobotStatsSQL, event.Time, event.SelfID, currentDate); err != nil {
			return fmt.Errorf("error updating robot status: %v", err)
		} else if affected, _ := result.RowsAffected(); affected == 0 {
			// 插入新记录，因为当天没有现有记录
			insertSQL := `
        INSERT INTO robot_status (self_id, date, online, message_received, message_sent, last_message_time, daily_dau)
        VALUES (?, ?, TRUE, 0, 0, ?, 1)`
			if _, err = tx.Exec(insertSQL, event.SelfID, currentDate, event.Time); err != nil {
				return fmt.Errorf("error inserting new robot status: %v", err)
			}
		}

		// 更新用户 included_in_group_count 状态为 FALSE 下次不运行includedInGroupCount分支
		updateUserGroupCountSQL := "UPDATE daily_user_stats SET included_in_group_count = FALSE WHERE user_id = ?"
		if _, err := tx.Exec(updateUserGroupCountSQL, event.UserID); err != nil {
			return fmt.Errorf("error updating user included_in_group_count: %v", err)
		}
	}

	return nil
}

// ProcessMetaEvent updates or inserts the robot status in the database based on MetaEvent data.
func ProcessMetaEvent(db *sql.DB, event structs.MetaEvent) error {
	currentDate := time.Now().Format("2006-01-02") // Get current date in YYYY-MM-DD format

	// Use INSERT OR REPLACE to handle the primary key constraint of self_id and date
	upsertSQL := `
    INSERT OR REPLACE INTO robot_status (
        self_id, date, online, message_received, message_sent, last_message_time)
    VALUES (?, ?, ?, ?, ?, ?);`

	_, err := db.Exec(upsertSQL,
		event.SelfID,
		currentDate,
		event.Status.Online,
		event.Status.Stat.MessageReceived,
		event.Status.Stat.MessageSent,
		event.Status.Stat.LastMessageTime)
	if err != nil {
		log.Printf("Error upserting robot status: %v", err)
		return fmt.Errorf("error upserting robot status: %w", err)
	}

	log.Println("Upserted robot status successfully for SelfID:", event.SelfID)
	return nil
}

// ProcessNoticeEvent 基于事件记录机器人信息
func ProcessNoticeEvent(db *sql.DB, event structs.NoticeEvent) error {
	currentDate := time.Now().Format("2006-01-02") // 获取当前日期

	if event.NoticeType == "group_increase" && event.SubType == "invite" {
		// 当收到邀请通知时增加邀请次数
		updateSQL := `
        UPDATE robot_status
        SET invites_received = invites_received + 1
        WHERE self_id = ? AND date = ?;`
		_, err := db.Exec(updateSQL, event.SelfID, currentDate)
		if err != nil {
			log.Printf("Error updating invites received count: %v", err)
			return fmt.Errorf("error updating invites received count: %w", err)
		}
		log.Println("Updated invites received count successfully for SelfID:", event.SelfID)
	} else if event.NoticeType == "group_decrease" && event.SubType == "kick_me" {
		// 当收到被踢通知时增加被踢次数
		updateSQL := `
        UPDATE robot_status
        SET kicks_received = kicks_received + 1
        WHERE self_id = ? AND date = ?;`
		_, err := db.Exec(updateSQL, event.SelfID, currentDate)
		if err != nil {
			log.Printf("Error updating kicks received count: %v", err)
			return fmt.Errorf("error updating kicks received count: %w", err)
		}
		log.Println("Updated kicks received count successfully for SelfID:", event.SelfID)
	}

	return nil
}
