package main

import (
	"database/sql"
	"time"
)

type booking struct {
	ID              int    `json:"id"`
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	Purpose         string `json:"purpose"`
	FacilityID      int    `json:"facility_id"`
	StartTime       string `json:"start_dt"`
	EndTime         string `json:"end_dt"`
	TransactionTime string `json:"transaction_dt"`
}

func (p *booking) getBooking(db *sql.DB) error {
	return db.QueryRow("SELECT user_id, email, purpose, facility_id, start_dt, end_dt, transaction_dt FROM booking.booking WHERE id=$1",
		p.ID).Scan(&p.UserID, &p.Email, &p.Purpose, &p.FacilityID, &p.StartTime, &p.EndTime, &p.TransactionTime)
}

func (p *booking) updateBooking(db *sql.DB) error {
	currentTime := time.Now()
	_, err :=
		db.Exec("UPDATE booking.booking SET user_id=$1, email=$2, purpose=$3, facility_id=$4, start_dt=$5, end_dt=$6, transaction_dt=$7 WHERE id=$8",
			p.UserID, p.Email, p.Purpose, p.FacilityID, p.StartTime, p.EndTime, currentTime, p.ID)

	return err
}

func (p *booking) deleteBooking(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM booking.booking WHERE id=$1", p.ID)

	return err
}

func (p *booking) createBooking(db *sql.DB) error {
	currentTime := time.Now()
	err := db.QueryRow(
		"INSERT INTO booking.booking(user_id, email, purpose, facility_id, start_dt, end_dt, transaction_dt) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		p.UserID, p.Email, p.Purpose, p.FacilityID, p.StartTime, p.EndTime, currentTime).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil
}

func getBookings(db *sql.DB, start, count int, userid string) ([]booking, error) {
	var rows *sql.Rows
	var err error
	if len(userid) > 0 {
		rows, err = db.Query(
			"SELECT id, user_id, email, purpose, facility_id, start_dt, end_dt, transaction_dt FROM booking.booking WHERE user_id=$1 LIMIT $2 OFFSET $3",
			userid, count, start)
	} else {
		rows, err = db.Query(
			"SELECT id, user_id, email, purpose, facility_id, start_dt, end_dt, transaction_dt FROM booking.booking LIMIT $1 OFFSET $2",
			count, start)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	bookings := []booking{}

	for rows.Next() {
		var p booking
		if err := rows.Scan(&p.ID, &p.UserID, &p.Email, &p.Purpose, &p.FacilityID, &p.StartTime, &p.EndTime, &p.TransactionTime); err != nil {
			return nil, err
		}
		bookings = append(bookings, p)
	}

	return bookings, nil
}

func getBookingsCount(db *sql.DB, userid string) (int, error) {

	var count int
	var err error
	if len(userid) > 0 {
		err = db.QueryRow("SELECT COUNT (id) FROM booking.booking WHERE user_id=$1", userid).Scan(&count)
	} else {
		err = db.QueryRow("SELECT COUNT (id) FROM booking.booking").Scan(&count)
	}

	if err != nil {
		return 0, err
	}

	return count, nil

}
