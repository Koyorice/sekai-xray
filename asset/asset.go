package asset

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func BundledAssets() BundledAssetCollection {
	jobs := map[string]func() BundledAssetMap{
		"mysekai_item":         wrapFetchBundledAssetFunc(fetchItems),
		"mysekai_fixture":      wrapFetchBundledAssetFunc(fetchFixtures),
		"mysekai_material":     wrapFetchBundledAssetFunc(fetchMaterials),
		"mysekai_music_record": wrapFetchBundledAssetFunc(fetchMusics),
	}

	result := make(BundledAssetCollection, len(jobs))

	var mu sync.Mutex
	var wg sync.WaitGroup

	for key, job := range jobs {
		wg.Go(func() {
			data := job()

			mu.Lock()
			result[key] = data
			mu.Unlock()
		})
	}

	wg.Wait()

	return result
}

func wrapFetchBundledAssetFunc[T BundledAsset](fetch func() map[int]T) func() BundledAssetMap {
	return func() BundledAssetMap {
		m := fetch()

		result := make(BundledAssetMap, len(m))

		for k, v := range m {
			result[k] = v
		}

		return result
	}
}

func fetchItems() map[int]*Item {
	items := fetchAssets[*Item]("mysekaiItems")

	for _, v := range items {
		v.Asset = fmt.Sprintf("mysekai/item_preview/item/%s/%s.png", v.Asset, v.Asset)
	}

	return items
}

func fetchFixtures() map[int]*Fixture {
	fixtures := fetchAssets[*Fixture]("mysekaiFixtures")

	for _, v := range fixtures {
		v.Asset = fmt.Sprintf("mysekai/item_preview/fixture/%s_%d/%s_%d.png", v.Asset, v.ID, v.Asset, v.ID)
	}

	return fixtures
}

func fetchMaterials() map[int]*Material {
	materials := fetchAssets[*Material]("mysekaiMaterials")

	for _, v := range materials {
		v.Asset = fmt.Sprintf("mysekai/item_preview/material/%s/%s.png", v.Asset, v.Asset)
	}

	return materials
}

func fetchMusics() map[int]*Music {
	musics := fetchAssets[*Music]("musics")
	musicRecords := fetchAssets[*MusicRecord]("mysekaiMusicRecords")

	for _, m := range musics {
		m.Asset = fmt.Sprintf("music/jacket/%s/%s.png", m.Asset, m.Asset)
	}

	assetMap := make(map[int]*Music, len(musicRecords))
	for _, record := range musicRecords {
		if record.MusicTrackType != "music" {
			continue
		}

		if record.MusicTrackType == "music" {
			assetMap[record.ID] = musics[record.ExternalID]
		}
	}

	return assetMap
}

func fetchAssets[T Asset](name string) map[int]T {
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/Team-Haruki/haruki-sekai-tc-master/refs/heads/main/master/%s.json", name))
	if err != nil {
		panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var assets []T
	err = json.Unmarshal(body, &assets)
	if err != nil {
		panic(err)
	}

	assetMap := make(map[int]T, len(assets))
	for i := range assets {
		assetMap[assets[i].GetID()] = assets[i]
	}

	return assetMap
}
