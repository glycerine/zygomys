package zygo

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	//"4d63.com/tz"
)

const RFC3339Micro = "2006-01-02T15:04:05.999999"

// convenience function for specifying Endx far in the future.
// Initialized in dfstix.go init() function.
var TheFuture time.Time

// nanospeconds since unix epoch utc.
type Utc64 int64

var ZeroUtc = Utc64(0)
var ZeroTime = time.Time{}

func (u Utc64) ToTime() time.Time {
	return time.Unix(0, int64(u))
}

func U(tm time.Time) Utc64 {
	return Utc64(tm.UnixNano())
}

const NanosPerSec = 1e9
const ZeroGmtOffset = 0
const BstGmtOffset = 1

// Date represents a UTC time zone day
type Date struct {
	Year  int `zid:"0"`
	Month int `zid:"1"`
	Day   int `zid:"2"`
}

// ParseMMDDYYYY converts a datestring '02/25/2016' into a Date{} struct.
func ParseMMDDYYYY(datestring string) (*Date, error) {
	parts := strings.Split(datestring, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("bad datestring '%s': did not have two slashes", datestring)
	}
	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse year", datestring)
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse month", datestring)
	}
	day, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse day", datestring)
	}

	if year < 1970 || year > 3000 {
		return nil, fmt.Errorf("year out of bounds: %v", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month out of bounds: %v", month)
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day out of bounds: %v", day)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

// ParseDate converts a datestring '2016/02/25' into a Date{} struct.
func ParseDate(datestring string, sep string) (*Date, error) {
	if sep == "" {
		sep = "/"
	}
	parts := strings.Split(datestring, sep)
	if len(parts) != 3 {
		return nil, fmt.Errorf("bad datestring '%s': did not have two '%s' separators",
			datestring, sep)
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse year", datestring)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse month", datestring)
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse day", datestring)
	}

	if year < 1970 || year > 3000 {
		return nil, fmt.Errorf("year out of bounds: %v", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month out of bounds: %v", month)
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day out of bounds: %v", day)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

var WestCoastUSLocation *time.Location
var EastCoastUSLocation *time.Location
var LondonLocation *time.Location
var UTCLocation = time.UTC

func init() {
	var err error
	WestCoastUSLocation, err = time.LoadLocation("America/Los_Angeles")
	panicOn(err)
	EastCoastUSLocation, err = time.LoadLocation("America/New_York")
	panicOn(err)
	LondonLocation, err = time.LoadLocation("Europe/London")
	panicOn(err)
}

// UTCDateFromTime returns the date after tm is moved to the UTC time zone.
func UTCDateFromTime(tm time.Time) *Date {
	y, m, d := tm.In(time.UTC).Date()
	return &Date{Year: y, Month: int(m), Day: d}
}

// NYCDateFromTime returns the date after tm is moved to the NYC (Eastern) time zone.
func NYCDateFromTime(tm time.Time) *Date {
	y, m, d := tm.In(EastCoastUSLocation).Date()
	return &Date{Year: y, Month: int(m), Day: d}
}

// Unix converts the date into an int64 representing the nanoseconds
// since the unix epoch for the ToGoTime() output of Date d.
func (d *Date) Unix() int64 {
	return d.ToGoTime().Unix()
}

// ToGoTime turns the date into UTC time.Time, at the 0 hrs 0 min 0 second start of the day.
func (d *Date) ToGoTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// ToGoTimeNYC turns the date into NYC time.Time, at the 0 hrs 0 min 0 second start of the day.
func (d *Date) ToGoTimeNYC() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, NYC)
}

// String turns the date into a string.
func (d Date) String() string {
	return fmt.Sprintf("%04d/%02d/%02d", d.Year, d.Month, d.Day)
}

func (d Date) NoSlashString() string {
	return fmt.Sprintf("%04d%02d%02d", d.Year, d.Month, d.Day)
}

func (d Date) StringNoSlashYYMMDD() string {
	yr := d.Year
	if yr < 2000 {
		yr = yr - 1900
	} else {
		yr = yr - 2000
	}
	return fmt.Sprintf("%02d%02d%02d", yr, d.Month, d.Day)
}

func (d Date) StringNoSlashCCYYMMDD() string {
	return fmt.Sprintf("%04d%02d%02d", d.Year, d.Month, d.Day)
}

// return true if a < b
func DateBefore(a *Date, b *Date) bool {
	if a.Year < b.Year {
		return true
	} else if a.Year > b.Year {
		return false
	}

	if a.Month < b.Month {
		return true
	} else if a.Month > b.Month {
		return false
	}

	if a.Day < b.Day {
		return true
	} else if a.Day > b.Day {
		return false
	}

	return false
}

// return true if a > b
func DateAfter(a *Date, b *Date) bool {
	if a.Year > b.Year {
		return true
	} else if a.Year < b.Year {
		return false
	}

	if a.Month > b.Month {
		return true
	} else if a.Month < b.Month {
		return false
	}

	if a.Day > b.Day {
		return true
	} else if a.Day < b.Day {
		return false
	}

	return false
}

// DatesEqual returns true if a and b are the exact same day.
func DatesEqual(a *Date, b *Date) bool {
	if a.Year == b.Year {
		if a.Month == b.Month {
			if a.Day == b.Day {
				return true
			}
		}
	}
	return false
}

// NextDate returns the next calendar day after d.
// You probably want NextDateNYC() instead.
/*func NextDate(d *Date) *Date {
	tm := d.ToGoTime()
	next := tm.AddDate(0, 0, 1)
	return UTCDateFromTime(next)
}
*/

func NextDateNYC(d *Date) *Date {
	tm := d.ToGoTimeNYC()
	next := tm.AddDate(0, 0, 1)
	return NYCDateFromTime(next)
}

func (d *Date) NextDate() *Date {
	tm := d.ToGoTime()
	next := tm.AddDate(0, 0, 1)
	return UTCDateFromTime(next)
}

func (d *Date) AddDays(n int) *Date {
	tm := d.ToGoTime()
	next := tm.AddDate(0, 0, n)
	return UTCDateFromTime(next)
}

// PrevDate returns the first calendar day prior to d.
func PrevDate(d *Date) *Date {
	tm := d.ToGoTime()
	next := tm.AddDate(0, 0, -1)
	return UTCDateFromTime(next)
}

func (d *Date) PrevDate() *Date {
	return PrevDate(d)
}

// PrevBusinessDate returns the previous business day prior to d.
// Useful to skip back past weekends and holidays.
func PrevBusinessDate(d *Date) *Date {
	tm := d.ToGoTimeNYC()
	prevTm := tm.AddDate(0, 0, -1)
	date := NYCDateFromTime(prevTm)
	for !IsBusinessDay(date) {
		prevTm = prevTm.AddDate(0, 0, -1)
		date = NYCDateFromTime(prevTm)
	}
	return date
}

func (d *Date) PrevBusinessDate() *Date {
	return PrevBusinessDate(d)
}

// NextBusinessDate returns the previous business day prior to d.
// Useful to skip back past weekends and holidays.
func NextBusinessDate(d *Date) *Date {
	tm := d.ToGoTimeNYC()
	nextTm := tm.AddDate(0, 0, 1)
	date := NYCDateFromTime(nextTm)
	for !IsBusinessDay(date) {
		nextTm = nextTm.AddDate(0, 0, 1)
		date = NYCDateFromTime(nextTm)
	}
	return date
}

func (d *Date) NextBusinessDate() *Date {
	return NextBusinessDate(d)
}

// TimeToDate returns the UTC Date associated with tm.
func TimeToDate(tm time.Time) Date {
	utc := tm.UTC()
	return Date{
		Year:  utc.Year(),
		Month: int(utc.Month()),
		Day:   utc.Day(),
	}
}

func ParseIQFeedTimestamp(timestamp string, loc *time.Location) (tm time.Time, dt *Date, err error) {

	// timestamp example: "2011-10-25T22:07:14"
	// YYYY-MM-DD is the date format
	tms := strings.Split(timestamp, "T")
	if len(tms) != 2 {
		return time.Time{}, nil, fmt.Errorf("timestamp did not have a 'T' in it: '%s'", timestamp)
	}
	dt, err = ParseDate(tms[0], "-")
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("on ParseDate('%s'): '%v'", tms[0], err)
	}

	parts := strings.Split(tms[1], ":")
	if len(parts) != 3 {
		return time.Time{}, nil, fmt.Errorf("timestamp time portion did not have two ':' in it: '%s'", tms[1])
	}

	hr, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("timestamp time portion had bad hour: '%s', in '%s'. err: '%s'", tms[1], timestamp, err)
	}

	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("timestamp time portion had bad minutes: '%s', in '%s'. err: '%s'", tms[1], timestamp, err)
	}

	sec, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("timestamp time portion had bad seconds: '%s', in '%s'. err: '%s'", tms[1], timestamp, err)
	}

	gotime := time.Date(dt.Year, time.Month(dt.Month), dt.Day, hr, min, sec, 0, loc)
	return gotime, dt, nil
}

