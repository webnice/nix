package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/webnice/nix/wrapper"
)

// Float64 Объект float64, допускающий значение nil.
type Float64 struct {
	Float64 float64 // Значение объекта.
	Valid   bool    // Valid равен true если значение не nil.
}

// NewFloat64 Создание нового объекта Float64.
func NewFloat64() Float64 {
	return Float64{
		Float64: 0,
		Valid:   false,
	}
}

// NewFloat64Value Создание нового действительного объекта Float64 из значения.
func NewFloat64Value(value float64) Float64 {
	return Float64{
		Float64: value,
		Valid:   true,
	}
}

// NewFloat64PointerValue Создание нового действительного объекта Float64 из ссылки на значение.
func NewFloat64PointerValue(ptr *float64) Float64 {
	if ptr == nil {
		return NewFloat64()
	}
	return NewFloat64Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
func (f *Float64) SetValid(value float64) { f.Float64, f.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (f *Float64) Reset() { f.Float64, f.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (f *Float64) NullIfDefault() Float64 {
	if f.Float64 == 0 {
		f.Reset()
	}
	return *f
}

// MustValue Возвращает значение в любом случае.
func (f *Float64) MustValue() float64 {
	if !f.Valid {
		return 0
	}
	return f.Float64
}

// Pointer Возвращает ссылку на значение.
func (f *Float64) Pointer() *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// Scan Реализация интерфейса Scanner.
func (f *Float64) Scan(value interface{}) (err error) {
	switch x := value.(type) {
	case nil:
		f.Valid = false
		return
	case float64:
		f.Float64 = value.(float64)
	default:
		buf := asString(x)
		f.Float64, err = strconv.ParseFloat(buf, 64)
	}
	f.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float64) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	}
	return f.Float64, nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (f *Float64) UnmarshalJSON(data []byte) (err error) {
	var v interface{}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case float64:
		f.Float64 = x
	case string:
		var str = x
		if len(str) == 0 {
			f.Valid = false
			return
		}
		f.Float64, err = strconv.ParseFloat(str, 64)
	case map[string]interface{}:
		err = json.Unmarshal(data, &f.Float64)
	case nil:
		f.Valid = false
		return
	default:
		err = fmt.Errorf("can't unmarshal %q into go value of type nix.Float64", reflect.TypeOf(v).Name())
	}
	f.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float64) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !f.Valid {
		data = []byte(nullString)
		return
	}
	if math.IsInf(f.Float64, 0) || math.IsNaN(f.Float64) {
		data, err = nil, &json.UnsupportedValueError{
			Value: reflect.ValueOf(f.Float64),
			Str:   strconv.FormatFloat(f.Float64, 'g', -1, 64),
		}
		return
	}
	data = []byte(fmt.Sprintf("%f", f.Float64))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (f *Float64) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		f.Float64, f.Valid = 0, false
		return
	case emptyString:
		f.Float64, f.Valid = 0, true
		return
	default:
		f.Float64, err = strconv.ParseFloat(string(text), 64)
	}
	f.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float64) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !f.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(fmt.Sprintf("%f", f.Float64))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (f *Float64) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Float64Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Float64Wrapper)
	if err = dec.Decode(item); err == nil {
		f.Float64, f.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float64) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Float64Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Float64Wrapper{Value: f.Float64, Valid: f.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
