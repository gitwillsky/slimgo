package images

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/nfnt/resize"
)

// 图片裁剪
func PicResize(fileName string, width uint, height uint) error {
	// open  pic file.
	c, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	// new reader
	r := bytes.NewReader(c)
	// new decoder
	i, f, e := image.Decode(r)
	if e != nil {
		return err
	}
	m := resize.Resize(width, height, i, resize.NearestNeighbor)
	out, err := os.Create(fileName)
	defer out.Close()
	if err != nil {
		return err
	}
	// encode content to file
	switch f {
	case "jpeg":
		jpeg.Encode(out, m, nil)
	case "png":
		png.Encode(out, m)
	case "gif":
		gif.Encode(out, m, nil)
	default:
		return nil
	}
	return nil
}
