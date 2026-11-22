package main

import (
	"context"
	"fmt"
	img "image"
	"image/png"
	"log"
	"os"
	"sekai-xray/image"
	"sekai-xray/parser"
	"sekai-xray/proxy"
	"sync"
	"time"

	"github.com/urfave/cli/v3"
)

type Map struct {
	Scale   float64
	Image   string
	Offset  parser.Point
	Reverse bool
}

var sekaiMap = map[int]Map{
	5: {
		Scale:   33.333,
		Image:   "grasslands",
		Offset:  parser.Point{X: 0, Y: -40},
		Reverse: true,
	},
	6: {
		Scale:   20.513,
		Image:   "beach",
		Offset:  parser.Point{X: 0, Y: 80},
		Reverse: false,
	},
	7: {
		Scale:   24.806,
		Image:   "flowergarden",
		Offset:  parser.Point{X: -62.015, Y: 20.672},
		Reverse: true,
	},
	8: {
		Scale:   21.333,
		Image:   "memorialplace",
		Offset:  parser.Point{X: 0, Y: -106.667},
		Reverse: false,
	},
}

func main() {
	cmd := &cli.Command{
		Usage: "sekai proxy and parsing tool",
		Commands: []*cli.Command{
			{
				Name:  "proxy",
				Usage: "Start proxy sekai server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "listen",
						Usage:    "proxy listen address",
						Aliases:  []string{"l"},
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					proxy.Start(cmd.String("listen"))
					return nil
				},
			},
			{
				Name:  "parse",
				Usage: "Start parse sekai json file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Usage:    "parse json file path",
						Aliases:  []string{"p"},
						Required: true,
					},
					&cli.IntSliceFlag{
						Name:     "fid",
						Usage:    "filter by item ID(s)",
						Required: false,
					},
					&cli.StringSliceFlag{
						Name:     "fit",
						Usage:    "filter by item type(s)",
						Required: false,
					},
					&cli.StringSliceFlag{
						Name:     "fin",
						Usage:    "filter by item name(s)",
						Required: false,
					},
					&cli.IntSliceFlag{
						Name:     "rid",
						Usage:    "mark rare by item ID(s)",
						Value:    []int{5, 12, 20, 24},
						Required: false,
					},
					&cli.StringSliceFlag{
						Name:     "rit",
						Usage:    "mark rare by item type(s)",
						Value:    []string{"mysekai_material"},
						Required: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					bgChan := make(map[int]chan img.Image, len(sekaiMap))
					for id, site := range sekaiMap {
						ch := make(chan img.Image, 1)
						bgChan[id] = ch

						go func(ch chan img.Image, bg string) {
							ch <- image.FetchImage(
								fmt.Sprintf(
									"https://raw.githubusercontent.com/Koyorice/sekai-xray/refs/heads/main/site/%s.png",
									bg,
								),
							)
						}(ch, site.Image)
					}

					path := cmd.String("path")
					rid := cmd.IntSlice("rid")
					rit := cmd.StringSlice("rit")

					sites := parser.Start(
						path, rid, rit,
						&parser.Filter{
							ID:   cmd.IntSlice("fid"),
							Type: cmd.StringSlice("fit"),
							Name: cmd.StringSlice("fin"),
						},
					)

					start := time.Now()
					var wg sync.WaitGroup
					for site, points := range sites {
						wg.Go(func() {
							sekai := sekaiMap[site]

							start := time.Now()
							combined := image.CombineImages(points, 32)
							elapsed := time.Since(start)
							log.Printf("Fetch site %d images in %s", site, elapsed)

							start = time.Now()
							bgImage := <-bgChan[site]
							elapsed = time.Since(start)
							log.Printf("Wait site %d bg image took %s", site, elapsed)

							start = time.Now()
							done := image.OverlayImages(bgImage, combined, sekai.Scale, sekai.Offset, sekai.Reverse)
							elapsed = time.Since(start)
							log.Printf("Draw site %d images in %s", site, elapsed)

							file, err := os.Create(fmt.Sprintf("%d.png", site))
							if err != nil {
								panic(err)
							}

							err = png.Encode(file, done)
							if err != nil {
								panic(err)
							}

							err = file.Close()
							if err != nil {
								panic(err)
							}
						})
					}

					wg.Wait()
					elapsed := time.Since(start)
					log.Printf("Save %d sites image in %s", len(sites), elapsed)

					return nil
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		panic(err)
	}
}
