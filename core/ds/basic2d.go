package ds

import "fmt"

type GeoLocation interface {
	Lat() float64
	Long() float64
	DistanceTo(location GeoLocation) float64
	IntersectsRectangle(Rectangle, int) bool
}

type Position struct {
	Latitude, Longitude float64
}

func NewPosition(lat float64, long float64) *Position {
	return &Position{Latitude: lat, Longitude: long}
}

func (p Position) IntersectsRectangle(r Rectangle, radiusInMetres int) bool {
	nearestCorner := r.GetNearestCorner(p)
	if DistanceOnEarth(p, nearestCorner) < float64(radiusInMetres) {
		return true
	}
	return false
}

func (p Position) DistanceTo(location GeoLocation) float64 {
	return DistanceOnEarth(p, location)
}

func (p Position) Lat() float64 {
	return p.Latitude
}

func (p Position) Long() float64 {
	return p.Longitude
}

//func NewPosition(Latitude, Longitude float64) Position {
//	return Position{Latitude, Longitude}
//}

type Rectangle interface {
	Corner1() GeoLocation
	Corner2() GeoLocation
	GetAllCorners() [4]GeoLocation
	GetNearestCorner(location GeoLocation) GeoLocation
	GetQuadrants() [4]Rectangle
}

type rectangle struct {
	corner1, corner2 GeoLocation
}

func (r rectangle) Corner1() GeoLocation {
	return r.corner1
}

func (r rectangle) Corner2() GeoLocation {
	return r.corner2
}

func NewRectangle(corner1, corner2 GeoLocation) Rectangle {
	return rectangle{corner1, corner2}
}

func (r rectangle) GetAllCorners() [4]GeoLocation {
	corner1, corner2 := r.corner1, r.corner2
	corner3 := NewPosition(corner1.Lat(), corner2.Long())
	corner4 := NewPosition(corner2.Lat(), corner1.Long())
	return [4]GeoLocation{corner1, corner2, corner3, corner4}
}

func (r rectangle) GetNearestCorner(location GeoLocation) GeoLocation {
	corners := r.GetAllCorners()
	nearestDist := 99999999999999999999.9
	var nearestCorner GeoLocation
	for _, corner := range corners {
		dist := EuclideanDistance(location, corner)
		if dist < nearestDist {
			nearestDist = dist
			nearestCorner = corner
		}
	}
	return nearestCorner
}

func (r rectangle) String() string {
	s := fmt.Sprintf("{%f,%f}", r.Corner1().Lat(), r.Corner1().Long())
	s += ", " + fmt.Sprintf("{%f,%f}", r.Corner2().Lat(), r.Corner2().Long())
	return s
}

func getMidPoint(p1, p2 GeoLocation) GeoLocation {
	midLat := (p1.Lat() + p2.Lat()) / 2
	midLong := (p1.Long() + p2.Long()) / 2
	return NewPosition(midLat, midLong)
}

func (r rectangle) GetQuadrants() [4]Rectangle {
	mid := getMidPoint(r.corner1, r.corner2)
	long1, lat1, long2, lat2 := r.corner1.Long(), r.corner1.Lat(), r.corner2.Long(), r.corner2.Lat()
	quad1 := NewRectangle(NewPosition(lat1, long1), mid)
	quad2 := NewRectangle(NewPosition(lat2, long1), mid)
	quad3 := NewRectangle(NewPosition(lat1, long2), mid)
	quad4 := NewRectangle(NewPosition(lat2, long2), mid)
	return [4]Rectangle{quad1, quad2, quad3, quad4}
}
