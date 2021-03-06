package main

import (
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	dir = flag.String("dir", ".", "Path to directory with template files to compile. "+
		"Only files with ext extension are compiled. See ext flag for details.\n"+
		"The compiler recursively processes all the subdirectories.\n"+
		"Compiled template files are placed near the original file with .go extension added.")

	ext = flag.String("ext", "qtpl", "Only files with this extension are compiled")
)

var logger = log.New(os.Stderr, "qtc: ", log.LstdFlags)

var filesCompiled int

func main() {
	flag.Parse()

	if len(*ext) == 0 {
		logger.Fatalf("ext cannot be empty")
	}
	if len(*dir) == 0 {
		*dir = "."
	}
	if (*ext)[0] != '.' {
		*ext = "." + *ext
	}

	logger.Printf("Compiling *%s template files in directory %q", *ext, *dir)
	compileDir(*dir)
	logger.Printf("Total files compiled: %d", filesCompiled)
}

func compileDir(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		logger.Fatalf("cannot compile files in %q: %s", path, err)
	}
	if !fi.IsDir() {
		logger.Fatalf("cannot compile files in %q: it is not directory", path)
	}
	d, err := os.Open(path)
	if err != nil {
		logger.Fatalf("cannot compile files in %q: %s", path, err)
	}
	defer d.Close()

	fis, err := d.Readdir(-1)
	if err != nil {
		logger.Fatalf("cannot read files in %q: %s", path, err)
	}

	var names []string
	for _, fi = range fis {
		name := fi.Name()
		if name == "." || name == ".." {
			continue
		}
		if !fi.IsDir() {
			names = append(names, name)
		} else {
			subPath := filepath.Join(path, name)
			compileDir(subPath)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		if strings.HasSuffix(name, *ext) {
			filename := filepath.Join(path, name)
			compileFile(filename)
		}
	}
}

func compileFile(infile string) {
	outfile := infile + ".go"
	logger.Printf("Compiling %q to %q...", infile, outfile)
	inf, err := os.Open(infile)
	if err != nil {
		logger.Fatalf("cannot open file %q: %s", infile, err)
	}

	tmpfile := outfile + ".tmp"
	outf, err := os.Create(tmpfile)
	if err != nil {
		logger.Fatalf("cannot create file %q: %s", tmpfile, err)
	}

	packageName, err := getPackageName(infile)
	if err != nil {
		logger.Fatalf("cannot determine package name for %q: %s", infile, err)
	}
	if err = parse(outf, inf, infile, packageName); err != nil {
		logger.Fatalf("error when parsing file %q: %s", infile, err)
	}
	if err = outf.Close(); err != nil {
		logger.Fatalf("error when closing file %q: %s", tmpfile, err)
	}
	if err = inf.Close(); err != nil {
		logger.Fatalf("error when closing file %q: %s", infile, err)
	}

	// prettify the output file
	uglyCode, err := ioutil.ReadFile(tmpfile)
	if err != nil {
		logger.Fatalf("cannot read file %q: %s", tmpfile, err)
	}
	prettyCode, err := format.Source(uglyCode)
	if err != nil {
		logger.Fatalf("error when formatting compiled code for %q: %s. See %q for details", infile, err, tmpfile)
	}
	if err = ioutil.WriteFile(outfile, prettyCode, 0666); err != nil {
		logger.Fatalf("error when writing file %q: %s", outfile, err)
	}
	if err = os.Remove(tmpfile); err != nil {
		logger.Fatalf("error when removing file %q: %s", tmpfile, err)
	}

	filesCompiled++
}
