package main

import "fmt"
import "os"

func main() {
	argv := os.Args[1:]
	argc := len(argv)
	
	var operation string;
	
	if argc > 0 {
		operation = argv[0]
	}	
	
	switch operation {
		case "embed":
			if argc != 4 {
				displayHelp()
				return
			}
			
			embedInImage(argv[1], argv[2], argv[3])
		case "extract":
			if argc != 3 {
				displayHelp()
				return
			}
			
			extractFromImage(argv[1], argv[2])
		default:
			displayHelp()
	}
}

func displayHelp() {
	fmt.Println("Usages:")
	fmt.Println("  steganography embed input_png  input_payload  output_png")
	fmt.Println("  steganography extract input_png  output_file")
	fmt.Println("")
}
