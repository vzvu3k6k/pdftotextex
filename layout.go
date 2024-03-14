package hyperpaper

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

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

func handleLineElement(element xml.StartElement, pageWidth, pageHeight int) (*Rect, error) {
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

	return &Rect{
		X:      float64(x) / float64(pageWidth),
		Y:      float64(y) / float64(pageHeight),
		Width:  float64(width) / float64(pageWidth),
		Height: float64(height) / float64(pageHeight),
	}, nil
}

func BuildVisibleRects(in io.Reader) ([]*Rect, error) {
	rects := []*Rect{}

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
