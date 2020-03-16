package dbmodel

import (
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/require"

	dbtest "isc.org/stork/server/database/test"
)

// This function creates multiple hosts used in tests which fetch and
// filter hosts.
func addTestHosts(t *testing.T, db *pg.DB) []Host {
	subnets := []Subnet{
		{
			ID:     1,
			Prefix: "192.0.2.0/24",
		},
		{
			ID:     2,
			Prefix: "2001:db8:1::/64",
		},
	}
	for i, s := range subnets {
		subnet := s
		err := AddSubnet(db, &subnet)
		require.NoError(t, err)
		require.NotZero(t, subnet.ID)
		subnets[i] = subnet
	}

	hosts := []Host{
		{
			SubnetID: 1,
			HostIdentifiers: []HostIdentifier{
				{
					Type:  "hw-address",
					Value: []byte{1, 2, 3, 4, 5, 6},
				},
				{
					Type:  "circuit-id",
					Value: []byte{1, 2, 3, 4},
				},
			},
			IPReservations: []IPReservation{
				{
					Address: "192.0.2.4/32",
				},
				{
					Address: "192.0.2.5/32",
				},
			},
		},
		{
			HostIdentifiers: []HostIdentifier{
				{
					Type:  "hw-address",
					Value: []byte{2, 3, 4, 5, 6, 7},
				},
				{
					Type:  "circuit-id",
					Value: []byte{2, 3, 4, 5},
				},
			},
			IPReservations: []IPReservation{
				{
					Address: "192.0.2.6/32",
				},
				{
					Address: "192.0.2.7/32",
				},
			},
		},
		{
			SubnetID: 2,
			HostIdentifiers: []HostIdentifier{
				{
					Type:  "hw-address",
					Value: []byte{1, 2, 3, 4, 5, 6},
				},
			},
			IPReservations: []IPReservation{
				{
					Address: "2001:db8:1::1/128",
				},
			},
		},
		{
			HostIdentifiers: []HostIdentifier{
				{
					Type:  "duid",
					Value: []byte{1, 2, 3, 4},
				},
			},
			IPReservations: []IPReservation{
				{
					Address: "2001:db8:1::2/128",
				},
			},
		},
	}

	for i, h := range hosts {
		host := h
		err := AddHost(db, &host)
		require.NoError(t, err)
		require.NotZero(t, host.ID)
		hosts[i] = host
	}
	return hosts
}

// This test verifies that the new host along with identifiers and reservations
// can be added to the database.
func TestAddHost(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Add a host with two identifiers and two reservations.
	host := &Host{
		HostIdentifiers: []HostIdentifier{
			{
				Type:  "hw-address",
				Value: []byte{1, 2, 3, 4, 5, 6},
			},
			{
				Type:  "circuit-id",
				Value: []byte{1, 1, 1, 1, 1},
			},
		},
		IPReservations: []IPReservation{
			{
				Address: "192.0.2.4",
			},
			{
				Address: "2001:db8:1::4",
			},
		},
	}
	err := AddHost(db, host)
	require.NoError(t, err)
	require.NotZero(t, host.ID)

	// Get the host from the database.
	returned, err := GetHost(db, host.ID)
	require.NoError(t, err)
	require.NotNil(t, returned)

	require.Equal(t, host.ID, returned.ID)
	require.Len(t, returned.HostIdentifiers, 2)
	require.Len(t, returned.IPReservations, 2)

	// Make sure that the returned host identifiers match.
	for i := range returned.HostIdentifiers {
		require.Contains(t, returned.HostIdentifiers, host.HostIdentifiers[i])
	}

	// Make sure that the returned reservations match.
	for i := range returned.IPReservations {
		require.Contains(t, returned.IPReservations[i].Address, host.IPReservations[i].Address)
	}
}

