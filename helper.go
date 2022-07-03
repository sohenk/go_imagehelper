package go_imagehelper

import (
	"bytes"
	"crypto/tls"
	"errors"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"

	"github.com/nfnt/resize"
)

func GetImgFromUrlToBytes(url string) (b []byte, filetype string, er error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	buf := &bytes.Buffer{}
	buf.ReadFrom(resp.Body)

	// retrieve a byte slice from bytes.Buffer
	data := buf.Bytes()
	_, filetype, err = image.Decode(bytes.NewBuffer(data))
	if err != nil {
		return nil, "", err
	}

	return data, filetype, nil
}

//Turn ioreader to image
func IoReaderToImage(reader io.Reader) (image.Image, string, error) {
	newimg, filetype, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}

	return newimg, filetype, nil
}

//resize img from bytes to bytes
func ResizeImgToByteFromBytes(img []byte, filetype string, width int64) ([]byte, string, error) {
	reader := bytes.NewReader(img)

	newimg, filetype, err := image.Decode(reader)
	if err != nil {
		return nil, "", errors.New("Picture Invalid")
	}
	thumbnailSize := int(width)
	var newImage image.Image
	if filetype != "gif" {
		newImage = resize.Resize(uint(thumbnailSize), 0, newimg, resize.Lanczos3)
	}
	w := new(bytes.Buffer)
	switch filetype {
	case "png":
		err = png.Encode(w, newImage)
	case "gif":
		gifs, err := ResizeGifToGifs(BytesToIoReader(img), thumbnailSize, 0)
		if err != nil {
			return nil, "", errors.New("Picture Invalid")
		}
		err = gif.EncodeAll(w, gifs)
	case "jpeg", "jpg":
		err = jpeg.Encode(w, newImage, nil)
	//case "bmp":
	//	err = bmp.Encode(w, newImage)
	//case "tiff":
	//	err = tiff.Encode(w, newImage, nil)
	default:
		// not sure how you got here but what are we going to do with you?
		// fmt.Println("Unknown image type: ", filetype)
		err = errors.New("Picture Invalid")
		//io.Copy(w, file)
	}
	if err != nil {

		return nil, "", err
	}
	// fmt.Println("filetype", filetype)
	return w.Bytes(), filetype, nil
}

// Resize the gif to another thumbnail gif
func ResizeGifToGifs(srcFile io.Reader, width int, height int) (*gif.GIF, error) {

	im, err := gif.DecodeAll(srcFile)

	if err != nil {
		return nil, err
	}

	if width == 0 {
		width = int(im.Config.Width * height / im.Config.Width)
	} else if height == 0 {
		height = int(width * im.Config.Height / im.Config.Width)
	}

	// reset the gif width and height
	im.Config.Width = width
	im.Config.Height = height

	firstFrame := im.Image[0].Bounds()
	img := image.NewRGBA(image.Rect(0, 0, firstFrame.Dx(), firstFrame.Dy()))

	// resize frame by frame
	for index, frame := range im.Image {
		b := frame.Bounds()
		draw.Draw(img, b, frame, b.Min, draw.Over)
		im.Image[index] = ImageToPaletted(resize.Resize(uint(width), uint(height), img, resize.NearestNeighbor))
	}
	//gif.Encode
	return im, nil
}

func ImageToPaletted(img image.Image) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, palette.Plan9)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}

//bytes to ioreader
func BytesToIoReader(img []byte) io.Reader {
	return bytes.NewReader(img)
}

//ioreadertoBytes
func IoReaderToBytes(reader io.Reader) ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.ReadFrom(reader)
	return buf.Bytes(), nil

}

func ByteToImage(imgbyte []byte) (image.Image, string, error) {
	reader := bytes.NewReader(imgbyte)
	newimg, filetype, err := image.Decode(reader)
	if err != nil {
		return nil, "", errors.New("Picture Invalid")
	}
	return newimg, filetype, nil
}
