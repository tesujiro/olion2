package olion

import termbox "github.com/nsf/termbox-go"

const (
	edge1 int = 1 << iota
	edge2
	edge3
	edge4
	edge5
	edge6
	edge7
)

const (
	digitMinus = edge4
	digit0     = edge1 | edge2 | edge3 | edge5 | edge6 | edge7
	digit1     = edge3 | edge6
	digit2     = edge1 | edge3 | edge4 | edge5 | edge7
	digit3     = edge1 | edge3 | edge4 | edge6 | edge7
	digit4     = edge2 | edge3 | edge4 | edge6
	digit5     = edge1 | edge2 | edge4 | edge6 | edge7
	digit6     = edge1 | edge2 | edge4 | edge5 | edge6 | edge7
	digit7     = edge1 | edge3 | edge6
	digit8     = edge1 | edge2 | edge3 | edge4 | edge5 | edge6 | edge7
	digit9     = edge1 | edge2 | edge3 | edge4 | edge6 | edge7
)

func digit(n int) int {
	switch n {
	case -1:
		return digitMinus
	case 0:
		return digit0
	case 1:
		return digit1
	case 2:
		return digit2
	case 3:
		return digit3
	case 4:
		return digit4
	case 5:
		return digit5
	case 6:
		return digit6
	case 7:
		return digit7
	case 8:
		return digit8
	case 9:
		return digit9
	}
	return 0
}

func disp_number(start Dot, n int) {
	color := colors.name("White").Attribute()
	setCell := func(dot *Dot, color Attribute) {
		termbox.SetCell(dot.X, dot.Y, ' ', termbox.ColorDefault, termbox.Attribute(color))
	}
	drawDigit := func(count int, digit int) {
		dot := Dot{start.X + count*4, start.Y}
		//   21113
		//   2   3
		//   24443
		//   5   6
		//   57776

		if digit&edge1 > 0 {
			setCell(&Dot{dot.X, dot.Y}, color)
			setCell(&Dot{dot.X + 1, dot.Y}, color)
			setCell(&Dot{dot.X + 2, dot.Y}, color)
		}
		if digit&edge2 > 0 {
			setCell(&Dot{dot.X, dot.Y}, color)
			setCell(&Dot{dot.X, dot.Y + 1}, color)
			setCell(&Dot{dot.X, dot.Y + 2}, color)
		}
		if digit&edge3 > 0 {
			setCell(&Dot{dot.X + 2, dot.Y}, color)
			setCell(&Dot{dot.X + 2, dot.Y + 1}, color)
			setCell(&Dot{dot.X + 2, dot.Y + 2}, color)
		}
		if digit&edge4 > 0 {
			setCell(&Dot{dot.X, dot.Y + 2}, color)
			setCell(&Dot{dot.X + 1, dot.Y + 2}, color)
			setCell(&Dot{dot.X + 2, dot.Y + 2}, color)
		}
		if digit&edge5 > 0 {
			setCell(&Dot{dot.X, dot.Y + 2}, color)
			setCell(&Dot{dot.X, dot.Y + 3}, color)
			setCell(&Dot{dot.X, dot.Y + 4}, color)
		}
		if digit&edge6 > 0 {
			setCell(&Dot{dot.X + 2, dot.Y + 2}, color)
			setCell(&Dot{dot.X + 2, dot.Y + 3}, color)
			setCell(&Dot{dot.X + 2, dot.Y + 4}, color)
		}
		if digit&edge7 > 0 {
			setCell(&Dot{dot.X, dot.Y + 4}, color)
			setCell(&Dot{dot.X + 1, dot.Y + 4}, color)
			setCell(&Dot{dot.X + 2, dot.Y + 4}, color)
		}
	}
	count := 0
	if n == 0 {
		drawDigit(count, digit(0))
		return
	} else if n < 0 {
		drawDigit(count, digit(-1))
		n = -n
		count++
	}
	var printNumber func(int, int) int
	printNumber = func(start int, d int) int {
		if d == 0 {
			return start
		}
		next := printNumber(start, d/10)
		drawDigit(next, digit(d%10))
		return next + 1
	}
	printNumber(count, n)
}

var alphabet [][]int

func init_alphabet() {
	alphabet = [][]int{
		{ // 'A'
			0, 0, 1, 0, 0,
			0, 1, 0, 1, 0,
			1, 0, 0, 0, 1,
			1, 1, 1, 1, 1,
			1, 0, 0, 0, 1,
		},
		{ // 'B'
			1, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 1, 1, 1, 0,
		},
		{ // 'C'
			0, 1, 1, 1, 1,
			1, 0, 0, 0, 0,
			1, 0, 0, 0, 0,
			1, 0, 0, 0, 0,
			0, 1, 1, 1, 1,
		},
	}

}
