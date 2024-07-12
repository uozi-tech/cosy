package redis

import (
    "git.uozi.org/uozi/cosy/settings"
    "github.com/stretchr/testify/assert"
    "math/rand"
    "testing"
    "time"
)

func generateRandomKey(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[seededRand.Intn(len(charset))]
    }
    return string(b)
}

func TestLPushAndLRange_Success(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    key := generateRandomKey(10)
    values := []interface{}{"value1", "value2", "value3"}

    err := LPush(key, values...)
    assert.NoError(t, err)

    result, err := LRange(key, 0, -1)
    assert.NoError(t, err)
    assert.Equal(t, []string{"value3", "value2", "value1"}, result)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func TestRPushAndLRange_Success(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    key := generateRandomKey(10)
    values := []interface{}{"value1", "value2", "value3"}

    err := RPush(key, values...)
    assert.NoError(t, err)

    result, err := LRange(key, 0, -1)
    assert.NoError(t, err)
    assert.Equal(t, []string{"value1", "value2", "value3"}, result)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func TestLLen_EmptyAndNonEmptyList(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    emptyKey := generateRandomKey(10)
    nonEmptyKey := generateRandomKey(10)

    err := LPush(nonEmptyKey, "value")
    assert.NoError(t, err)

    emptyListLength, err := LLen(emptyKey)
    assert.NoError(t, err)
    assert.Equal(t, int64(0), emptyListLength)

    nonEmptyListLength, err := LLen(nonEmptyKey)
    assert.NoError(t, err)
    assert.Equal(t, int64(1), nonEmptyListLength)

    // Clean up
    err = Del(nonEmptyKey)
    assert.NoError(t, err)
}

func TestLPop_NonEmptyList(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    key := generateRandomKey(10)
    err := LPush(key, "value1", "value2")
    assert.NoError(t, err)

    result, err := LPop(key)
    assert.NoError(t, err)
    assert.Equal(t, "value2", result) // Assuming LPush inserts in reverse order

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func TestRPop_NonEmptyList(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    key := generateRandomKey(10)
    err := RPush(key, "value1", "value2")
    assert.NoError(t, err)

    result, err := RPop(key)
    assert.NoError(t, err)
    assert.Equal(t, "value2", result)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func TestLRem_RemoveExistingAndNonExistingElement(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    key := generateRandomKey(10)
    err := LPush(key, "value1", "value2", "value1")
    assert.NoError(t, err)

    removedCount, err := LRem(key, "value1")
    assert.NoError(t, err)
    assert.Equal(t, int64(2), removedCount)

    removedCountNonExisting, err := LRem(key, "value3")
    assert.NoError(t, err)
    assert.Equal(t, int64(0), removedCountNonExisting)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func TestInsertIntoList_InsertBeforeAndAfter(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    key := generateRandomKey(10)
    err := LPush(key, "pivot")
    assert.NoError(t, err)

    _, err = InsertIntoList(key, "pivot", "beforePivot", true)
    assert.NoError(t, err)

    _, err = InsertIntoList(key, "pivot", "afterPivot", false)
    assert.NoError(t, err)

    result, err := LRange(key, 0, -1)
    assert.NoError(t, err)
    assert.Equal(t, []string{"beforePivot", "pivot", "afterPivot"}, result)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func TestLIndex_ValidAndInvalidIndex(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    key := generateRandomKey(10)
    err := LPush(key, "value1", "value2")
    assert.NoError(t, err)

    result, err := LIndex(key, 0)
    assert.NoError(t, err)
    assert.Equal(t, "value2", result) // Assuming LPush inserts in reverse order

    _, err = LIndex(key, 10)
    assert.Error(t, err) // Expecting an error for an out-of-range index

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func GetlistpageSuccesswithfullpage(t *testing.T) {
    key := generateRandomKey(10)
    // Assuming LPush inserts in reverse order and we want the page to reflect that
    values := []interface{}{"value1", "value2", "value3", "value4", "value5"}
    for _, value := range values {
        err := LPush(key, value)
        assert.NoError(t, err)
    }

    pageIndex := int64(1)
    pageSize := int64(3)
    expectedPage := []string{"value2", "value1"}

    result, err := GetListPage(key, pageIndex, pageSize)
    assert.NoError(t, err)
    assert.Equal(t, expectedPage, result)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func GetlistpageSuccesswithpartialpage(t *testing.T) {
    key := generateRandomKey(10)
    values := []interface{}{"value1", "value2", "value3"}
    for _, value := range values {
        err := LPush(key, value)
        assert.NoError(t, err)
    }

    pageIndex := int64(0)
    pageSize := int64(5)
    expectedPage := []string{"value3", "value2", "value1"}

    result, err := GetListPage(key, pageIndex, pageSize)
    assert.NoError(t, err)
    assert.Equal(t, expectedPage, result)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func GetlistpageEmptylistreturnsemptypage(t *testing.T) {
    key := generateRandomKey(10)

    pageIndex := int64(0)
    pageSize := int64(3)

    result, err := GetListPage(key, pageIndex, pageSize)
    assert.NoError(t, err)
    assert.Empty(t, result)
}

func TestGetListPage(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    GetlistpageSuccesswithfullpage(t)
    GetlistpageSuccesswithpartialpage(t)
    GetlistpageEmptylistreturnsemptypage(t)
}

func GetListReturnsAllElements(t *testing.T) {
    key := generateRandomKey(10)
    values := []interface{}{"element1", "element2", "element3"}
    for _, value := range values {
        err := LPush(key, value)
        assert.NoError(t, err)
    }

    result, err := GetList(key)
    assert.NoError(t, err)
    assert.Equal(t, []string{"element3", "element2", "element1"}, result)

    // Clean up
    err = Del(key)
    assert.NoError(t, err)
}

func GetListEmptyListReturnsEmptySlice(t *testing.T) {
    key := generateRandomKey(10)

    result, err := GetList(key)
    assert.NoError(t, err)
    assert.Empty(t, result)
}

func GetListNonExistentKeyReturnsEmptySlice(t *testing.T) {
    key := generateRandomKey(10) // Key that has not been used

    result, err := GetList(key)
    assert.NoError(t, err)
    assert.Empty(t, result)
}

func TestGetList(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    GetListReturnsAllElements(t)
    GetListEmptyListReturnsEmptySlice(t)
    GetListNonExistentKeyReturnsEmptySlice(t)
}
