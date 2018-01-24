package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func randomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func randomKikiFile() string {
	files := []string{"./misc/grey_scale/kiki-0.png",
		"./misc/grey_scale/kiki_1.png",
		"./misc/grey_scale/kiki_2.png",
		"./misc/grey_scale/kiki_3.png",
		"./misc/grey_scale/kiki_4.png",
		"./misc/grey_scale/kiki_5.png",
		"./misc/grey_scale/kiki_6.png",
		"./misc/grey_scale/kiki_7.png"}
	idx := randomInt(0, 7)
	return files[idx]
}

func handleProfileImage(c *gin.Context) {
	// imgfile, err := os.Open("./static/kiki_0.png")
	imgfile, err := os.Open(randomKikiFile())

	img, err := png.Decode(imgfile)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	var colors = make(map[string]int)

	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	new_img := image.NewRGBA(image.Rect(0, 0, w, h))
	nr := uint8(randomInt(0, 255))
	ng := uint8(randomInt(0, 255))
	nb := uint8(randomInt(0, 255))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			changed := false
			r, g, b, a := img.At(x, y).RGBA()
			if 0 != r && 0 != g && 0 != b && 0 != a {

				_c := fmt.Sprintf("r=%v g=%v b=%v a=%v", r, g, b, a)
				if _, ok := colors[_c]; !ok {
					colors[_c] = 0
				}
				colors[_c]++

				if 47031 == r && 47031 == g && 47031 == b && 65535 == a {
					new_img.Set(x, y, color.RGBA{nr, ng, nb, 255})
					changed = true
				}
			}

			if !changed {
				new_img.Set(x, y, img.At(x, y))
			}
		}
	}

	for v := range colors {
		fmt.Println(v, colors[v])
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, new_img)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	mimeType := "image/png"

	c.Data(http.StatusOK, mimeType, buf.Bytes())
}
