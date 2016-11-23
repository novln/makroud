package sqlxx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByParams(t *testing.T) {
	is := assert.New(t)

	db, _, shutdown := dbConnection(t)
	defer shutdown()

	user := User{}
	require.NoError(t, GetByParams(db, &user, map[string]interface{}{"username": "jdoe"}))

	is.Equal(1, user.ID)
	is.Equal("jdoe", user.Username)
	is.True(user.IsActive)
	is.NotZero(user.CreatedAt)
	is.NotZero(user.UpdatedAt)
}

// func TestFindByParams(t *testing.T) {
// 	is := assert.New(t)

// 	db, _, shutdown := dbConnection(t)
// 	defer shutdown()

// 	// TODO: handle interface conversion for []sqlxx.User
// 	users := []User{}
// 	require.NoError(t, FindByParams(db, &users, nil))

// 	is.Len(users, 1)

// 	user := users[0]
// 	is.Equal(1, user.ID)
// 	is.Equal("jdoe", user.Username)
// 	is.True(user.IsActive)
// 	is.NotZero(user.CreatedAt)
// 	is.NotZero(user.UpdatedAt)
// }
