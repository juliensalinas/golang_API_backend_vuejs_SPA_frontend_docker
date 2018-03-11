/*
Send to frontend a list of companies types existing in companysocialprofile.type
*/

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

// Added "2" at the end because another CompSocProfRow is aleady exported
// do not want to spend time finding another naming rule
type CompSocProfRow2 struct {
	Type string `json:"typeName"`
}

// getAllCompaniesTypes queries db to retrieve a distinct list of all companies sizes
// in companysocialprofile table
func getAllCompaniesTypes(w http.ResponseWriter) ([]CompSocProfRow2, error) {

	var compSocProfRows2 []CompSocProfRow2
	var err error

	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return compSocProfRows2, err
	}
	defer db.Close()

	sqlStmt := "SELECT DISTINCT(type) FROM companysocialprofile WHERE type <> ''"

	rows, err := db.Query(sqlStmt)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return compSocProfRows2, err
	}
	defer rows.Close()

	for rows.Next() {
		var compSocProfRow2 CompSocProfRow2
		if err := rows.Scan(&compSocProfRow2.Type); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return compSocProfRows2, err
		}
		compSocProfRows2 = append(compSocProfRows2, compSocProfRow2)
	}

	return compSocProfRows2, err
}

// ReturnCompaniesTypesList loads all companies types from db and send it in JSON to frontend
func ReturnCompaniesTypesList(w http.ResponseWriter, r *http.Request) {

	compSocProfRows2, err := getAllCompaniesTypes(w)
	if err != nil {
		return
	}

	if len(compSocProfRows2) == 0 {
		log.Println("No result found\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	returnedJson, err := json.Marshal(compSocProfRows2)
	if err != nil {
		err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
