package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"net/http"
	"sekai-xray/parser"
	"sort"
	"sync"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type ItemImage struct {
	Point  parser.Point
	Image  image.Image
	Rared  bool
	Count  int
	Width  int
	Height int
}

type RenderItem struct {
	Point image.Point
	Image ItemImage
}

type RenderTask struct {
	Rect  image.Rectangle
	Image image.Image
	Count int
}

func OverlayImages(bg image.Image, images []ItemImage, scale float64, offset parser.Point, reverse bool) *image.NRGBA {
	bounds := bg.Bounds()
	result := image.NewNRGBA(bounds)
	border := make([]int, bounds.Dy())
	centerX := float64(bounds.Dx()) / 2.0
	centerY := float64(bounds.Dy()) / 2.0

	draw.Draw(result, bounds, bg, image.Point{}, draw.Src)

	renders := make([]RenderItem, len(images))
	for i, img := range images {
		point := img.Point

		sx := point.X * scale
		sy := point.Y * scale

		if reverse {
			sx, sy = -sy, sx
		}

		cx := centerX + sx
		cy := centerY - sy

		ox := cx + offset.X
		oy := cy + offset.Y

		w, h := img.Width, img.Height

		x := int(ox) - w/2
		y := int(oy) - h/2

		renders[i] = RenderItem{
			Point: image.Pt(x, y),
			Image: img,
		}
	}

	sort.Slice(renders, func(i, j int) bool {
		pi, pj := renders[i].Point, renders[j].Point

		if pi.X != pj.X {
			return pi.X < pj.X
		}

		if pi.Y != pj.Y {
			return pi.Y < pj.Y
		}

		ii, ij := renders[i].Image, renders[j].Image

		if ii.Count != ij.Count {
			return ii.Count < ij.Count
		}

		return ii.Rared && !ij.Rared
	})

	maskLP := image.NewAlpha(bounds)
	maskAQ := image.NewAlpha(bounds)

	tasks := make([]RenderTask, len(renders))
	for i, render := range renders {
		img := render.Image
		point := render.Point

		x, y := point.X, point.Y
		w, h := img.Width, img.Height

		if x < border[y] {
			x = border[y]
		}
		border[y] = x + w

		rect := image.Rect(x, y, x+w, y+h)

		if img.Rared {
			draw.Draw(maskLP, rect, image.NewUniform(color.Alpha{A: 255}), image.Point{}, draw.Over)
		} else {
			draw.Draw(maskAQ, rect, image.NewUniform(color.Alpha{A: 255}), image.Point{}, draw.Over)
		}

		tasks[i] = RenderTask{
			Rect:  rect,
			Image: img.Image,
			Count: img.Count,
		}
	}

	bgLP := image.NewUniform(color.NRGBA{R: 255, G: 85, B: 255, A: 128})
	bgAQ := image.NewUniform(color.NRGBA{R: 85, G: 255, B: 255, A: 128})
	draw.DrawMask(result, bounds, bgLP, image.Point{}, maskLP, image.Point{}, draw.Over)
	draw.DrawMask(result, bounds, bgAQ, image.Point{}, maskAQ, image.Point{}, draw.Over)

	for _, task := range tasks {
		draw.Draw(result, task.Rect, task.Image, image.Point{}, draw.Over)

		point := fixed.P(task.Rect.Min.X+2, task.Rect.Min.Y+10)

		fontDrawer := &font.Drawer{
			Dot:  point,
			Dst:  result,
			Src:  image.Black,
			Face: basicfont.Face7x13,
		}

		fontDrawer.DrawString(fmt.Sprintf("%d", task.Count))
	}

	return result
}

func CombineImages(points parser.PointMap, size int) []ItemImage {
	var mu sync.Mutex
	var wg sync.WaitGroup

	images := make([]ItemImage, 0, len(points)*2)
	for point, items := range points {
		for _, item := range items {
			wg.Go(func() {
				origin := FetchImage(
					fmt.Sprintf(
						"https://storage.exmeaning.com/sekai-tw-assets/%s",
						item.Asset,
					),
				)

				resized := imaging.Resize(
					origin,
					size, size,
					imaging.Lanczos,
				)

				img := ItemImage{
					Point:  point,
					Image:  resized,
					Rared:  item.Rared,
					Count:  item.Count,
					Width:  resized.Bounds().Dx(),
					Height: resized.Bounds().Dy(),
				}

				mu.Lock()
				images = append(images, img)
				mu.Unlock()
			})
		}
	}

	wg.Wait()

	return images
}

func FetchImage(url string) image.Image {
	return cacheOrFetch(url, func() image.Image {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp.Body)

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			panic(err)
		}

		return img
	})
}
