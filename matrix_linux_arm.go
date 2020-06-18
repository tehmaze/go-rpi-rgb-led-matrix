package rgbmatrix

/*
#cgo CFLAGS: -std=c99 -I${SRCDIR}/vendor/rpi-rgb-led-matrix/include -DSHOW_REFRESH_RATE
#cgo LDFLAGS: -lrgbmatrix -L${SRCDIR}/vendor/rpi-rgb-led-matrix/lib -lstdc++ -lm
#include <led-matrix-c.h>

void led_matrix_swap(struct RGBLedMatrix *matrix, struct LedCanvas *offscreen_canvas,
                     int width, int height, const uint32_t pixels[]) {


  int i, x, y;
  uint32_t color;
  for (x = 0; x < width; ++x) {
    for (y = 0; y < height; ++y) {
      i = x + (y * width);
      color = pixels[i];

      led_canvas_set_pixel(offscreen_canvas, x, y,
        (color >> 16) & 255, (color >> 8) & 255, color & 255);
    }
  }

  offscreen_canvas = led_matrix_swap_on_vsync(matrix, offscreen_canvas);
}

void set_show_refresh_rate(struct RGBLedMatrixOptions *o, int show_refresh_rate) {
  o->show_refresh_rate = show_refresh_rate != 0 ? 1 : 0;
}

void set_disable_hardware_pulsing(struct RGBLedMatrixOptions *o, int disable_hardware_pulsing) {
  o->disable_hardware_pulsing = disable_hardware_pulsing != 0 ? 1 : 0;
}

void set_inverse_colors(struct RGBLedMatrixOptions *o, int inverse_colors) {
  o->inverse_colors = inverse_colors != 0 ? 1 : 0;
}
*/
import "C"
import (
	"fmt"
	"image/color"
	"unsafe"
)

func (c *HardwareConfig) toC() *C.struct_RGBLedMatrixOptions {
	o := &C.struct_RGBLedMatrixOptions{}
	o.rows = C.int(c.Rows)
	o.cols = C.int(c.Cols)
	o.chain_length = C.int(c.ChainLength)
	o.parallel = C.int(c.Parallel)
	o.pwm_bits = C.int(c.PWMBits)
	o.pwm_lsb_nanoseconds = C.int(c.PWMLSBNanoseconds)
	o.brightness = C.int(c.Brightness)
	o.scan_mode = C.int(c.ScanMode)
	o.hardware_mapping = C.CString(c.HardwareMapping)

	if c.ShowRefreshRate == true {
		C.set_show_refresh_rate(o, C.int(1))
	} else {
		C.set_show_refresh_rate(o, C.int(0))
	}

	if c.DisableHardwarePulsing == true {
		C.set_disable_hardware_pulsing(o, C.int(1))
	} else {
		C.set_disable_hardware_pulsing(o, C.int(0))
	}

	if c.InverseColors == true {
		C.set_inverse_colors(o, C.int(1))
	} else {
		C.set_inverse_colors(o, C.int(0))
	}

	return o
}

func newRGBLedMatrix(config *HardwareConfig) (Matrix, error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error creating matrix: %v", r)
			}
		}
	}()

	w, h := config.geometry()
	m := C.led_matrix_create_from_options(config.toC(), nil, nil)
	b := C.led_matrix_create_offscreen_canvas(m)
	c = &rgbLedMatrix{
		Config: config,
		width:  w, height: h,
		matrix: m,
		buffer: b,
		leds:   make([]C.uint32_t, w*h),
	}
	if m == nil {
		return nil, fmt.Errorf("unable to allocate memory")
	}

	return c, nil
}

// rgbLedMatrix matrix representation for ws281x
type rgbLedMatrix struct {
	Config *HardwareConfig

	height int
	width  int
	matrix *C.struct_RGBLedMatrix
	buffer *C.struct_LedCanvas
	leds   []C.uint32_t
}

// Initialize initialize library, must be called once before other functions are
// called.
func (c *rgbLedMatrix) Initialize() error {
	return nil
}

// Geometry returns the width and the height of the matrix
func (c *rgbLedMatrix) Geometry() (width, height int) {
	return c.width, c.height
}

// Apply set all the pixels to the values contained in leds
func (c *rgbLedMatrix) Apply(leds []color.Color) error {
	for position, l := range leds {
		c.Set(position, l)
	}

	return c.Render()
}

// Render update the display with the data from the LED buffer
func (c *rgbLedMatrix) Render() error {
	w, h := c.Config.geometry()

	C.led_matrix_swap(
		c.matrix,
		c.buffer,
		C.int(w), C.int(h),
		(*C.uint32_t)(unsafe.Pointer(&c.leds[0])),
	)

	c.leds = make([]C.uint32_t, w*h)
	return nil
}

// At return an Color which allows access to the LED display data as
// if it were a sequence of 24-bit RGB values.
func (c *rgbLedMatrix) At(position int) color.Color {
	return uint32ToColor(c.leds[position])
}

// Set set LED at position x,y to the provided 24-bit color value.
func (c *rgbLedMatrix) Set(position int, color color.Color) {
	c.leds[position] = C.uint32_t(colorToUint32(color))
}

// Close finalizes the ws281x interface
func (c *rgbLedMatrix) Close() error {
	C.led_matrix_delete(c.matrix)
	return nil
}

func colorToUint32(c color.Color) uint32 {
	if c == nil {
		return 0
	}

	// A color's RGBA method returns values in the range [0, 65535]
	red, green, blue, _ := c.RGBA()
	return (red>>8)<<16 | (green>>8)<<8 | blue>>8
}

func uint32ToColor(u C.uint32_t) color.Color {
	return color.RGBA{
		uint8(u>>16) & 255,
		uint8(u>>8) & 255,
		uint8(u>>0) & 255,
		0,
	}
}
