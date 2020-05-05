package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/fogleman/gg"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// Render background as context
func drawBackground(dc *gg.Context, backgroundImageFilename string) error {
	backgroundImage, err := gg.LoadImage(backgroundImageFilename)
	if err != nil {
		return err
	}
	dc.DrawImage(backgroundImage, 0, 0)
	return nil
}

// Add a semi-transparent overlay
func addOverlay(dc *gg.Context) {
	margin := 20.0
	x := margin
	y := margin
	w := float64(dc.Width()) - (2.0 * margin)
	h := float64(dc.Height()) - (2.0 * margin)
	dc.SetColor(color.RGBA{0, 0, 0, 204})
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()
}

func addLogo(dc *gg.Context) error {
	fontPath := filepath.Join("fonts", "Lato", "Lato-Regular.ttf")
	if err := dc.LoadFontFace(fontPath, 70); err != nil {
		return err
	}

	dc.SetColor(color.White)
	s := "JDHEYBURN"
	marginX := 50.0
	// marginY := -10.0
	// _, textHeight := dc.MeasureString(s)
	// x = float64(dc.Width()) - textWidth - marginX
	// y = float64(dc.Height()) - textHeight - marginY - 100
	y := 90.0
	x := marginX
	// y = marginY
	dc.DrawString(s, x, y)

	return nil
}

func addDomainText(dc *gg.Context) error {
	textColor := color.White
	fontPath := filepath.Join("fonts", "Lato", "Lato-Regular.ttf")
	if err := dc.LoadFontFace(fontPath, 60); err != nil {
		return err
	}
	r, g, b, _ := textColor.RGBA()
	mutedColor := color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(200),
	}
	dc.SetColor(mutedColor)
	marginY := 30.0
	s := "https://jdheyburn.co.uk/"
	_, textHeight := dc.MeasureString(s)
	x := 70.0
	y := float64(dc.Height()) - float64(textHeight) - marginY
	dc.DrawString(s, x, y)
	return nil
}

func addTitle(dc *gg.Context) error {

	title := "Post title here"
	textShadowColor := color.Black
	textColor := color.White
	fontPath := filepath.Join("fonts", "Open_Sans", "OpenSans-Bold.ttf")
	if err := dc.LoadFontFace(fontPath, 90); err != nil {
		return err
	}
	textRightMargin := 60.0
	textTopMargin := 90.0
	x := textRightMargin
	y := textTopMargin
	maxWidth := float64(dc.Width()) - textRightMargin - textRightMargin
	dc.SetColor(textShadowColor)
	dc.DrawStringWrapped(title, x+1, y+1, 0, 0, maxWidth, 1.5, gg.AlignLeft)
	dc.SetColor(textColor)
	dc.DrawStringWrapped(title, x, y, 0, 0, maxWidth, 1.5, gg.AlignLeft)

	return nil
}

type circle struct {
	p image.Point
	r int
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

func addMe(dc *gg.Context) error {

	// Read image from file that already exists
	existingImageFile, err := os.Open("apple-icon-180x180.png")
	if err != nil {
		panic("error")
		// Handle error
	}
	defer existingImageFile.Close()

	// Calling the generic image.Decode() will tell give us the data
	// and type of image it is as a string. We expect "png"
	_, imageType, err := image.Decode(existingImageFile)
	if err != nil {
		panic("error")
	}
	// fmt.Println(imageData)
	fmt.Println(imageType)

	// We only need this because we already read from the file
	// We have to reset the file pointer back to beginning
	existingImageFile.Seek(0, 0)

	// Alternatively, since we know it is a png already
	// we can call png.Decode() directly
	src, err := png.Decode(existingImageFile)
	if err != nil {
		// Handle error
	}

	dst := image.NewRGBA(src.Bounds())

	p := &image.Point{90, 90}

	draw.DrawMask(dst, dst.Bounds(), src, image.ZP, &circle{*p, 85}, image.ZP, draw.Over)

	outputFile, err := os.Create("test.png")
	if err != nil {
		// Handle error
	}

	// Encode takes a writer interface and an image interface
	// We pass it the File and the RGBA
	err = png.Encode(outputFile, dst)
	if err != nil {

	}

	// Don't forget to close files
	outputFile.Close()

	meImage, err := gg.LoadImage("test.png")
	if err != nil {
		return err
	}
	dc.DrawImage(meImage, 0, 0)
	// dc.SetRGB(200, 200, 0)
	// dc.DrawCircle(400, 400, 100)
	// dc.Fill()
	// dc.DrawCircle(0, 0, 20)

	return nil
}

func run() error {
	dc := gg.NewContext(1200, 628)

	backgroundImageFilename := "background.jpg"
	outputFilename := "output.png"

	if err := drawBackground(dc, backgroundImageFilename); err != nil {
		return errors.Wrap(err, "load background image")
	}

	addOverlay(dc)

	if err := addLogo(dc); err != nil {
		return errors.Wrap(err, "render logo")
	}

	if err := addDomainText(dc); err != nil {
		return errors.Wrap(err, "render domain text")
	}

	if err := addTitle(dc); err != nil {
		return errors.Wrap(err, "render title")
	}

	if err := addMe(dc); err != nil {
		return errors.Wrap(err, "render me")
	}

	// Save the image file
	if err := dc.SavePNG(outputFilename); err != nil {
		return errors.Wrap(err, "save png")
	}

	return nil

}
