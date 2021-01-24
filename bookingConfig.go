package main

import (
	"database/sql"
)

type bookingConfig struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (p *bookingConfig) getBookingConfig(db *sql.DB) error {
	return db.QueryRow("SELECT key, value FROM booking.booking_config WHERE id=$1",
		p.ID).Scan(&p.Key, &p.Value)
}

func (p *bookingConfig) updateBookingConfig(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE booking.booking_config SET key=$1, value=$2 WHERE id=$3",
			p.Key, p.Value, p.ID)

	return err
}

func getBookingConfigs(db *sql.DB, start, count int) ([]bookingConfig, error) {
	rows, err := db.Query(
		"SELECT id, key, value FROM booking.booking_config LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	bookingConfigs := []bookingConfig{}

	for rows.Next() {
		var p bookingConfig
		if err := rows.Scan(&p.ID, &p.Key, &p.Value); err != nil {
			return nil, err
		}
		bookingConfigs = append(bookingConfigs, p)
	}

	return bookingConfigs, nil
}
