package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getLastQrCodeArrayFromShell(path string) []string {
	dat, err := ioutil.ReadFile(path)
	check(err)

	lines := strings.Split(string(dat), "\n")
	qrStartLine := -1
	qrEndLine := -1

	// Look for the QR code indicators.
	for i, line := range lines {
		if strings.Contains(line, "===BEGIN QR CODE===") {
			qrStartLine = i
		}
		if strings.Contains(line, "===END QR CODE===") {
			qrEndLine = i
		}
	}

	// Either the start or end of a QR code could not be read.
	if qrStartLine == -1 || qrEndLine == -1 {
		return nil
	}

	// The last QR code start occurs after the last QR code end.
	if qrStartLine > qrEndLine {
		return nil
	}

	// return a QR code
	return lines[qrStartLine+1 : qrEndLine]
}

func stringArrayEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func waitForFileChange(filePath string) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func generateQrCodeArrays(path string) <-chan []string {
	// Make an output channel to send the QR codes to.
	out := make(chan []string)

	var lastQrCodeArray []string

	go func() {
		for {
			// Wait until shell.txt is written to.
			waitForFileChange(path)

			// Get the last QR code array from shell.txt.
			qrCodeArray := getLastQrCodeArrayFromShell(path)

			// If no QR array is returned, stop processing.
			if qrCodeArray == nil {
				continue
			}

			// Make sure this QR code array was not the last QR code array to
			// be processed.
			if stringArrayEquals(qrCodeArray, lastQrCodeArray) {
				continue
			}

			// Save this QR code array as the last QR code array.
			lastQrCodeArray = qrCodeArray

			// Send a QR code array to the output channel.
			out <- qrCodeArray
		}
	}()

	return out
}

func translateQrCodeArrayToBlackWhiteChars(qrs <-chan []string) <-chan [][]string {
	out := make(chan [][]string)

	go func() {
		for qr := range qrs {
			// Prepare a 2D byte array.
			qrByteArray := make([][]string, 0)

			// Get the top left character. This will be black.
			bChar := qr[0][0]

			// Get the character at 1,1. This will be white.
			wChar := qr[1][1]

			for _, row := range qr {
				// Prepare a byte array for this row.
				rowByteArray := make([]string, len(row))

				// Iterate over the cells of the row and set them either B, W, or ?
				for i := 0; i < len(row); i++ {
					cell := row[i]
					if cell == bChar {
						rowByteArray[i] = "B"
					} else if cell == wChar {
						rowByteArray[i] = "W"
					} else {
						rowByteArray[i] = "?"
					}
				}

				// Add the new row to the new QR array.
				qrByteArray = append(qrByteArray, rowByteArray)
			}

			// Send the new array to the output channel.
			out <- qrByteArray
		}
	}()

	return out
}

func main() {
	path := flag.String("path", "shell.txt", "path to the shell.txt file")

	// Start the pipeline by generating a channel of QR code arrays.
	qrCodes := generateQrCodeArrays(*path)

	// Translate the characters of the array to "B" and "W".
	bwChars := translateQrCodeArrayToBlackWhiteChars(qrCodes)

	for qca := range bwChars {
		for _, row := range qca {
			rowStr := ""
			for _, cell := range row {
				rowStr = rowStr + cell
			}
			log.Println(rowStr)
		}
	}

	// TODO: take the qr code and translate the array to an image.Image with the 1's being black and the 0's being white

	// TODO: parse the image with https://github.com/kdar/goquirc

	// TODO: put the result into the clipboard

	// TODO: determine a command to send to the game

	// TODO: send commands to a game
	//       - windows: https://play.golang.org/p/kwfYDhhiqk
	//       - mac: github.com/everdev/mack

}
