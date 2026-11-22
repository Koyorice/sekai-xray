package asset

type BundledAssetMap map[int]BundledAsset
type BundledAssetCollection map[string]BundledAssetMap

type Asset interface {
	GetID() int
}

type BundledAsset interface {
	Asset
	GetAsset() string
	GetDisplay() string
}

type MusicRecord struct {
	ID             int    `json:"id"`
	ExternalID     int    `json:"externalId"`
	MusicTrackType string `json:"mysekaiMusicTrackType"`
}

func (a *MusicRecord) GetID() int { return a.ID }

type Music struct {
	ID      int    `json:"id"`
	Asset   string `json:"assetbundleName"`
	Display string `json:"title"`
}

func (a *Music) GetID() int         { return a.ID }
func (a *Music) GetAsset() string   { return a.Asset }
func (a *Music) GetDisplay() string { return a.Display }

type Item struct {
	ID      int    `json:"id"`
	Asset   string `json:"iconAssetbundleName"`
	Display string `json:"name"`
}

func (a *Item) GetID() int         { return a.ID }
func (a *Item) GetAsset() string   { return a.Asset }
func (a *Item) GetDisplay() string { return a.Display }

type Fixture struct {
	ID      int    `json:"id"`
	Asset   string `json:"assetbundleName"`
	Display string `json:"name"`
}

func (a *Fixture) GetID() int         { return a.ID }
func (a *Fixture) GetAsset() string   { return a.Asset }
func (a *Fixture) GetDisplay() string { return a.Display }

type Material struct {
	ID      int    `json:"id"`
	Asset   string `json:"iconAssetbundleName"`
	Display string `json:"name"`
}

func (a *Material) GetID() int         { return a.ID }
func (a *Material) GetAsset() string   { return a.Asset }
func (a *Material) GetDisplay() string { return a.Display }
