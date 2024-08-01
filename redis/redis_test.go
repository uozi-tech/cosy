package redis

import (
    "git.uozi.org/uozi/cosy/settings"
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

func TestRedis(t *testing.T) {
    settings.Init("../app.ini")
    Init()

    err := Set("test", "test", 10*time.Second)
    if err != nil {
        t.Error(err)
    }
    ok, err := Exists("test")
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, true, ok)
    v, err := Get("test")
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, "test", v)
    assert.LessOrEqual(t, 10*time.Second, TTL("test"))

    inc, err := Incr("test_incr")
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, int64(1), inc)

    incStr, err := Get("test_incr")
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, "1", incStr)

    inc, err = Incr("test_incr")
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, int64(2), inc)

    decr, err := Decr("test_incr")
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, int64(1), decr)

    keys, err := Keys("test*")
    if err != nil {
        t.Error(err)
        return
    }
    assert.ObjectsAreEqual([]string{
        "test",
        "test_incr",
    }, keys)

    err = Del("test", "test_incr")
    if err != nil {
        t.Error(err)
    }
    v, _ = Get("test")
    assert.Equal(t, "", v)
    v, _ = Get("test_incr")
    assert.Equal(t, "", v)

    err = SetEx("test", "test", 10*time.Second)
    if err != nil {
        t.Error(err)
    }
    v, _ = Get("test")
    assert.Equal(t, "test", v)

    err = SetNx("test1", "test1", 10*time.Second)
    if err != nil {
        t.Error(err)
    }
    v, _ = Get("test1")
    assert.Equal(t, "test1", v)

    err = SetNx("test1", "test2", 10*time.Second)
    if err != nil {
        t.Error(err)
    }
    v, _ = Get("test1")
    assert.Equal(t, "test1", v)

    err = Set("test_do", "test_do", 10*time.Second)
    if err != nil {
        t.Error(err)
    }
    actual, err := Do("get", "cosy:test_do")
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, "test_do", actual)
}

func TestEval(t *testing.T) {
    settings.Init("../app.ini")
    Init()
    // Define test cases
    tests := []struct {
        name           string
        script         string
        numKeys        int
        keys           []string
        args           []interface{}
        expectedResult interface{}
        expectedError  error
    }{
        {
            name:           "Test with valid script and no keys or args",
            script:         "return 1+1",
            numKeys:        0,
            keys:           nil,
            args:           nil,
            expectedResult: int64(2),
            expectedError:  nil,
        },
        {
            name:           "Test with valid script and keys",
            script:         "return redis.call('SET', KEYS[1], ARGV[1])",
            numKeys:        1,
            keys:           []string{"key1"},
            args:           []interface{}{1},
            expectedResult: "OK",
            expectedError:  nil,
        },
        {
            name:           "Test with valid script and args",
            script:         "return redis.call('INCRBY', KEYS[1], ARGV[1])",
            numKeys:        1,
            keys:           []string{"key1"},
            args:           []interface{}{10},
            expectedResult: int64(11),
            expectedError:  nil,
        },
    }

    // Iterate through test cases
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            // Call the function under test
            result, err := Eval(tc.script, tc.numKeys, tc.keys, tc.args)

            // Check if the error is what we expect (if any)
            assert.Equal(t, tc.expectedError, err)

            // Check if the result is what we expect
            assert.Equal(t, tc.expectedResult, result)
        })
    }
}
