package main

import (
	"database/sql"
	"fmt"
	"forum/controller/handlers"
	"forum/controller/logging"
	"forum/controller/server"
	"log"
	"net/http"
	"strings"

	"forum/controller/cookies"
	"forum/model/data"
	forumDB "forum/model/functions"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// Ouverture de la connexion à la base SQLite
	db, err := sql.Open("sqlite3", "./model/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	forumDB.Initialisation(db)

	// Delete all previous sessions
	_, err = forumDB.DeleteAllSessions(db)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// Parse templates
	templates := server.ParseTemplates("./view/assets/templates/*.html")

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./view/assets/statics"))
	mux.Handle("/statics/", http.StripPrefix("/statics/", fs))

	// tmp := fetch all posts
	allPosts, err := forumDB.FetchPosts(db)
	if err != nil {
		log.Fatal(err)
	}
	data.CombinedData = data.AllData{
		ToDisplay: data.ToDisplay{
			Posts: allPosts,
		},
		Username: "",
	}

	handlers.RegisterRoutes(mux, templates, db)

	// dev/testing route
	mux.HandleFunc("/dev", func(w http.ResponseWriter, r *http.Request) {
		devHandler(w, r, db)
	})

	// cookie db init - temporary
	cookies.SetDB(db)

	// handlers db init - temporary
	handlers.SetDB(db)

	// tmp - filter handler
	mux.HandleFunc("/filter", handlers.FilterHandler)

	// Init custom logger
	logging.Init()

	logging.Logger.Println("Server starting : http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logging.Logger.Fatal(err)
	}
}

func devHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "text/plain")

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tables = append(tables, name)
	}

	for _, table := range tables {
		fmt.Fprintf(w, "=== TABLE: %s ===\n", table)

		query := fmt.Sprintf("SELECT * FROM %s;", table)
		tblRows, err := db.Query(query)
		if err != nil {
			fmt.Fprintf(w, "(error reading table: %v)\n\n", err)
			continue
		}

		// Get column names
		cols, err := tblRows.Columns()
		if err != nil {
			fmt.Fprintf(w, "(error reading columns: %v)\n\n", err)
			tblRows.Close()
			continue
		}
		fmt.Fprintf(w, "Columns: %s\n", strings.Join(cols, ", "))

		// Make a slice of interface{} to scan each row dynamically
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}

		for tblRows.Next() {
			if err := tblRows.Scan(ptrs...); err != nil {
				fmt.Fprintf(w, "(error scanning row: %v)\n", err)
				continue
			}

			var out []string
			for _, v := range vals {
				if v == nil {
					out = append(out, "NULL")
				} else {
					out = append(out, fmt.Sprintf("%v", v))
				}
			}
			fmt.Fprintf(w, "%s\n", strings.Join(out, " | "))
		}

		fmt.Fprintln(w)
		tblRows.Close()
	}

	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusOK)
}
