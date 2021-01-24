package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	a.Initialize(
		"facilityadmin",
		"faci1ityAdmin",
		"localhost",
		"5432",
		"facility_booking")

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS booking.booking
(
	id SERIAL,
	user_id text,
	email text,
	purpose text,
	facility_id integer,
	start_dt timestamptz,
	end_dt timestamptz,
	transaction_dt timestamptz,
	CONSTRAINT booking_pkey PRIMARY KEY (id)
)`

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM booking.booking")
	a.DB.Exec("ALTER SEQUENCE booking.booking_id_seq RESTART WITH 1")
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/bookings", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/booking/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Booking not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Booking not found'. Got '%s'", m["error"])
	}
}

func TestCreateBooking(t *testing.T) {

	clearTable()

	var jsonStr = []byte(`{"user_id":"test", "email": "test@email.com", "purpose": "nil", "facility_id": 1, "start_dt": "2021-01-24 10:00:00+08", "end_dt": "2021-01-24 18:00:00+08", "transaction_dt": "2021-01-23 10:00:00+08"}`)
	req, _ := http.NewRequest("POST", "/booking", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != 1.0 {
		t.Errorf("Expected booking ID to be '1'. Got '%v'", m["id"])
	}

	if m["user_id"] != "test" {
		t.Errorf("Expected booking user_id to be 'test'. Got '%v'", m["user_id"])
	}

	if m["email"] != "test@email.com" {
		t.Errorf("Expected booking email to be 'test@email'. Got '%v'", m["email"])
	}

	if m["purpose"] != "nil" {
		t.Errorf("Expected booking purpose to be 'nil'. Got '%v'", m["purpose"])
	}

	if m["facility_id"] != 1.0 {
		t.Errorf("Expected facility_id to be '1'. Got '%v'", m["facility_id"])
	}

	if m["start_dt"] != "2021-01-24 10:00:00+08" {
		t.Errorf("Expected booking start_dt to be '2021-01-24 10:00:00+08'. Got '%v'", m["start_dt"])
	}

}

func addBookings(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO booking.booking(user_id, email, purpose, facility_id, start_dt, end_dt, transaction_dt) VALUES($1, $2, $3, $4, $5, $6, $7)", "user_"+strconv.Itoa(i), "user_"+strconv.Itoa(i)+"@email", "nil", 1, "2021-01-24 10:00:00+08", "2021-01-24 10:00:00+08", "2021-01-24 10:00:00+08")
	}
}

func TestGetBooking(t *testing.T) {
	clearTable()
	addBookings(1)

	req, _ := http.NewRequest("GET", "/booking/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateBooking(t *testing.T) {

	clearTable()
	addBookings(1)

	req, _ := http.NewRequest("GET", "/booking/1", nil)
	response := executeRequest(req)
	var originalBooking map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalBooking)

	var jsonStr = []byte(`{"user_id":"updated", "email": "updated@email.com", "purpose": "updated", "facility_id": 2, "start_dt": "2021-01-24 11:00:00+08", "end_dt": "2021-01-24 17:00:00+08", "transaction_dt": "2021-01-23 11:00:00+08"}`)
	req, _ = http.NewRequest("PUT", "/booking/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["user_id"] != "updated" {
		t.Errorf("Expected the user_id to change from '%v' to 'updated'. Got '%v'", originalBooking["user_id"], m["user_id"])
	}

	if m["email"] != "updated@email.com" {
		t.Errorf("Expected the email to change from '%v' to 'updated@email.com'. Got '%v'", originalBooking["email"], m["email"])
	}

	if m["purpose"] != "updated" {
		t.Errorf("Expected the purpose to change from '%v' to 'updated'. Got '%v'", originalBooking["purpose"], m["purpose"])
	}

	if m["facility_id"] != 2.0 {
		t.Errorf("Expected the email to change from '%v' to 2. Got '%v'", originalBooking["facility_id"], m["facility_id"])
	}

	if m["start_dt"] != "2021-01-24 11:00:00+08" {
		t.Errorf("Expected the user_id to change from '%v' to '2021-01-24 11:00:00+08'. Got '%v'", originalBooking["start_dt"], m["start_dt"])
	}

	if m["end_dt"] != "2021-01-24 17:00:00+08" {
		t.Errorf("Expected the email to change from '%v' to '2021-01-24 17:00:00+08'. Got '%v'", originalBooking["end_dt"], m["end_dt"])
	}

	if m["transaction_dt"] != "2021-01-23 11:00:00+08" {
		t.Errorf("Expected the email to change from '%v' to '2021-01-23 11:00:00+08'. Got '%v'", originalBooking["transaction_dt"], m["transaction_dt"])
	}

	if m["id"] != originalBooking["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalBooking["id"], m["id"])
	}
}

func TestDeleteBooking(t *testing.T) {
	clearTable()
	addBookings(1)

	req, _ := http.NewRequest("GET", "/booking/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/booking/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/booking/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
