package main

import (
	"image/draw"
	"fmt"
	"image"
	"image/png"
	"image/color"
	"os"
)

// TODO: Further documentation
// TODO: Sort out the uint/int/int32/uint32 types
// TODO: Fix support for non 32-bit PNG images

func embedInImage(imagePath string, payloadPath string, outputPath string) {
	fmt.Println(fmt.Sprintf("Attempting to embed payload \"%v\" into \"%v\", ouputting to \"%v\"", payloadPath, imagePath, outputPath))

	// Open files
	img, payload, outImg := openFilesForEmbedding(imagePath, payloadPath, outputPath)

	// Set position pointer and the last pixel in the source image
	pos := 0
	endPos := img.Bounds().Max.X * img.Bounds().Max.Y

	// Retrieve payload size
	fi, err := payload.Stat()
	if err != nil {
        fmt.Println("Error: Could not obtain payload length")
        fmt.Println(err)
        os.Exit(1)
	}

	payloadSize := fi.Size()

	// Embed payload size in image as (u?)int32
	for ; pos < 4; pos++ {
		embedByteInImage(img, outImg, byte(payloadSize >> pos * 8), pos)
	}

	if pos + int(payloadSize) > endPos {
		fmt.Println("Warning: The payload size exceeds the image's size, some data may be lost")
	}

	// While there is space left in the image, and there is payload data left to be written
	for ; (pos + 4) < int(payloadSize) && pos < endPos; pos++ {
		// Create buffer and read a byte from the payload
		buff := make([]byte, 1)
    	_, err := payload.Read(buff)
		if err != nil {
			fmt.Println("Error: Couldn't read byte from payload")
			fmt.Println(err)
			os.Exit(1)
		}

		// embed byte
		embedByteInImage(img, outImg, buff[0], pos)
	}

	// Output as PNG
	f, _ := os.Create(outputPath)
	png.Encode(f, outImg)
}

func extractFromImage(inputPath string, outputPath string) {
	fmt.Println(fmt.Sprintf("Attempting to extract payload from \"%v\", ouputting to \"%v\"", inputPath, outputPath))
	
	// Open input and output files
	img, outFile := openFilesForExtraction(inputPath, outputPath)

	// Position tracker, total pixel count, payload length
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y	
	var length uint32 = 0;
	var pos int32 = 0;
	var endPos = int32(width * height)

	// Read the payload length from the first 4 bytes in the image
	for pos < 4  {
		length |= uint32(readByteFromImage(img, width, height, pos)) << (pos * 8)
		pos++
	}

	// Increase by four to take into account the length-bytes themselves
	length += 4;
	
	for pos < endPos && pos < int32(length) {
		outFile.Write([]byte{readByteFromImage(img, width, height, pos)})
		pos++
	}
}

func readByteFromImage(img image.Image, width int, height int, pos int32) byte {
	var b byte = 0
	color := img.At(int(pos) % width, int(pos) / width).(color.NRGBA)
	bytes := []byte{byte(color.A), byte(color.R), byte(color.G), byte(color.B)}
	
	for i := 0; i < 8; i++ {
		var mask byte = 0b10
		if i < 4 {
			mask = 0b01
		}

		setBitValue(&b, i, (bytes[i % 4] & mask) == mask)
	}
	
	return b
}

func embedByteInImage(img image.Image, outImg *image.NRGBA, inByte byte, pos int) {
	x, y := int(pos) % img.Bounds().Max.X, int(pos) / img.Bounds().Max.X
	col := img.At(x, y).(color.NRGBA)
	bytes := []byte{byte(col.A), byte(col.R), byte(col.G), byte(col.B)}

	// For each bit
	for i := 0; i < 8; i++ {
		// Set mask to check the specific bit (LE, RTL)
		var mask byte = 0b1 << i
		inBytePos := 0b1
		if i < 4 {
			inByte = 0b0
		}

		// Set the LSB (or 1st to LSB for the second set of 4 bits) to 1 if inByte matches the mask, 0 otherwise
		setBitValue(&bytes[i % 4], inBytePos, (inByte & mask) == mask);
	}

	// Write pixel to image
	newColor := color.NRGBA{bytes[1], bytes[2], bytes[3], bytes[0]}
	outImg.Set(x, y, newColor);	
}

func setBitValue(b *byte, position int, condition bool) {
	// Calculate the correct mask, e.g. pos = 0, mask = 0b1; pos = 4, mask = 0b10000
	var mask byte = 0b1 << position

	// Set bit value, if condition == true then bitwise OR the mask, otherwise bitwise AND the inverted mask
	if condition {
		*b |= mask
	} else {
		*b &= ^mask
	}
}

func openFilesForExtraction(inputPath string, outputPath string) (image.Image, *os.File) {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
    inputFile, err := os.Open(inputPath)

    if err != nil {
        fmt.Println("Error: File could not be opened")
		fmt.Println(err)
        os.Exit(1)
    }

    defer inputFile.Close()
	
	img, _, err := image.Decode(inputFile)

    if err != nil {
        fmt.Println(fmt.Sprintf("Error: %v", err))
		fmt.Println(err)
        os.Exit(1)
    }

	outFile, err := os.Create(outputPath)
    if err != nil {
        fmt.Println("Error: File could not be opened")
		fmt.Println(err)
        os.Exit(1)
    }
    defer outFile.Close()
	
	return img, outFile
}

func openFilesForEmbedding(inputPath string, payloadPath string, outputPath string) (image.Image, *os.File, *image.NRGBA) {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

    inputFile, err := os.Open(inputPath)
    if err != nil {
        fmt.Println("Error: File could not be opened")
        os.Exit(1)
    }

    defer inputFile.Close()
	
	img, _, err := image.Decode(inputFile)
    if err != nil {
        fmt.Println(fmt.Sprintf("Error: %v", err))
        os.Exit(1)
    }

	payload, err := os.Open(payloadPath)
    if err != nil {
        fmt.Println("Error: File could not be opened")
		fmt.Println(err)
        os.Exit(1)
    }

    //defer payload.Close()

    outImage := image.NewNRGBA(img.Bounds())
	draw.Draw(outImage, outImage.Bounds(), img, image.Point{0, 0}, draw.Src);
	
	
	return img, payload, outImage
}