// +build !arm

package rgbmatrix

import "github.com/tehmaze/go-rpi-rgb-led-matrix/emulator"

func newRGBLedMatrix(config *HardwareConfig) (Matrix, error) {
	w, h := config.geometry()
	return emulator.NewEmulator(w, h, emulator.DefaultPixelPitch, !config.SkipInit), nil
}
