package main

import (
	"flag"
	"image"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
)

var (
	jpegQuality int
	aspect      string
	smart       bool
)

func parseWidthHeight(aspect string) (w int, h int) {
	xidx := strings.IndexByte(aspect, 'x')
	if xidx < 0 {
		log.Fatalf("invalid aspect syntax: %s", aspect)
	}
	var err error
	w, err = strconv.Atoi(aspect[:xidx])
	if err != nil {
		log.Fatal(err)
	}
	h, err = strconv.Atoi(aspect[xidx+1:])
	if err != nil {
		log.Fatal(err)
	}
	return w, h
}

func parseAspect(aspect string) float64 {
	w, h := parseWidthHeight(aspect)
	return float64(w) / float64(h)
}

func cropWidth(img image.Image, w int) image.Rectangle {
	dw2 := (img.Bounds().Dx() - w) / 2
	res := image.Rectangle{
		Min: image.Point{
			X: img.Bounds().Min.X + dw2,
			Y: img.Bounds().Min.Y,
		},
		Max: image.Point{
			X: img.Bounds().Max.X - dw2,
			Y: img.Bounds().Max.Y,
		},
	}
	return res
}

func cropHeight(img image.Image, h int) image.Rectangle {
	dh2 := (img.Bounds().Dy() - h) / 2
	res := image.Rectangle{
		Min: image.Point{
			X: img.Bounds().Min.X,
			Y: img.Bounds().Min.Y + dh2,
		},
		Max: image.Point{
			X: img.Bounds().Max.X,
			Y: img.Bounds().Max.Y - dh2,
		},
	}
	return res
}

func cropSmart(img image.Image, w, h int) image.Rectangle {
	a := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	res, err := a.FindBestCrop(img, w, h)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func adjustAspect(img image.Image) image.Image {
	imgAspect := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
	outAspect := parseAspect(aspect)
	if math.Abs(imgAspect-outAspect) > 0.001 {
		if imgAspect > outAspect {
			outWidth := int(outAspect * float64(img.Bounds().Dy()))
			var rect image.Rectangle
			if smart {
				rect = cropSmart(img, outWidth, img.Bounds().Dy())
			} else {
				rect = cropWidth(img, outWidth)
			}
			img = transform.Crop(img, rect)
		} else if imgAspect < outAspect {
			outHeight := int(float64(img.Bounds().Dx()) / outAspect)
			var rect image.Rectangle
			if smart {
				rect = cropSmart(img, img.Bounds().Dx(), outHeight)
			} else {
				rect = cropHeight(img, outHeight)
			}
			img = transform.Crop(img, rect)
		}
	}
	return img
}

func main() {
	flag.IntVar(&jpegQuality, "q", 90, "Set JPEG output quality (1–100)")
	flag.StringVar(&aspect, "a", "", "Set output image aspect ratio")
	flag.BoolVar(&smart, "smart", false, "Use smartcrop to adjust aspect ratio")
	flag.Parse()
	img, err := imgio.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	if aspect != "" {
		img = adjustAspect(img)
	}
	err = imgio.Save(flag.Arg(1), img, imgio.JPEGEncoder(jpegQuality))
	if err != nil {
		log.Fatal(err)
	}
}
