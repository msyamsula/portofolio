package textrendering

import (
	"github.com/msyamsula/portofolio/other-works/design-pattern/flyweight/text_rendering/font"
	"github.com/msyamsula/portofolio/other-works/design-pattern/flyweight/text_rendering/font/color"
	"github.com/msyamsula/portofolio/other-works/design-pattern/flyweight/text_rendering/font/family"
	"github.com/msyamsula/portofolio/other-works/design-pattern/flyweight/text_rendering/font/size"
	"github.com/msyamsula/portofolio/other-works/design-pattern/flyweight/text_rendering/font/weight"
)

type FontFlyweight interface {
	IsFontFlyweight() bool
}

type Text struct {
	Characters []font.Character

	FontFlyWeights map[string]FontFlyweight

	// fontColorMap  map[string]color.FontColor
	// fontWeightMap map[string]weight.FontWeight
	// fontFamilyMap map[string]family.FontFamily
	// fontSizeMap   map[string]size.FontSize
}

func NewText() *Text {
	return &Text{
		Characters: []font.Character{}, // fontColorMap:  make(map[string]color.FontColor),
		// fontWeightMap: make(map[string]weight.FontWeight),
		// fontFamilyMap: make(map[string]family.FontFamily),
		// fontSizeMap:   make(map[string]size.FontSize),
		FontFlyWeights: make(map[string]FontFlyweight),
	}
}

// func (t *Text) AddCharactersFlyweight(char byte, x, y int, c, w, s, f string) {
// 	fontColor, ok := t.fontColorMap[c]
// 	if !ok {
// 		fontColor = color.NewColor(c)
// 		t.fontColorMap[c] = fontColor
// 	}

// 	fontWeight, ok := t.fontWeightMap[w]
// 	if !ok {
// 		fontWeight = weight.NewFontWeight(w)
// 		t.fontWeightMap[w] = fontWeight
// 	}

// 	fontFamily, ok := t.fontFamilyMap[f]
// 	if !ok {
// 		fontFamily = family.NewFontFamily(f)
// 		t.fontFamilyMap[f] = fontFamily
// 	}

// 	fontSize, ok := t.fontSizeMap[s]
// 	if !ok {
// 		fontSize = size.NewFontSize(s)
// 		t.fontSizeMap[s] = fontSize
// 	}

// 	nc := font.Character{
// 		Char: char,
// 		Position: font.Position{
// 			X: x,
// 			Y: y,
// 		},
// 		Color:  fontColor,
// 		Family: fontFamily,
// 		Size:   fontSize,
// 		Weight: fontWeight,
// 	}

// 	t.Characters = append(t.Characters, nc)
// }

func (t *Text) AddCharacters(char byte, x, y int, c, w, s, f string) {
	fontColor := t.FontFlyWeights["color"]
	if fontColor == nil {
		fontColor = color.NewColor(c)
		t.FontFlyWeights["color"] = fontColor
	}
	fontWeight := t.FontFlyWeights["weight"]
	if fontWeight == nil {
		fontWeight = weight.NewFontWeight(w)
		t.FontFlyWeights["weight"] = fontWeight
	}
	fontFamily := t.FontFlyWeights["family"]
	if fontFamily == nil {
		fontFamily = family.NewFontFamily(f)
		t.FontFlyWeights["family"] = fontFamily
	}
	fontSize := t.FontFlyWeights["size"]
	if fontSize == nil {
		fontSize = size.NewFontSize(s)
		t.FontFlyWeights["size"] = fontSize
	}
	t.Characters = append(t.Characters, font.Character{
		Char: char,
		Position: font.Position{
			X: x,
			Y: y,
		},
		Color:  fontColor.(color.FontColor),
		Family: fontFamily.(family.FontFamily),
		Size:   fontSize.(size.FontSize),
		Weight: fontWeight.(weight.FontWeight),
	})
}
