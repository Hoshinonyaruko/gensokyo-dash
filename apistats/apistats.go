package apistats

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hoshinonyaruko/gensokyo-dashboard/config"
)

type APIStatus struct {
	APIPaths        string  `json:"apiPaths"`
	APINames        string  `json:"apiNames"`
	Online          bool    `json:"online"`
	ResponseTime    int     `json:"responseTime,omitempty"`
	ChecksPerformed int     `json:"checksPerformed,omitempty"`
	ChecksFailed    int     `json:"checksFailed,omitempty"`
	SuccessRate     float64 `json:"successRate,omitempty"`
	Date            string  `json:"date"`
}

// MonitorAPIs regularly checks the API endpoints and updates the database.
func MonitorAPIs(db *sql.DB, cfg config.Config) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		log.Printf("Monitoring started, monitoring %d APIs\n", len(cfg.ApisInfos))

		// Wait for the first tick to fire immediately, confirming that the ticker works.
		<-ticker.C

		for t := range ticker.C {
			log.Printf("Ticker triggered at %v", t)
			if len(cfg.ApisInfos) == 0 {
				log.Println("No APIs to monitor.")
				continue
			}

			today := time.Now().Format("2006-01-02")
			for _, api := range cfg.ApisInfos {
				fmt.Printf("Checking API: %s\n", api.APIPaths)
				response, err := http.Get(api.APIPaths)
				if err != nil {
					log.Printf("Failed to reach API %s: %v", api.APINames, err)
					incrementAPIStatus(db, api.APIPaths, today, false)
					continue
				}

				// Handle response and close immediately.
				log.Printf("API %s is online, responded with status code: %d", api.APINames, response.StatusCode)
				incrementAPIStatus(db, api.APIPaths, today, true)
				response.Body.Close() // Close response body immediately after processing
			}
		}
	}()
}

func incrementAPIStatus(db *sql.DB, apiURL, date string, success bool) {
	var sqlStr string
	if success {
		// 更新已存在的行，或者插入一个新行，增加成功请求的次数
		sqlStr = `
        INSERT INTO api_status (api_url, date, online, checks_performed, response_time, checks_failed) 
        VALUES (?, ?, TRUE, 1, 1, 0)
        ON CONFLICT(api_url, date) DO UPDATE SET
            online = TRUE,
            checks_performed = checks_performed + 1,
            response_time = response_time + 1`
	} else {
		// 更新已存在的行，或者插入一个新行，增加失败的请求次数
		sqlStr = `
        INSERT INTO api_status (api_url, date, online, checks_performed, checks_failed) 
        VALUES (?, ?, FALSE, 1, 1)
        ON CONFLICT(api_url, date) DO UPDATE SET
            online = FALSE,
            checks_performed = checks_performed + 1,
            checks_failed = checks_failed + 1`
	}

	_, err := db.Exec(sqlStr, apiURL, date)
	if err != nil {
		log.Printf("Error updating or inserting API status for %s on %s: %v", apiURL, date, err)
	}
}

// FetchAPIStatuses fetches the status for all APIs listed in the config for the last N days.
func FetchAPIStatuses(db *sql.DB, cfg config.Config, days int) ([]APIStatus, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	var allStatuses []APIStatus

	for _, api := range cfg.ApisInfos {
		query := `
        SELECT date, online, response_time, checks_performed, checks_failed
        FROM api_status
        WHERE api_url = ? AND date BETWEEN ? AND ?
        ORDER BY date DESC`
		rows, err := db.Query(query, api.APIPaths, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Printf("Error querying api_status for %s: %v", api.APIPaths, err)
			continue // Skip to the next API if there's an error querying this one
		}

		for rows.Next() {
			var status APIStatus
			err = rows.Scan(&status.Date, &status.Online, &status.ResponseTime, &status.ChecksPerformed, &status.ChecksFailed)
			if err != nil {
				log.Printf("Error reading api status rows for %s: %v", api.APIPaths, err)
				break // Break out of this API's loop on error
			}
			status.APIPaths = api.APIPaths
			status.APINames = api.APINames
			if status.ChecksPerformed > 0 { // Calculate success rate if there were checks performed
				status.SuccessRate = float64(status.ChecksPerformed-status.ChecksFailed) / float64(status.ChecksPerformed) * 100
			}
			allStatuses = append(allStatuses, status)
		}
		rows.Close()
	}

	return allStatuses, nil
}
