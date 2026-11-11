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

// Float32 Объект float32, допускающий значение nil.
type Float32 struct {
	Float32 float32 // Значение объекта.
	Valid   bool    // Valid равен true если значение не nil.
}

// NewFloat32 Создание нового объекта Float32.
func NewFloat32() Float32 {
	return Float32{
		Float32: 0,
		Valid:   false,
	}
}

// NewFloat32Value Создание нового действительного объекта Float32 из значения.
func NewFloat32Value(value float32) Float32 {
	return Float32{
		Float32: value,
		Valid:   true,
	}
}

// NewFloat32PointerValue Создание нового действительного объекта Float32 из ссылки на значение.
func NewFloat32PointerValue(ptr *float32) Float32 {
	if ptr == nil {
		return NewFloat32()
	}
	return NewFloat32Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (f *Float32) SetValid(value float32) { f.Float32, f.Valid = value, true; f.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (f *Float32) Set(value float32) { f.Float32, f.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (f *Float32) Reset() { f.Float32, f.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (f *Float32) NullIfDefault() Float32 {
	if f.Float32 == 0 {
		f.Reset()
	}
	return *f
}

// MustValue Возвращает значение в любом случае.
func (f *Float32) MustValue() float32 {
	if !f.Valid {
		return 0
	}
	return f.Float32
}

// Pointer Возвращает ссылку на значение.
func (f *Float32) Pointer() *float32 {
	if !f.Valid {
		return nil
	}
	return &f.Float32
}

// Scan Реализация интерфейса Scanner.
func (f *Float32) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		f.Valid = false
		return
	case float64:
		f.Float32 = float32(x)
	case float32:
		f.Float32 = x
	default:
		buf := asString(x)
		var n float64
		n, err = strconv.ParseFloat(buf, 32)
		f.Float32 = float32(n)
	}
	f.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float32) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	}
	return float64(f.Float32), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (f *Float32) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case float64:
		f.Float32 = float32(x)
	case string:
		var str = x
		if len(str) == 0 {
			f.Valid = false
			return
		}
		var n float64
		n, err = strconv.ParseFloat(str, 32)
		f.Float32 = float32(n)
	case map[string]any:
		err = json.Unmarshal(data, &f.Float32)
	case nil:
		f.Valid = false
		return
	default:
		err = fmt.Errorf(errFloat32Unmarshal, reflect.TypeOf(v).Name())
	}
	f.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float32) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !f.Valid {
		data = []byte(nullString)
		return
	}
	if math.IsInf(float64(f.Float32), 0) || math.IsNaN(float64(f.Float32)) {
		data, err = nil, &json.UnsupportedValueError{
			Value: reflect.ValueOf(f.Float32),
			Str:   strconv.FormatFloat(float64(f.Float32), 'g', -1, 32),
		}
		return
	}
	data = []byte(fmt.Sprintf("%f", f.Float32))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (f *Float32) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		f.Float32, f.Valid = 0, false
		return
	case emptyString:
		f.Float32, f.Valid = 0, true
		return
	default:
		var n float64
		n, err = strconv.ParseFloat(string(text), 32)
		f.Float32 = float32(n)
	}
	f.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float32) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !f.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(fmt.Sprintf("%f", f.Float32))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (f *Float32) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Float32Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Float32Wrapper)
	if err = dec.Decode(item); err == nil {
		f.Float32, f.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (f Float32) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Float32Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Float32Wrapper{Value: f.Float32, Valid: f.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
