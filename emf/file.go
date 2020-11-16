package emf

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/lokks307/emftoimg/fontname"
)

type EmfFile struct {
	Header  *HeaderRecord
	Records []Recorder
	Eof     *EofRecord
}

func ReadFile(data []byte) (*EmfFile, error) {
	reader := bytes.NewReader(data)
	file := &EmfFile{}

	for reader.Len() > 0 {
		rec, err := readRecord(reader)

		if err != nil {
			return nil, err
		}

		switch rec := rec.(type) {
		case *HeaderRecord:
			file.Header = rec
		case *EofRecord:
			file.Eof = rec
		default:
			file.Records = append(file.Records, rec)
		}
	}

	return file, nil
}

type context struct {
	draw2dimg.GraphicContext
	img     draw.Image
	objects map[uint32]interface{}

	w, h int

	wo, vo *PointL
	we, ve *SizeL
	mm     uint32
}

func (f *EmfFile) initContext(w, h int) *context {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	gc := draw2dimg.NewGraphicContext(img)

	draw2d.SetFontFolder("C:\\Windows\\Fonts")
	draw2d.SetFontNamer(func(f draw2d.FontData) string {
		fmt.Println("font in=", f.Name)
		res, err := fontname.Find(f.Name)
		fmt.Println("font out=", res)
		if err != nil || res == "." {
			return "malgun.ttf"
		}
		return res
	})

	return &context{
		GraphicContext: *gc,
		img:            img,
		w:              w,
		h:              h,
		mm:             MM_TEXT,
		objects:        make(map[uint32]interface{}),
	}
}

func (ctx context) applyTransformation() {
	if ctx.we == nil || ctx.ve == nil {
		return
	}

	switch ctx.mm {
	case MM_ISOTROPIC, MM_ANISOTROPIC:
		sx := float64(ctx.ve.Cx) / float64(ctx.we.Cx)
		sy := float64(ctx.ve.Cy) / float64(ctx.we.Cy)
		ctx.Scale(sx, sy)
	default:
		sx := float64(ctx.w) / float64(ctx.we.Cx)
		sy := float64(ctx.h) / float64(ctx.we.Cy)
		ctx.Scale(sx, sy)

	}
}

func (f *EmfFile) Draw() image.Image {

	bounds := f.Header.Original.Bounds

	// inclusive-inclusive bounds
	width := int(bounds.Width()) + 1
	height := int(bounds.Height()) + 1

	ctx := f.initContext(width, height)

	if bounds.Left != 0 || bounds.Top != 0 {
		ctx.Translate(-float64(bounds.Left), -float64(bounds.Top))
	}

	for _, rec := range f.Records {
		rec.Draw(ctx)
	}

	return ctx.img
}
