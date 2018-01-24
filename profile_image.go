package main

import (
	"bytes"
	"hash/fnv"
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

var r *rand.Rand

func init() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r = rand.New(s1)
}

func randomInt(min, max int) int {
	return r.Intn(max-min) + min
}

func randomKikiFile() string {
	files := []string{"./misc/grey_scale/kiki_0.png",
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
	id := c.Param("id")
	if len(id) > 1 {
		alg := fnv.New32a()
		alg.Write([]byte(id))
		s1 := rand.NewSource(int64(alg.Sum32()))
		r = rand.New(s1)
	} else {
		s1 := rand.NewSource(time.Now().UnixNano())
		r = rand.New(s1)
	}

	// imgfile, err := os.Open("./static/kiki_0.png")
	imgfile, err := os.Open(randomKikiFile())
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	img, err := png.Decode(imgfile)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	// var colors = make(map[string]int)

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

	// for v := range colors {
	// 	fmt.Println(v, colors[v])
	// }

	buf := new(bytes.Buffer)
	err = png.Encode(buf, new_img)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	mimeType := "image/png"

	c.Data(http.StatusOK, mimeType, buf.Bytes())
}
