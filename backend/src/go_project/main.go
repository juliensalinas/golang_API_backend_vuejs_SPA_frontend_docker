/*

Backend of the webapp. RESTful API only.

Nice source about Go RESTful APIs: https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql
*/

package main

import (
	"errors"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"strings"
)

// Initialize db parameters
var localHost string = getLocalHost()
var remoteHost string = getRemoteHost()

// Initialize db parameters
const (
	// Local DB:
	localPort     = 5432
	localUser     = "my_local_user"
	localPassword = "my_local_pass"
	localDbname   = "my_local_db"

	// Remote DB:
	remotePort     = 5432
	remoteUser     = "my_remote_user"
	remotePassword = "my_remote_pass"
	remoteDbname   = "my_remote_db"
)

// getLogFilePath gets log file path from env var set by Docker run
func getLogFilePath() string {
	envContent := os.Getenv("LOG_FILE_PATH")
	return envContent
}

// getLocalHost gets local db host from env var set by Docker run.
// If no env var set, set it to localhost.
func getLocalHost() string {
	envContent := os.Getenv("LOCAL_DB_HOST")
	if envContent == "" {
		envContent = "127.0.0.1"
	}
	return envContent
}

// getRemoteHost gets remote db host from env var set by Docker run.
// If no env var set, set it to localhost.
func getRemoteHost() string {
	envContent := os.Getenv("REMOTE_DB_HOST")
	if envContent == "" {
		envContent = "127.0.0.1"
	}
	return envContent
}

// getRemoteHost gets remote db host from env var set by Docker run.
// If no env var set, set it to localhost.
func getCorsAllowedOrigin() string {
	envContent := os.Getenv("CORS_ALLOWED_ORIGIN")
	if envContent == "" {
		envContent = "http://localhost:8080"
	}
	return envContent
}

// getUserEmail gets user email of the person who will receive the results
// from env var set by Docker run.
// If no env var set, set it to admin.
func getUserEmail() string {
	envContent := os.Getenv("USER_EMAIL")
	if envContent == "" {
		envContent = "admin@example.com"
	}
	return envContent
}

// CustErr adds a custom message to any error message and
// formats it nicely
func CustErr(err error, msg string) error {

	var newErr strings.Builder
	newErr.WriteString("\n")
	newErr.WriteString(msg)
	newErr.WriteString("\nDetailed error:\n")
	newErr.WriteString(err.Error())

	return errors.New(newErr.String())

}

func main() {

	// Log everything to file or console depending on user preference.
	// Directory was created first by "Docker run" thanks to the -v option.
	log.SetFlags(log.LstdFlags | log.Lshortfile)            // add line number to logger
	if logFilePath := getLogFilePath(); logFilePath != "" { // write to log file only if logFilePath is set
		f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Using gorilla/mux for passing parameters in url like {missionnumber}
	router := mux.NewRouter()

	// CORS is a nightmare. Using gorilla/handlers seemed to work for
	// GET requests but not for POST requests.
	// The rs/cors library is a great solution.
	// https://stackoverflow.com/questions/40985920/making-golang-gorilla-cors-handler-work
	// Here is a nice explaination about CORS:
	// https://husobee.github.io/golang/cors/2015/09/26/cors.html
	c := cors.New(cors.Options{
		AllowedOrigins: []string{getCorsAllowedOrigin()},
	})
	handler := c.Handler(router)

	// Set routes
	router.HandleFunc("/get-contacts-levels-list", ReturnContactsLevelsList).Methods("GET")
	router.HandleFunc("/get-contacts-functions-list", ReturnContactsFunctionsList).Methods("GET")
	router.HandleFunc("/get-companies-types-list", ReturnCompaniesTypesList).Methods("GET")
	router.HandleFunc("/get-companies-sizes-list", ReturnCompaniesSizesList).Methods("GET")
	router.HandleFunc("/get-contacts-industries-list", ReturnContactsIndustriesList).Methods("GET")
	router.HandleFunc("/get-companies-industries-list", ReturnCompaniesIndustriesList).Methods("GET")
	router.HandleFunc("/get-countries-list", ReturnCountriesList).Methods("GET")
	router.HandleFunc("/get-companies-and-contacts", ReturnCompaniesAndContacts).Methods("POST")
	router.HandleFunc("/get-emails-checked-by-john/mission-number/{missionnumber}", ReturnEmailsCheckedByPA).Methods("GET")

	// Launch server
	err := http.ListenAndServe(":8000", handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
