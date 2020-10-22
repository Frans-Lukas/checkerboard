package mapDrawer

import (
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

type MapInfo struct {
	name string
	draw.Image
	sizeX       int
	sizeY       int
	playerImage image.Image
	//clientImage draw.Image
}

func SetupMap(name string, sizeX int, sizeY int) MapInfo {

	newImage := image.NewRGBA(image.Rectangle{Min: image.Pt(0, 0), Max: image.Pt(sizeX, sizeY)})

	f, err := os.Open(constants.PlayerImage)
	if err != nil {
		println("SetupMap: failed to load image: ", err.Error())
	}

	newPlayerImage, _, err := image.Decode(f)
	newPlayerImage = resize.Resize(uint(sizeX/constants.MAP_SIZE), uint(sizeX/constants.MAP_SIZE), newPlayerImage, resize.Lanczos3)
	if err != nil {
		println("SetupMap: failed to decode image: ", err.Error())
	}

	return MapInfo{name: name, Image: newImage, sizeX: sizeX, sizeY: sizeY, playerImage: newPlayerImage}
}

func (thisMap *MapInfo) ClearMap() {
	white := color.RGBA{R: 100, G: 100, B: 100, A: 0xff}
	gray := color.RGBA{R: 50, G: 50, B: 50, A: 0xff}
	for x := 0; x < thisMap.sizeX; x++ {
		for y := 0; y < thisMap.sizeY; y++ {
			if x%((thisMap.sizeX/constants.MAP_SIZE)*2) < thisMap.sizeX/constants.MAP_SIZE {
				if y%((thisMap.sizeX/constants.MAP_SIZE)*2) < thisMap.sizeX/constants.MAP_SIZE {
					thisMap.Image.Set(x, y, white)
				} else {
					thisMap.Image.Set(x, y, gray)
				}
			} else {
				if y%((thisMap.sizeX/constants.MAP_SIZE)*2) < thisMap.sizeX/constants.MAP_SIZE {
					thisMap.Image.Set(x, y, gray)
				} else {
					thisMap.Image.Set(x, y, white)
				}
			}
		}
	}
}

func (thisMap *MapInfo) DrawClient(posX int, posY int, isPlayer bool) {
	startPoint := image.Pt(posX*(thisMap.sizeX/constants.MAP_SIZE), posY*(thisMap.sizeX/constants.MAP_SIZE))
	if isPlayer {
		r := image.Rectangle{Min: startPoint, Max: startPoint.Add(startPoint.Add(thisMap.playerImage.Bounds().Size()))}
		draw.Draw(thisMap.Image, r, thisMap.playerImage, thisMap.playerImage.Bounds().Min, draw.Src)
	}
}

func (thisMap *MapInfo) SaveMapAsPNG() {
	file, err := os.Create(thisMap.name + ".png")
	if err != nil {
		println("SaveMapAsPNG: ", err.Error())
	}
	png.Encode(file, thisMap.Image)
}
