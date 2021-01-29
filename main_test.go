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
	clearBookingTable()
	resetRecord()
	clearFacilityDetailTable()
	os.Exit(code)
}

const bookingTableCreationQuery = `CREATE TABLE IF NOT EXISTS booking.booking
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

const bookingConfigTableCreationQuery = `CREATE TABLE IF NOT EXISTS booking.booking_config
(
	id SERIAL,
	key text NOT NULL,
	value text,
	CONSTRAINT booking_config_pkey PRIMARY KEY (id)
)`

const facilityDetailTableCreationQuery = `CREATE TABLE IF NOT EXISTS booking.facility_detail
(
	id SERIAL,
	name text UNIQUE,
	level integer,
	description text,
	status text,
	transaction_dt timestamptz,
	CONSTRAINT facility_detail_pkey PRIMARY KEY (id)
)`

const accountTableCreationQuery = `CREATE TABLE IF NOT EXISTS booking.account
(
    id SERIAL,
    user_id text NOT NULL UNIQUE,
	admin boolean,
	email text,
    password text NOT NULL,
	CONSTRAINT account_pkey PRIMARY KEY (id)
)`

func ensureTableExists() {
	if _, err := a.DB.Exec(bookingTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(bookingConfigTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(facilityDetailTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(accountTableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearBookingTable() {
	a.DB.Exec("DELETE FROM booking.booking")
	a.DB.Exec("ALTER SEQUENCE booking.booking_id_seq RESTART WITH 1")
	a.DB.Exec("DELETE FROM booking.facility_detail")
	a.DB.Exec("ALTER SEQUENCE booking.facility_detail_id_seq RESTART WITH 1")
}

func clearFacilityDetailTable() {
	a.DB.Exec("DELETE FROM booking.facility_detail")
	a.DB.Exec("ALTER SEQUENCE booking.facility_detail_id_seq RESTART WITH 1")
}

func resetRecord() {
	a.DB.Exec("UPDATE booking.booking_config SET key='max_hr_per_booking', value='2' WHERE id=1")
}

func TestEmptyBookingTable(t *testing.T) {
	clearBookingTable()

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

func TestGetNonExistentBooking(t *testing.T) {
	clearBookingTable()

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

	clearBookingTable()

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
		t.Errorf("Expected user_id to be 'test'. Got '%v'", m["user_id"])
	}

	if m["email"] != "test@email.com" {
		t.Errorf("Expected email to be 'test@email'. Got '%v'", m["email"])
	}

	if m["purpose"] != "nil" {
		t.Errorf("Expected purpose to be 'nil'. Got '%v'", m["purpose"])
	}

	if m["facility_id"] != 1.0 {
		t.Errorf("Expected facility_id to be '1'. Got '%v'", m["facility_id"])
	}

	if m["start_dt"] != "2021-01-24 10:00:00+08" {
		t.Errorf("Expected start_dt to be '2021-01-24 10:00:00+08'. Got '%v'", m["start_dt"])
	}

	if m["end_dt"] != "2021-01-24 18:00:00+08" {
		t.Errorf("Expected end_dt to be '2021-01-24 18:00:00+08'. Got '%v'", m["end_dt"])
	}

	if m["transaction_dt"] != "2021-01-23 10:00:00+08" {
		t.Errorf("Expected transaction_dt to be '2021-01-23 10:00:00+08'. Got '%v'", m["transaction_dt"])
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
	clearBookingTable()
	addBookings(1)

	req, _ := http.NewRequest("GET", "/booking/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateBooking(t *testing.T) {

	clearBookingTable()
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
	clearBookingTable()
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

func TestGetBookingConfigs(t *testing.T) {

	req, _ := http.NewRequest("GET", "/bookingConfigs", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var result []map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)

	if len(result) != 4 {
		t.Errorf("Expected 4 record. Got %v", len(result))
	}
}

const firstConfigKey = "max_hr_per_booking"
const firstConfigValue = "2"

func TestGetBookingConfig(t *testing.T) {

	req, _ := http.NewRequest("GET", "/bookingConfig/1", nil)
	response := executeRequest(req)

	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)

	checkResponseCode(t, http.StatusOK, response.Code)

	if result["key"] != firstConfigKey {
		t.Errorf("Expected 'max_hr_per_booking'. Got %v", result["key"])
	}
	if result["value"] != firstConfigValue {
		t.Errorf("Expected '2'. Got %v", result["value"])
	}
}

func TestUpdateBookingConfig(t *testing.T) {

	req, _ := http.NewRequest("GET", "/bookingConfig/1", nil)
	response := executeRequest(req)
	var originalBookingConfig map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalBookingConfig)

	var jsonStr = []byte(`{"key":"max_hr_per_booking_updated", "value": "3"}`)
	req, _ = http.NewRequest("PUT", "/bookingConfig/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["key"] != "max_hr_per_booking_updated" {
		t.Errorf("Expected the key to change from '%v' to 'max_hr_per_booking_updated'. Got '%v'", originalBookingConfig["key"], m["key"])
	}

	if m["value"] != "3" {
		t.Errorf("Expected the value to change from '%v' to '3'. Got '%v'", originalBookingConfig["value"], m["value"])
	}

	if m["id"] != originalBookingConfig["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalBookingConfig["id"], m["id"])
	}
}

func TestEmptyFacilityDetailTable(t *testing.T) {
	clearFacilityDetailTable()

	req, _ := http.NewRequest("GET", "/facilityDetails", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentFacilityDetail(t *testing.T) {
	clearFacilityDetailTable()

	req, _ := http.NewRequest("GET", "/facilityDetail/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Facility detail not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Facility detail not found'. Got '%s'", m["error"])
	}
}

func TestCreateFacilityDetail(t *testing.T) {
	clearFacilityDetailTable()

	var jsonStr = []byte(`{"name":"Meeting Room L1-01", "level": "1", "description": "Meeting Room", "status": "OPEN", "transaction_dt": "2021-01-23 10:00:00+08"}`)
	req, _ := http.NewRequest("POST", "/facilityDetail", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != 1.0 {
		t.Errorf("Expected Facility Detail ID to be 1. Got '%v'", m["id"])
	}

	if m["name"] != "Meeting Room L1-01" {
		t.Errorf("Expected name to be 'Meeting Room L1-01'. Got '%v'", m["name"])
	}

	if m["level"] != "1" {
		t.Errorf("Expected level to be '1'. Got '%v'", m["level"])
	}

	if m["description"] != "Meeting Room" {
		t.Errorf("Expected description to be 'Meeting Room'. Got '%v'", m["description"])
	}

	if m["status"] != "OPEN" {
		t.Errorf("Expected status to be 'OPEN'. Got '%v'", m["status"])
	}

	if m["transaction_dt"] != "2021-01-23 10:00:00+08" {
		t.Errorf("Expected transaction_dt to be '2021-01-23 10:00:00+08'. Got '%v'", m["transaction_dt"])
	}

}

func addFacilityDetail(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO booking.facility_detail(name, level, description, status, transaction_dt) VALUES($1, $2, $3, $4, $5)", "Meeting Room L"+strconv.Itoa(i), "L"+strconv.Itoa(i), "Meeting Room", "OPEN", "2021-01-24 10:00:00+08")
	}
}

func TestGetFacilityDetail(t *testing.T) {
	clearFacilityDetailTable()
	addFacilityDetail(1)

	req, _ := http.NewRequest("GET", "/facilityDetail/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateFacilityDetail(t *testing.T) {

	clearFacilityDetailTable()
	addFacilityDetail(1)

	req, _ := http.NewRequest("GET", "/facilityDetail/1", nil)
	response := executeRequest(req)
	var originalFacilityDetail map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalFacilityDetail)

	var jsonStr = []byte(`{"name":"Meeting Room L1-01", "level": "1", "description": "Meeting Rm", "status": "OPEN", "transaction_dt": "2021-01-23 12:00:00+08"}`)
	req, _ = http.NewRequest("PUT", "/facilityDetail/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "Meeting Room L1-01" {
		t.Errorf("Expected the name to change from '%v' to 'Meeting Room L1-01'. Got '%v'", originalFacilityDetail["name"], m["name"])
	}

	if m["level"] != "1" {
		t.Errorf("Expected the level to change from '%v' to '1'. Got '%v'", originalFacilityDetail["level"], m["level"])
	}

	if m["description"] != "Meeting Rm" {
		t.Errorf("Expected the description to change from '%v' to 'Meeting Rm'. Got '%v'", originalFacilityDetail["description"], m["description"])
	}

	if m["status"] != "OPEN" {
		t.Errorf("Expected the status to change from '%v' to 'OPEN'. Got '%v'", originalFacilityDetail["status"], m["status"])
	}

	if m["transaction_dt"] != "2021-01-23 12:00:00+08" {
		t.Errorf("Expected the email to change from '%v' to '2021-01-23 12:00:00+08'. Got '%v'", originalFacilityDetail["transaction_dt"], m["transaction_dt"])
	}

	if m["id"] != originalFacilityDetail["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalFacilityDetail["id"], m["id"])
	}
}

func TestDeleteFacilityDetail(t *testing.T) {
	clearFacilityDetailTable()
	addFacilityDetail(1)

	req, _ := http.NewRequest("GET", "/facilityDetail/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/facilityDetail/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/facilityDetail/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetBookingsCount(t *testing.T) {
	clearBookingTable()
	req, _ := http.NewRequest("GET", "/bookingsCount", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var count int
	json.Unmarshal(response.Body.Bytes(), &count)

	if count != 0 {
		t.Errorf("Expected the count to be 0. Got %d", count)
	}
	addBookings(5)

	req, _ = http.NewRequest("GET", "/bookingsCount", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &count)

	if count != 5 {
		t.Errorf("Expected the count to be 5. Got %d", count)
	}
}

func TestGetBookingConfigsCount(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bookingConfigsCount", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var count int
	json.Unmarshal(response.Body.Bytes(), &count)

	if count != 4 {
		t.Errorf("Expected the count to be 4. Got %d", count)
	}
}

func TestGetFacilityDetailsCount(t *testing.T) {
	clearFacilityDetailTable()
	req, _ := http.NewRequest("GET", "/facilityDetailsCount", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var count int
	json.Unmarshal(response.Body.Bytes(), &count)

	if count != 0 {
		t.Errorf("Expected the count to be 0. Got %d", count)
	}
	addFacilityDetail(5)

	req, _ = http.NewRequest("GET", "/facilityDetailsCount", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &count)

	if count != 5 {
		t.Errorf("Expected the count to be 5. Got %d", count)
	}
}

func addTestAccount() {
	a.DB.Exec("INSERT INTO booking.account(user_id, admin, email, password) VALUES ($1, $2, $3, crypt($4, gen_salt('bf')))", "testAccount", false, "testAccount@mail.com", "TestAccountPassword")
}

func removeTestAccount() {
	a.DB.Exec("DELETE FROM booking.account WHERE user_id=$1", "testAccount")
}

func TestLogin(t *testing.T) {
	addTestAccount()

	var jsonStr = []byte(`{"user_id":"testAccount", "password": "wrongpassword"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)

	jsonStr = []byte(`{"user_id":"testAccount", "password": "TestAccountPassword"}`)
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["user_id"] != "testAccount" {
		t.Errorf("Expected the user_id to be testAccount. Got '%v'", m["user_id"])
	}

	if m["admin"] != false {
		t.Errorf("Expected the admin to be false. Got '%v'", m["admin"])
	}

	if m["email"] != "testAccount@mail.com" {
		t.Errorf("Expected the email to be testAccount@mail.com. Got '%v'", m["email"])
	}

	removeTestAccount()
}
