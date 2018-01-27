package main

import (
	"bytes"
	"encoding/base64"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cache "github.com/robfig/go-cache"
)

var r *rand.Rand
var imageCaching *cache.Cache

func init() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r = rand.New(s1)
	imageCaching = cache.New(5*time.Minute, 10*time.Minute)
}

func randomInt(min, max int) int {
	return r.Intn(max-min) + min
}

func randomKikiFile() string {
	files := []string{"static/grey_scale/kiki_0.png",
		"static/grey_scale/kiki_1.png",
		"static/grey_scale/kiki_2.png",
		"static/grey_scale/kiki_3.png",
		"static/grey_scale/kiki_4.png",
		"static/grey_scale/kiki_5.png",
		"static/grey_scale/kiki_6.png",
		"static/grey_scale/kiki_7.png"}
	return files[r.Intn(len(files))]
}

func handleProfileImage(c *gin.Context) {
	id := c.Param("id")
	if len(id) > 1 {
		imageBytesInterface, found := imageCaching.Get(id)
		if found {
			logger.Log.Debugf("using cache for /kiki%s", id)
			imageBytes, _ := base64.StdEncoding.DecodeString(imageBytesInterface.(string))
			c.Data(http.StatusOK, "image/png", imageBytes)
			return
		}
		alg := fnv.New32a()
		alg.Write([]byte(id))
		s1 := rand.NewSource(int64(alg.Sum32()))
		r = rand.New(s1)
	} else {
		s1 := rand.NewSource(time.Now().UnixNano())
		r = rand.New(s1)
	}

	// imgfile, err := os.Open(randomKikiFile())
	imgfile, err := Asset(randomKikiFile())
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}
	imgfileBytes := bytes.NewReader(imgfile)
	img, err := png.Decode(imgfileBytes)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	colors := []Color{Red, Orange, Yellow, Green, Blue, Purple, Pink, Monochrome}
	randColor := New(colors[r.Intn(len(colors))], LIGHT)
	newR, newG, newB, newA := randColor.RGBA()
	newRint := uint8(255 * float64(newR) / 65535)
	newGint := uint8(255 * float64(newG) / 65535)
	newBint := uint8(255 * float64(newB) / 65535)
	newAint := uint8(255 * float64(newA) / 65535)

	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	new_img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			changed := false
			r, g, b, a := img.At(x, y).RGBA()
			if 0 != r && 0 != g && 0 != b && 0 != a {
				if 47031 == r && 47031 == g && 47031 == b && 65535 == a {
					new_img.Set(x, y, color.RGBA{newRint, newGint, newBint, newAint})
					changed = true
				}
			}

			if !changed {
				new_img.Set(x, y, img.At(x, y))
			}
		}
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, new_img)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	if len(id) > 1 {
		img := base64.StdEncoding.EncodeToString(buf.Bytes())
		imageCaching.Set(id, img, 5*time.Minute)
	}
	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

// THE FOLLOWING IS FROM github.com/hansrodtang/randomcolor
// LICENSED FROM CC0 1.0 Universal

// Monochrome hue ranges
var Monochrome = Color{
	HueRange:    Range{0, 0},
	LowerBounds: []Range{{0, 0}, {100, 0}},
}

// Red hue ranges
var Red = Color{
	HueRange:    Range{-26, 18},
	LowerBounds: []Range{{20, 100}, {30, 92}, {40, 89}, {50, 85}, {60, 78}, {70, 70}, {80, 60}, {90, 55}, {100, 50}},
}

// Orange hue ranges
var Orange = Color{
	HueRange:    Range{19, 46},
	LowerBounds: []Range{{20, 100}, {30, 93}, {40, 88}, {50, 86}, {60, 85}, {70, 70}, {100, 70}},
}

// Yellow hue ranges
var Yellow = Color{
	HueRange:    Range{47, 62},
	LowerBounds: []Range{{25, 100}, {40, 94}, {50, 89}, {60, 86}, {70, 84}, {80, 82}, {90, 80}, {100, 75}},
}

// Green hue ranges
var Green = Color{
	HueRange:    Range{63, 178},
	LowerBounds: []Range{{30, 100}, {40, 90}, {50, 85}, {60, 81}, {70, 74}, {80, 64}, {90, 50}, {100, 40}},
}

// Blue hue ranges
var Blue = Color{
	HueRange:    Range{179, 257},
	LowerBounds: []Range{{20, 100}, {30, 86}, {40, 80}, {50, 74}, {60, 60}, {70, 52}, {80, 44}, {90, 39}, {100, 35}},
}

// Purple hue ranges
var Purple = Color{
	HueRange:    Range{258, 282},
	LowerBounds: []Range{{20, 100}, {30, 87}, {40, 79}, {50, 70}, {60, 65}, {70, 59}, {80, 52}, {90, 45}, {100, 42}},
}

// Pink hue ranges
var Pink = Color{
	HueRange:    Range{283, 334},
	LowerBounds: []Range{{20, 100}, {30, 90}, {40, 86}, {60, 84}, {80, 80}, {90, 75}, {100, 73}},
}

// Random hue ranges
var Random = Color{
	HueRange:    Range{0, 360},
	LowerBounds: []Range{},
}

