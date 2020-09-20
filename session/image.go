package session

import (
	"bytes"
	"github.com/disintegration/imaging"
)

var formatExts = map[string]imaging.Format{
	"jpg":  imaging.JPEG,
	"jpeg": imaging.JPEG,
	"png":  imaging.PNG,
	"gif":  imaging.GIF,
	"tif":  imaging.TIFF,
	"tiff": imaging.TIFF,
	"bmp":  imaging.BMP,
}

var anchorExts = map[string]imaging.Anchor{
	"center":      imaging.Center,
	"topleft":     imaging.TopLeft,
	"top":         imaging.Top,
	"topright":    imaging.TopRight,
	"left":        imaging.Left,
	"right":       imaging.Right,
	"bottomleft":  imaging.BottomLeft,
	"bottom":      imaging.Bottom,
	"bottomRight": imaging.BottomRight,
}

func Blur(bucketName, objectName string, sigma float64, suffix string) ([]byte, error) {
	data, err := GetObject(bucketName, objectName)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)
	initialImg, err := imaging.Decode(buffer, imaging.AutoOrientation(true))
	processedImg := imaging.Blur(initialImg, sigma)
	err = imaging.Encode(buffer, processedImg, formatExts[suffix])
	if err != nil {
		return nil, err
	}
	processedData := buffer.Bytes()
	return processedData, nil
}

func Resize(bucketName, objectName string, width, height int, suffix string) ([]byte, error) {
	data, err := GetObject(bucketName, objectName)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)
	initialImg, err := imaging.Decode(buffer, imaging.AutoOrientation(true))
	processedImg := imaging.Resize(initialImg, width, height, imaging.Lanczos)
	err = imaging.Encode(buffer, processedImg, formatExts[suffix])
	if err != nil {
		return nil, err
	}
	processedData := buffer.Bytes()
	return processedData, nil
}

func CropAnchor(bucketName, objectName string, width, height int, anchor string, suffix string) ([]byte, error) {
	data, err := GetObject(bucketName, objectName)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)
	initialImg, err := imaging.Decode(buffer, imaging.AutoOrientation(true))
	processedImg := imaging.CropAnchor(initialImg, width, height, anchorExts[anchor])
	err = imaging.Encode(buffer, processedImg, formatExts[suffix])
	if err != nil {
		return nil, err
	}
	processedData := buffer.Bytes()
	return processedData, nil
}
