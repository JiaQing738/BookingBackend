package main

import (
	"database/sql"
	"time"
)

type facilityDetail struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Level           string `json:"level"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	TransactionTime string `json:"transaction_dt"`
}

func (p *facilityDetail) getFacilityDetail(db *sql.DB) error {
	return db.QueryRow("SELECT name, level, description, status, transaction_dt FROM booking.facility_detail WHERE id=$1",
		p.ID).Scan(&p.Name, &p.Level, &p.Description, &p.Status, &p.TransactionTime)
}

func (p *facilityDetail) updateFacilityDetail(db *sql.DB) error {
	currentTime := time.Now()
	_, err :=
		db.Exec("UPDATE booking.facility_detail SET name=$1, level=$2, description=$3, status=$4, transaction_dt=$5 WHERE id=$6",
			p.Name, p.Level, p.Description, p.Status, currentTime, p.ID)

	return err
}

func (p *facilityDetail) deleteFacilityDetail(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM booking.facility_detail WHERE id=$1", p.ID)

	return err
}

func (p *facilityDetail) createFacilityDetail(db *sql.DB) error {
	currentTime := time.Now()
	err := db.QueryRow(
		"INSERT INTO booking.facility_detail(name, level, description, status, transaction_dt) VALUES($1, $2, $3, $4, $5) RETURNING id",
		p.Name, p.Level, p.Description, p.Status, currentTime).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil
}

func getFacilityDetails(db *sql.DB, start, count int) ([]facilityDetail, error) {
	rows, err := db.Query(
		"SELECT id, name, level, description, status, transaction_dt FROM booking.facility_detail LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	facilityDetails := []facilityDetail{}

	for rows.Next() {
		var p facilityDetail
		if err := rows.Scan(&p.ID, &p.Name, &p.Level, &p.Description, &p.Status, &p.TransactionTime); err != nil {
			return nil, err
		}
		facilityDetails = append(facilityDetails, p)
	}

	return facilityDetails, nil
}

func getFacilityDetailsCount(db *sql.DB) (int, error) {

	var count int
	err := db.QueryRow("SELECT COUNT (id) FROM booking.facility_detail").Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}
