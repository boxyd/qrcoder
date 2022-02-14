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
	defaultPxSize   = 1024
	defaultFontSize = 150
)

// Generate creates a new QR code.
// Returns a []byte formatted as .png; requires an url for the QR code,
// label that will be printed on the code and an index for the QR code.
//
// Label is limited to 5 characters for now as the image is tailored
// to that size.
func Generate(url, label string, index int) ([]byte, error) {
	return GenerateWithColor(url, label, index, 1, color.Black, color.White)
}

// GenerateWithMagnitude creates a new QR code,
// but smaller by a magnitude.
// Same as Generate, returns a []byte formatted as png.
//
// Magnitude is inverted, meaning the bigger the magnitude,
// the smaller the QR code will be.
func GenerateWithMagnitude(url, label string, index, magnitude int) ([]byte, error) {
	return GenerateWithColor(url, label, index, magnitude, color.Black, color.White)
}

// GenerateWithColor creates a new QR code using custom colors.
// Returns a []byte formatted as png.
// Same as GenerateWithMagnitude, extended with foreground and background color
// which has to be of type color.Color.
func GenerateWithColor(url, label string, index, magnitude int, foreground, background color.Color) ([]byte, error) {
	if len(label) > 5 {
		return nil, fmt.Errorf("Label of %v characters is too big, limited to 5 characters", len(label))
	}

	newQR, err := qrcode.New(url, qrcode.High)
	if err != nil {
		return nil, err
	}

	newQR.ForegroundColor = foreground
	newQR.BackgroundColor = background

	pngf, err := newQR.PNG(defaultPxSize / magnitude)
	if err != nil {
		return nil, err
	}

	imgf, _, err := image.Decode(bytes.NewReader(pngf))
	if err != nil {
		return nil, err
	}

	img := ImageToRGBA(imgf)
	labelLeft, err := generateLabel(img, foreground, 0, defaultFontSize, magnitude, label)
	if err != nil {
		return nil, err
	}

	indexToString := indexer(index)

	labelRight, err := generateLabel(img, foreground, 0, defaultFontSize, magnitude, indexToString)
	if err != nil {
		return nil, err
	}

	combineLabels := combineToRight(labelLeft, labelRight, magnitude)
	finalImage := combineToBottom(img, combineLabels, magnitude)

	buf := new(bytes.Buffer)

	err = png.Encode(buf, finalImage)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generateLabel(img *image.RGBA, color color.Color, x, y, magnitude int, label string) (*image.RGBA, error) {
	point := fixed.Point26_6{X: fixed.Int26_6((x / magnitude) * 64), Y: fixed.Int26_6((y / magnitude) * 64)}

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

	// TODO: Below needs to be recalculated as now it's just a magical number that works.
	magicSize := 670 / magnitude
	newImgSize := (defaultFontSize + 20) / magnitude
	textImg := image.NewRGBA(image.Rect(0, 0, magicSize, newImgSize))

	d := &font.Drawer{
		Dst: textImg,
		Src: image.NewUniform(color),
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    float64(defaultFontSize / magnitude),
			DPI:     dpi,
			Hinting: h,
		}),
		Dot: point,
	}
	d.DrawString(label)

	return textImg, nil
}

func combineToBottom(img1, img2 *image.RGBA, magnitude int) *image.RGBA {
	pxSize := defaultPxSize / magnitude
	fontSize := defaultFontSize / magnitude

	sp := image.Point{X: 0, Y: img1.Bounds().Dy()}
	rt := image.Rectangle{sp, sp.Add(img2.Bounds().Size())}
	r := image.Rectangle{image.Point{0, 0}, image.Point{pxSize, pxSize + fontSize}}
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, img1.Bounds(), img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, rt, img2, image.Point{0, 0}, draw.Src)

	return rgba
}

func combineToRight(img1, img2 *image.RGBA, magnitude int) *image.RGBA {
	pxSize := defaultPxSize / magnitude
	fontSize := defaultFontSize / magnitude

	sp := image.Point{X: img1.Bounds().Dx(), Y: 0}
	rt := image.Rectangle{sp, sp.Add(img2.Bounds().Size())}
	r := image.Rectangle{image.Point{0, 0}, image.Point{pxSize, pxSize + fontSize}}
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, img1.Bounds(), img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, rt, img2, image.Point{0, 0}, draw.Src)

	return rgba
}

// ImageToRGBA is a utilify function that allows for conversion
// from image.Image to *image.RGBA
func ImageToRGBA(im image.Image) *image.RGBA {
	dst := image.NewRGBA(im.Bounds())
	draw.Draw(dst, im.Bounds(), im, im.Bounds().Min, draw.Src)
	return dst
}

func indexer(index int) string {
	return fmt.Sprintf("%04d", index)
}
