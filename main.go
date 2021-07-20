package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"gpsConvert/gps"
)


func main() {
	f := flag.String("f", "", "input file")
	i := flag.String("i", "", "wgs/bd/gcj")
	o := flag.String("o", "", "wgs/bd/gcj")
	flag.Parse()
	if *f == "" {
		log.Fatal("no specific input file")
		return
	}
	if *i == *o {
		log.Fatal("input gps type equal to output gps type, nothing to do")
		return
	}
	file, err := os.Open(*f)
	if err != nil {
		log.Fatal("could not open file, please check file path!")
		return
	}
	outDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	outFile := outDir + "/" + "gpsconvert.txt"
	os.Remove(outFile)
	output, err := os.OpenFile(outFile, os.O_CREATE|os.O_APPEND,0644)
	if err != nil {
		fmt.Println("OpenFile err")
		log.Fatal(err)
	}
	defer output.Close()
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, " ", "", -1)
		words := strings.Split(line, ",")
		longf, err := strconv.ParseFloat(words[0], 64)
		if err != nil {
			log.Fatal("parse gps longitude fail, please check file content!")
			return
		}
		latf, err := strconv.ParseFloat(words[1], 64)
		if err != nil {
			log.Fatal("parse gps latitude fail, please check file content!")
			return
		}
		var olat, olong float64
		var g gps.GPS
		switch *i {
		case "wgs":
			g = gps.NewWGS(latf, longf)
		case "gcj":
			g = gps.NewGCJ(latf, longf)
		case "bd":
			g = gps.NewBD(latf, longf)
		}
		switch *o {
		case "wgs":
			olat, olong = g.ConvertToWGS()
		case "gcj":
			olat, olong = g.ConvertToGCJ()
		case "bd":
			olat, olong = g.ConvertToBD()
		}
		outline := fmt.Sprintf("%f,%f\r\n", olong, olat)
		output.WriteString(outline)
	}
}
