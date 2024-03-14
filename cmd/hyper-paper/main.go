package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/vzvu3k6k/hyperpaper"
)

func handleErr(err error) {
	log.Fatal(err)
}

func lookupAttrValue(attrs []xml.Attr, name string) (string, bool) {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value, true
		}
	}
	return "", false
}

// handlePageElementはPAGE要素の属性を処理する。
// PAGE要素のWIDTHとHEIGHT属性をそれぞれ返す。属性が存在しない場合、エラーを返す。
func handlePageElement(element xml.StartElement) (int, int, error) {
	pageWidthValue, ok := lookupAttrValue(element.Attr, "WIDTH")
	if !ok {
		return 0, 0, fmt.Errorf("page width is not set")
	}
	pageHeightValue, ok := lookupAttrValue(element.Attr, "HEIGHT")
	if !ok {
		return 0, 0, fmt.Errorf("page height is not set")
	}

	pageWidth, err := strconv.Atoi(pageWidthValue)
	if err != nil {
		return 0, 0, err
	}
	pageHeight, err := strconv.Atoi(pageHeightValue)
	if err != nil {
		return 0, 0, err
	}

	return pageWidth, pageHeight, nil
}

func handleLineElement(element xml.StartElement, pageWidth, pageHeight int) (*hyperpaper.Rect, error) {
	typeAttr, ok := lookupAttrValue(element.Attr, "TYPE")
	if !ok {
		return nil, fmt.Errorf("LINE element has no TYPE attribute")
	}
	if typeAttr != "本文" {
		return nil, nil
	}

	xAttrValue, ok := lookupAttrValue(element.Attr, "X")
	if !ok {
		return nil, fmt.Errorf("LINE element has no X attribute")
	}
	x, err := strconv.Atoi(xAttrValue)
	if err != nil {
		return nil, err
	}

	yAttrValue, ok := lookupAttrValue(element.Attr, "Y")
	if !ok {
		return nil, fmt.Errorf("LINE element has no Y attribute")
	}
	y, err := strconv.Atoi(yAttrValue)
	if err != nil {
		return nil, err
	}

	widthAttrValue, ok := lookupAttrValue(element.Attr, "WIDTH")
	if !ok {
		return nil, fmt.Errorf("LINE element has no WIDTH attribute")
	}
	width, err := strconv.Atoi(widthAttrValue)
	if err != nil {
		return nil, err
	}

	heightAttrValue, ok := lookupAttrValue(element.Attr, "HEIGHT")
	if !ok {
		return nil, fmt.Errorf("LINE element has no HEIGHT attribute")
	}
	height, err := strconv.Atoi(heightAttrValue)
	if err != nil {
		return nil, err
	}

	return &hyperpaper.Rect{
		X:      float64(x) / float64(pageWidth),
		Y:      float64(y) / float64(pageHeight),
		Width:  float64(width) / float64(pageWidth),
		Height: float64(height) / float64(pageHeight),
	}, nil
}

func buildVisibleRects(in io.Reader) ([]*hyperpaper.Rect, error) {
	rects := []*hyperpaper.Rect{}

	var pageWidth, pageHeight int

	d := xml.NewDecoder(in)
	for {
		token, err := d.Token()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case xml.StartElement:
			switch token.Name.Local {
			case "PAGE":
				pageWidth, pageHeight, err = handlePageElement(token)
				if err != nil {
					return nil, err
				}
			case "LINE":
				rect, err := handleLineElement(token, pageWidth, pageHeight)
				if err != nil {
					return nil, err
				}
				if rect == nil {
					continue
				}
				rects = append(rects, rect)
				// s, _ := lookupAttrValue(token.Attr, "STRING")
				// fmt.Printf("layout text: %s\n", s)

				// fmt.Printf("bb: %+v\n", boundingBox.Text)
			}
		case xml.EndElement:
			//do something
		case xml.CharData:
			//do something
		case xml.Comment:
			//do something
		case xml.ProcInst:
			//do something
		case xml.Directive:
			//do something
		default:
			panic("unknown xml token.")
		}
	}
	return rects, nil
}

type boundingBox struct {
	Page int
	Rect *hyperpaper.Rect
	Text string
}

// loadBoundingBoxesは`pdftotext -tsv`によって出力されたバウンディングボックスの情報を読み込む。
func loadBoundingBoxes(in io.Reader) ([]*boundingBox, error) {
	boundingBoxes := []*boundingBox{}

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

		boundingBoxes = append(boundingBoxes, &boundingBox{
			Page: page,
			Text: text,
			Rect: &hyperpaper.Rect{
				X:      x / pageWidth,
				Y:      y / pageHeight,
				Width:  width / pageWidth,
				Height: height / pageHeight,
			},
		})
	}

	return boundingBoxes, nil
}

func main() {
	boundingBoxesFile, err := os.Open("C3-5.tsv")
	if err != nil {
		handleErr(err)
	}
	defer boundingBoxesFile.Close()

	boundingBoxes, err := loadBoundingBoxes(boundingBoxesFile)
	if err != nil {
		handleErr(err)
	}

	layoutFile, err := os.Open("C5-5-1.sorted.xml")
	if err != nil {
		handleErr(err)
	}
	defer layoutFile.Close()

	visibleRects, err := buildVisibleRects(layoutFile)
	if err != nil {
		handleErr(err)
	}

	for _, b := range boundingBoxes {
		if b.Page != 1 {
			continue
		}
		for _, r := range visibleRects {
			if hyperpaper.IsOverlapping(b.Rect, r) {
				fmt.Printf("visible: %s\n", b.Text)
				break
			}
		}
	}
}
