package parser

import (
	"encoding/json"
	"log"
	"os"
	"sekai-xray/asset"
	"slices"
	"time"
)

type SiteMap map[int]PointMap
type PointMap map[Point]ItemDropMap
type ItemDropMap map[int]*ItemDropValue

func Start(path string, rid []int, rit []string, filter *Filter) SiteMap {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var sekai SekaiData
	err = json.Unmarshal(b, &sekai)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	assets := asset.BundledAssets()
	elapsed := time.Since(start)
	log.Printf("Fetch %d assets in %s", len(assets), elapsed)

	start = time.Now()
	result := parseAndFilter(sekai.UpdatedResources.HarvestMaps, rid, rit, filter, assets)
	elapsed = time.Since(start)
	log.Printf("Parsed and filtered %d result in %s", len(result), elapsed)

	return result
}

func parseAndFilter(sites []HarvestMap, rid []int, rit []string, filter *Filter, assets asset.BundledAssetCollection) SiteMap {
	siteMap := make(SiteMap, len(sites))
	for _, site := range sites {
		pointMap := make(PointMap, len(site.ResourceDrops))
		for _, res := range site.ResourceDrops {
			if filter.filterID(res.ID) || filter.filterType(res.Type) {
				continue
			}

			resAsset := assets[res.Type][res.ID]
			if resAsset == nil {
				log.Printf("Skipping unsupported resource %d %s", res.ID, res.Type)
				continue
			}

			if filter.filterName(resAsset.GetDisplay()) {
				continue
			}

			resPoint := Point{
				X: float64(res.X),
				Y: float64(res.Z),
			}

			dropMap, ok := pointMap[resPoint]
			if !ok {
				dropMap = make(ItemDropMap, 3)
				pointMap[resPoint] = dropMap
			}

			item, ok := dropMap[res.ID]
			if ok {
				item.Count += res.Count
			} else {
				shouldMarkRared := slices.Contains(rid, res.ID) && slices.Contains(rit, res.Type)

				if shouldMarkRared {
					log.Printf("\x1b[31m%d - %s\x1b[0m", site.ID, resAsset.GetDisplay())
				}

				dropMap[res.ID] = &ItemDropValue{
					Count: res.Count,
					Rared: shouldMarkRared,
					Asset: resAsset.GetAsset(),
				}
			}
		}

		siteMap[site.ID] = pointMap
	}

	return siteMap
}
