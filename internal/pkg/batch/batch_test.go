// package batch

// import (
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// )

// /* Submits 10 jobs into a pool that supports 2 simultaneous jobs, and
//    tests that only two of the jobs ran at a time by capturing the time
//    that they ran and comparing against the start time. */
// func TestBatchPool (t *testing.T) {
// 	pool := New(2)
// 	var times []time.Time
// 	for i := 0; i <= 10; i++ {
// 		pool.Submit(func() {
// 			times = append(times, time.Now())
// 			time.Sleep(1 * time.Second)
// 		})
// 	}
// 	startTime := time.Now()
// 	pool.Run()
// 	assert.Equal(t, 0 * time.Second, times[0].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 0 * time.Second, times[1].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 1 * time.Second, times[2].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 1 * time.Second, times[3].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 2 * time.Second, times[4].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 2 * time.Second, times[5].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 3 * time.Second, times[6].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 3 * time.Second, times[7].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 4 * time.Second, times[8].Sub(startTime).Round(time.Second))
// 	assert.Equal(t, 4 * time.Second, times[9].Sub(startTime).Round(time.Second))
// }
