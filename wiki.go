package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	// Import the database driver that's used for MariaDB
	"github.com/go-sql-driver/mysql"

	// Markdown to HTML module

	// Random username module
	"github.com/lucasepe/codename"
)

// Defining global variables in this block.  The GO DB tutorial doesn't recommend declaring the DB as global.
// It suggests passing it through functions, however, the http Handler functions don't allow for that so for ease it's defined as global.
var (
	db   *sql.DB
	tmpl *template.Template
)

type Page struct {
	PageID      int64
	Title       string
	Body        string
	CreatedDate time.Time
	UpdatedDate time.Time
	UserID      int64
	AuthorName  string
	UpdatedBy   int64
}

type User struct {
	UserID         int64
	AuthorUserName string
	AuthorName     string
	CreatedDate    string
}

// You can only pass one value through templating.  Some of the http pages created use data from multiple queries.
// As a workaround the results from each of these are included in a 'Payload' struct of structs which are reference via the Go templating.
type Payload struct {
	RecentList []Page
	ViewPage   Page
}

func main() {
	// Create the DB connection
	dbCXN()

	// To ensure that styling and images are served a http fileserver for static content is defined
	http.Handle("/assets/", http.FileServer(http.Dir(".")))

	// Parse all .html files into the templating
	var err error
	tmpl, err = tmpl.ParseGlob("html/*.html")
	if err != nil {
		log.Println(err)
	}

	// Handle http requests
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/read/", readHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	// Output some basic log information to console
	log.Println("Application being served at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dbCXN() {
	// Database connection properties
	cfg := mysql.Config{
		User:   "app",
		Passwd: "jVqb2aren2Gm", //Hardcoded for testing
		Net:    "tcp",
		Addr:   "172.17.0.2:3306",
		DBName: "wiki",
		// Setting ParseTime to true allows the DATEIME types from MariaDB to be stored in Go's Time.Time data type
		ParseTime: true,
		// Credit to a Stackoverflow article for highlighting that AllowNativePasswords is requred for DB Authentication
		AllowNativePasswords: true,
	}

	// Get the database handle
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	// ping the database to ensure connectivity
	pingErr := db.Ping()
	if err != nil {
		log.Fatal(pingErr)
	}
	log.Println("Database is reachable!")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// This handler handles any requests for root `\` and serves index.html
	p := Payload{recentPages(), loadPage(1)}
	tmpl.ExecuteTemplate(w, "index.html", p)
	log.Printf("Index page served.\n")
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	// This handler allows users to read a wiki page
	pID, err := strconv.ParseInt(r.URL.Path[len("/read/"):], 10, 64) // [len("/read/"):] slices `read` from the URL path
	if err != nil {
		log.Println(err)
	} else {
		p := Payload{recentPages(), loadPage(pID)}
		tmpl.ExecuteTemplate(w, "read.html", p)
		log.Printf("Page ID=%v served.\n", p.ViewPage.PageID)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	// This handler allows users to edit a wiki page
	pID, err := strconv.ParseInt(r.URL.Path[len("/edit/"):], 10, 64) // [len("/edit/"):] slices `read` from the URL path
	if err != nil {
		log.Println(err)
	}
	p := Payload{recentPages(), loadPage(pID)}
	tmpl.ExecuteTemplate(w, "edit.html", p)
	log.Printf("Edit template for page ID=%v served.\n", p.ViewPage.PageID)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	// This handler allows users to save a wiki page after editing it
	pID, err := strconv.ParseInt(r.URL.Path[len("/save/"):], 10, 64) // [len("/save/"):] slices `save` from the URL path
	if err != nil {
		log.Println(err)
	}
	p := loadPage(pID)
	p.Title = r.FormValue("title")
	p.Body = r.FormValue("body")
	save := editPage(p)
	log.Println(save)

	http.Redirect(w, r, "/read/"+strconv.FormatInt(pID, 10), http.StatusFound) //conver Page ID (In64 in decimal format) to string
}

func createPage(p Page) int64 {
	result, err := db.Exec("INSERT INTO pages (title, body, createdDate, updatedDate, userID) VALUES (?, ?, ?, ?, ?)", p.Title, p.Body, p.CreatedDate, p.UpdatedDate, p.UserID)
	if err != nil {
		log.Println(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
	}
	log.Printf("New DB record created for page ID=%v.\n", id)

	return id
}

func loadPage(id int64) Page {
	var p Page
	query := "SELECT p.pageID, p.title, p.body, p.createdDate, p.updatedDate, p.userID, p.updatedBy, u.realName from pages p INNER JOIN user u on p.userID = u.userID WHERE p.pageID = ?;"
	row := db.QueryRow(query, id)
	if err := row.Scan(&p.PageID, &p.Title, &p.Body, &p.CreatedDate, &p.UpdatedDate, &p.UserID, &p.UpdatedBy, &p.AuthorName); err != nil {
		log.Printf("Database read error: %s", err)
	} else {
		log.Printf("DB record for page ID=%v retrieved.\n", id)
	}

	return p
}

func editPage(p Page) sql.Result {
	userID := randomUser()
	query := "UPDATE pages SET title = ?, body = ?, updatedDate = ?, updatedBy = ? WHERE pageID = ?;"
	result, err := db.Exec(query, p.Title, p.Body, time.Now(), userID, p.PageID)
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Printf("DB record for page ID=%v updated by user=%v.\n", p.PageID, userID)

	return result
}

func recentPages() []Page {
	var pages []Page

	// Get the 10 most recent wiki pages
	rows, err := db.Query("SELECT p.pageID, p.title, p.body, p.createdDate, p.updatedDate, p.userID, p.updatedBy, u.realName FROM pages p INNER JOIN user u on p.userID = u.userID ORDER BY createdDate DESC LIMIT 10;")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	// Populate the Page struct with query results
	for rows.Next() {
		var p Page
		if err := rows.Scan(&p.PageID, &p.Title, &p.Body, &p.CreatedDate, &p.UpdatedDate, &p.UserID, &p.UpdatedBy, &p.AuthorName); err != nil {
			log.Println(err)
			return nil
		}
		pages = append(pages, p)
		log.Printf("Recent page list retrieved\n")
	}

	return pages
}

func randomUser() int64 {
	// Creates a random user
	rng, err := codename.DefaultRNG()
	if err != nil {
		log.Println(err)
	}
	uName := codename.Generate(rng, 0)
	s := strings.Split(uName, "-")
	rName := s[0] + " " + s[len(s)-1]

	// Add user to database
	query := "INSERT INTO user (realName, userName, createdDate) VALUES (?, ?, ?);"
	result, err := db.Exec(query, rName, uName, time.Now())
	if err != nil {
		log.Println(err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
	}

	return userID
}

func addUser() {
	// Add user code here
}

func dbTests() {
	// // Test page load function
	// loadPage(1)

	// // Test create page function
	// new := createPage(Page{
	// 	Title:       "Lorem Ipsum ++",
	// 	Body:        []byte("Sed nec bibendum quam. Nulla auctor euismod sapien, at congue orci tincidunt nec. Sed a tortor pretium, ornare nisl nec, volutpat lacus. Etiam tincidunt nulla ligula, id pretium felis vulputate et. Suspendisse potenti. Nunc non metus eu felis semper pellentesque in dictum magna. Proin quis dignissim eros, vitae interdum lectus. Aenean tempor at dolor quis lacinia. Mauris aliquam massa et lacus dapibus convallis. Donec eget est libero."),
	// 	CreatedDate: time.Now(),
	// 	UpdatedDate: time.Now(),
	// 	userID:    1,
	// })
	// log.Printf("Page ID '%v' created!", new)

	// // Test update page function
	// edt := editPage(Page{
	// 	PageID:      15,
	// 	Title:       "Lorem Ipsum Edited",
	// 	Body:        []byte("Etiam porta euismod ligula. Morbi varius, dui a finibus vestibulum, risus leo varius odio, vel volutpat elit purus ultrices neque. Vivamus libero risus, gravida vitae nulla ut, gravida suscipit tellus. Donec dapibus placerat orci, sit amet lobortis justo semper in. Nulla tincidunt diam quis viverra malesuada. Curabitur tristique enim eu semper egestas. Mauris interdum malesuada pretium. Etiam enim ligula, tristique in viverra sed, fringilla sit amet orci. Cras mauris libero, accumsan at egestas non, posuere sed felis. Fusce nec est dui. Cras sed auctor lacus. Praesent vel vehicula metus, non ultrices lectus. Duis non molestie ipsum. "),
	// 	UpdatedDate: time.Now(),
	// })
	// log.Printf("Page ID '%v' updated!", edt)

	// // Test last 10 pages
	// pages := recentPages()
	// for _, p := range pages {
	// 	fmt.Printf("%v, %s \n", p.PageID, p.Title)
	// }
	// serve http on port 8080
}
