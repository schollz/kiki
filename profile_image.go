package main

import (
	"bytes"
	// "fmt"
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

func handleProfileImage(c *gin.Context) {
	imgfile, err := os.Open("./static/kiki_0.png")

	img, err := png.Decode(imgfile)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

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
				// fmt.Println(r, g, b, a)
				if 22873 == r && 18761 == g && 65535 == b && 65535 == b {
					new_img.Set(x, y, color.RGBA{nr, ng, nb, 255})
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

	mimeType := "image/png"

	// c.Data(http.StatusOK, mimeType, imageBytes)
	c.Data(http.StatusOK, mimeType, buf.Bytes())
}
