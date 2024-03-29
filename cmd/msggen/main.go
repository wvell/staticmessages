package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wvell/messages"
)

func main() {
	var pkg, src, target string

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v", err)
		os.Exit(1)
	}

	flag.StringVar(&pkg, "pkg", "", "Package name for the generated code.")
	flag.StringVar(&src, "src", cwd, "Location where the .yml files are stored (only .yml files are parsed).")
	flag.StringVar(&target, "target", cwd, "Location where the go translation files should be written.")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage of msggen:\n\n")
		fmt.Fprint(os.Stderr, `msggen generates translation files based on .yml files.

To generate go translation files in the current working directory:
	# Inside myproject/translations
	$ msggen -pkg translations

Note: Files are never automaticly removed, use a scritp to remove old translation files before generating new ones.

`)
		fmt.Fprintf(os.Stderr, "Options:\n")

		flag.PrintDefaults()
	}
	flag.Parse()

	if pkg == "" {
		fmt.Fprintln(os.Stderr, "Package name is required.")
		os.Exit(1)
	}

	// Read all .yml files from src.
	files, err := os.ReadDir(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yml") {
			filename := filepath.Join(src, file.Name())
			f, err := os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
				os.Exit(1)
			}

			strippedFilename := strings.TrimSuffix(file.Name(), ".yml")

			// Parse yml file.
			parsed, err := messages.Parse(strippedFilename, f)
			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filename, err)
				os.Exit(1)
			}

			targetFile := filepath.Join(target, strippedFilename+".go")

			// Generate go code.
			f, err = os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening target file %s: %v\n", targetFile, err)
				os.Exit(1)
			}

			err = messages.Write(parsed, pkg, f)
			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to file %s: %v\n", targetFile, err)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stdout, "Generated %s\n", targetFile)
		}
	}
}
