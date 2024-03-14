package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math"
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

func almostEqual(a, b float64) bool {
	const epsilon = 1e-3 * 5
	return math.Abs(a-b) < epsilon
}

func lookupBoundingBox(boxes []*BoundingBox, page int, x, y, width, height float64) (*BoundingBox, bool) {
	for _, box := range boxes {
		if box.Page == page &&
			almostEqual(box.X, x) &&
			almostEqual(box.Y, y) &&
			almostEqual(box.Width, width) &&
			almostEqual(box.Height, height) {
			return box, true
		}
	}
	return nil, false
}

func buildHTML(in io.Reader, boundingBoxes []*BoundingBox) error {
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
				typeAttr, ok := lookupAttrValue(token.Attr, "TYPE")
				if !ok {
					return fmt.Errorf("LINE element has no TYPE attribute")
				}
				if typeAttr == "本文" {
					stringAttr, ok := lookupAttrValue(token.Attr, "STRING")
					if !ok {
						return fmt.Errorf("LINE element has no STRING attribute")
					}
					_ = stringAttr
				}

				xAttrValue, ok := lookupAttrValue(token.Attr, "X")
				if !ok {
					return fmt.Errorf("LINE element has no X attribute")
				}
				x, err := strconv.Atoi(xAttrValue)
				if err != nil {
					return err
				}

				yAttrValue, ok := lookupAttrValue(token.Attr, "Y")
				if !ok {
					return fmt.Errorf("LINE element has no Y attribute")
				}
				y, err := strconv.Atoi(yAttrValue)
				if err != nil {
					return err
				}

				widthAttrValue, ok := lookupAttrValue(token.Attr, "WIDTH")
				if !ok {
					return fmt.Errorf("LINE element has no WIDTH attribute")
				}
				width, err := strconv.Atoi(widthAttrValue)
				if err != nil {
					return err
				}

				heightAttrValue, ok := lookupAttrValue(token.Attr, "HEIGHT")
				if !ok {
					return fmt.Errorf("LINE element has no HEIGHT attribute")
				}
				height, err := strconv.Atoi(heightAttrValue)
				if err != nil {
					return err
				}

				boundingBox, ok := lookupBoundingBox(
					boundingBoxes,
					1,
					float64(x)/float64(pageWidth),
					float64(y)/float64(pageHeight),
					float64(width)/float64(pageWidth),
					float64(height)/float64(pageHeight),
				)
				if !ok {
					return fmt.Errorf("bounding box for %q not found", token)
				}
				fmt.Printf("bb: %+v\n", boundingBox.Text)
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
	Page   int
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
			Page:   page,
			X:      x / pageWidth,
			Y:      y / pageHeight,
			Width:  width / pageWidth,
			Height: height / pageHeight,
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

	if err := buildHTML(layoutFile, boxes); err != nil {
		handleErr(err)
	}
}
