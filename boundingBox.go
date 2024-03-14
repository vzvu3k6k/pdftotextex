package hyperpaper

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

type BoundingBox struct {
	Page int
	Rect *Rect
	Text string
}

// LoadBoundingBoxesは`pdftotext -tsv`によって出力されたバウンディングボックスの情報を読み込む。
func LoadBoundingBoxes(in io.Reader) ([]*BoundingBox, error) {
	boundingBoxes := []*BoundingBox{}

	r := csv.NewReader(in)
	r.Comma = '\t'

	// ヘッダ行を読み飛ばす。
	_, err := r.Read()
	if err != nil {
		return nil, err
	}

	// 最初の行でページサイズを取得する。
	record, err := r.Read()
	if err != nil {
		return nil, err
	}
	if record[11] != "###PAGE###" {
		return nil, fmt.Errorf("expected ###PAGE###, but got %s", record[11])
	}
	if record[10] != "-1" {
		return nil, fmt.Errorf("expected conf is -1, but got %s", record[10])
	}
	pageWidth, err := strconv.ParseFloat(record[8], 64)
	if err != nil {
		return nil, err
	}
	pageHeight, err := strconv.ParseFloat(record[9], 64)
	if err != nil {
		return nil, err
	}

	// CSVの各行をBoundingBox構造体に詰め替える。
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// conf = -1の行は特殊な意味を持つので読み飛ばす。
		if record[10] == "-1" {
			continue
		}

		page, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, err
		}
		x, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			return nil, err
		}
		y, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			return nil, err
		}
		width, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			return nil, err
		}
		height, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			return nil, err
		}
		text := record[11]

		boundingBoxes = append(boundingBoxes, &BoundingBox{
			Page: page,
			Text: text,
			Rect: &Rect{
				X:      x / pageWidth,
				Y:      y / pageHeight,
				Width:  width / pageWidth,
				Height: height / pageHeight,
			},
		})
	}

	return boundingBoxes, nil
}
