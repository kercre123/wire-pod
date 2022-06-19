package util_test

import (
	"testing"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/util"

	"github.com/stretchr/testify/assert"
)

func TestDoOnce(t *testing.T) {
	i := 0
	f := func() {
		i = i + 1
	}
	once := util.NewDoOnce(f)
	assert.Equal(t, i, 0)
	for x := 0; x < 10; x++ {
		once.Do()
		assert.Equal(t, i, 1)
	}
}

func TestCanSelect(t *testing.T) {
	// nil channel blocks
	assert.False(t, util.CanSelect(nil))

	ch := make(chan struct{})
	// can't select channel with nothing on it
	assert.False(t, util.CanSelect(ch))

	// can select after sending to it
	go func() {
		ch <- struct{}{}
	}()
	time.Sleep(time.Millisecond)
	assert.True(t, util.CanSelect(ch))
	// now that it's been pulled from, can't select
	assert.False(t, util.CanSelect(ch))

	// can select after closing
	close(ch)
	assert.True(t, util.CanSelect(ch))
	assert.True(t, util.CanSelect(ch))
}

func TestChanWriter(t *testing.T) {
	ch := make(chan []byte)
	cw := util.NewChanWriter(ch)

	write := func(buf []byte) {
		go func() {
			n, err := cw.Write(buf)
			assert.Equal(t, n, len(buf))
			assert.Nil(t, err)
		}()
		time.Sleep(time.Millisecond)
	}
	write(nil)
	assert.Nil(t, <-ch)
	write([]byte{1})
	assert.Equal(t, <-ch, []byte{1})
}

func TestTimeFunc(t *testing.T) {
	dur := util.TimeFuncMs(func() {})
	assert.True(t, dur < 1)
	dur = util.TimeFuncMs(func() {
		time.Sleep(5 * time.Millisecond)
	})
	assert.True(t, dur >= 5.0)
}

// func TestSleepSelect(t *testing.T) {
// 	ch := make(chan struct{})
// 	assert.False(t, util.SleepSelect(time.Millisecond, ch))

// 	time.AfterFunc(time.Millisecond, func() {
// 		close(ch)
// 	})
// 	assert.True(t, util.SleepSelect(8*time.Millisecond, ch))

// 	// canceled contexts should also trigger
// 	assert.False(t, util.SleepSelect(time.Millisecond, context.Background().Done()))
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
// 	defer cancel()
// 	assert.False(t, util.SleepSelect(time.Millisecond, ctx.Done()))
// 	assert.True(t, util.SleepSelect(15*time.Millisecond, ctx.Done()))
// }
