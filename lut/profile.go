package lut

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Profile struct {
	UniqueCameraModel           string       `json:"UniqueCameraModel"`
	ProfileName                 string       `json:"ProfileName"`
	ProfileCopyright            string       `json:"ProfileCopyright"`
	ProfileEmbedPolicy          string       `json:"ProfileEmbedPolicy"`
	ProfileCalibrationSignature string       `json:"ProfileCalibrationSignature"`
	CalibrationIlluminant1      string       `json:"CalibrationIlluminant1"`
	CalibrationIlluminant2      string       `json:"CalibrationIlluminant2"`
	ColorMatrix1                [][3]float64 `json:"ColorMatrix1"`
	ColorMatrix2                [][3]float64 `json:"ColorMatrix2"`
	ForwardMatrix1              [][3]float64 `json:"ForwardMatrix1"`
	ForwardMatrix2              [][3]float64 `json:"ForwardMatrix2"`
	DefaultBlackRender          string       `json:"DefaultBlackRender"`
	BaselineExposureOffset      float64      `json:"BaselineExposureOffset"`
	ProfileLookTableDims        []int        `json:"ProfileLookTableDims"`
	ProfileLookTableEncoding    string       `json:"ProfileLookTableEncoding"`
	ProfileLookTable            []*LookupRow `json:"ProfileLookTable"`
	ProfileToneCurve            [][2]float64 `json:"ProfileToneCurve"`
}

func (p *Profile) Diff(
	other *Profile,
	ignoreToneCurveValues [][2]float64,
) (*Profile, error) {
	if other == nil {
		fmt.Printf("other: %#v\n", other)
		return nil, errors.New("other profile is nil")
	}

	n := &Profile{}
	err := p.DeepCopy(n)
	if err != nil {
		return nil, err
	}

	n.ProfileName = fmt.Sprintf(
		"Correct %s to %s", p.ProfileName, other.ProfileName,
	)

	if len(n.ProfileLookTable) != len(other.ProfileLookTable) {
		return nil, errors.New("ProfileLookTable length mismatch")
	}

	for i, row := range n.ProfileLookTable {
		row.HueDiv = (row.HueDiv * -1) + other.ProfileLookTable[i].HueDiv
		row.SatDiv = (row.SatDiv * -1) + other.ProfileLookTable[i].SatDiv
		row.ValDiv = (row.ValDiv * -1) + other.ProfileLookTable[i].ValDiv
		row.HueShift = (row.HueShift * -1) + other.ProfileLookTable[i].HueShift
		row.SatScale = (row.SatScale * -1) + other.ProfileLookTable[i].SatScale
		row.ValScale = (row.ValScale * -1) + other.ProfileLookTable[i].ValScale
	}

	curveA := filterToneCurve(n.ProfileToneCurve, ignoreToneCurveValues)
	curveB := filterToneCurve(other.ProfileToneCurve, ignoreToneCurveValues)

	if len(curveA) != len(curveB) {
		return nil, errors.New("ProfileToneCurve length mismatch")
	}

	for i, row := range curveA {
		for j, val := range row {
			curveA[i][j] = (val * -1) + curveB[i][j]
		}
	}

	return n, nil
}

func (p *Profile) DeepCopy(x *Profile) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &x)
}

type LookupRow struct {
	HueDiv   int     `json:"HueDiv"`
	SatDiv   int     `json:"SatDiv"`
	ValDiv   int     `json:"ValDiv"`
	HueShift float64 `json:"HueShift"`
	SatScale float64 `json:"SatScale"`
	ValScale float64 `json:"ValScale"`
}

func filterToneCurve(
	curve [][2]float64,
	filters [][2]float64,
) [][2]float64 {
	var out [][2]float64
	for _, row := range curve {
		ignore := false
		for _, val := range filters {
			if row[0] != val[0] && row[1] != val[1] {
				ignore = true
				break
			}
		}

		if !ignore {
			out = append(out, row)
		}
	}

	return out
}
