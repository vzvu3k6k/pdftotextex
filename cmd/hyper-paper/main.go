package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
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
	var pageWidth, pageHeight int
	var err error

	for _, attr := range element.Attr {
		switch attr.Name.Local {
		case "WIDTH":
			pageWidth, err = strconv.Atoi(attr.Value)
			if err != nil {
				return 0, 0, err
			}
		case "HEIGHT":
			pageHeight, err = strconv.Atoi(attr.Value)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	// WIDTHとHEIGHTが存在しない場合は0のままなのでエラーを返す。
	if pageWidth == 0 || pageHeight == 0 {
		return 0, 0, fmt.Errorf("page width or height is not set")
	}

	return pageWidth, pageHeight, nil
}

func buildHTML(in io.Reader) error {
	var pageWidth, pageHeight int

	d := xml.NewDecoder(in)
	for {
		token, err := d.Token()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			switch token.Name.Local {
			case "PAGE":
				pageWidth, pageHeight, err = handlePageElement(token)
				if err != nil {
					return err
				}
			case "LINE":

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
	fmt.Printf("pageWidth: %d, pageHeight: %d\n", pageWidth, pageHeight)
	return nil
}

type BoundingBox struct {
	Page   float64
	X      float64
	Y      float64
	Width  float64
	Height float64
	Text   string
}

// loadBoundingBoxesは`pdftotext -tsv`によって出力されたバウンディングボックスの情報を読み込む。
func loadBoundingBoxes(in io.Reader) ([]*BoundingBox, error) {
	boundingBoxes := []*BoundingBox{}

	r := csv.NewReader(in)
	r.Comma = '\t'

	// ヘッダ行を読み飛ばす。
	_, err := r.Read()
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

		page, err := strconv.ParseFloat(record[1], 64)
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
			Page:   page,
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
			Text:   text,
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

	boxes, err := loadBoundingBoxes(boundingBoxesFile)
	if err != nil {
		handleErr(err)
	}

	layoutFile, err := os.Open("C5-5-1.sorted.xml")
	if err != nil {
		handleErr(err)
	}
	defer layoutFile.Close()

	buildHTML(layoutFile)
}
