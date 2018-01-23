package web

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"regexp"
	"strings"
)

// ContentType returns the type of content
func ContentType(filename string) string {
	switch {
	case strings.Contains(filename, ".css"):
		return "text/css"
	case strings.Contains(filename, ".jpg"):
		return "image/jpeg"
	case strings.Contains(filename, ".png"):
		return "image/png"
	case strings.Contains(filename, ".js"):
		return "application/javascript"
	case strings.Contains(filename, ".xml"):
		return "application/xml"
	}
	return "text/html"
}

// CaptureBase64Images takes HTML and replaces all base64 representations (except GIF) with a filename. It returns the new HTML and also a map of the new file names and their associated data (converted from base64 to binary).
func CaptureBase64Images(startingHTML string) (newHTML string, images map[string][]byte, err error) {
	images = make(map[string][]byte)
	newHTML = startingHTML
	r, err := regexp.Compile(`src="(data:image\/[^;]+;base64[^"]+)"`)
	if err != nil {
		return
	}
	for _, img := range r.FindAllString(startingHTML, -1) {
		base64data := after(img, "base64,")
		base64data = base64data[:len(base64data)-1]
		imageMimeType := between(img, "data:", ";base64")

		// https://stackoverflow.com/questions/46022262/covert-base64-string-to-jpg-golang
		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64data))
		m, _, err2 := image.Decode(reader)
		if err2 != nil {
			log.Error(err2)
			err = err2
			return
		}
		// bounds := m.Bounds()
		// fmt.Println(bounds, formatString)

		//Encode from image format to writer
		h := sha256.New()
		h.Write([]byte(base64data))
		filename := fmt.Sprintf("%x.%s", h.Sum(nil), strings.TrimLeft(imageMimeType, "image/"))
		f := bytes.NewBuffer(nil)

		switch imageMimeType {
		case "image/jpeg":
			err = jpeg.Encode(f, m, nil)
			if err != nil {
				break
			}
		case "image/png":
			err = png.Encode(f, m)
			if err != nil {
				break
			}
		default:
			continue
		}
		images[filename] = f.Bytes()
		newHTML = strings.Replace(newHTML, img, fmt.Sprintf(`class="img-fluid" src="/img/%s"`, filename), -1)

	}
	if err != nil {
		return
	}
	return
}

func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func after(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}
