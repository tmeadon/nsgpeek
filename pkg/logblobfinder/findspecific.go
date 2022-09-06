package logblobfinder

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

var (
	blobPathRe     = regexp.MustCompile(`.*(y=\d{4})\/?(m=\d{2})?\/?(d=\d{2})?\/?(h=\d{2})?\/?(m=\d{2})?\/?`)
	blobPathElemRe = regexp.MustCompile(`[ymdh]=(\d{2,4})`)
)

func (f *Finder) FindSpecific(start time.Time, end time.Time) ([]azure.Blob, error) {
	logPrefix, err := f.findNsgBlobPrefix()
	if err != nil {
		return nil, err
	}

	if logPrefix == "" {
		return nil, ErrBlobPrefixNotFound
	}

	return f.findTimeBlobs(start, end, logPrefix)
}

func (f *Finder) findTimeBlobs(start time.Time, end time.Time, logPrefix string) ([]azure.Blob, error) {
	blobs := make([]azure.Blob, 0)

	_, childPrefixes, err := f.ListBlobDirectory(logPrefix)
	if err != nil {
		return nil, err
	}

	for _, prefix := range childPrefixes {
		elems, err := extractBlobPathElements(prefix)
		if err != nil {
			return nil, fmt.Errorf("failed to extract time elements from blob path '%v': %w", prefix, err)
		}

		if elems.Year == nil {
			return nil, fmt.Errorf("unexpected nil value for year element in blob path '%v", prefix)
		}

		if shouldListPrefixes(start, end, elems) {
			b, err := f.findTimeBlobs(start, end, prefix)
			if err != nil {
				return nil, err
			}
			blobs = append(blobs, b...)
		}

		if shouldListBlobs(start, end, elems) {
			b, err := f.ListBlobs(prefix)
			if err != nil {
				return nil, err
			}
			blobs = append(blobs, b...)
		}
	}

	return blobs, nil
}

type blobPathTimeElements struct {
	Year, Month, Day, Hour *int
}

func extractBlobPathElements(prefix string) (*blobPathTimeElements, error) {
	elems := new(blobPathTimeElements)

	match := blobPathRe.FindStringSubmatch(prefix)

	if match[1] != "" {
		y, err := extractTimeElemInt(match[1])
		if err != nil {
			return nil, err
		}
		elems.Year = &y
	}

	if match[2] != "" {
		m, err := extractTimeElemInt(match[2])
		if err != nil {
			return nil, err
		}
		elems.Month = &m
	}

	if match[3] != "" {
		d, err := extractTimeElemInt(match[3])
		if err != nil {
			return nil, err
		}
		elems.Day = &d
	}

	if match[4] != "" {
		h, err := extractTimeElemInt(match[4])
		if err != nil {
			return nil, err
		}
		elems.Hour = &h
	}

	return elems, nil
}

func extractTimeElemInt(element string) (int, error) {
	match := blobPathElemRe.FindStringSubmatch(element)
	val, err := strconv.Atoi(match[1])
	return val, err
}

func shouldListPrefixes(start time.Time, end time.Time, elems *blobPathTimeElements) bool {
	if elems.Month == nil {
		return correctYear(start, end, elems)
	} else if elems.Day == nil {
		return correctMonth(start, end, elems)
	} else if elems.Hour == nil {
		return correctDay(start, end, elems)
	}
	return false
}

func shouldListBlobs(start time.Time, end time.Time, elems *blobPathTimeElements) bool {
	if elems.Year != nil && elems.Month != nil && elems.Day != nil && elems.Hour != nil {
		return correctYear(start, end, elems) &&
			correctMonth(start, end, elems) &&
			correctDay(start, end, elems) &&
			correctHour(start, end, elems)
	}
	return false
}

func correctYear(start time.Time, end time.Time, elems *blobPathTimeElements) bool {
	return *elems.Year >= start.Year() && *elems.Year <= end.Year()
}

func correctMonth(start time.Time, end time.Time, elems *blobPathTimeElements) bool {
	t := timeFromYearMonth(*elems.Year, *elems.Month)
	startMonth := timeFromYearMonth(start.Year(), int(start.Month()))
	endMonth := timeFromYearMonth(end.Year(), int(end.Month()))
	return tInRange(startMonth, endMonth, t)
}

func timeFromYearMonth(year int, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
}

func correctDay(start time.Time, end time.Time, elems *blobPathTimeElements) bool {
	t := timeFromYearMonthDay(*elems.Year, *elems.Month, *elems.Day)
	startDay := timeFromYearMonthDay(start.Year(), int(start.Month()), start.Day())
	endDay := timeFromYearMonthDay(end.Year(), int(end.Month()), end.Day())
	return tInRange(startDay, endDay, t)
}

func timeFromYearMonthDay(year int, month int, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}

func correctHour(start time.Time, end time.Time, elems *blobPathTimeElements) bool {
	t := timeFromYearMonthDayHour(*elems.Year, *elems.Month, *elems.Day, *elems.Hour)
	startHour := timeFromYearMonthDayHour(start.Year(), int(start.Month()), start.Day(), start.Hour())
	endHour := timeFromYearMonthDayHour(end.Year(), int(end.Month()), end.Day(), end.Hour())
	return tInRange(startHour, endHour, t)
}

func timeFromYearMonthDayHour(year int, month int, day int, hour int) time.Time {
	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local)
}

func tInRange(start time.Time, end time.Time, t time.Time) bool {
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}
