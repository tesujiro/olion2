package olion

import (
	"context"
	"math/rand"
	"testing"
	"time"

	termbox "github.com/nsf/termbox-go"
)

func TestGetObjects(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	ctx, cancel := context.WithCancel(context.Background())
	n := 10
	spc := NewSpace(ctx, cancel, n)
	expected := n
	actual := len(spc.GetObjects())
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}
