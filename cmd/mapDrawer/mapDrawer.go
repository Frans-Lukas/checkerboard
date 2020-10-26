package mapDrawer

import (
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	objects2 "github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"unsafe"
)

type MapInfo struct {
	name string
	draw.Image
	gdkwin      *gdk.Window
	drawingarea *gtk.DrawingArea
	pixmap      *gdk.Pixmap
	gc          *gdk.GC
	sizeX       int
	sizeY       int
}

func SetupMap(name string, sizeX int, sizeY int) *MapInfo {

	newImage := image.NewRGBA(image.Rectangle{Min: image.Pt(0, 0), Max: image.Pt(sizeX, sizeY)})

	thisMap := MapInfo{name: name, Image: newImage, sizeX: sizeX, sizeY: sizeY}

	///////////////////////////////////////////////////

	gtk.Init(&os.Args)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle(name)
	window.Connect("destroy", gtk.MainQuit)

	vbox := gtk.NewVBox(true, 0)
	vbox.SetBorderWidth(5)
	thisMap.drawingarea = gtk.NewDrawingArea()

	var p1, p2 gdk.Point
	p1.X = -1
	p1.Y = -1
	colors := []string{
		"black",
		"gray",
		"blue",
		"purple",
		"red",
		"orange",
		"yellow",
		"green",
		"darkgreen",
	}

	thisMap.drawingarea.Connect("configure-event", func() {
		if thisMap.pixmap != nil {
			thisMap.pixmap.Unref()
		}
		allocation := thisMap.drawingarea.GetAllocation()
		thisMap.pixmap = gdk.NewPixmap(thisMap.drawingarea.GetWindow().GetDrawable(), allocation.Width, allocation.Height, 24)
		thisMap.gc = gdk.NewGC(thisMap.pixmap.GetDrawable())
		thisMap.gc.SetRgbFgColor(gdk.NewColor("white"))
		thisMap.pixmap.GetDrawable().DrawRectangle(thisMap.gc, true, 0, 0, -1, -1)
		thisMap.gc.SetRgbFgColor(gdk.NewColor(colors[0]))
		thisMap.gc.SetRgbBgColor(gdk.NewColor("white"))
	})

	thisMap.drawingarea.Connect("motion-notify-event", func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		mev := *(**gdk.EventMotion)(unsafe.Pointer(&arg))
		var mt gdk.ModifierType
		if mev.IsHint != 0 {
			thisMap.gdkwin.GetPointer(&p2.X, &p2.Y, &mt)
		} else {
			p2.X, p2.Y = int(mev.X), int(mev.Y)
		}
		if p1.X != -1 && p2.X != -1 && (gdk.EventMask(mt)&gdk.BUTTON_PRESS_MASK) != 0 {
			thisMap.pixmap.GetDrawable().DrawLine(thisMap.gc, p1.X, p1.Y, p2.X, p2.Y)
			thisMap.gdkwin.Invalidate(nil, false)
		}
		colors = append(colors[1:], colors[0])
		thisMap.gc.SetRgbFgColor(gdk.NewColor(colors[0]))
		p1 = p2
	})

	thisMap.drawingarea.Connect("expose-event", func() {
		thisMap.Redraw()
	})

	glib.TimeoutAdd(1000, thisMap.Redraw, thisMap.drawingarea)

	thisMap.drawingarea.SetEvents(int(gdk.POINTER_MOTION_MASK | gdk.POINTER_MOTION_HINT_MASK | gdk.BUTTON_PRESS_MASK))
	vbox.Add(thisMap.drawingarea)

	window.Add(vbox)
	window.SetSizeRequest(sizeX, sizeY)
	window.ShowAll()

	thisMap.gdkwin = thisMap.drawingarea.GetWindow()

	go func() {
		gtk.Main()
	}()

	return &thisMap
}

