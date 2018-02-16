package olion

import "testing"

func InitTest_Color() error {
	return InitColor()
}

func TestColorAttribute(t *testing.T) {
	cases := []struct {
		name     string
		expected Attribute
	}{
		{name: "Red", expected: 10},
	}
	for _, c := range cases {
		actual := colors.name(c.name).Attribute()
		if actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}
