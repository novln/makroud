package sqlxx_test

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/ulule/sqlxx"
)

func TestSave_Save(t *testing.T) {
	env := setup(t)
	defer env.teardown()

	user := User{Username: "thoas"}

	queries, err := sqlxx.SaveWithQueries(env.driver, &user)
	assert.NoError(t, err)
	assert.NotNil(t, queries)
	assert.Len(t, queries, 1)
	assert.Contains(t, queries[0].Query, "INSERT INTO")

	assert.NotZero(t, user.ID)
	assert.Equal(t, true, user.IsActive)
	assert.NotZero(t, user.UpdatedAt)

	user.Username = "gilles"

	queries, err = sqlxx.SaveWithQueries(env.driver, &user)
	assert.NoError(t, err)
	assert.Contains(t, queries[0].Query, "UPDATE users SET")
	assert.Contains(t, queries[0].Query, "username = :username")

	m := map[string]interface{}{"username": "gilles"}

	query := `
	SELECT count(*)
	FROM %s
	WHERE username = :username
	`

	stmt, err := env.driver.PrepareNamed(fmt.Sprintf(query, user.TableName()))
	assert.Nil(t, err)

	var count int
	err = stmt.Get(&count, m)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}