// Test that the host can be updated and that this update includes extending
// the list of reservations and identifiers.
func TestUpdateHostExtend(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Add the host with two reservations and two identifiers.
	host := &Host{
		HostIdentifiers: []HostIdentifier{
			{
				Type:  "hw-address",
				Value: []byte{1, 2, 3, 4, 5, 6},
			},
			{
				Type:  "circuit-id",
				Value: []byte{1, 1, 1, 1, 1},
			},
		},
		IPReservations: []IPReservation{
			{
				Address: "192.0.2.4/32",
			},
			{
				Address: "2001:db8:1::4/128",
			},
		},
	}
	err := AddHost(db, host)
	require.NoError(t, err)
	require.NotZero(t, host.ID)

	// Modify the value of the first identifier.
	host.HostIdentifiers[0].Value = []byte{6, 5, 4, 3, 2, 1}
	// Modify the identifier type of the second identifier.
	host.HostIdentifiers[1].Type = "client-id"
	// Add one more identifier.
	host.HostIdentifiers = append(host.HostIdentifiers, HostIdentifier{
		Type:  "flex-id",
		Value: []byte{2, 2, 2, 2, 2},
	})

	// Modify the first reservation.
	host.IPReservations[0].Address = "192.0.3.4/32"
	// Add one more reservation.
	host.IPReservations = append(host.IPReservations, IPReservation{
		Address: "3000::/64",
	})

	// Not only does updating the host modify the host value but also adds
	// or removes reservations and identifiers.
	err = UpdateHost(db, host)
	require.NoError(t, err)
	require.NotZero(t, host.ID)

	// Get the updated host.
	returned, err := GetHost(db, host.ID)
	require.NoError(t, err)
	require.NotNil(t, returned)

	require.Len(t, returned.HostIdentifiers, 3)
	require.Len(t, returned.IPReservations, 3)

	// Make sure that the identifiers and reservations were modified.
	require.ElementsMatch(t, returned.HostIdentifiers, host.HostIdentifiers)
	require.ElementsMatch(t, returned.IPReservations, host.IPReservations)
}

// Test that the host can be updated and that some reservations and
// host identifiers are deleted as a result of this update.
func TestUpdateHostShrink(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	host := &Host{
		HostIdentifiers: []HostIdentifier{
			{
				Type:  "hw-address",
				Value: []byte{1, 2, 3, 4, 5, 6},
			},
			{
				Type:  "circuit-id",
				Value: []byte{1, 1, 1, 1, 1},
			},
		},
		IPReservations: []IPReservation{
			{
				Address: "192.0.2.4/32",
			},
			{
				Address: "2001:db8:1::4/128",
			},
		},
	}
	err := AddHost(db, host)
	require.NoError(t, err)
	require.NotZero(t, host.ID)

	// Remove one host identifier and one reservation.
	host.HostIdentifiers = host.HostIdentifiers[0:1]
	host.IPReservations = host.IPReservations[1:]

	// Updating the host should result in removal of this identifier
	// and the reservation from the database.
	err = UpdateHost(db, host)
	require.NoError(t, err)
	require.NotZero(t, host.ID)

	// Get the updated host.
	returned, err := GetHost(db, host.ID)
	require.NoError(t, err)
	require.NotNil(t, returned)

	// Verify that only one identifier and one reservation have left.
	require.Len(t, returned.HostIdentifiers, 1)
	require.Len(t, returned.IPReservations, 1)

	require.Equal(t, "hw-address", returned.HostIdentifiers[0].Type)
	require.Equal(t, []byte{1, 2, 3, 4, 5, 6}, returned.HostIdentifiers[0].Value)

	require.Equal(t, "2001:db8:1::4/128", returned.IPReservations[0].Address)
}

// Test that all hosts or all hosts having IP reservations of specified family
// can be fetched.
func TestGetAllHosts(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Add four hosts. Two with IPv4 and two with IPv6 reservations.
	hosts := addTestHosts(t, db)

	// Fetch all hosts having IPv4 reservations.
	returned, err := GetAllHosts(db, 4)
	require.NoError(t, err)
	require.Len(t, returned, 2)

	require.Contains(t, returned, hosts[0])
	require.Contains(t, returned, hosts[1])

	// Fetch all hosts having IPv6 reservations.
	returned, err = GetAllHosts(db, 6)
	require.NoError(t, err)
	require.Len(t, returned, 2)

	require.Contains(t, returned, hosts[2])
	require.Contains(t, returned, hosts[3])

	// Fetch all hosts.
	returned, err = GetAllHosts(db, 0)
	require.NoError(t, err)
	require.Len(t, returned, 4)

	for _, host := range hosts {
		require.Contains(t, returned, host)
	}
}

// Test that hosts can be fetched by subnet ID.
func TestGetHostsBySubnetID(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Add four hosts. Two with IPv4 and two with IPv6 reservations.
	hosts := addTestHosts(t, db)

	// Fetch host having IPv4 reservations.
	returned, err := GetHostsBySubnetID(db, 1)
	require.NoError(t, err)
	require.Len(t, returned, 1)
	require.Contains(t, returned, hosts[0])

	// Fetch host having IPv6 reservations.
	returned, err = GetHostsBySubnetID(db, 2)
	require.NoError(t, err)
	require.Len(t, returned, 1)
	require.Contains(t, returned, hosts[2])
}

// Test that page of the hosts can be fetched without filtering.
func TestGetHostsByPageNoFiltering(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Add four hosts. Two with IPv4 and two with IPv6 reservations.
	_ = addTestHosts(t, db)

	returned, _, err := GetHostsByPage(db, 0, 10, nil, nil)
	require.NoError(t, err)
	require.Len(t, returned, 4)
}

