package main

import (
	"bytes"
	"fmt"
	"strconv"
)
import "log"

import shp "github.com/jonas-p/go-shp"

var ACRES string = "ACRE"

func printSummary(shapeReader *shp.Reader, n int) {
	// print attributes
	name := map[int]string{
		1:  "PIN",
		21: "Acres",
		10: "Owner",
	}
	toPrint := []int{1, 21, 10}
	//toPrint := []int{1}
	for _, attrIndex := range toPrint {
		val := shapeReader.ReadAttribute(n, attrIndex)
		fmt.Printf("%v: %v\t", name[attrIndex], val)
	}
}

func printAllFields(shapeReader *shp.Reader, n int, fields []shp.Field) {
	for k, f := range fields {
		val := shapeReader.ReadAttribute(n, k)
		fmt.Printf("\t%v: %v\n", f, val)
	}
}

func main() {
	// open a shapefile for reading
	shapeReader, err := shp.Open("data/parcel_taxdata.shp")
	if err != nil {
		log.Fatal(err)
	}
	defer shapeReader.Close()

	// fields from the attribute table (DBF)
	fields := shapeReader.Fields()

	var found bool
	for _, f := range fields {
		fs := string(bytes.TrimRight(f.Name[:], "\x00"))
		if fs == ACRES {
			found = true
		}
	}
	if !found {
		log.Fatalf("%x is not a field in the shapefile.", ACRES)
	}

	// loop through all features in the shapefile
	matches := make(map[string]int, 0)
	for shapeReader.Next() {
		n, _ := shapeReader.Shape()
		//n, shape := shapeReader.Shape()

		v := shapeReader.ReadAttribute(n, 3)
		vf, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Fatalf("%s invalid: %s", v, err)
		}
		if vf <= 10000 {
			continue
		}

		pin := shapeReader.ReadAttribute(n, 1)
		matches[pin] = n

		if len(matches) >= 5 {
			break
		}
	}

	for _, n := range matches {
		// print feature
		//fmt.Println(reflect.TypeOf(shape).Elem(), shape.BBox())

		printSummary(shapeReader, n)
		printAllFields(shapeReader, n, fields)
		fmt.Println()

	}

}
