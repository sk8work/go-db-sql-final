package main

import (
	"database/sql"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?", number)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	/// Проверяем, что посылка имеет статус "registered"
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&currentStatus)
	if err != nil {
		return err
	}

	if currentStatus != ParcelStatusRegistered {
		return sql.ErrNoRows
	}

	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// Проверяем, что посылка имеет статус "registered"
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&currentStatus)
	if err != nil {
		return err
	}

	if currentStatus != ParcelStatusRegistered {
		return sql.ErrNoRows
	}

	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
	if err != nil {
		return err
	}
	return nil
}
