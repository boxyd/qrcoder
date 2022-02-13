package qrcoder

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"path"

	"github.com/goki/freetype/truetype"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	pxSize   = 1024
	fontSize = 150.0
)

func Generate(url, label string, index int) ([]byte, error) {
	return GenerateWithColor(url, label, index, color.Black, color.White)
}

func GenerateWithColor(url, label string, index int, foreground, background color.Color) ([]byte, error) {
	newQR, err := qrcode.New(url, qrcode.High)
	if err != nil {
		return nil, err
	}

	newQR.ForegroundColor = foreground
	newQR.BackgroundColor = background

	pngf, err := newQR.PNG(pxSize)
	if err != nil {
		return nil, err
	}

	imgf, _, err := image.Decode(bytes.NewReader(pngf))
	if err != nil {
		return nil, err
	}

	img := ImageToRGBA(imgf)
	labelLeft, err := generateLabel(img, foreground, 0, int(fontSize), label)
	if err != nil {
		return nil, err
	}

	indexToString := indexer(index)

	labelRight, err := generateLabel(img, foreground, 0, int(fontSize), indexToString)
	if err != nil {
		return nil, err
	}

	combineLabels := combineToRight(labelLeft, labelRight)
	finalImage := combineToBottom(img, combineLabels)

	buf := new(bytes.Buffer)

	err = png.Encode(buf, finalImage)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generateLabel(img *image.RGBA, color color.Color, x, y int, label string) (*image.RGBA, error) {
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}

	// Read the font data.
	pwd, err := os.Getwd()
	if err != nil {
		return &image.RGBA{}, err
	}
	path := path.Join(pwd, "VCR_OSD_MONO_1.001.ttf")

	fontBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return &image.RGBA{}, err
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return &image.RGBA{}, err
	}

	h := font.HintingNone
	dpi := 72.0

	// NOTE: Below needs to be recalculated as now it's just a magical number that works.
	magicSize := 670
	textImg := image.NewRGBA(image.Rect(0, 0, magicSize, int(fontSize)+20))

	d := &font.Drawer{
		Dst: textImg,
		Src: image.NewUniform(color),
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    fontSize,
			DPI:     dpi,
			Hinting: h,
		}),
		Dot: point,
	}
	d.DrawString(label)

	return textImg, nil
}

func combineToBottom(img1, img2 *image.RGBA) *image.RGBA {
	sp := image.Point{X: 0, Y: img1.Bounds().Dy()}
	rt := image.Rectangle{sp, sp.Add(img2.Bounds().Size())}
	r := image.Rectangle{image.Point{0, 0}, image.Point{pxSize, pxSize + int(fontSize)}}
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, img1.Bounds(), img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, rt, img2, image.Point{0, 0}, draw.Src)

	return rgba
}

func combineToRight(img1, img2 *image.RGBA) *image.RGBA {
	sp := image.Point{X: img1.Bounds().Dx(), Y: 0}
	rt := image.Rectangle{sp, sp.Add(img2.Bounds().Size())}
	r := image.Rectangle{image.Point{0, 0}, image.Point{pxSize, pxSize + int(fontSize)}}
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, img1.Bounds(), img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, rt, img2, image.Point{0, 0}, draw.Src)

	return rgba
}

func ImageToRGBA(im image.Image) *image.RGBA {
	dst := image.NewRGBA(im.Bounds())
	draw.Draw(dst, im.Bounds(), im, im.Bounds().Min, draw.Src)
	return dst
}

func indexer(index int) string {
	return fmt.Sprintf("%04d", index)
}