// Test that page of the hosts can be fetched with filtering by subnet id.
func TestGetHostsByPageSubnet(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Add four hosts. Two with IPv4 and two with IPv6 reservations.
	hosts := addTestHosts(t, db)

	// Get global hosts only.
	subnetID := int64(0)
	returned, _, err := GetHostsByPage(db, 0, 10, &subnetID, nil)
	require.NoError(t, err)
	require.Len(t, returned, 2)
	require.Contains(t, returned, hosts[1])
	require.Contains(t, returned, hosts[3])

	// Get hosts associated with subnet id 1.
	subnetID = int64(1)
	returned, _, err = GetHostsByPage(db, 0, 10, &subnetID, nil)
	require.NoError(t, err)
	require.Len(t, returned, 1)
	require.Contains(t, returned, hosts[0])

	// Get hosts associated with subnet id 2.
	subnetID = int64(2)
	returned, _, err = GetHostsByPage(db, 0, 10, &subnetID, nil)
	require.NoError(t, err)
	require.Len(t, returned, 1)
	require.Contains(t, returned, hosts[2])
}

// Test that page of the hosts can be filtered by IP reservations.
func TestGetHostsByPageFilteringText(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Add four hosts. Two with IPv4 and two with IPv6 reservations.
	hosts := addTestHosts(t, db)

	filterText := "0.2.4"
	returned, _, err := GetHostsByPage(db, 0, 10, nil, &filterText)
	require.NoError(t, err)
	require.Len(t, returned, 1)
	require.Contains(t, returned, hosts[0])

	filterText = "192.0.2"
	returned, _, err = GetHostsByPage(db, 0, 10, nil, &filterText)
	require.NoError(t, err)
	require.Len(t, returned, 2)
	require.Contains(t, returned, hosts[0])
	require.Contains(t, returned, hosts[1])

	filterText = "0"
	returned, _, err = GetHostsByPage(db, 0, 10, nil, &filterText)
	require.NoError(t, err)
	require.Len(t, returned, 4)

	require.ElementsMatch(t, returned, hosts)
}

// Test that the host and its identifiers and reservations can be
// deleted.
func TestDeleteHost(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	host := &Host{
		HostIdentifiers: []HostIdentifier{
			{
				Type:  "hw-address",
				Value: []byte{1, 2, 3, 4, 5, 6},
			},
		},
		IPReservations: []IPReservation{
			{
				Address: "192.0.2.4/32",
			},
		},
	}
	err := AddHost(db, host)
	require.NoError(t, err)
	require.NotZero(t, host.ID)

	err = DeleteHost(db, host.ID)
	require.NoError(t, err)

	returned, err := GetHost(db, host.ID)
	require.NoError(t, err)
	require.Nil(t, returned)
}

// Test that an app can be associated with a host.
func TestAddAppToHost(t *testing.T) {
	db, _, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	// Insert apps and hosts into the database.
	apps := addTestSubnetApps(t, db)
	hosts := addTestHosts(t, db)

	// Associate the first host with the first app.
	host := hosts[0]
	err := AddAppToHost(db, &host, apps[0], "api")
	require.NoError(t, err)

	// Fetch the host from the database.
	returned, err := GetHost(db, host.ID)
	require.NoError(t, err)
	require.NotNil(t, returned)

	// Make sure that the host includes the local host information which
	// associates the app with the host.
	require.Len(t, returned.LocalHosts, 1)
	require.Equal(t, "api", returned.LocalHosts[0].DataSource)
	require.EqualValues(t, apps[0].ID, returned.LocalHosts[0].AppID)
	// When fetching one selected host the app information should be also
	// returned.
	require.NotNil(t, returned.LocalHosts[0].App)

	// Get all hosts.
	returnedList, err := GetAllHosts(db, 0)
	require.NoError(t, err)
	require.Len(t, returnedList, 4)
	require.Len(t, returnedList[0].LocalHosts, 1)
	require.Equal(t, "api", returnedList[0].LocalHosts[0].DataSource)
	require.EqualValues(t, apps[0].ID, returnedList[0].LocalHosts[0].AppID)
	// When fetching all hosts, the detailed app information should not be
	// returned.
	require.Nil(t, returnedList[0].LocalHosts[0].App)

	// Get the first host by reserved IP address.
	filterText := "192.0.2.4"
	returnedList, _, err = GetHostsByPage(db, 0, 10, nil, &filterText)
	require.NoError(t, err)
	require.Len(t, returnedList, 1)
	require.Len(t, returnedList[0].LocalHosts, 1)
	require.Equal(t, "api", returnedList[0].LocalHosts[0].DataSource)
	require.EqualValues(t, apps[0].ID, returnedList[0].LocalHosts[0].AppID)
	// This time the detailed app information should not be included.
	require.Nil(t, returnedList[0].LocalHosts[0].App)
}