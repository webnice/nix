package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/webnice/nix/v2/wrapper"
)

// Duration Объект time.Duration, допускающий значение nil.
type Duration struct {
	Duration time.Duration // Значение объекта.
	Valid    bool          // Valid равен true если значение не nil.
}

// NewDuration Создание нового объекта Duration.
func NewDuration() Duration {
	return Duration{
		Duration: 0,
		Valid:    false,
	}
}

// NewDurationValue Создание нового действительного объекта Duration из значения.
func NewDurationValue(value time.Duration) Duration {
	return Duration{
		Duration: value,
		Valid:    true,
	}
}

// NewDurationPointerValue Создание нового действительного объекта Duration из ссылки на значение.
func NewDurationPointerValue(ptr *time.Duration) Duration {
	if ptr == nil {
		return NewDuration()
	}
	return NewDurationValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (d *Duration) SetValid(value time.Duration) {
	d.Duration, d.Valid = value, true
	d.NullIfDefault()
}

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (d *Duration) Set(value time.Duration) { d.Duration, d.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (d *Duration) Reset() { d.Duration, d.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (d *Duration) NullIfDefault() Duration {
	if d.Duration == 0 {
		d.Reset()
	}
	return *d
}

// MustValue Возвращает значение в любом случае.
func (d *Duration) MustValue() time.Duration {
	if !d.Valid {
		return 0
	}
	return d.Duration
}

// Pointer Возвращает ссылку на значение.
func (d *Duration) Pointer() *time.Duration {
	if !d.Valid {
		return nil
	}
	return &d.Duration
}

// Scan Реализация интерфейса Scanner.
func (d *Duration) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		d.Duration, d.Valid = 0, false
		return
	case time.Duration:
		d.Duration = x
	case int64:
		d.Duration = time.Duration(x)
	default:
		buf := asString(x)
		d.Duration, err = time.ParseDuration(buf)
	}
	d.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (d Duration) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return int64(d.Duration), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (d *Duration) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case float64:
		var n int64
		n, err = strconv.ParseInt(string(data), 10, 64)
		d.Duration = time.Duration(n)
	case string:
		var str = x
		if len(str) == 0 {
			d.Valid = false
			return
		}
		d.Duration, err = time.ParseDuration(str)
	case map[string]any:
		err = json.Unmarshal(data, &d.Duration)
	case nil:
		d.Valid = false
		return
	default:
		err = fmt.Errorf(errDurationUnmarshal, reflect.TypeOf(v).Name())
	}
	d.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (d Duration) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !d.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatInt(int64(d.Duration), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (d *Duration) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		d.Duration, d.Valid = 0, false
		return
	case emptyString:
		d.Duration, d.Valid = 0, true
		return
	default:
		d.Duration, err = time.ParseDuration(str)
	}
	d.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (d Duration) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !d.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(d.Duration.String())

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (d *Duration) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.DurationWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.DurationWrapper)
	if err = dec.Decode(item); err == nil {
		d.Duration, d.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (d Duration) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.DurationWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.DurationWrapper{Value: d.Duration, Valid: d.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
