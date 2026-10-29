package i18n

import (
	"time"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/number"
)

// Number formats v with at most the given decimals, such as 1,234.5 in
// English and 1 234,5 in Canadian French.
func (p *Printer) Number(v float64, decimals int) string {
	return p.numbers.Sprint(number.Decimal(v, number.MaxFractionDigits(decimals)))
}

// sizeUnits holds the catalog IDs of decimal size units, as macOS and most
// download tools count them.
var sizeUnits = []string{"size.bytes", "size.kilobytes", "size.megabytes", "size.gigabytes", "size.terabytes"}

// Size formats a byte count with at most one decimal in the largest unit
// that keeps the number at 1 or more, such as 1.5 MB, in the plural form of
// that number. Bytes take no decimal.
func (p *Printer) Size(bytes int64) string {
	value, unit := float64(bytes), 0
	for value >= 1000 && unit < len(sizeUnits)-1 {
		value /= 1000
		unit++
	}
	decimals := 1
	if unit == 0 {
		decimals = 0
	}
	return p.measure(sizeUnits[unit], value, decimals)
}

// durationUnits holds the catalog IDs of duration units, from the
// millisecond to the hour, with the length of each.
var durationUnits = []struct {
	id     string
	length time.Duration
}{{"duration.milliseconds", time.Millisecond}, {"duration.seconds", time.Second}, {"duration.minutes", time.Minute}, {"duration.hours", time.Hour}}

// Duration formats d in the largest unit, up to the hour, that keeps the
// number at 1 or more, with at most one decimal, such as 1.5 s or 2 min,
// in the plural form of that number. Milliseconds take no decimal.
func (p *Printer) Duration(d time.Duration) string {
	unit := 0
	for unit < len(durationUnits)-1 && d >= durationUnits[unit+1].length {
		unit++
	}
	decimals := 1
	if unit == 0 {
		decimals = 0
	}
	return p.measure(durationUnits[unit].id, float64(d)/float64(durationUnits[unit].length), decimals)
}

// measure returns the text of the unit message id, with N filled by v
// formatted with at most the given decimals, in the plural form of that
// number as it shows, such as 2 for 1.96. The plural rules read the digits
// that plain writes, which round v as N does, ties to even included, so the
// form and the text never disagree.
func (p *Printer) measure(id string, v float64, decimals int) string {
	return p.localize(&goi18n.LocalizeConfig{
		MessageID:    id,
		PluralCount:  p.plain.Sprint(number.Decimal(v, number.MaxFractionDigits(decimals), number.NoSeparator())),
		TemplateData: map[string]any{"N": p.Number(v, decimals)},
	})
}

// Date formats a day as ISO 8601, such as 2026-10-06, in every language:
// it reads the same in English and is the Canadian French standard.
func (p *Printer) Date(t time.Time) string { return t.Format(time.DateOnly) }

// clipLimit is the number of characters that a message quotes of a value
// that it refuses. A valid run ID, at most 64 characters, stays whole.
const clipLimit = 80

// Clip returns the first clipLimit characters of a value from the command
// line, then …, when the value is longer. A message that refuses the value
// quotes it so, as Output in contract.md asks. A hint or a next command keeps
// the whole value.
func Clip(value string) string {
	count := 0
	for i := range value {
		if count == clipLimit {
			return value[:i] + "…"
		}
		count++
	}
	return value
}
