package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
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

// Used to generate logo - commented while using image
// func addLogo(dc *gg.Context) error {
// 	fontPath := filepath.Join("fonts", "Lato", "Lato-Regular.ttf")
// 	if err := dc.LoadFontFace(fontPath, 70); err != nil {
// 		return err
// 	}

// 	dc.SetColor(color.White)
// 	s := "JDHEYBURN"
// 	marginX := 50.0
// 	marginY := -50.0
// 	textWidth, textHeight := dc.MeasureString(s)
// 	x := float64(dc.Width()) - textWidth - marginX
// 	// y := float64(dc.Height()) - textHeight - marginY
// 	y := textHeight - marginY
// 	// y := 90.0
// 	// x := marginX
// 	// y = marginY
// 	dc.DrawString(s, x, y)

// 	return nil
// }

func addDomainText(dc *gg.Context) error {
	textColor := color.White
	fontPath := filepath.Join("fonts", "Source_Code_Pro", "SourceCodePro-Medium.ttf")
	// fontPath := filepath.Join("fonts", "Ubuntu_Mono", "UbuntuMono-Bold.ttf")
	// fontPath := filepath.Join("fonts", "Courier_Prime", "CourierPrime-Regular.ttf")
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
	marginY := 10.0
	s := "https://jdheyburn.co.uk/"
	_, textHeight := dc.MeasureString(s)
	x := 70.0
	y := float64(dc.Height()) - float64(textHeight) - marginY
	dc.DrawString(s, x, y)
	return nil
}

func validateHeight(dc *gg.Context, s string, maxWidth, lineSpacing float64) error {

	wrapped := dc.WordWrap(s, maxWidth)

	h := float64(len(wrapped)) * dc.FontHeight() * lineSpacing
	h -= (lineSpacing - 1) * dc.FontHeight()

	// Hardcoded based on experimenting - need a better way of determining this
	maxHeight := 400.0
	if h > maxHeight {
		return errors.New("title is too long - too many lines")
	}

	return nil
}

func addTitle(dc *gg.Context) error {

	// TODO add this to variable
	title := "Who Goes Blogging 6:\n3 Steps to Better Hugo RSS Feeds!"
	textShadowColor := color.Black
	textColor := color.White
	fontPath := filepath.Join("fonts", "Merriweather", "Merriweather-Regular.ttf")
	if err := dc.LoadFontFace(fontPath, 90); err != nil {
		return err
	}
	textRightMargin := 60.0
	textTopMargin := 90.0
	x := textRightMargin
	y := textTopMargin
	maxWidth := float64(dc.Width()) - textRightMargin - textRightMargin
	lineSpacing := 1.65

	if err := validateHeight(dc, title, maxWidth, lineSpacing); err != nil {
		return err
	}

	dc.SetColor(textShadowColor)
	dc.DrawStringWrapped(title, x+1, y+1, 0, 0, maxWidth, lineSpacing, gg.AlignLeft)
	dc.SetColor(textColor)
	dc.DrawStringWrapped(title, x, y, 0, 0, maxWidth, lineSpacing, gg.AlignLeft)

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

func decodeImage(f *os.File) (image.Image, error) {
	// Calling the generic image.Decode() will tell give us the data
	// and type of image it is as a string. We expect "png"
	_, t, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	// We only need this because we already read from the file
	// We have to reset the file pointer back to beginning
	if _, err = f.Seek(0, 0); err != nil {
		return nil, err
	}

	switch t {
	case "jpeg":
		return jpeg.Decode(f)
	case "png":
		return png.Decode(f)
	}
	return nil, errors.New(fmt.Sprintf("unsupported image type: %s", t))
}

func outputFile(dst *image.RGBA, outputPath string) (err error) {
	f, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrap(err, "create output file")
	}

	// Note to self - not sure how to wrap around existing err
	defer func() {
		if ferr := f.Close(); ferr != nil && err == nil {
			err = errors.Wrap(ferr, "closing output file")
		}
	}()

	// Encode takes a writer interface and an image interface
	// We pass it the File and the RGBA
	return png.Encode(f, dst)
}

func closeQuietly(v interface{}) {
	if d, ok := v.(io.Closer); ok {
		_ = d.Close()
	}
}

func circleCropMe(imagePath string) (string, error) {

	existingImageFile, err := os.Open(filepath.Clean(imagePath))
	if err != nil {
		return "", errors.Wrap(err, "opening source image file")
	}

	// TODO better way to close properly with error checking - ignoring for now since this is a read-only operation
	defer closeQuietly(existingImageFile)

	src, err := decodeImage(existingImageFile)
	if err != nil {
		return "", errors.Wrap(err, "decoding source image file")
	}

	src = resize.Resize(180, 0, src, resize.Lanczos3)
	dst := image.NewRGBA(src.Bounds())
	c := &image.Point{src.Bounds().Dx() / 2, src.Bounds().Dy() / 2}

	// Assuming square image to draw the circle around
	r := src.Bounds().Dx() / 2
	mask := &circle{*c, r}
	draw.DrawMask(dst, dst.Bounds(), src, image.Point{}, mask, image.Point{}, draw.Over)

	outputFname := "circular-me.png"
	if err := outputFile(dst, outputFname); err != nil {
		return "", errors.Wrap(err, "saving cropped image")
	}

	return outputFname, nil
}

func addMe(dc *gg.Context) error {

	// Read image from file that already exists
	croppedMe, err := circleCropMe("me.jpg")
	if err != nil {
		return err
	}

	meImage, err := gg.LoadImage(croppedMe)
	if err != nil {
		return err
	}

	marginX := 10
	x := dc.Width() - (meImage.Bounds().Dx() / 2) - marginX
	y := dc.Height() - (meImage.Bounds().Dy() / 2)

	dc.SetHexColor("#ffffff")
	dc.FillPreserve()
	dc.SetLineWidth(10)
	dc.DrawCircle(float64(x), float64(y), float64(meImage.Bounds().Dx()/2))
	dc.Stroke()

	dc.RotateAbout(gg.Radians(10), float64(x), float64(y))
	dc.DrawImageAnchored(meImage, int(x), int(y), 0.5, 0.5)

	return nil
}

func run() error {

	dc := gg.NewContext(1200, 628)

	backgroundImageFilename := "background.jpg"
	outputFilename := "output.png"

	if err := drawBackground(dc, backgroundImageFilename); err != nil {
		return errors.Wrap(err, "load background image")
	}

	// Uncomment below for blank background
	// dc.DrawRectangle(0, 0, 1200, 628)
	// dc.SetColor(color.White)
	// dc.Fill()
	addOverlay(dc)

	// if err := addLogo(dc); err != nil {
	// 	return errors.Wrap(err, "render logo")
	// }

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
