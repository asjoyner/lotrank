package main

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)
import "log"

import shp "github.com/jonas-p/go-shp"

var ACRES string = "GISACRES"
var NCPIN string = "PARNO"
var OWNER string = "OWNNAME"
var LOCATIONS = []Location{
	{1202328.750000, 735207.628255, 100, "MDS"},
	{1245792.937475, 790258.814687, 100, "Google"},
}
var NOTFORSALE = []string{
	"STATE OF NORTH CAROLINA",
	"FOOTHILLS REGIONAL AIRPORT",
	"CITY OF MORGANTON",
	"STATE OF N C",
	"STATE HOSPITAL",
}

type Location struct {
	x      float64
	y      float64
	weight float64
	name   string
}

type ShapeRecord struct {
	n     int
	acres float64
	shape shp.Shape
}

type Parcel struct {
	pin       string
	acres     float64
	owner     string
	score     float64
	shapes    []ShapeRecord
	distances []float64
}

func (p *Parcel) String() string {
	return fmt.Sprintf("%6.2f  PIN: %v  Acres: %7.2f  Owner: %v", p.score, p.pin, p.acres, p.owner)
}

// ByScore implements sort.Interface for []Parcel based on
// the Age field.
type ByScore []*Parcel

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].score < a[j].score }

func printSummary(shapeReader *shp.Reader, n int) {
	// print attributes
	name := map[int]string{
		25: "PIN",
		4:  "Acres",
		24: "Owner",
	}
	toPrint := []int{25, 4, 24}
	for _, attrIndex := range toPrint {
		val := shapeReader.ReadAttribute(n, attrIndex)
		fmt.Printf("%v: %v\t", name[attrIndex], val)
	}
	fmt.Println()
}

func printFields(shapeReader *shp.Reader, n int, fields []shp.Field) {
	for i, f := range fields {
		val := shapeReader.ReadAttribute(n, i)
		fmt.Printf("\t%v: %v\n", f, val)
	}
}

func fieldNum(name string, fields []shp.Field) (int, error) {
	for i, f := range fields {
		fs := string(bytes.TrimRight(f.Name[:], "\x00"))
		if fs == name {
			return i, nil
		}
	}
	return 0, fmt.Errorf("%x is not a field in the shapefile.", name)
}

func unworkable(p *Parcel) bool {
	for _, s := range NOTFORSALE {
		if strings.Contains(p.owner, s) {
			return true
		}
	}
	for _, d := range p.distances {
		if d > 20 {
			return true
		}
	}
	return false
}

func main() {
	// open a shapefile for reading
	//shapeReader, err := shp.Open("data/parcel_taxdata.shp")
	shapeReader, err := shp.Open("data/nc_burke_parcels_poly_2015_09_04.shp")
	if err != nil {
		log.Fatal(err)
	}
	defer shapeReader.Close()

	// fields from the attribute table (DBF)
	fields := shapeReader.Fields()

	// Look up the field numbers of a few key fields
	acreField, err := fieldNum(ACRES, fields)
	if err != nil {
		log.Fatalf("%x is not a field in the shapefile.", ACRES)
	}
	pinField, err := fieldNum(NCPIN, fields)
	if err != nil {
		log.Fatalf("%x is not a field in the shapefile.", ACRES)
	}
	ownerField, err := fieldNum(OWNER, fields)
	if err != nil {
		log.Fatalf("%x is not a field in the shapefile.", ACRES)
	}

	// loop through all polys in the shapefile, group by pin, calculate acreage
	parcels := make(map[string]*Parcel, 0)
	for shapeReader.Next() {
		n, shape := shapeReader.Shape()
		pin := strings.Trim(shapeReader.ReadAttribute(n, pinField), "\x00")

		v := strings.Trim(shapeReader.ReadAttribute(n, acreField), "\x00")
		acres, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Fatalf("%s invalid: %s", v, err)
		}

		p, ok := parcels[pin]
		if !ok {
			p = &Parcel{pin: pin}
		}
		p.shapes = append(p.shapes, ShapeRecord{n, acres, shape})
		p.acres += acres
		parcels[pin] = p
	}

	// filter parcels whose combined acerage is too small
	var matches []*Parcel
	for _, parcel := range parcels {
		if parcel.acres <= 30 {
			continue
		}

		matches = append(matches, parcel)

	}

	// Score the choices
	for _, parcel := range matches {
		parcel.score = 1000
		// print aproximate location of feature
		//fmt.Println(reflect.TypeOf(shape).Elem(), shape.BBox())

		var xtot, ytot float64
		for _, sr := range parcel.shapes {
			box := sr.shape.BBox()
			xtot += (box.MaxX + box.MinX) / 2
			ytot += (box.MaxY + box.MinY) / 2
		}
		xMid := xtot / float64(len(parcel.shapes))
		yMid := ytot / float64(len(parcel.shapes))
		//fmt.Printf("Midpoint: %f, %f\n", xMid, yMid)

		for _, l := range LOCATIONS {
			a := math.Abs(xMid - l.x)
			b := math.Abs(yMid - l.y)
			distance := math.Sqrt(math.Pow(a, 2)+math.Pow(b, 2)) / 5280
			parcel.distances = append(parcel.distances, distance)
			parcel.score -= distance * l.weight
		}

	}

	sort.Sort(ByScore(matches))
	for _, parcel := range matches {
		parcel.owner = strings.Trim(shapeReader.ReadAttribute(parcel.shapes[0].n, ownerField), "\x00")

		// Skip some parcels from the result set
		if unworkable(parcel) {
			continue
		}

		fmt.Println(parcel)
		//for i, d := range parcel.distances {
		//	fmt.Printf("\tDistance to %s: %f\n", LOCATIONS[i].name, d)
		//}
		//printSummary(shapeReader, srs[0].n)
		//printFields(shapeReader, srs[0].n, fields)
		//for sr := range srs {
		//	printFields(shapeReader, sr.n, [])
		//}
	}

}
