package main

import (
	"bytes"
	"image/color"
	"strings"

	"github.com/golang/freetype/truetype"

	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"math"
	"math/rand"
	"time"

	"image"
	"image/draw"
	"image/jpeg"
	"image/png"

	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/vgimg"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// getFont returns the truetype.Font for the given font name or an error.
func getFont(name string) (*truetype.Font, error) {

	data, err := ioutil.ReadFile(name)

	font, err := truetype.Parse(data)

	return font, err
}

// WaterMark for adding a watermark on the image
func WaterMark(img image.Image, markText string, fontsFile string) (image.Image, error) {
	// image's length to canvas's length
	bounds := img.Bounds()
	w := vg.Length(bounds.Max.X) * vg.Inch / vgimg.DefaultDPI
	h := vg.Length(bounds.Max.Y) * vg.Inch / vgimg.DefaultDPI
	diagonal := vg.Length(math.Sqrt(float64(w*w + h*h)))

	// create a canvas, which width and height are diagonal
	c := vgimg.New(diagonal, diagonal)

	// draw image on the center of canvas
	rect := vg.Rectangle{}
	rect.Min.X = diagonal/2 - w/2
	rect.Min.Y = diagonal/2 - h/2
	rect.Max.X = diagonal/2 + w/2
	rect.Max.Y = diagonal/2 + h/2
	c.DrawImage(rect, img)

	loadFont, _ := getFont(fontsFile)
	vg.AddFont("SourceHanSansK", loadFont)

	fontStyle, _ := vg.MakeFont("SourceHanSansK", vg.Inch*0.15)

	// repeat the markText
	markTextWidth := fontStyle.Width(markText)
	unitText := markText
	for markTextWidth <= diagonal {
		markText += "   " + unitText
		markTextWidth = fontStyle.Width(markText)
	}

	// set the color of markText
	//c.SetColor(color.RGBA{0, 0, 0, 122})
	c.SetColor(color.RGBA{255, 255, 255, 80})
	//c.SetColor(color.White)

	// set a random angle between 0 and π/2
	//θ := math.Pi * rand.Float64() / 4
	//c.Rotate(θ)

	// set the lineHeight and add the markText
	lineHeight := fontStyle.Extents().Height * 1
	for offset := -2 * diagonal; offset < 2*diagonal; offset += lineHeight {
		c.FillString(fontStyle, vg.Point{X: 0, Y: offset}, markText)
	}

	// canvas writeto jpeg
	// canvas.img is private
	// so use a buffer to transfer
	jc := vgimg.PngCanvas{Canvas: c}
	buff := new(bytes.Buffer)
	jc.WriteTo(buff)
	img, _, err := image.Decode(buff)
	if err != nil {
		return nil, err
	}

	// get the center point of the image
	ctp := int(diagonal * vgimg.DefaultDPI / vg.Inch / 2)

	// cutout the marked image
	size := bounds.Size()
	bounds = image.Rect(ctp-size.X/2, ctp-size.Y/2, ctp+size.X/2, ctp+size.Y/2)
	rv := image.NewRGBA(bounds)
	draw.Draw(rv, bounds, img, bounds.Min, draw.Src)
	return rv, nil
}

// MarkingPicture for marking picture with text
func MarkingPicture(filepath, text string, fonts string) (image.Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	img, err = WaterMark(img, text, fonts)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func writeTo(img image.Image, ext string) (rv *bytes.Buffer, err error) {
	ext = strings.ToLower(ext)
	rv = new(bytes.Buffer)
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(rv, img, &jpeg.Options{Quality: 100})
	case ".png":
		err = png.Encode(rv, img)
	}
	return rv, err
}

func addWatermark(target string, output string, text string, fonts string) {
	img, err := MarkingPicture(target, text, fonts)
	if err != nil {
		// fmt.Printf("MarkingPicture failed, target: %v, err: %v", target, err)
		return
	}

	ext := path.Ext(target)
	f, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	buff, err := writeTo(img, ext)
	if err != nil {
		panic(err)
	}
	if _, err = buff.WriteTo(f); err != nil {
		panic(err)
	}
}

func main() {

	var target string
	var output string
	var fonts string
	var text string

	flag.StringVar(&target, "target", "target.png", "target image file path. only supports .png .jpg .jpeg")
	flag.StringVar(&output, "output", "", "output image file path. only supports .png .jpg .jpeg")
	flag.StringVar(&fonts, "fonts", "fonts/SourceHanSansK-Normal.ttf", "fonts file.")
	flag.StringVar(&text, "text", "仅限域名备案使用", "watermark text")
	flag.Parse()

	fmt.Printf("begin add watermark... target: %v, target: %v, fonts: %v, text: %v\n", target, target, fonts, text)

	if stat, err := os.Stat(target); err == nil && stat.IsDir() {
		// target is a directory
		files, _ := ioutil.ReadDir(target)
		for _, fn := range files {
			filepath := path.Join(target, fn.Name())
			output = strings.Split(path.Base(filepath), ".")[0] + "_marked" + path.Ext(filepath)
			addWatermark(filepath, output, text, fonts)
		}
	} else {
		// target is a single image file
		if output == "" {
			output = strings.Split(path.Base(target), ".")[0] + "_marked" + path.Ext(target)
		}
		addWatermark(target, output, text, fonts)
	}

	fmt.Printf("add watermark finish.\n")

}
