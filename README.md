# Purpose

Application dedicated to presenting data from our various databases in a user friendly way. It consists of 2 Docker containers:

* backend: Go exposing a RESTful API with the Go native net/http and json libs + gorilla/mux for routing
* frontend: Nginx serving static Vue.js/axios/Vuetify.js

# Features

* user has to enter credentials in order to use the SPA
* user can select various interfaces on the left pane in order to retrieve data from various db tables
* user can decide either to only count results returned by db or get the full results
* if results returned by db are lightweight enough then they are returned by the API and displayed within the SPA app inside a nice data table. User can also decide to export it as CSV.
* if results are too heavyweight, then results are sent to user by email within a .zip archive
* as input criteria, the user can enter text or CSV files listing a big amount of criteria
* some user inputs are select lists whose values are loaded dynamically from db

# More explanations

I wrote [this blog article](https://juliensalinas.com/en/golang-API-backend-vuejs-SPA-frontend-docker-modern-application/) in order to comment/explain the application.