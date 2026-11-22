package parser

import (
	"encoding/json"
	"fmt"
)

type Point struct {
	X float64
	Y float64
}

type ItemDropValue struct {
	Count int
	Rared bool
	Asset string
}

type ResourceDrop struct {
	Type  string
	ID    int
	X     int
	Z     int
	Count int
}

type HarvestMap struct {
	ID            int
	ResourceDrops []ResourceDrop
}

type UpdatedResources struct {
	HarvestMaps []HarvestMap `json:"userMysekaiHarvestMaps"`
}

type SekaiData struct {
	UpdatedResources UpdatedResources `json:"updatedResources"`
}

func (h *HarvestMap) UnmarshalJSON(data []byte) error {
	var err error
	var raw []json.RawMessage

	err = json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	if len(raw) != 3 {
		return fmt.Errorf("invalid harvest array length: expected 3, got %d", len(raw))
	}

	err = json.Unmarshal(raw[0], &h.ID)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw[2], &h.ResourceDrops)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResourceDrop) UnmarshalJSON(data []byte) error {
	var err error
	var raw []json.RawMessage

	err = json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	if len(raw) != 9 {
		return fmt.Errorf("invalid resource drop array length: expected 9, got %d", len(raw))
	}

	err = json.Unmarshal(raw[0], &r.Type)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw[1], &r.ID)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw[2], &r.X)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw[3], &r.Z)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw[7], &r.Count)
	if err != nil {
		return err
	}

	return nil
}
