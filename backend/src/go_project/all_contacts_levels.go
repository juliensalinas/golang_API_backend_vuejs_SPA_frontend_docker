/*
Send to frontend a list of job levels existing in the job_level table
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

type JobLevelRow struct {
	Name string `json:"levelName"`
}

// getAllContactsLevels queries db to retrieve a distinct list of all job levels
// in job_level table
func getAllContactsLevels(w http.ResponseWriter) ([]JobLevelRow, error) {

	var jobLevelRows []JobLevelRow
	var err error

	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return jobLevelRows, err
	}
	defer db.Close()

	sqlStmt := "SELECT name FROM job_level"

	rows, err := db.Query(sqlStmt)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return jobLevelRows, err
	}
	defer rows.Close()

	for rows.Next() {
		var jobLevelRow JobLevelRow
		if err := rows.Scan(&jobLevelRow.Name); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return jobLevelRows, err
		}
		jobLevelRows = append(jobLevelRows, jobLevelRow)
	}

	return jobLevelRows, err
}

// ReturnContactsLevelsList loads all contacts job levels from db and send it in JSON to frontend
func ReturnContactsLevelsList(w http.ResponseWriter, r *http.Request) {

	jobLevelRows, err := getAllContactsLevels(w)
	if err != nil {
		return
	}

	if len(jobLevelRows) == 0 {
		log.Println("No result found\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	returnedJson, err := json.Marshal(jobLevelRows)
	if err != nil {
		err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
