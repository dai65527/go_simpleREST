package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

var db *sql.DB

type Item struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

func main() {
	var err error

	// Open data base
	if os.Getenv("DB_MIDDLEWARE") == "mysql" {
		db, err = sql.Open("mysql", "todoapi:todopass@tcp(tododb:3306)/todo")
	} else {
		db, err = sql.Open("sqlite", "./database.db")
	}
	if err != nil {
		log.Fatal(err)
	}

	// Init db
	err = initDB(db)
	if err != nil {
		log.Fatal(err)
	}

	// リクエストハンドラの追加
	http.HandleFunc("/items", itemsHandler)    // `/items`の処理（）
	http.HandleFunc("/items/", itemsIdHandler) // `/items/:id`と`/items/:id/done`の処理

	err = http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

/*** リクエストハンドラ ***/

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // localhost:3000からのオリジン間アクセスを許可する

	switch r.Method {
	case "GET":
		getAllItems(w, r) // 全てのitemの取得
	case "POST":
		addNewItem(w, r) // 新しいitemの追加
	case "DELETE":
		deleteDoneItems(w) // 実行済みitemの削除
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")               // Content-Typeヘッダの使用を許可する
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS") // pre-flightリクエストに対応する
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func itemsIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // localhost:3000からのオリジン間アクセスを許可する

	// ルートパラメータの取得（例: `/items/1/done` -> ["items", "1", "done"]）
	params := getRouteParams(r)
	if len(params) < 2 || len(params) > 3 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// itemのidをintで取得
	id, err := strconv.Atoi(params[1])
	if err != nil || id < 1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if len(params) == 2 {
		updateItem(id, w, r)
	} else if params[2] == "done" {
		updateDone(id, w, r)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func updateItem(id int, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		deleteOneItem(id, w)
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")    // Content-Typeヘッダの使用を許可する
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS") // pre-flightリクエストに対応する
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func updateDone(id int, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		doneItem(id, w, r)
	case "DELETE":
		unDoneItem(id, w, r)
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")         // Content-Typeヘッダの使用を許可する
		w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE, OPTIONS") // pre-flightリクエストに対応する
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

/*** データベース操作 ***/

// データベース初期化
func initDB(db *sql.DB) error {
	var sql string
	if os.Getenv("DB_MIDDLEWARE") == "mysql" {
		sql = `
			CREATE TABLE IF NOT EXISTS items (
				id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
				name TEXT NOT NULL,
				done BOOLEAN NOT NULL DEFAULT 0
			);`
	} else {
		sql = `
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT 0
		);`
	}
	_, err := db.Exec(sql)
	return err
}

// 全アイテムの取得
func getAllItems(w http.ResponseWriter, r *http.Request) {
	var items []Item
	rows, err := db.Query(`SELECT * FROM items;`)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	for rows.Next() {
		var item Item
		rows.Scan(&item.Id, &item.Name, &item.Done)
		items = append(items, item)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(items); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, buf.String())
}

// 新しいアイテムを追加
func addNewItem(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Name string `json:"name"`
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	_, err = db.Exec(`INSERT INTO items (name, done) values (?, ?)`, reqBody.Name, false)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

// 1つのアイテムを削除
func deleteOneItem(id int, w http.ResponseWriter) {
	_, err := db.Exec(`DELETE FROM items WHERE id=?`, id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// 全ての実行済みアイテムを削除
func deleteDoneItems(w http.ResponseWriter) {
	_, err := db.Exec(`DELETE FROM items WHERE done=true`)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// アイテムを実行済みにする
func doneItem(id int, w http.ResponseWriter, r *http.Request) {
	if id == -1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	_, err := db.Exec(`UPDATE items SET done=true where id=?`, id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

// アイテムを未実行にする
func unDoneItem(id int, w http.ResponseWriter, r *http.Request) {
	if id == -1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	_, err := db.Exec(`UPDATE items SET done=false where id=?`, id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

/*** その他 ***/

func getRouteParams(r *http.Request) []string {
	splited := strings.Split(r.RequestURI, "/")
	var params []string
	for i := 0; i < len(splited); i++ {
		if len(splited[i]) != 0 {
			params = append(params, splited[i])
		}
	}
	return params
}
