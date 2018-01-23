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

//var alphabet [][]int
var alphabet map[rune]int

func init_alphabet() {
	alphabet = make(map[rune]int)
	alphabet['A'] = 0x4547F1
	alphabet['B'] = 0x1E8FA3E
	alphabet['C'] = 0xF8420F
	alphabet['D'] = 0x1E8C63E
	alphabet['E'] = 0x1F8721F
	alphabet['F'] = 0x1F87210
	alphabet['G'] = 0xF84E2E
	alphabet['H'] = 0x118FE31
	alphabet['I'] = 0x1F2109F
	alphabet['J'] = 0x1F10A4C
	alphabet['K'] = 0x1197251
	alphabet['L'] = 0x108421F
	alphabet['M'] = 0x11DD631
	alphabet['N'] = 0x11CD671
	alphabet['O'] = 0xE8C62E
	alphabet['P'] = 0x1E8FA10
	alphabet['Q'] = 0xE8D64D
	alphabet['R'] = 0x1E8FA31
	alphabet['S'] = 0xF8383E
	alphabet['T'] = 0x1F21084
	alphabet['U'] = 0x118C62E
	alphabet['V'] = 0x118C544
	alphabet['W'] = 0x118D771
	alphabet['X'] = 0x1151151
	alphabet['Y'] = 0x1151084
	alphabet['Z'] = 0x1F1111F
	alphabet['0'] = 0xE9D72E
	alphabet['1'] = 0xC2108E
	alphabet['2'] = 0x1E0BA1F
	alphabet['3'] = 0x1E0983E
	alphabet['4'] = 0x654BE2
	alphabet['5'] = 0x1F8783E
	alphabet['6'] = 0xF87A2E
	alphabet['7'] = 0x1F08888
	alphabet['8'] = 0xE8BA2E
	alphabet['9'] = 0xE8BC3E
	alphabet['.'] = 0x000004
	alphabet[':'] = 0x020004
	alphabet['-'] = 0x007C00

}

func disp_string(start Dot, str string) {
	if len(alphabet) == 0 {
		init_alphabet()
	}
	color := colors.name("White").Attribute()
	setCell := func(dot *Dot, color Attribute) {
		termbox.SetCell(dot.X, dot.Y, ' ', termbox.ColorDefault, termbox.Attribute(color))
	}
	drawRune := func(count int, r rune) int {
		dot := Dot{start.X + count*6, start.Y}
		font := alphabet[r]
		//debug.Printf("rune=%c font=%x\n", r, font)
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				if font&(1<<uint(24-y*5-x)) > 0 {
					setCell(&Dot{X: dot.X + x, Y: dot.Y + y}, color)
				}
			}
		}
		return count + 1
	}
	next := 0
	for _, r := range str {
		next = drawRune(next, r)
	}
}
