package textemulator

import (
	"image/color"
	"io"
	"sync"

	"github.com/gdamore/tcell"
)

type Emulator struct {
	Palette color.Palette
	Title   string
	screen  tcell.Screen
	w, h    int
	buffer  []color.Color
	once    sync.Once
	quit    chan struct{}
}

func (emu *Emulator) Geometry() (width, height int) {
	return emu.w, emu.h
}

func (emu *Emulator) At(position int) color.Color {
	if position < 0 || position >= emu.w*emu.h {
		return emu.Palette[0]
	}
	return emu.buffer[position]
}

func (emu *Emulator) Set(position int, c color.Color) {
	if position >= 0 && position < emu.w*emu.h {
		emu.buffer[position] = c
	}
}

func (emu *Emulator) Apply(colors []color.Color) error {
	return nil
}

func (emu *Emulator) Render() error {
	select {
	case <-emu.quit:
		return io.EOF
	default:
	}

	emu.screen.Clear()
	emu.screen.SetCell((emu.w-len(emu.Title))/2, 0, tcell.StyleDefault, []rune(emu.Title)...)
	for y := 0; y < emu.h; y += 2 {
		o := y * emu.w
		for x := 0; x < emu.w; x++ {
			y0 := emu.buffer[o+x]
			y1 := emu.buffer[o+x+emu.w]
			r0, g0, b0, _ := y0.RGBA()
			r1, g1, b1, _ := y1.RGBA()
			if r0 == r1 && g0 == g1 && b0 == b1 {
				style := tcell.StyleDefault.
					Foreground(tcell.NewRGBColor(int32(r0), int32(g0), int32(b0)))
				emu.screen.SetCell(x, 1+y/2, style, '█')
			} else {
				style := tcell.StyleDefault.
					Background(tcell.NewRGBColor(int32(r0), int32(g0), int32(b0))).
					Foreground(tcell.NewRGBColor(int32(r1), int32(g1), int32(b1)))
				emu.screen.SetCell(x, 1+y/2, style, '▄')
			}
		}
	}
	emu.screen.Show()
	return nil
}

func (emu *Emulator) Close() error {
	close(emu.quit)
	emu.screen.Fini()
	return nil
}

func (emu *Emulator) events() {
	go func() {
		for {
			ev := emu.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(emu.quit)
					return
				case tcell.KeyCtrlL:
					emu.screen.Sync()
				}
			case *tcell.EventResize:
				emu.screen.Sync()
			}
		}
	}()
}

func New(w, h int) (*Emulator, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err = s.Init(); err != nil {
		s.Fini()
		return nil, err
	}
	s.Resize(0, 0, w, h/2)

	buffer := make([]color.Color, w*h)
	for i := range buffer {
		buffer[i] = color.Black
	}

	emu := &Emulator{
		Palette: DefaultPalette,
		screen:  s,
		w:       w,
		h:       h,
		buffer:  buffer,
		quit:    make(chan struct{}),
	}
	go emu.events()
	return emu, nil
}

var DefaultPalette = color.Palette{
	color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff},
	color.RGBA{R: 0x11, G: 0x08, B: 0x00, A: 0xff},
	color.RGBA{R: 0x22, G: 0x11, B: 0x00, A: 0xff},
	color.RGBA{R: 0x33, G: 0x19, B: 0x00, A: 0xff},
	color.RGBA{R: 0x44, G: 0x22, B: 0x00, A: 0xff},
	color.RGBA{R: 0x55, G: 0x2a, B: 0x00, A: 0xff},
	color.RGBA{R: 0x66, G: 0x33, B: 0x00, A: 0xff},
	color.RGBA{R: 0x77, G: 0x3b, B: 0x00, A: 0xff},
	color.RGBA{R: 0x88, G: 0x44, B: 0x00, A: 0xff},
	color.RGBA{R: 0x99, G: 0x4c, B: 0x00, A: 0xff},
	color.RGBA{R: 0xaa, G: 0x55, B: 0x00, A: 0xff},
	color.RGBA{R: 0xbb, G: 0x5d, B: 0x00, A: 0xff},
	color.RGBA{R: 0xcc, G: 0x66, B: 0x00, A: 0xff},
	color.RGBA{R: 0xdd, G: 0x6e, B: 0x00, A: 0xff},
	color.RGBA{R: 0xee, G: 0x77, B: 0x00, A: 0xff},
	color.RGBA{R: 0xff, G: 0x7f, B: 0x00, A: 0xff},
}
