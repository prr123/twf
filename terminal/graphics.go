package terminal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	bold      = "1"
	faint     = "2"
	reverse   = "7"
	nobold    = "21"
	nofaint   = "22"
	noreverse = "27"
	and       = ";"
	style     = "m"
)

var escapeRegex *regexp.Regexp

func init() {
	escapeRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
}

type Color interface {
	ToAnsiFg() string
	ToAnsiBg() string
}

type color3Bit struct {
	value    int
	isBright bool
}

type color8Bit struct {
	value int
}

type color24Bit struct {
	r int
	g int
	b int
}

func (c *color3Bit) ToAnsiFg() string {
	if c.isBright {
		return fmt.Sprintf("9%d", c.value)
	} else {
		return fmt.Sprintf("3%d", c.value)
	}
}

func (c *color3Bit) ToAnsiBg() string {
	if c.isBright {
		return fmt.Sprintf("10%d", c.value)
	} else {
		return fmt.Sprintf("4%d", c.value)
	}
}

func color3BitFromString(s string) (Color, error) {
	c := color3Bit{}
	if strings.HasPrefix(s, "bright") {
		s = s[len("bright"):]
		c.isBright = true
	}
	switch s {
	case "black":
		c.value = 0
	case "red":
		c.value = 1
	case "green":
		c.value = 2
	case "yellow":
		c.value = 3
	case "blue":
		c.value = 4
	case "magenta":
		c.value = 5
	case "cyan":
		c.value = 6
	case "white":
		c.value = 7
	default:
		return nil, errors.New("Could not interpret color string.")
	}
	return &c, nil
}

func ColorFromString(s string) (Color, error) {
	return color3BitFromString(s)
}

type Graphics struct {
	FgColor Color
	BgColor Color

	Bold    bool
	Reverse bool
}

func (g *Graphics) ToAnsi(reset bool) string {
	changes := []string{}
	if g.Bold {
		changes = append(changes, bold)
	}
	if g.Reverse {
		changes = append(changes, reverse)
	}
	if g.FgColor != nil {
		changes = append(changes, g.FgColor.ToAnsiFg())
	}
	if g.BgColor != nil {
		changes = append(changes, g.BgColor.ToAnsiBg())
	}

	result := escape + strings.Join(changes, and) + style
	if reset {
		return escape + style + result
	} else {
		return result
	}
}

type Line interface {
	Append(string, *Graphics) Line
	AppendRaw(string) Line
	Length() int
	Text() string
}

type line struct {
	line            strings.Builder
	length          int
	maxLength       int
	defaultGraphics *Graphics
}

func NewLine(defaultGraphics *Graphics, maxLength int) Line {
	return &line{defaultGraphics: defaultGraphics, maxLength: maxLength}
}

func (l *line) Append(s string, graphics *Graphics) Line {
	if l.length >= l.maxLength {
		return l
	}
	if l.length+len(s) > l.maxLength {
		s = s[:l.maxLength-l.length]
	}
	l.length += len(s)
	if graphics != nil {
		l.line.WriteString(graphics.ToAnsi(false))
	}
	l.line.WriteString(s)
	l.line.WriteString(l.defaultGraphics.ToAnsi(true))
	return l
}

func (l *line) AppendRaw(s string) Line {
	matches := escapeRegex.FindAllStringIndex(s, -1)
	prevIndex := 0
	for i := 0; i < len(matches); i++ {
		piece := s[prevIndex:matches[i][0]]
		if l.length+len(piece) > l.maxLength {
			piece = piece[:l.maxLength-l.length]
		}
		l.length += len(piece)
		l.line.WriteString(piece)
		l.line.WriteString(s[matches[i][0]:matches[i][1]])
		prevIndex = matches[i][1]
	}
	piece := s[prevIndex:]
	if l.length+len(piece) > l.maxLength {
		piece = piece[:l.maxLength-l.length]
	}
	l.length += len(piece)
	l.line.WriteString(piece)
	l.line.WriteString(l.defaultGraphics.ToAnsi(true))
	return l
}

func (l *line) Length() int {
	return l.length
}

func (l *line) Text() string {
	return l.line.String()
}