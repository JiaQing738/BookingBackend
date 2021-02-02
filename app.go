package main

import (
	"database/sql"
	"fmt"
	"log"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// App struct exposes references to the router and the database
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize the Postgresql DB connection
func (a *App) Initialize(user, password, host, port, dbname string) {
	connectionString := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		user,
		password,
		host,
		port,
		dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

// Run to Listen to port 8010
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) getBooking(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	p := booking{ID: id}
	if err := p.getBooking(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Booking not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) getBookings(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	userid := r.FormValue("user_id")

	if count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	bookings, err := getBookings(a.DB, start, count, userid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, bookings)
}

func (a *App) getBookingsCount(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	userid := r.FormValue("user_id")

	count, err := getBookingsCount(a.DB, userid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, count)
}

func (a *App) createBooking(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var p booking
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	var count int
	var err error
	count, err = p.getOverlappingBookings(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if count > 0 {
		respondWithError(w, http.StatusInternalServerError, "Overlap Bookings")
		return
	}

	err = p.createBooking(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateBooking(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var p booking
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateBooking(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteBooking(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Booking ID")
		return
	}

	p := booking{ID: id}
	if err := p.deleteBooking(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getBookingConfigs(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	bookingConfigs, err := getBookingConfigs(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, bookingConfigs)
}

func (a *App) getBookingConfig(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid booking config ID")
		return
	}

	p := bookingConfig{ID: id}
	if err := p.getBookingConfig(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Booking Config not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) updateBookingConfig(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid booking config ID")
		return
	}

	var p bookingConfig
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateBookingConfig(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) getBookingConfigsCount(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	count, err := getBookingConfigsCount(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, count)
}

func (a *App) getFacilityDetail(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid facility detail ID")
		return
	}

	p := facilityDetail{ID: id}
	if err := p.getFacilityDetail(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Facility detail not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) getFacilityDetails(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	status := r.FormValue("status")

	if count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	facilityDetails, err := getFacilityDetails(a.DB, start, count, status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, facilityDetails)
}

func (a *App) createFacilityDetail(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var p facilityDetail
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createFacilityDetail(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateFacilityDetail(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid facility detail ID")
		return
	}

	var p facilityDetail
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateFacilityDetail(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteFacilityDetail(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid facility detail ID")
		return
	}

	p := facilityDetail{ID: id}
	if err := p.deleteFacilityDetail(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	b := booking{FacilityID: id}
	if err := b.deleteAllBookingByFacilityID(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getFacilityDetailsCount(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	status := r.FormValue("status")

	count, err := getFacilityDetailsCount(a.DB, status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, count)
}

func (a *App) authenticate(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var p login
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	account, err := p.authenticate(a.DB)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusOK, "Login failed")
			return
		}

		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, account)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (a *App) optionsEnableCors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/bookings", a.getBookings).Methods("GET")
	a.Router.HandleFunc("/booking", a.createBooking).Methods("POST")
	a.Router.HandleFunc("/booking/{id:[0-9]+}", a.getBooking).Methods("GET")
	a.Router.HandleFunc("/booking/{id:[0-9]+}", a.updateBooking).Methods("PUT")
	a.Router.HandleFunc("/booking/{id:[0-9]+}", a.deleteBooking).Methods("DELETE")
	a.Router.HandleFunc("/bookingConfigs", a.getBookingConfigs).Methods("GET")
	a.Router.HandleFunc("/bookingConfig/{id:[0-9]+}", a.getBookingConfig).Methods("GET")
	a.Router.HandleFunc("/bookingConfig/{id:[0-9]+}", a.updateBookingConfig).Methods("PUT")
	a.Router.HandleFunc("/facilityDetails", a.getFacilityDetails).Methods("GET")
	a.Router.HandleFunc("/facilityDetail", a.createFacilityDetail).Methods("POST")
	a.Router.HandleFunc("/facilityDetail/{id:[0-9]+}", a.getFacilityDetail).Methods("GET")
	a.Router.HandleFunc("/facilityDetail/{id:[0-9]+}", a.updateFacilityDetail).Methods("PUT")
	a.Router.HandleFunc("/facilityDetail/{id:[0-9]+}", a.deleteFacilityDetail).Methods("DELETE")
	a.Router.HandleFunc("/booking", a.optionsEnableCors).Methods(http.MethodOptions)
	a.Router.HandleFunc("/booking/{id:[0-9]+}", a.optionsEnableCors).Methods(http.MethodOptions)
	a.Router.HandleFunc("/bookingConfig/{id:[0-9]+}", a.optionsEnableCors).Methods(http.MethodOptions)
	a.Router.HandleFunc("/facilityDetail", a.optionsEnableCors).Methods(http.MethodOptions)
	a.Router.HandleFunc("/facilityDetail/{id:[0-9]+}", a.optionsEnableCors).Methods(http.MethodOptions)
	a.Router.HandleFunc("/bookingsCount", a.getBookingsCount).Methods("GET")
	a.Router.HandleFunc("/bookingConfigsCount", a.getBookingConfigsCount).Methods("GET")
	a.Router.HandleFunc("/facilityDetailsCount", a.getFacilityDetailsCount).Methods("GET")
	a.Router.HandleFunc("/login", a.authenticate).Methods("POST")
	a.Router.HandleFunc("/login", a.optionsEnableCors).Methods(http.MethodOptions)
	a.Router.Use(mux.CORSMethodMiddleware(a.Router))
}
