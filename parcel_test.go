package main

import (
	"database/sql"
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
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// get
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, storedParcel)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// check deleted
	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db := setupDB(t)
	defer cleanupDB(t, db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	db := setupDB(t)
	defer cleanupDB(t, db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set status
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, updatedParcel.Status)
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
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		expected, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, expected, parcel)
	}
	// done
}
