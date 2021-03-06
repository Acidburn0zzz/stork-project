package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	dbmodel "isc.org/stork/server/database/model"
)

// Helper function checking if the user belonging to the specified group
// has access to the resource.
func authorizeAccept(t *testing.T, groupID int, path string) bool {
	// Create user with ID 5 and specified group id if the group id is
	// positive.
	user := &dbmodel.SystemUser{
		ID: 5,
	}
	if groupID > 0 {
		user.Groups = []*dbmodel.SystemGroup{
			{
				ID: groupID,
			},
		}
	}

	// Create request with the specified path and authorize.
	req, _ := http.NewRequest("GET", "http://example.org/api"+path, nil)
	ok, err := Authorize(user, req)
	require.NoError(t, err)

	return ok
}

// Verify that users belonging to the super-admin and admin group
// has appropriate access privileges.
func TestAuthorize(t *testing.T) {
	// admin group have limited access to the users' management
	require.False(t, authorizeAccept(t, 2, "/users?start=0&limit=10"))
	require.False(t, authorizeAccept(t, 2, "/users/list"))
	require.False(t, authorizeAccept(t, 2, "/users/4/password"))
	require.False(t, authorizeAccept(t, 2, "/users//4/password/"))
	require.True(t, authorizeAccept(t, 2, "/users/5"))
	require.True(t, authorizeAccept(t, 2, "/users/5/password"))
	require.True(t, authorizeAccept(t, 2, "/users//5//password"))

	// super-admin has no such restrictions
	require.True(t, authorizeAccept(t, 1, "/users?start=0&limit=10"))
	require.True(t, authorizeAccept(t, 1, "/users/list"))
	require.True(t, authorizeAccept(t, 1, "/users/4/password"))
	require.True(t, authorizeAccept(t, 1, "/users//4/password/"))
	require.True(t, authorizeAccept(t, 1, "/users/5"))
	require.True(t, authorizeAccept(t, 1, "/users/5/password"))
	require.True(t, authorizeAccept(t, 1, "/users//5//password"))

	// admin group have no restriction on machines
	require.True(t, authorizeAccept(t, 2, "/machines/1/"))

	// but someone who belongs to no groups would not be able
	// to access machines
	require.False(t, authorizeAccept(t, 0, "/machines/1/"))

	// the same in case of someone belonging to non existing group
	require.False(t, authorizeAccept(t, 3, "/machines/1/"))
}
