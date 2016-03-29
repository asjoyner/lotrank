package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

	shp "github.com/jonas-p/go-shp"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Usage: %s <parcel pin>", os.Args[0])
	}
	pin := flag.Arg(0)

	// open a shapefile for reading
	shapeReader, err := shp.Open("data/parcel_taxdata.shp")
	if err != nil {
		log.Fatal(err)
	}
	defer shapeReader.Close()

	// loop through all features in the shapefile
	for shapeReader.Next() {
		//n, _ := shapeReader.Shape()
		n, shape := shapeReader.Shape()

		v := shapeReader.ReadAttribute(n, 1)
		if v != pin {
			continue
		}
		fmt.Println(reflect.TypeOf(shape).Elem(), shape.BBox())
		switch shape := shape.(type) {
		default:
			fmt.Printf("unsupported shape type: %T\n", shape)
		case *shp.Polygon:
			for _, p := range shape.Points {
				fmt.Printf("\t%+v\n", p)
			}
		}
	}
}
