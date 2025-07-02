package font

import (
	"github.com/msyamsula/portofolio/design_pattern/flyweight/text_rendering/font/color"
	"github.com/msyamsula/portofolio/design_pattern/flyweight/text_rendering/font/family"
	"github.com/msyamsula/portofolio/design_pattern/flyweight/text_rendering/font/size"
	"github.com/msyamsula/portofolio/design_pattern/flyweight/text_rendering/font/weight"
)

type Position struct {
	X, Y int
}

// type Style struct {
// 	color *color.FontColor
// }

type Character struct {
	Char     byte
	Position Position

	Color  color.FontColor
	Family family.FontFamily
	Size   size.FontSize
	Weight weight.FontWeight
}
