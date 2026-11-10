package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/webnice/nix/wrapper"
)

// Int8 Объект int8, допускающий значение nil.
type Int8 struct {
	Int8  int8 // Значение объекта.
	Valid bool // Valid равен true если значение не nil.
}

// NewInt8 Создание нового объекта Int8.
func NewInt8() Int8 {
	return Int8{
		Int8:  0,
		Valid: false,
	}
}

// NewInt8Value Создание нового действительного объекта Int8 из значения.
func NewInt8Value(value int8) Int8 {
	return Int8{
		Int8:  value,
		Valid: true,
	}
}

// NewInt8PointerValue Создание нового действительного объекта Int8 из ссылки на значение.
func NewInt8PointerValue(ptr *int8) Int8 {
	if ptr == nil {
		return NewInt8()
	}
	return NewInt8Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (i *Int8) SetValid(value int8) { i.Int8, i.Valid = value, true; i.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (i *Int8) Set(value int8) { i.Int8, i.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (i *Int8) Reset() { i.Int8, i.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (i *Int8) NullIfDefault() Int8 {
	if i.Int8 == 0 {
		i.Reset()
	}
	return *i
}

// MustValue Возвращает значение в любом случае.
func (i *Int8) MustValue() int8 {
	if !i.Valid {
		return 0
	}
	return i.Int8
}

// Pointer Возвращает ссылку на значение.
func (i *Int8) Pointer() *int8 {
	if !i.Valid {
		return nil
	}
	return &i.Int8
}

// Scan Реализация интерфейса Scanner.
func (i *Int8) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		i.Int8, i.Valid = 0, false
		return
	case int8:
		i.Int8 = value.(int8)
	default:
		buf := asString(x)
		var n int64
		n, err = strconv.ParseInt(buf, 10, 8)
		i.Int8 = int8(n)
	}
	i.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int8) Value() (driver.Value, error) {
	if !i.Valid {
		return nil, nil
	}
	return int64(i.Int8), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (i *Int8) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case float64:
		var n int64
		n, err = strconv.ParseInt(string(data), 10, 8)
		i.Int8 = int8(n)
	case string:
		var str = x
		if len(str) == 0 {
			i.Valid = false
			return
		}
		var n int64
		n, err = strconv.ParseInt(str, 10, 8)
		i.Int8 = int8(n)
	case map[string]any:
		err = json.Unmarshal(data, &i.Int8)
	case nil:
		i.Valid = false
		return
	default:
		err = fmt.Errorf("can't unmarshal %q into go value of type nix.Int8", reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int8) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatInt(int64(i.Int8), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (i *Int8) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		i.Int8, i.Valid = 0, false
		return
	case emptyString:
		i.Int8, i.Valid = 0, true
		return
	default:
		var n int64
		n, err = strconv.ParseInt(str, 10, 8)
		i.Int8 = int8(n)
	}
	i.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int8) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatInt(int64(i.Int8), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (i *Int8) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Int8Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Int8Wrapper)
	if err = dec.Decode(item); err == nil {
		i.Int8, i.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int8) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Int8Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Int8Wrapper{Value: i.Int8, Valid: i.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
