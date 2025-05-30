package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE parcel (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER NOT NULL,
			status TEXT NOT NULL,
			address TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	return db
}

func cleanupDB(t *testing.T, db *sql.DB) {
	err := db.Close()
	require.NoError(t, err)
}

func TestAddGetDelete(t *testing.T) {
	db := setupDB(t)
	defer cleanupDB(t, db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	if assert.NoError(t, err) {
		assert.NotZero(t, id)
	}
	parcel.Number = id

	// get
	storedParcel, err := store.Get(id)
	if assert.NoError(t, err) {
		assert.Equal(t, parcel, storedParcel)
	}

	// delete
	err = store.Delete(id)
	assert.NoError(t, err)

	// check deleted
	_, err = store.Get(id)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db := setupDB(t)
	defer cleanupDB(t, db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	if assert.NoError(t, err) {
		assert.NotZero(t, id)
	}

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	if assert.NoError(t, err) {
		updatedParcel, err := store.Get(id)
		if assert.NoError(t, err) {
			assert.Equal(t, newAddress, updatedParcel.Address)
		}
	}
}

func TestSetStatus(t *testing.T) {
	db := setupDB(t)
	defer cleanupDB(t, db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	if assert.NoError(t, err) {
		assert.NotZero(t, id)
	}

	// set status
	err = store.SetStatus(id, ParcelStatusSent)
	if assert.NoError(t, err) {
		updatedParcel, err := store.Get(id)
		if assert.NoError(t, err) {
			assert.Equal(t, ParcelStatusSent, updatedParcel.Status)
		}
	}
}

func TestGetByClient(t *testing.T) {
	db := setupDB(t)
	defer cleanupDB(t, db)

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if assert.NoError(t, err) {
			assert.NotZero(t, id)
			parcels[i].Number = id
			parcelMap[id] = parcels[i]
		}
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if assert.NoError(t, err) {
		assert.Len(t, storedParcels, len(parcels))

		// check
		for _, parcel := range storedParcels {
			expected, exists := parcelMap[parcel.Number]
			if assert.True(t, exists) {
				assert.Equal(t, expected, parcel)
			}
		}
	}
}
