package funk

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceOf(t *testing.T) {
	is := assert.New(t)

	f := &Foo{
		ID:        1,
		FirstName: "Dark",
		LastName:  "Vador",
		Age:       30,
		Bar: &Bar{
			Name: "Test",
		},
	}

	result := SliceOf(f)

	resultType := reflect.TypeOf(result)

	is.True(resultType.Kind() == reflect.Slice)
	is.True(resultType.Elem().Kind() == reflect.Ptr)

	elemType := resultType.Elem().Elem()

	is.True(elemType.Kind() == reflect.Struct)

	value := reflect.ValueOf(result)

	is.Equal(value.Len(), 1)

	_, ok := value.Index(0).Interface().(*Foo)

	is.True(ok)
}

func TestRandomInt(t *testing.T) {
	is := assert.New(t)

	is.True(RandomInt(0, 10) <= 10)
}

func TestShard(t *testing.T) {
	is := assert.New(t)

	tokey := "e89d66bdfdd4dd26b682cc77e23a86eb"

	is.Equal(Shard(tokey, 1, 2, false), []string{"e", "8", "e89d66bdfdd4dd26b682cc77e23a86eb"})
	is.Equal(Shard(tokey, 2, 2, false), []string{"e8", "9d", "e89d66bdfdd4dd26b682cc77e23a86eb"})
	is.Equal(Shard(tokey, 2, 3, true), []string{"e8", "9d", "66", "bdfdd4dd26b682cc77e23a86eb"})
}

func TestRandomString(t *testing.T) {
	is := assert.New(t)

	is.Len(RandomString(10), 10)

	result := RandomString(10, []rune("abcdefg"))

	is.Len(result, 10)

	for _, char := range result {
		is.True(char >= []rune("a")[0] && char <= []rune("g")[0])
	}
}
