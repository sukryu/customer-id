package entities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/customer-id.git/internal/domain/entities"
)

func TestNewCustomer(t *testing.T) {
	cust, err := entities.NewCustomer("cust123", map[string]string{"drink": "coffee"})
	assert.NoError(t, err)
	assert.Equal(t, "cust123", cust.CustomerID)
	assert.NotEmpty(t, cust.LastSeen)
}

func TestNewBeacon(t *testing.T) {
	beacon, err := entities.NewBeacon("550e8400-e29b-41d4-a716-446655440000", "store100", 100, 3, "Table 3", entities.StatusActive)
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", beacon.BeaconID)
	assert.Equal(t, entities.StatusActive, beacon.Status)
}

func TestNewBeaconData(t *testing.T) {
	bd, err := entities.NewBeaconData("550e8400-e29b-41d4-a716-446655440000", 100, 3, -50)
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", bd.UUID())
	assert.Equal(t, int32(100), bd.Major())
	assert.Equal(t, int32(3), bd.Minor())
	assert.Equal(t, int32(-50), bd.RSSI())

	err = bd.Validate()
	assert.NoError(t, err)
}

func TestNewBeaconDataInvalidRSSI(t *testing.T) {
	_, err := entities.NewBeaconData("550e8400-e29b-41d4-a716-446655440000", 100, 3, 10) // RSSI > 0
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rssi must be between -100 and 0")
}

func TestNewLocation(t *testing.T) {
	loc, err := entities.NewLocation("Table 3", entities.LocationTypeTable)
	assert.NoError(t, err)
	assert.Equal(t, "Table 3", loc.Name())
	assert.Equal(t, entities.LocationTypeTable, loc.Type())

	err = loc.Validate()
	assert.NoError(t, err)
}

func TestNewLocationInvalidType(t *testing.T) {
	_, err := entities.NewLocation("Invalid", "invalid") // Invalid type
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location type")
}
