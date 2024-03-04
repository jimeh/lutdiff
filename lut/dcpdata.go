package lut

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type DCPData struct {
	XMLName                      xml.Name           `xml:"dcpData" json:"dcpdata,omitempty"`
	ProfileName                  string             `xml:"ProfileName" json:"profileName,omitempty"`
	CalibrationIlluminant1       string             `xml:"CalibrationIlluminant1" json:"calibrationIlluminant1,omitempty"`
	CalibrationIlluminant2       string             `xml:"CalibrationIlluminant2" json:"calibrationIlluminant2,omitempty"`
	ColorMatrix1                 *DCPColorMatrix    `xml:"ColorMatrix1" json:"colorMatrix1,omitempty"`
	ColorMatrix2                 *DCPColorMatrix    `xml:"ColorMatrix2" json:"colorMatrix2,omitempty"`
	ForwardMatrix1               *DCPColorMatrix    `xml:"ForwardMatrix1" json:"forwardMatrix1,omitempty"`
	ForwardMatrix2               *DCPColorMatrix    `xml:"ForwardMatrix2" json:"forwardMatrix2,omitempty"`
	ReductionMatrix1             DCPReductionMatrix `xml:"ReductionMatrix1" json:"reductionMatrix1,omitempty"`
	ReductionMatrix2             DCPReductionMatrix `xml:"ReductionMatrix2" json:"reductionMatrix2,omitempty"`
	Copyright                    string             `xml:"Copyright" json:"copyright,omitempty"`
	EmbedPolicy                  string             `xml:"EmbedPolicy" json:"embedPolicy,omitempty"`
	ProfileHueSatMapEncoding     string             `xml:"ProfileHueSatMapEncoding" json:"profileHueSatMapEncoding,omitempty"`
	LookTable                    *DCPLookTable      `xml:"LookTable" json:"lookTable,omitempty"`
	ToneCurve                    *DCPToneCurve      `xml:"ToneCurve" json:"toneCurve,omitempty"`
	ProfileCalibrationSignature  string             `xml:"ProfileCalibrationSignature" json:"profileCalibrationSignature,omitempty"`
	UniqueCameraModelRestriction string             `xml:"UniqueCameraModelRestriction" json:"uniqueCameraModelRestriction,omitempty"`
	ProfileLookTableEncoding     string             `xml:"ProfileLookTableEncoding" json:"profileLookTableEncoding,omitempty"`
	BaselineExposureOffset       string             `xml:"BaselineExposureOffset" json:"baselineExposureOffset,omitempty"`
	DefaultBlackRender           string             `xml:"DefaultBlackRender" json:"defaultBlackRender,omitempty"`
}

func (dd *DCPData) DeepCopy(x *DCPData) error {
	b, err := xml.Marshal(dd)
	if err != nil {
		return err
	}

	return xml.Unmarshal(b, &x)
}

func (dd *DCPData) Diff(
	other *DCPData,
	skipToneCurve bool,
	ignoreToneCurveValues [][2]float64,
) (*DCPData, error) {
	if other == nil {
		return nil, errors.New("other profile is nil")
	}

	n := &DCPData{}
	err := dd.DeepCopy(n)
	if err != nil {
		return nil, err
	}

	n.ProfileName = fmt.Sprintf(
		"Correct %s to %s", dd.ProfileName, other.ProfileName,
	)

	if len(n.LookTable.Elements) != len(other.LookTable.Elements) {
		return nil, errors.New("ProfileLookTable length mismatch")
	}

	for i, row := range n.LookTable.Elements {
		row.HueDiv = (row.HueDiv * -1) + other.LookTable.Elements[i].HueDiv
		row.SatDiv = (row.SatDiv * -1) + other.LookTable.Elements[i].SatDiv
		row.ValDiv = (row.ValDiv * -1) + other.LookTable.Elements[i].ValDiv
		row.HueShift = (row.HueShift * -1) + other.LookTable.Elements[i].HueShift
		row.SatScale = (row.SatScale * -1) + other.LookTable.Elements[i].SatScale
		row.ValScale = (row.ValScale * -1) + other.LookTable.Elements[i].ValScale
	}

	if skipToneCurve {
		return n, nil
	}

	curveA := n.ToneCurve.Elements.Filter(ignoreToneCurveValues)
	curveB := other.ToneCurve.Elements.Filter(ignoreToneCurveValues)
	fmt.Printf("len(curveA): %#v\n", len(curveA))

	if len(curveA) != len(curveB) {
		return nil, errors.New("ProfileToneCurve length mismatch")
	}

	for i, row := range curveA {
		row.N = i
		row.H = (row.H * -1) + curveB[i].H
		row.V = (row.V * -1) + curveB[i].V
	}

	n.ToneCurve.Elements = curveA

	return n, nil
}

type DCPColorMatrix struct {
	Rows     int                      `xml:"Rows,attr" json:"rows,omitempty"`
	Cols     int                      `xml:"Cols,attr" json:"cols,omitempty"`
	Elements []*DCPColorMatrixElement `xml:"Element" json:"elements,omitempty"`
}

type DCPColorMatrixElement struct {
	Text float64 `xml:",chardata" json:"text,omitempty"`
	Row  int     `xml:"Row,attr" json:"row,omitempty"`
	Col  int     `xml:"Col,attr" json:"col,omitempty"`
}

type DCPReductionMatrix struct {
	Rows int `xml:"Rows,attr" json:"rows,omitempty"`
	Cols int `xml:"Cols,attr" json:"cols,omitempty"`
}

type DCPLookTable struct {
	HueDivisions int                    `xml:"hueDivisions,attr" json:"hueDivisions,omitempty"`
	SatDivisions int                    `xml:"satDivisions,attr" json:"satDivisions,omitempty"`
	ValDivisions int                    `xml:"valDivisions,attr" json:"valDivisions,omitempty"`
	Elements     []*DCPLookTableElement `xml:"Element" json:"elements,omitempty"`
}

type DCPLookTableElement struct {
	HueDiv   int     `xml:"HueDiv,attr" json:"hueDiv,omitempty"`
	SatDiv   int     `xml:"SatDiv,attr" json:"satDiv,omitempty"`
	ValDiv   int     `xml:"ValDiv,attr" json:"valDiv,omitempty"`
	HueShift float64 `xml:"HueShift,attr" json:"hueShift,omitempty"`
	SatScale float64 `xml:"SatScale,attr" json:"satScale,omitempty"`
	ValScale float64 `xml:"ValScale,attr" json:"valScale,omitempty"`
}

type DCPToneCurve struct {
	Size     int                  `xml:"Size,attr" json:"size,omitempty"`
	Elements DCPToneCurveElements `xml:"Element" json:"elements,omitempty"`
}

type DCPToneCurveElements []*DCPToneCurveElement

func (els DCPToneCurveElements) Filter(filters [][2]float64) DCPToneCurveElements {
	var out []*DCPToneCurveElement
	for _, row := range els {
		ignore := false
		for _, val := range filters {
			if row.H == val[0] && row.V == val[1] {
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

type DCPToneCurveElement struct {
	N int     `xml:"N,attr" json:"n,omitempty"`
	H float64 `xml:"h,attr" json:"h,omitempty"`
	V float64 `xml:"v,attr" json:"v,omitempty"`
}
