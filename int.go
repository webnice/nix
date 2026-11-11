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

// Int Объект int, допускающий значение nil.
type Int struct {
	Int   int  // Значение объекта.
	Valid bool // Valid равен true если значение не nil.
}

// NewInt Создание нового объекта Int.
func NewInt() Int {
	return Int{
		Int:   0,
		Valid: false,
	}
}

// NewIntValue Создание нового действительного объекта Int из значения.
func NewIntValue(value int) Int {
	return Int{
		Int:   value,
		Valid: true,
	}
}

// NewIntPointerValue Создание нового действительного объекта Int из ссылки на значение.
func NewIntPointerValue(ptr *int) Int {
	if ptr == nil {
		return NewInt()
	}
	return NewIntValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (i *Int) SetValid(value int) { i.Int, i.Valid = value, true; i.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (i *Int) Set(value int) { i.Int, i.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (i *Int) Reset() { i.Int, i.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (i *Int) NullIfDefault() Int {
	if i.Int == 0 {
		i.Reset()
	}
	return *i
}

// MustValue Возвращает значение в любом случае.
func (i *Int) MustValue() int {
	if !i.Valid {
		return 0
	}
	return i.Int
}

// Pointer Возвращает ссылку на значение.
func (i *Int) Pointer() *int {
	if !i.Valid {
		return nil
	}
	return &i.Int
}

// Scan Реализация интерфейса Scanner.
func (i *Int) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		i.Int, i.Valid = 0, false
		return
	case int:
		i.Int = value.(int)
	case int64:
		i.Int = int(value.(int64))
	default:
		buf := asString(x)
		var n int64
		n, err = strconv.ParseInt(buf, 10, 64)
		i.Int = int(n)
	}
	i.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int) Value() (driver.Value, error) {
	if !i.Valid {
		return nil, nil
	}
	return int64(i.Int), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (i *Int) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case float64:
		var n int64
		n, err = strconv.ParseInt(string(data), 10, 64)
		i.Int = int(n)
	case string:
		var str = x
		if len(str) == 0 {
			i.Valid = false
			return
		}
		var n int64
		n, err = strconv.ParseInt(str, 10, 64)
		i.Int = int(n)
	case map[string]any:
		err = json.Unmarshal(data, &i.Int)
	case nil:
		i.Valid = false
		return
	default:
		err = fmt.Errorf(errIntUnmarshal, reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatInt(int64(i.Int), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (i *Int) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		i.Int, i.Valid = 0, false
		return
	case emptyString:
		i.Int, i.Valid = 0, true
		return
	default:
		var n int64
		n, err = strconv.ParseInt(str, 10, 64)
		i.Int = int(n)
	}
	i.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatInt(int64(i.Int), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (i *Int) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.IntWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.IntWrapper)
	if err = dec.Decode(item); err == nil {
		i.Int, i.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i Int) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.IntWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.IntWrapper{Value: i.Int, Valid: i.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
