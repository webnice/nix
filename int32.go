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

// Int32 Объект int32, допускающий значение nil.
type Int32 struct {
	Int32 int32 // Значение объекта.
	Valid bool  // Valid равен true если значение не nil.
}

// NewInt32 Создание нового объекта Int32.
func NewInt32() Int32 {
	return Int32{
		Int32: 0,
		Valid: false,
	}
}

// NewInt32Value Создание нового действительного объекта Int32 из значения.
func NewInt32Value(value int32) Int32 {
	return Int32{
		Int32: value,
		Valid: true,
	}
}

// NewInt32PointerValue Создание нового действительного объекта Int32 из ссылки на значение.
func NewInt32PointerValue(ptr *int32) Int32 {
	if ptr == nil {
		return NewInt32()
	}
	return NewInt32Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
func (i *Int32) SetValid(value int32) { i.Int32, i.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (i *Int32) Reset() { i.Int32, i.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (i *Int32) NullIfDefault() Int32 {
	if i.Int32 == 0 {
		i.Reset()
	}
	return *i
}

// MustValue Возвращает значение в любом случае.
func (i *Int32) MustValue() int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

// Pointer Возвращает ссылку на значение.
func (i *Int32) Pointer() *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// Scan Реализация интерфейса Scanner.
func (i *Int32) Scan(value interface{}) (err error) {
	switch x := value.(type) {
	case nil:
		i.Int32, i.Valid = 0, false
		return
	case int32:
		i.Int32 = value.(int32)
	default:
		buf := asString(x)
		var n int64
		n, err = strconv.ParseInt(buf, 10, 32)
		i.Int32 = int32(n)
	}
	i.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int32) Value() (driver.Value, error) {
	if !i.Valid {
		return nil, nil
	}
	return int64(i.Int32), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (i *Int32) UnmarshalJSON(data []byte) (err error) {
	var v interface{}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case float64:
		var n int64
		n, err = strconv.ParseInt(string(data), 10, 32)
		i.Int32 = int32(n)
	case string:
		var str = x
		if len(str) == 0 {
			i.Valid = false
			return
		}
		var n int64
		n, err = strconv.ParseInt(str, 10, 32)
		i.Int32 = int32(n)
	case map[string]interface{}:
		err = json.Unmarshal(data, &i.Int32)
	case nil:
		i.Valid = false
		return
	default:
		err = fmt.Errorf("can't unmarshal %q into go value of type nix.Int32", reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int32) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatInt(int64(i.Int32), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (i *Int32) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		i.Int32, i.Valid = 0, false
		return
	case emptyString:
		i.Int32, i.Valid = 0, true
		return
	default:
		var n int64
		n, err = strconv.ParseInt(str, 10, 32)
		i.Int32 = int32(n)
	}
	i.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int32) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatInt(int64(i.Int32), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (i *Int32) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Int32Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Int32Wrapper)
	if err = dec.Decode(item); err == nil {
		i.Int32, i.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int32) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Int32Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Int32Wrapper{Value: i.Int32, Valid: i.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
