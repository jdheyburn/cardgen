package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
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

func circleCropMe(imagePath string) (string, error) {

	existingImageFile, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer existingImageFile.Close()

	// Calling the generic image.Decode() will tell give us the data
	// and type of image it is as a string. We expect "png"
	_, imageData, err := image.Decode(existingImageFile)
	if err != nil {
		return "", err
	}

	// We only need this because we already read from the file
	// We have to reset the file pointer back to beginning
	existingImageFile.Seek(0, 0)

	// Decode to image.Image
	var src image.Image
	if imageData == "jpeg" {
		src, err = jpeg.Decode(existingImageFile)
	} else if imageData == "png" {
		src, err = png.Decode(existingImageFile)
	} else {
		err = errors.New(fmt.Sprintf("unsupported imageData: %s", imageData))
	}

	if err != nil {
		return "", err
	}

	src = resize.Resize(180, 0, src, resize.Lanczos3)

	dst := image.NewRGBA(src.Bounds())

	centerPoint := &image.Point{src.Bounds().Dx() / 2, src.Bounds().Dy() / 2}
	// Assuming square image
	r := src.Bounds().Dx() / 2
	circleMask := &circle{*centerPoint, r}
	draw.DrawMask(dst, dst.Bounds(), src, image.ZP, circleMask, image.ZP, draw.Over)

	// rotatedDst := imaging.Rotate(dst, 45.0, color.Transparent)

	outputFname := "circular-me.png"
	outputFile, err := os.Create(outputFname)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()

	// Encode takes a writer interface and an image interface
	// We pass it the File and the RGBA
	err = png.Encode(outputFile, dst)
	if err != nil {
		return "", err
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

	// gg Draw circle then
	fmt.Println(meImage.Bounds().Dx())

	// TODO work out numbers properly (Don't hack it)
	dc.DrawCircle(1086.5, 115, float64(meImage.Bounds().Dx()/2))
	// dc.SetRGB(200, 200, 0)

	dc.SetHexColor("#ffffff")
	dc.FillPreserve()
	dc.SetLineWidth(10)
	dc.Stroke()
	// dc.Rotate(gg.Radians(10))

	// TODO work out numbers properly (Don't hack it)
	// dc.DrawImageAnchored(meImage, 1090, -75, 0.5, 0.5)
	dc.DrawImage(meImage, 1020, 50)

	// iw, ih := meImage.Bounds().Dx(), meImage.Bounds().Dy()
	// dc.SetHexColor("#0000ff")
	// dc.SetLineWidth(2)
	// dc.Rotate(gg.Radians(10))
	// dc.Draw Rectangle(100, 0, float64(iw), float64(ih)/2+20.0)
	// dc.StrokePreserve()
	// dc.Clip()
	// dc.DrawImageAnchored(meImage, 100, 0, 0.0, 0.0)

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