var colors = []Color{Monochrome, Red, Orange, Yellow, Green, Blue, Purple, Pink}

// ColorInfo returns the hue range that matches the supplied hue.
// If no range can be found it returns Monochrome.
func ColorInfo(hue int) Color {
	if hue >= 334 && hue <= 360 {
		hue = hue - 360
	}

	for _, color := range colors {
		if hue >= color.HueRange[0] && hue <= color.HueRange[1] {
			return color
		}
	}
	return Monochrome
}

// HSV represents a cylindrical coordinate of points in an RGB color model.
// Values are in the range 0 to 1.
type HSV struct {
	H, S, V float64
}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the HSV.
func (c HSV) RGBA() (uint32, uint32, uint32, uint32) {

	var R, G, B float64

	hI := math.Floor(c.H * 6)
	f := c.H*6 - hI
	p := c.V * (1.0 - c.S)
	q := c.V * (1.0 - f*c.S)
	t := c.V * (1.0 - (1.0-f)*c.S)

	switch hI {
	case 0:
		R, G, B = c.V, t, p
	case 1:
		R, G, B = q, c.V, p
	case 2:
		R, G, B = p, c.V, t
	case 3:
		R, G, B = p, q, c.V
	case 4:
		R, G, B = t, p, c.V
	case 5:
		R, G, B = c.V, p, q
	}

	r := uint8((R * 255) + 0.5)
	g := uint8((G * 255) + 0.5)
	b := uint8((B * 255) + 0.5)

	return uint32(r) * 0x101, uint32(g) * 0x101, uint32(b) * 0x101, 0xffff
}

// Luminosity stores the level of luminosity for the generated color.
type Luminosity int

const (
	// LIGHT is used to generate light colors
	LIGHT Luminosity = iota
	// DARK is used to generate dark colors
	DARK
	// BRIGHT is used to generator bright colors
	BRIGHT
	// RANDOM is used to generate colors of random luminosity
	RANDOM
)

// New returns a random color in the specified hue and luminosity.
func New(hue Color, lum Luminosity) color.Color {

	c := HSV{}
	c.H = setHue(hue)
	c.S = setSaturation(c, hue, lum)
	c.V = setBrightness(c, lum)

	if c.H == 0 {
		c.H = 1
	}
	if c.H == 360 {
		c.H = 359
	}

	// Rebase the h,s,v values
	c.H = c.H / 360
	c.S = c.S / 100
	c.V = c.V / 100

	return c
}

// Range represents a range between lower (Range[0]) and upper bounds (Range[1]).
type Range [2]int

// Color represents a color in a specified range
type Color struct {
	HueRange    Range
	LowerBounds []Range
}

// SaturationRange returns the minimum and maximum saturation for the color.
func (c Color) SaturationRange() Range {
	sMin := c.LowerBounds[0][0]
	sMax := c.LowerBounds[len(c.LowerBounds)-1][0]

	return Range{sMin, sMax}
}

// BrightnessRange returns the minimum and maximum brigthness for the color.
func (c Color) BrightnessRange() Range {
	bMin := c.LowerBounds[len(c.LowerBounds)-1][1]
	bMax := c.LowerBounds[0][1]

	return Range{bMin, bMax}
}

func setHue(c Color) float64 {
	hue := randWithin(c.HueRange[0], c.HueRange[1])

	if hue < 0 {
		hue = 360 + hue
	}
	return float64(hue)
}

func setSaturation(hsv HSV, hue Color, lum Luminosity) float64 {
	if hue.HueRange == Monochrome.HueRange {
		return 0
	}

	saturationRange := ColorInfo(int(hsv.H)).SaturationRange()

	var sMin = saturationRange[0]
	var sMax = saturationRange[1]

	switch lum {
	case BRIGHT:
		sMin = 55
	case DARK:
		sMin = sMax - 10
	case LIGHT:
		sMax = 55
	case RANDOM:
		return float64(randWithin(0, 100))
	}

	return float64(randWithin(sMin, sMax))
}

func setBrightness(hsv HSV, lum Luminosity) float64 {
	bMin := getMinimumBrightness(hsv)
	bMax := 100

	switch lum {
	case DARK:
		bMax = bMin + 20
	case LIGHT:
		bMin = (bMax + bMin) / 2
	case BRIGHT:
		//
	default:
		bMin = 0
		bMax = 100
	}
	return float64(randWithin(bMin, bMax))
}

func getMinimumBrightness(hsv HSV) int {
	var lowerBounds []Range
	lowerBounds = ColorInfo(int(hsv.H)).LowerBounds
	for i := 0; i < (len(lowerBounds) - 1); i++ {
		s1 := float64(lowerBounds[i][0])
		v1 := float64(lowerBounds[i][1])

		s2 := float64(lowerBounds[i+1][0])
		v2 := float64(lowerBounds[i+1][1])

		if hsv.S >= s1 && hsv.S <= s2 {
			m := (v2 - v1) / (s2 - s1)
			b := v1 - m*s1
			return int(m*hsv.S + b)
		}
	}
	return 0
}

func randWithin(first, last int) int {
	return first + r.Intn(last+1-first)
}
