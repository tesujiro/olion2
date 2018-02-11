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
	//ctx, cancel = context.WithCancel(context.Background())
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

func TestAddObj(t *testing.T) {
	objects := 10
	spc := NewSpace(ctx, cancel, objects)
	spc.AddObj(newStar(time.Now(), 10, Coordinates{}))
	expected := objects + 1
	actual := len(spc.GetObjects())
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestVanish(t *testing.T) {
	objects := 10
	spc := NewSpace(ctx, cancel, objects)
	expected := objects
	spc.Vanish(spc.GetObjects()[0])
	actual := len(spc.GetObjects())
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}