func AbsDaysBetween(a, b *Date) int {
	n := DaysBetweenAminusB(a, b)
	if n < 0 {
		return -n
	}
	return n
}

func DaysBetweenAminusB(a, b *Date) int {
	return int(int64(a.ToGoTime().Sub(b.ToGoTime())) / int64(time.Hour*24))
}

func MonthNumberToCapString(month int) string {
	switch month {
	case 1:
		return "JAN"
	case 2:
		return "FEB"
	case 3:
		return "MAR"
	case 4:
		return "APR"
	case 5:
		return "MAY"
	case 6:
		return "JUN"
	case 7:
		return "JUL"
	case 8:
		return "AUG"
	case 9:
		return "SEP"
	case 10:
		return "OCT"
	case 11:
		return "NOV"
	case 12:
		return "DEC"
	}
	panic(fmt.Sprintf("unrecognized month %v", month))
	return ""
}

func ParseYYMMDDNoSlash(datestring string) (*Date, error) {
	if len(datestring) != 6 {
		return nil, fmt.Errorf("bad datestring '%s': did not have 6 characters", datestring)
	}
	year, err := strconv.Atoi(datestring[:2])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse year", datestring)
	}
	// Warning: this only works for 1990 - 2090 :-)
	if year < 90 {
		year += 2000
	} else {
		year += 1900
	}
	month, err := strconv.Atoi(datestring[2:4])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse month", datestring)
	}
	day, err := strconv.Atoi(datestring[4:6])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse day", datestring)
	}

	if year < 1970 || year > 3000 {
		return nil, fmt.Errorf("year out of bounds: %v", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month out of bounds: %v", month)
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day out of bounds: %v", day)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

func MustParseYYYYMMDDNoSlash(datestring string) *Date {
	d, err := ParseYYYYMMDDNoSlash(datestring)
	panicOn(err)
	return d
}

func ParseYYYYMMDDNoSlash(datestring string) (*Date, error) {
	if len(datestring) != 8 {
		return nil, fmt.Errorf("bad datestring '%s': did not have 8 characters", datestring)
	}
	year, err := strconv.Atoi(datestring[:4])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse year", datestring)
	}
	month, err := strconv.Atoi(datestring[4:6])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse month", datestring)
	}
	day, err := strconv.Atoi(datestring[6:])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse day", datestring)
	}

	if year < 1970 || year > 3000 {
		return nil, fmt.Errorf("year out of bounds: %v", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month out of bounds: %v", month)
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day out of bounds: %v", day)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

func ParseYYDDWithMonthSupplied(datestring string, month int) (*Date, error) {
	if len(datestring) != 4 {
		return nil, fmt.Errorf("bad YYDD datestring '%s': did not have 4 characters", datestring)
	}
	year, err := strconv.Atoi(datestring[:2])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse year", datestring)
	}
	// Warning: this only works for 1990 - 2090 :-)
	if year < 90 {
		year += 2000
	} else {
		year += 1900
	}
	day, err := strconv.Atoi(datestring[2:4])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse day", datestring)
	}

	if year < 1970 || year > 3000 {
		return nil, fmt.Errorf("year out of bounds: %v", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month out of bounds: %v", month)
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day out of bounds: %v", day)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

/// from time.go

// returns true iff d occurs strictly before d2
func (d Date) Before(d2 Date) bool {
	switch {
	case d.Year < d2.Year:
		return true
	case d.Year > d2.Year:
		return false
	}
	switch {
	case d.Month < d2.Month:
		return true
	case d.Month > d2.Month:
		return false
	}
	switch {
	case d.Day < d2.Day:
		return true
	case d.Day > d2.Day:
		return false
	}

	// same day
	return false
}

// returns true iff d occurs strictly after d2
func (d Date) After(d2 Date) bool {
	switch {
	case d.Year > d2.Year:
		return true
	case d.Year < d2.Year:
		return false
	}
	switch {
	case d.Month > d2.Month:
		return true
	case d.Month < d2.Month:
		return false
	}
	switch {
	case d.Day > d2.Day:
		return true
	case d.Day < d2.Day:
		return false
	}

	// same day
	return false
}

// returns true iff d and d2 are the same day
func (d Date) Equals(d2 Date) bool {
	if d.Year == d2.Year && d.Month == d2.Month && d.Day == d2.Day {
		return true
	}
	return false
}

// To3339(): Stringify a Date as, for example, "2015-12-25" for Dec 25, 2015.
// Tame comes from constant RFC3339 = "2006-01-02T15:04:05" in go stdlib time pkg,
// which is the template we use for time string parsing.
func (d Date) To3339() string {
	return fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d Date) ToSlash() string {
	return fmt.Sprintf("%d/%02d/%02d", d.Year, d.Month, d.Day)
}

// for midnight at the given date in "2015/06/11" format and tz location,
// get the Utc nanoseconds since unix epoch.
func ToUtcDate(date *Date, tz *time.Location) (Utc64, time.Time) {

	inttm, err, tm, _ := DecodeLocalTimeString(tz, date, "00:00:00")
	if err != nil {
		panic(err)
	}
	return Utc64(inttm), tm
}

// At the given date and tz location,
// for the timestring in "HH:MM:SS.999999999" format, get the Utc nanoseconds since unix epoch.
func ToUtc(date *Date, timestring string, tz *time.Location) (Utc64, time.Time) {

	inttm, err, tm, _ := DecodeLocalTimeString(tz, date, timestring)
	if err != nil {
		panic(err)
	}
	return Utc64(inttm), tm
}

func ToUtc1(date *Date, timestring string, tz *time.Location) time.Time {

	_, err, tm, _ := DecodeLocalTimeString(tz, date, timestring)
	if err != nil {
		panic(err)
	}
	return tm.UTC()
}

func ToNyc1(date *Date, timestring string, tz *time.Location) time.Time {

	_, err, tm, _ := DecodeLocalTimeString(tz, date, timestring)
	if err != nil {
		panic(err)
	}
	return tm.In(NYC)
}

func GmtOffsetStringToInt(gmtOffsetString string) (int, error) {
	offset := gmtOffsetString
	if offset[0] == '+' {
		offset = offset[1:]
	}
	off, err := strconv.Atoi(offset)
	if err != nil {
		return 0, fmt.Errorf("GmtOffsetStringToInt() error: gmtOffsetString '%s' could not be converted to integer: '%s'", gmtOffsetString, err)
	}
	return off, nil
}

func GmtOffsetStringToLocation(gmtOffsetString string) (loc *time.Location, hoursEastOfGmt int, err error) {

	if len(gmtOffsetString) == 0 {
		return nil, 0, fmt.Errorf("GmtOffsetStringToLocation() error: gmtOffsetString must not be empty string")
	}

	off, err := GmtOffsetStringToInt(gmtOffsetString)
	if err != nil {
		return nil, 0, err
	}

	// FixedZone wants seconds east of utc, we have hours east. Multiply by 3600 to convert.
	return time.FixedZone(gmtOffsetString, off*3600), off, nil
}

func DecodeTimeFromTrthCsvUtc(dateString, utcTimeString string) (time.Time, *Date, error) {

	date, err := DashMonthDate(dateString)
	if err != nil {
		return ZeroTime, &Date{}, fmt.Errorf("DecodeTimeFromTrthCsvUtc() error: DashMonthDate() failed with: '%s'", err)
	}

	tm, err := DecodeUtcTimeString(date, utcTimeString)
	return tm, date, err
}

// e.g. DecodeTmGmtOffset("+1", "11-JUN-2015", "16:15:05.725633") -> 1434035705725633000
func DecodeTimeFromTrthCsvLocalOffset(gmtOffsetString, dateString, localTimeString string) (Utc64, error, time.Time, *Date) {

	loc, _, err := GmtOffsetStringToLocation(gmtOffsetString)
	if err != nil {
		return ZeroUtc, fmt.Errorf("DecodeTmGmtOffset() error: GmtOffsetStringToLocation() failed with: '%s'", err), time.Time{}, nil
	}

	date, err := DashMonthDate(dateString)
	if err != nil {
		return ZeroUtc, fmt.Errorf("DecodeTmGmtOffset() error: GmtOffsetStringToLocation() failed with: '%s'", err), time.Time{}, nil
	}

	return DecodeLocalTimeString(loc, date, localTimeString)
}

// Produces one Utc64, an int64 holding the number of nanoseconds since the unix epoch at utc.
// localTimeString example: "15:04:05.999999999"
func DecodeLocalTimeString(loc *time.Location, date *Date, localTimeString string) (Utc64, error, time.Time, *Date) {

	date3339 := date.To3339()

	const RFC3339Micro = "2006-01-02T15:04:05.999999"
	tmFormat := RFC3339Micro
	decodeMe := date3339 + "T" + localTimeString

	tm, err := time.ParseInLocation(tmFormat, decodeMe, loc)
	if err != nil {
		return ZeroUtc, fmt.Errorf(`DecodeLocalTimeString() error: time.Parse("%s", "%s") failed with error: '%s'`, tmFormat, decodeMe, err), time.Time{}, nil
	}

	return Utc64(tm.UnixNano()), nil, tm, date
}

// Produces one Utc, an int64 holding the number of nanoseconds since the unix epoch at utc.
// utcTimeString example: "15:04:05.999999999"
func DecodeUtcTimeString(date *Date, utcTimeString string) (time.Time, error) {

	date3339 := date.To3339()

	tmFormat := RFC3339Micro
	decodeMe := date3339 + "T" + utcTimeString

	// we want to parse in UTC, not local time, as UTC time
	// is the time in the TRTH CSV files.
	// NOT:
	// tm, err := time.ParseInLocation(tmFormat, decodeMe, loc)
	// YES:
	tm, err := time.Parse(tmFormat, decodeMe)
	if err != nil {
		return ZeroTime, fmt.Errorf(`DecodeGmtTimeString() error: time.Parse("%s", "%s") failed with error: '%s'`, tmFormat, decodeMe, err)
	}

	return tm, nil
}

// Convert from "11-JUN-2015" -> Date
func DashMonthDate(csvDate string) (*Date, error) {
	splt := strings.Split(csvDate, "-")
	if len(splt) != 3 {
		return nil, fmt.Errorf("CsvDateTo3339('%s') error: argument did not have two dashes in it", csvDate)
	}

	day, err := strconv.Atoi(splt[0])
	if err != nil || day < 1 || day > 31 {
		return nil, fmt.Errorf("CsvDateTo3339('%s') error: expecting DD-MON-YYYY, but could not convert DD part before first dash to an integer day (1-31): '%s'", csvDate, err)
	}

	year, err := strconv.Atoi(splt[2])
	if err != nil || year < 1900 || year > 3000 {
		return nil, fmt.Errorf("CsvDateTo3339('%s') error: expecting DD-MON-YYYY, but could not convert YYYY part after second dash to an integer year (1900-3000): '%s'", csvDate, err)
	}

	if len(splt[1]) != 3 {
		return nil, fmt.Errorf("CsvDateTo3339('%s') error: expecting DD-MON-YYYY, but the month part between the two dashes did not have 3 characters.", csvDate)
	}
	month, err := convertMonth(splt[1])
	if err != nil {
		return nil, fmt.Errorf("CsvDateTo3339('%s') error: expecting DD-MON-YYYY, but convertMonth on the month part failed with: '%s'", csvDate, err)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

// Convert from 20151225 -> Date
func FromIntDate(intDate int) (Date, error) {
	year := intDate / 10000
	month := (intDate % 10000) / 100
	day := intDate % 100
	return Date{Year: year, Month: month, Day: day}, nil
}

func convertMonth(month string) (int, error) {
	switch month {
	case "JAN":
		return 1, nil
	case "FEB":
		return 2, nil
	case "MAR":
		return 3, nil
	case "APR":
		return 4, nil
	case "MAY":
		return 5, nil
	case "JUN":
		return 6, nil
	case "JUL":
		return 7, nil
	case "AUG":
		return 8, nil
	case "SEP":
		return 9, nil
	case "OCT":
		return 10, nil
	case "NOV":
		return 11, nil
	case "DEC":
		return 12, nil
	default:
		return -1, fmt.Errorf("convertMonth() errror: unrecognized month: '%s'", month)
	}
}

// Convert from "2015/06/11" -> Date
func SlashDateTo3339(slashDate string) (*Date, error) {
	splt := strings.Split(slashDate, "/")
	if len(splt) != 3 || len(splt[0]) != 4 || len(splt[1]) != 2 || len(splt[2]) != 2 {
		return nil, fmt.Errorf("SlashDateTo3339('%s') error: argument did not have two '/' in it, or components of wrong length. expecting 'YYYY/MM/DD'.", slashDate)
	}

	day, err := strconv.Atoi(splt[2])
	if err != nil || day < 1 || day > 31 {
		return nil, fmt.Errorf("SlashDateTo3339('%s') error: expecting YYYY/MM/DD, but could not convert DD part after second '/' to an integer day (1-31): '%s'", slashDate, err)
	}

	year, err := strconv.Atoi(splt[0])
	if err != nil || year < 1900 || year > 3000 {
		return nil, fmt.Errorf("SlashDateTo3339('%s') error: expecting YYYY/MM/DD, but could not convert YYYY part before first '/' to an integer year (1900-3000): '%s'", slashDate, err)
	}

	month, err := strconv.Atoi(splt[1])
	if err != nil || month < 1 || month > 12 {
		return nil, fmt.Errorf("SlashDateTo3339('%s') error: expecting YYYY/MM/DD, but could not convert MM part between the two slashes to an integer month (1-12): '%s'", slashDate, err)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

// Convert from "2015/06/11" -> Date
func ParseDateWithSlashOrPanic(s string) *Date {
	dt, err := SlashDateTo3339(s)
	panicOn(err)
	return dt
}

// Convert from "2015-06-11" -> Date
func DashDateTo3339(slashDate string) (Date, error) {
	splt := strings.Split(slashDate, "-")
	if len(splt) != 3 || len(splt[0]) != 4 || len(splt[1]) != 2 || len(splt[2]) != 2 {
		return Date{}, fmt.Errorf("DashDateTo3339('%s') error: argument did not have two '-' in it, or components of wrong length. expecting 'YYYY-MM-DD'.", slashDate)
	}

	day, err := strconv.Atoi(splt[2])
	if err != nil || day < 1 || day > 31 {
		return Date{}, fmt.Errorf("DashDateTo3339('%s') error: expecting YYYY-MM-DD, but could not convert DD part after second '-' to an integer day (1-31): '%s'", slashDate, err)
	}

	year, err := strconv.Atoi(splt[0])
	if err != nil || year < 1900 || year > 3000 {
		return Date{}, fmt.Errorf("DashDateTo3339('%s') error: expecting YYYY-MM-DD, but could not convert YYYY part before first '-' to an integer year (1900-3000): '%s'", slashDate, err)
	}

	month, err := strconv.Atoi(splt[1])
	if err != nil || month < 1 || month > 12 {
		return Date{}, fmt.Errorf("DashDateTo3339('%s') error: expecting YYYY-MM-DD, but could not convert MM part between the two slashes to an integer month (1-12): '%s'", slashDate, err)
	}

	return Date{Year: year, Month: month, Day: day}, nil
}

// version that panics, fine for tests.
func DashDate(slashDate string) Date {
	sd, err := DashDateTo3339(slashDate)
	panicOn(err)
	return sd
}

// version that panics, fine for tests.
func SlashDate(slashDate string) *Date {
	sd, err := SlashDateTo3339(slashDate)
	panicOn(err)
	return sd
}

const nanoseconds_per_day = 60 * 60 * 24 * 1e9
const seconds_per_day = 60 * 60 * 24

// Return the difference between two dates in days: a - b. Rounds halves up (away from zero).
func DayDifference(a, b *Date) int {
	//pp("DayDifference(a='%s', b='%s')", a, b)
	ad, _ := ToUtc(a, "00:00:00", UtcTz)
	bd, _ := ToUtc(b, "00:00:00", UtcTz)

	asec := int64(ad) / 1e9
	bsec := int64(bd) / 1e9
	secdiff := float64(asec - bsec)
	fractionalDay := secdiff / float64(seconds_per_day)
	//pp("fractionalDay = '%v'", fractionalDay)
	roundedDay := math.Round(fractionalDay)
	return int(roundedDay)
}

func ValidTimeOfDay(localTimeString string, date *Date) (okay bool, err error) {
	_, err, _, _ = DecodeLocalTimeString(UtcTz, date, localTimeString)
	if err == nil {
		okay = true
	}
	return okay, err
}

func OnSameDay(a, b time.Time) bool {
	da := NYCDateFromTime(a)
	db := NYCDateFromTime(b)
	return DatesEqual(da, db)
}
