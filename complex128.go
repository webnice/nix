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

// Complex128 Объект complex128, допускающий значение nil.
type Complex128 struct {
	Complex128 complex128 // Значение объекта.
	Valid      bool       // Valid равен true если значение не nil.
}

// NewComplex128 Создание нового объекта Complex128.
func NewComplex128() Complex128 {
	return Complex128{
		Complex128: 0,
		Valid:      false,
	}
}

// NewComplex128Value Создание нового действительного объекта Complex128 из значения.
func NewComplex128Value(value complex128) Complex128 {
	return Complex128{
		Complex128: value,
		Valid:      true,
	}
}

// NewComplex128PointerValue Создание нового действительного объекта Complex128 из ссылки на значение.
func NewComplex128PointerValue(ptr *complex128) Complex128 {
	if ptr == nil {
		return NewComplex128()
	}
	return NewComplex128Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
func (c *Complex128) SetValid(value complex128) { c.Complex128, c.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (c *Complex128) Reset() { c.Complex128, c.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (c *Complex128) NullIfDefault() Complex128 {
	if c.Complex128 == 0 {
		c.Reset()
	}
	return *c
}

// MustValue Возвращает значение в любом случае.
func (c *Complex128) MustValue() complex128 {
	if !c.Valid {
		return 0
	}
	return c.Complex128
}

// Pointer Возвращает ссылку на значение.
func (c *Complex128) Pointer() *complex128 {
	if !c.Valid {
		return nil
	}
	return &c.Complex128
}

// Scan Реализация интерфейса Scanner.
func (c *Complex128) Scan(value interface{}) (err error) {
	switch x := value.(type) {
	case nil:
		c.Complex128, c.Valid = 0, false
		return
	case complex128:
		c.Complex128 = complex128(x)
	default:
		buf := asString(x)
		var n complex128
		n, err = strconv.ParseComplex(buf, 128)
		c.Complex128 = complex128(n)
	}
	c.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex128) Value() (driver.Value, error) {
	if !c.Valid {
		return nil, nil
	}
	return strconv.FormatComplex(complex128(c.Complex128), 'g', -1, 128), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (c *Complex128) UnmarshalJSON(data []byte) (err error) {
	var v interface{}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case string:
		var str = x
		if len(str) == 0 {
			c.Valid = false
			return
		}
		var n complex128
		n, err = strconv.ParseComplex(str, 128)
		c.Complex128 = complex128(n)
	case map[string]interface{}:
		err = json.Unmarshal(data, &c.Complex128)
	case nil:
		c.Valid = false
		return
	default:
		err = fmt.Errorf("can't unmarshal %q into go value of type nix.Complex128", reflect.TypeOf(v).Name())
	}
	c.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex128) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !c.Valid {
		data = []byte(nullString)
		return
	}
	data, err = json.Marshal(strconv.FormatComplex(complex128(c.Complex128), 'g', -1, 128))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (c *Complex128) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		c.Complex128, c.Valid = 0, false
		return
	case emptyString:
		c.Complex128, c.Valid = 0, true
		return
	default:
		var n complex128
		n, err = strconv.ParseComplex(str, 128)
		c.Complex128 = complex128(n)
	}
	c.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex128) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !c.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatComplex(complex128(c.Complex128), 'g', -1, 128))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (c *Complex128) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Complex128Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Complex128Wrapper)
	if err = dec.Decode(item); err == nil {
		c.Complex128, c.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex128) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Complex128Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Complex128Wrapper{Value: c.Complex128, Valid: c.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
