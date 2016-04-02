package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	shp "github.com/jonas-p/go-shp"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Usage: %s <parcel pin>", os.Args[0])
	}
	pin := flag.Arg(0)

	// open a shapefile for reading
	shapeReader, err := shp.Open("data/nc_burke_parcels_poly_2015_09_04.shp")
	if err != nil {
		log.Fatal(err)
	}
	defer shapeReader.Close()

	// loop through all features in the shapefile
	for shapeReader.Next() {
		//n, _ := shapeReader.Shape()
		n, shape := shapeReader.Shape()

		// pin is field 1 in burke county, 25 in caldwell
		v := strings.Trim(shapeReader.ReadAttribute(n, 25), "\x00")
		if v != pin {
			continue
		}

		// Print the edges of the bounding box
		fmt.Println(reflect.TypeOf(shape).Elem(), shape.BBox())
		// Print the assocaited record details of each shape
		for k, f := range shapeReader.Fields() {
			val := strings.Trim(shapeReader.ReadAttribute(n, k), "\x00")
			fmt.Printf("\t%v: %q\n", f, val)
		}
		// Print the midpoint of each shape
		box := shape.BBox()
		fmt.Printf("Midpoint: %f, %f\n", (box.MaxX+box.MinX)/2, (box.MaxY+box.MinY)/2)

		// Print out each point in the polygon's shape
		//switch shape := shape.(type) {
		//default:
		//	fmt.Printf("unsupported shape type: %T\n", shape)
		//case *shp.Polygon:
		//	for _, p := range shape.Points {
		//		fmt.Printf("\t%+v\n", p)
		//	}
		//}
	}
}
