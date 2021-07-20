package gps

import "math"

const (
	PI float64 = 3.14159265358979324
	X_PI = PI * 3000.0 / 180.0
)

type GPS interface {
	NewGPS(lat, long float64)
	ConvertToGCJ() (float64, float64)
	ConvertToBD() (float64, float64)
	ConvertToWGS() (float64, float64)
}

type gpsBase struct {
	lat float64
	long float64
}

type wgs struct {
	gpsBase
}


type gcj struct {
	gpsBase
}

type bd struct {
	gpsBase
}


func (g gpsBase) NewGPS(lat, long float64) {
	g.lat = lat
	g.long = long
}

func NewGPS(lat, long float64) gpsBase {
	return gpsBase{
		lat,
		long,
	}
}

func NewWGS(lat, long float64) GPS {
	return wgs{NewGPS(lat, long)}
}

func NewGCJ(lat, long float64) GPS {
	return gcj{NewGPS(lat, long)}
}

func NewBD(lat, long float64) GPS {
	return bd{NewGPS(lat, long)}
}


func transformLat(x, y float64) float64{
	ret := -100.0 + 2.0 * x + 3.0 * y + 0.2 * y * y + 0.1 * x * y + 0.2 * math.Sqrt(math.Abs(x))
	ret += (20.0 * math.Sin(6.0 * x *PI) + 20.0 * math.Sin(2.0 * x *PI)) * 2.0 / 3.0
	ret += (20.0 * math.Sin(y *PI) + 40.0 * math.Sin(y / 3.0 *PI)) * 2.0 / 3.0
	ret += (160.0 * math.Sin(y / 12.0 *PI) + 320 * math.Sin(y *PI/ 30.0)) * 2.0 / 3.0
	return ret
}

func transformLong(x, y float64) float64{
	ret := 300.0 + x + 2.0 * y + 0.1 * x * x + 0.1 * x * y + 0.1 * math.Sqrt(math.Abs(x))
	ret += (20.0 * math.Sin(6.0 * x *PI) + 20.0 * math.Sin(2.0 * x *PI)) * 2.0 / 3.0
	ret += (20.0 * math.Sin(x *PI) + 40.0 * math.Sin(x / 3.0 *PI)) * 2.0 / 3.0
	ret += (150.0 * math.Sin(x / 12.0 *PI) + 300 * math.Sin(x *PI/ 30.0)) * 2.0 / 3.0
	return ret
}

func (g gpsBase) outOfChina() bool {

	if g.long < 72.004 || g.long > 137.8347 {
		return true
	}

	if g.lat < 0.8293 || g.lat > 55.8271 {
		return true
	}

	return false
}

func (g gpsBase) delta() (float64, float64) {
	a := 6378245.0 //  a: 卫星椭球坐标投影到平面地图坐标系的投影因子。
	ee := 0.00669342162296594323 //  ee: 椭球的偏心率。
	dLat := transformLat(g.long - 105.0, g.lat - 35.0)
	dLong := transformLong(g.long - 105.0, g.lat - 35.0)
	radLat := g.lat / 180 * PI
	magic := math.Sin(radLat)
	magic = 1 - ee * magic * magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180) / (( a * (1 - ee)) / (magic * sqrtMagic) * PI)
	dLong = (dLong * 180.0) / (a / sqrtMagic * math.Cos(radLat) * PI)

	return dLat, dLong
}

func (wgs wgs) ConvertToGCJ() (float64, float64){
	if wgs.outOfChina() {
		return wgs.lat, wgs.long
	}
	dLat, dLong := wgs.delta()
	return wgs.lat + dLat, wgs.long + dLong
}

func (wgs wgs) ConvertToBD() (float64, float64){
	gcjLat, gcjLong := wgs.ConvertToGCJ()
	gcj := NewGCJ(gcjLat, gcjLong)
	return gcj.ConvertToBD()
}

func (wgs wgs) ConvertToWGS() (float64, float64){
	return wgs.lat, wgs.long
}

func (gcj gcj) ConvertToWGS() (float64, float64){
	if gcj.outOfChina() {
		return gcj.lat, gcj.long
	}
	dLat, dLong := gcj.delta()
	return gcj.lat - dLat, gcj.long - dLong
}

func (gcj gcj) ConvertToWGSExact() (float64, float64){
	initdelta := 0.01
	threshold := 0.000000001
	dLat := initdelta
	dLong := initdelta
	mLat := gcj.lat - dLat
	mLong := gcj.long - dLong
	pLat := gcj.lat + dLat
	pLong := gcj.long + dLong
	wgsLat := 0.0
	wgsLong := 0.0
	i := 0
	for true {
		wgsLat = (mLat + pLat) / 2.0
		wgsLong = (mLong + pLong) / 2.0
		wgs := NewWGS(wgsLat, wgsLong)
		lat, long := wgs.ConvertToGCJ()
		dLat = lat - gcj.lat
		dLong = long - gcj.long

		if math.Abs(dLat) < threshold && math.Abs(dLong) < threshold {
			break
		}

		if dLat > 0 {
			pLat = wgsLat
		}else{
			mLat = wgsLat
		}

		if dLong > 0 {
			pLong = wgsLong
		}else{
			mLong = wgsLong
		}

		i++

		if i > 10000 {
			break
		}
	}

	return wgsLat, wgsLong
}

func (gcj gcj) ConvertToBD() (float64, float64){
	x := gcj.long
	y := gcj.lat
	z := math.Sqrt(x * x + y * y) + 0.00002 * math.Sin(y *X_PI)
	theta := math.Atan2(y, x) + 0.000003 * math.Cos(x *X_PI)
	bdLon := z * math.Cos(theta) + 0.0065
	bdLat := z * math.Sin(theta) + 0.006
	return bdLat, bdLon
}

func (gcj gcj) ConvertToGCJ() (float64, float64){
	return gcj.lat, gcj.long
}

func (bd bd) ConvertToGCJ() (float64, float64){
	x := bd.long - 0.0065
	y := bd.lat - 0.006
	z := math.Sqrt(x * x + y * y) - 0.00002 * math.Sin(y *X_PI)
	theta := math.Atan2(y, x) - 0.000003 * math.Cos(x *X_PI)
	gcjLong := z * math.Cos(theta)
	gcjLat := z * math.Sin(theta)
	return gcjLat, gcjLong
}

func (bd bd) ConvertToWGS() (float64, float64){
	gcjLat, gcjLong := bd.ConvertToGCJ()
	gcj := NewGCJ(gcjLat, gcjLong)
	return gcj.ConvertToWGS()
}

func (bd bd) ConvertToBD() (float64, float64){
	return bd.lat, bd.long
}
