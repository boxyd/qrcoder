package qrcoder

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	// NOTE: generating almost all QR codes
	for i := 1; i <= 20; i++ {
		url := fmt.Sprintf("http://35.176.98.167:8080/box/%04d.html", i)
		//	testImage := getImage("./assets/google.png")

		actual, err := Generate(url, "boxyd", i)

		path := fmt.Sprintf("./assets/box_%04d.png", i)
		img, err := os.Create(path)
		if err != nil {
			panic(err)
		}

		decoded, _, _ := image.Decode(bytes.NewReader(actual))

		err = png.Encode(img, decoded)
		if err != nil {
			panic(err)
		}

		log.Printf("Generated box %04d\n---------------\n", i)

		//	assert.Equal(t, testImage, actual)
		assert.NotEmpty(t, actual)
		assert.NoError(t, err)
	}
}

func TestGenerateWithColor(t *testing.T) {
	url := "http://35.176.98.167:8080/boxes/0001.html"
	orange := color.RGBA{255, 187, 1, 0xff}

	grey := color.RGBA{54, 54, 54, 0xff}

	actual, err := GenerateWithColor(url, "boxyd", 1, orange, grey)

	/*
		img, err := os.Create("./test_image_0001.png")
		if err != nil {
			panic(err)
		}

		decoded, _, _ := image.Decode(bytes.NewReader(actual))

		err = png.Encode(img, decoded)
		if err != nil {
			panic(err)
		}
	*/

	assert.NotEmpty(t, actual)
	assert.NoError(t, err)
}

func getImage(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestIndexer(t *testing.T) {
	t.Run("Index = 1", func(t *testing.T) {
		expected := "0001"
		index := 1

		actual := indexer(index)
		assert.Equal(t, expected, actual)
	})

	t.Run("Index = 10", func(t *testing.T) {
		expected := "0010"
		index := 10

		actual := indexer(index)
		assert.Equal(t, expected, actual)
	})

	t.Run("Index = 100", func(t *testing.T) {
		expected := "0100"
		index := 100

		actual := indexer(index)
		assert.Equal(t, expected, actual)
	})

	t.Run("Index = 4567", func(t *testing.T) {
		expected := "4567"
		index := 4567

		actual := indexer(index)
		assert.Equal(t, expected, actual)
	})
}
