package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// パッケージ変数はあとでconfig化する
var DBSOURCE string = "./database.db"
var db *sql.DB

type Item struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

func main() {
	var err error

	// Open data base
	db, err = sql.Open("sqlite", DBSOURCE) // 後でMySQLにする
	if err != nil {
		log.Fatal(err)
	}

	// Init db
	err = initDB(db)
	if err != nil {
		log.Fatal(err)
	}

	// ハンドラの追加
	http.HandleFunc("/items/", handleItems)
	http.HandleFunc("/done/", handleDone)

	err = http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000") // localhost:3000からのアクセスを許可する
	switch r.Method {
	case "GET":
		sendItems(w)
	case "POST":
		addNewItems(w, r)
	case "DELETE":
		deleteItems(w, r)
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")               // Content-Typeヘッダの使用を許可する
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS") // pre-flightリクエストに対応する
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
}

func handleDone(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000") // localhost:3000からのアクセスを許可する
	switch r.Method {
	case "PUT":
		doneItems(w, r)
	case "DELETE":
		unDoneItems(w, r)
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")         // Content-Typeヘッダの使用を許可する
		w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE, OPTIONS") // pre-flightリクエストに対応する
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
}

func initDB(db *sql.DB) error {
	const sql = `
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT 0
		);`
	_, err := db.Exec(sql)
	return err
}

func routeParameter(r *http.Request, n int) (string, error) {
	splited := strings.Split(r.RequestURI, "/")
	var params []string
	for i := 0; i < len(splited); i++ {
		if len(splited[i]) != 0 {
			params = append(params, splited[i])
		}
	}

	if len(params) <= n {
		return "", errors.New("cannot find parameter")
	}
	return params[n], nil
}

func routeParameterInt(r *http.Request, n int) int {
	idStr, err := routeParameter(r, n)
	if err != nil {
		return -1
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1
	}
	return id
}

func sendItems(w http.ResponseWriter) {
	var items []Item
	rows, err := db.Query(`SELECT * FROM items;`)
	if err != nil {
		http.Error(w, "Internal Sever Error", http.StatusInternalServerError)
	}
	for rows.Next() {
		var item Item
		rows.Scan(&item.Id, &item.Name, &item.Done)
		items = append(items, item)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(items); err != nil {
		http.Error(w, "Internal Sever Error", http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
}

func addNewItems(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Name string `json:"name"`
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad Request", 400)
	}
	_, err = db.Exec(`INSERT INTO items (name, done) values (?, ?)`, reqBody.Name, false)
	if err != nil {
		log.Print(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

func deleteItems(w http.ResponseWriter, r *http.Request) {
	id := routeParameterInt(r, 1)
	if id != -1 {
		deleteOneItem(w, id)
	} else {
		deleteDoneItems(w)
	}
}

func deleteOneItem(w http.ResponseWriter, id int) {
	_, err := db.Exec(`DELETE FROM items WHERE id=?`, id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func deleteDoneItems(w http.ResponseWriter) {
	_, err := db.Exec(`DELETE FROM items WHERE done=true`)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func doneItems(w http.ResponseWriter, r *http.Request) {
	id := routeParameterInt(r, 1)
	if id == -1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	_, err := db.Exec(`UPDATE items SET done=true where id=?`, id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

func unDoneItems(w http.ResponseWriter, r *http.Request) {
	id := routeParameterInt(r, 1)
	if id == -1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	_, err := db.Exec(`UPDATE items SET done=false where id=?`, id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
