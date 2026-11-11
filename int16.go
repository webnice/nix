package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/webnice/nix/v2/wrapper"
)

// Int16 Объект int16, допускающий значение nil.
type Int16 struct {
	Int16 int16 // Значение объекта.
	Valid bool  // Valid равен true если значение не nil.
}

// NewInt16 Создание нового объекта Int16.
func NewInt16() Int16 {
	return Int16{
		Int16: 0,
		Valid: false,
	}
}

// NewInt16Value Создание нового действительного объекта Int16 из значения.
func NewInt16Value(value int16) Int16 {
	return Int16{
		Int16: value,
		Valid: true,
	}
}

// NewInt16PointerValue Создание нового действительного объекта Int16 из ссылки на значение.
func NewInt16PointerValue(ptr *int16) Int16 {
	if ptr == nil {
		return NewInt16()
	}
	return NewInt16Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (i *Int16) SetValid(value int16) { i.Int16, i.Valid = value, true; i.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (i *Int16) Set(value int16) { i.Int16, i.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (i *Int16) Reset() { i.Int16, i.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (i *Int16) NullIfDefault() Int16 {
	if i.Int16 == 0 {
		i.Reset()
	}
	return *i
}

// MustValue Возвращает значение в любом случае.
func (i *Int16) MustValue() int16 {
	if !i.Valid {
		return 0
	}
	return i.Int16
}

// Pointer Возвращает ссылку на значение.
func (i *Int16) Pointer() *int16 {
	if !i.Valid {
		return nil
	}
	return &i.Int16
}

// Scan Реализация интерфейса Scanner.
func (i *Int16) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		i.Int16, i.Valid = 0, false
		return
	case int16:
		i.Int16 = value.(int16)
	default:
		buf := asString(x)
		var n int64
		n, err = strconv.ParseInt(buf, 10, 16)
		i.Int16 = int16(n)
	}
	i.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int16) Value() (driver.Value, error) {
	if !i.Valid {
		return nil, nil
	}
	return int64(i.Int16), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (i *Int16) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case float64:
		var n int64
		n, err = strconv.ParseInt(string(data), 10, 16)
		i.Int16 = int16(n)
	case string:
		var str = x
		if len(str) == 0 {
			i.Valid = false
			return
		}
		var n int64
		n, err = strconv.ParseInt(str, 10, 16)
		i.Int16 = int16(n)
	case map[string]any:
		err = json.Unmarshal(data, &i.Int16)
	case nil:
		i.Valid = false
		return
	default:
		err = fmt.Errorf(errInt16Unmarshal, reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int16) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatInt(int64(i.Int16), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (i *Int16) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		i.Int16, i.Valid = 0, false
		return
	case emptyString:
		i.Int16, i.Valid = 0, true
		return
	default:
		var n int64
		n, err = strconv.ParseInt(str, 10, 16)
		i.Int16 = int16(n)
	}
	i.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int16) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatInt(int64(i.Int16), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (i *Int16) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Int16Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Int16Wrapper)
	if err = dec.Decode(item); err == nil {
		i.Int16, i.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int16) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Int16Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Int16Wrapper{Value: i.Int16, Valid: i.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
