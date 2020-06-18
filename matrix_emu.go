// +build !arm

package rgbmatrix

import "github.com/tehmaze/go-rpi-rgb-led-matrix/textemulator"

func newRGBLedMatrix(config *HardwareConfig) (Matrix, error) {
	return textemulator.New(config.geometry())
}
