package provider

import (
	"math"

	"github.com/vatsimnerd/geoidx"
)

const (
	airportSizeNM = 3.0
	planeSizeNM   = 0.005
)

func nmToLatLon(latSizeNM float64, lngSizeNM float64, atLatitude float64) (lng float64, lat float64) {
	// for latitude 60nm is 1˚
	lat = (1.0 / 60) * latSizeNM

	// for longitude it depends on current latitude
	// first let's convert latitude degree to radians
	latitudeRad := (atLatitude * 2 * math.Pi) / 360

	// calculate size
	lng = (1.0 / 60) * lngSizeNM
	// and make a latitude correction
	lng = lng / math.Abs(math.Cos(latitudeRad))
	return
}

// square makes square bounds of a given size
// Lat and Lng represent top left angle of the square
func square(lat float64, lng float64, sizeNM float64) geoidx.Rect {
	lngSize, latSize := nmToLatLon(sizeNM, sizeNM, lat)
	return geoidx.MakeRect(lng, lat, lng+lngSize, lat+latSize)
}

// squareCentered makes square bounds of a given size
// Lat and Lng represent center of the square
func squareCentered(lat float64, lng float64, sizeNM float64) geoidx.Rect {
	lngSize, latSize := nmToLatLon(sizeNM/2, sizeNM/2, lat)
	lat = lat - latSize
	lng = lng - lngSize
	return square(lat, lng, sizeNM)
}
