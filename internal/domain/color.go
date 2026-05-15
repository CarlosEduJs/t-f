package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type OKLCHColor struct {
	L, C, H, Alpha float64
}

func ParseOKLCH(s string) (*OKLCHColor, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "oklch(") || !strings.HasSuffix(s, ")") {
		return nil, false
	}
	inner := strings.TrimSpace(s[6 : len(s)-1])

	var mainPart, alphaPart string
	if idx := strings.IndexByte(inner, '/'); idx != -1 {
		mainPart = strings.TrimSpace(inner[:idx])
		alphaPart = strings.TrimSpace(inner[idx+1:])
	} else {
		mainPart = strings.TrimSpace(inner)
	}

	fields := strings.Fields(mainPart)
	if len(fields) < 3 {
		return nil, false
	}

	L, err1 := strconv.ParseFloat(fields[0], 64)
	C, err2 := strconv.ParseFloat(fields[1], 64)
	H, err3 := strconv.ParseFloat(fields[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, false
	}

	alpha := 1.0
	if alphaPart != "" {
		a, err := strconv.ParseFloat(alphaPart, 64)
		if err == nil {
			alpha = a
		}
	}

	return &OKLCHColor{L: L, C: C, H: H, Alpha: alpha}, true
}

func (c *OKLCHColor) ToHEX() string {
	hRad := c.H * math.Pi / 180
	a := c.C * math.Cos(hRad)
	b := c.C * math.Sin(hRad)

	l_ := c.L + 0.3963377774*a + 0.2158037573*b
	m_ := c.L - 0.1055613458*a - 0.0638541728*b
	s_ := c.L - 0.0894841775*a - 1.2914855480*b

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	rLin := 4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	gLin := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	bLin := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s

	r := linearToSRGB(rLin)
	g := linearToSRGB(gLin)
	bv := linearToSRGB(bLin)

	r = clampU8(r)
	g = clampU8(g)
	bv = clampU8(bv)

	ri := int(math.Round(r * 255))
	gi := int(math.Round(g * 255))
	bi := int(math.Round(bv * 255))

	if c.Alpha < 1 {
		ai := int(math.Round(c.Alpha * 255))
		return fmt.Sprintf("#%02x%02x%02x%02x", ri, gi, bi, ai)
	}
	return fmt.Sprintf("#%02x%02x%02x", ri, gi, bi)
}

func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return c * 12.92
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

func clampU8(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func IsColorValue(value string) bool {
	v := strings.TrimSpace(value)
	if strings.HasPrefix(v, "#") {
		return true
	}
	if strings.HasPrefix(v, "oklch(") || strings.HasPrefix(v, "rgb(") ||
		strings.HasPrefix(v, "rgba(") || strings.HasPrefix(v, "hsl(") ||
		strings.HasPrefix(v, "hsla(") {
		return true
	}
	return false
}

func IsDimensionValue(value string) bool {
	v := strings.TrimSpace(value)
	if strings.HasPrefix(v, "calc(") {
		return true
	}
	if strings.Contains(v, "px") || strings.Contains(v, "rem") ||
		strings.Contains(v, "em") {
		return true
	}
	return false
}

func ConvertColorToHEX(value string) string {
	c, ok := ParseOKLCH(value)
	if !ok {
		if strings.HasPrefix(value, "#") {
			return value
		}
		return value
	}
	return c.ToHEX()
}