func (thisMap *MapInfo) Redraw() bool {
	if thisMap.pixmap == nil {
		return true
	}
	tmpImage := gtk.NewImageFromFile(thisMap.name + ".png")
	thisMap.pixmap.GetDrawable().DrawPixbuf(thisMap.gc, tmpImage.GetPixbuf(), 0, 0, 0, 0, -1, -1, gdk.RGB_DITHER_NONE, 0, 0)
	//gdkwin.GetDrawable().DrawDrawable(gc, &gdk.Drawable{GDrawable: newPlayerImage}, 0, 0, 0, 0, -1, -1)
	thisMap.gdkwin.GetDrawable().DrawDrawable(thisMap.gc, thisMap.pixmap.GetDrawable(), 0, 0, 0, 0, -1, -1)
	return true
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

func (thisMap *MapInfo) DrawClient(posX int, posY int, imagePath string) {
	startPoint := image.Pt(posX*(thisMap.sizeX/constants.MAP_SIZE), posY*(thisMap.sizeX/constants.MAP_SIZE))

	f, err := os.Open(imagePath)
	if err != nil {
		println("SetupMap: failed to load image: ", err.Error())
	}

	drawableImage, _, err := image.Decode(f)

	if err != nil {
		println("SetupMap: failed to decode image: ", err.Error())
	}

	drawableImage = resize.Resize(uint(thisMap.sizeX/constants.MAP_SIZE), uint(thisMap.sizeX/constants.MAP_SIZE), drawableImage, resize.Lanczos3)

	r := image.Rectangle{Min: startPoint, Max: startPoint.Add(startPoint.Add(drawableImage.Bounds().Size()))}
	draw.Draw(thisMap.Image, r, drawableImage, drawableImage.Bounds().Min, draw.Src)

}

func (thisMap *MapInfo) DrawCellBoundaries(cell objects2.Cell) {
	topLeft := image.Pt(int(cell.PosX)*(thisMap.sizeX/constants.MAP_SIZE), int(cell.PosY)*(thisMap.sizeX/constants.MAP_SIZE))
	bottomLeft := image.Pt(int(cell.PosX)*(thisMap.sizeX/constants.MAP_SIZE), int(cell.PosY+cell.Height)*(thisMap.sizeX/constants.MAP_SIZE))
	topRight := image.Pt(int(cell.PosX+cell.Width)*(thisMap.sizeX/constants.MAP_SIZE), int(cell.PosY)*(thisMap.sizeX/constants.MAP_SIZE))
	bottomRight := image.Pt(int(cell.PosX+cell.Width)*(thisMap.sizeX/constants.MAP_SIZE), int(cell.PosY+cell.Height)*(thisMap.sizeX/constants.MAP_SIZE))

	borderColor := color.RGBA{R: 100, G: 0, B: 0, A: 0xff}

	for i := topLeft; i.X < topRight.X; i.X++ {
		thisMap.Image.Set(i.X, i.Y-1, borderColor)
		thisMap.Image.Set(i.X, i.Y, borderColor)
		thisMap.Image.Set(i.X, i.Y+1, borderColor)
	}

	for i := topLeft; i.Y < bottomLeft.Y; i.Y++ {
		thisMap.Image.Set(i.X-1, i.Y, borderColor)
		thisMap.Image.Set(i.X, i.Y, borderColor)
		thisMap.Image.Set(i.X+1, i.Y, borderColor)
	}

	for i := bottomLeft; i.X < bottomRight.X; i.X++ {
		thisMap.Image.Set(i.X, i.Y-1, borderColor)
		thisMap.Image.Set(i.X, i.Y, borderColor)
		thisMap.Image.Set(i.X, i.Y+1, borderColor)
	}

	for i := topRight; i.Y < bottomRight.Y; i.Y++ {
		thisMap.Image.Set(i.X-1, i.Y, borderColor)
		thisMap.Image.Set(i.X, i.Y, borderColor)
		thisMap.Image.Set(i.X+1, i.Y, borderColor)
	}
}

func (thisMap *MapInfo) SaveMapAsPNG() {
	file, err := os.Create(thisMap.name + ".png")
	if err != nil {
		println("SaveMapAsPNG: ", err.Error())
	}
	png.Encode(file, thisMap.Image)

	gtk.EventsPending()
}
