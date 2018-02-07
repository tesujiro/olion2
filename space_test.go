package olion

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	termbox "github.com/nsf/termbox-go"
)

var cancel func()
var ctx context.Context

func InitTest() {
	rand.Seed(time.Now().UnixNano())
	InitColor()
	ctx, cancel = context.WithCancel(context.Background())
}

func TestMain(m *testing.M) {
	// ここにテストの初期化処理
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	InitTest()

	code := m.Run()

	// ここでテストのお片づけ
	defer os.Exit(code)
	defer termbox.Close()
}

func TestGetObjects(t *testing.T) {
	cases := []struct {
		objects int
	}{
		{objects: 0},
		{objects: 10},
		{objects: 20},
	}
	for _, c := range cases {
		n := c.objects
		spc := NewSpace(ctx, cancel, n)
		expected := n
		actual := len(spc.GetObjects())
		if actual != expected {
			t.Errorf("got %v\nwant %v", actual, expected)
		}
	}
}
