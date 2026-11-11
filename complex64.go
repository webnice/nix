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

// Complex64 Объект complex64, допускающий значение nil.
type Complex64 struct {
	Complex64 complex64 // Значение объекта.
	Valid     bool      // Valid равен true если значение не nil.
}

// NewComplex64 Создание нового объекта Complex64.
func NewComplex64() Complex64 {
	return Complex64{
		Complex64: 0,
		Valid:     false,
	}
}

// NewComplex64Value Создание нового действительного объекта Complex64 из значения.
func NewComplex64Value(value complex64) Complex64 {
	return Complex64{
		Complex64: value,
		Valid:     true,
	}
}

// NewComplex64PointerValue Создание нового действительного объекта Complex64 из ссылки на значение.
func NewComplex64PointerValue(ptr *complex64) Complex64 {
	if ptr == nil {
		return NewComplex64()
	}
	return NewComplex64Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (c *Complex64) SetValid(value complex64) { c.Complex64, c.Valid = value, true; c.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (c *Complex64) Set(value complex64) { c.Complex64, c.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (c *Complex64) Reset() { c.Complex64, c.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (c *Complex64) NullIfDefault() Complex64 {
	if c.Complex64 == 0 {
		c.Reset()
	}
	return *c
}

// MustValue Возвращает значение в любом случае.
func (c *Complex64) MustValue() complex64 {
	if !c.Valid {
		return 0
	}
	return c.Complex64
}

// Pointer Возвращает ссылку на значение.
func (c *Complex64) Pointer() *complex64 {
	if !c.Valid {
		return nil
	}
	return &c.Complex64
}

// Scan Реализация интерфейса Scanner.
func (c *Complex64) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		c.Complex64, c.Valid = 0, false
		return
	case complex128:
		c.Complex64 = complex64(x)
	case complex64:
		c.Complex64 = x
	default:
		buf := asString(x)
		var n complex128
		n, err = strconv.ParseComplex(buf, 64)
		c.Complex64 = complex64(n)
	}
	c.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex64) Value() (driver.Value, error) {
	if !c.Valid {
		return nil, nil
	}
	return strconv.FormatComplex(complex128(c.Complex64), 'g', -1, 64), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (c *Complex64) UnmarshalJSON(data []byte) (err error) {
	var v any

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
		n, err = strconv.ParseComplex(str, 64)
		c.Complex64 = complex64(n)
	case map[string]any:
		err = json.Unmarshal(data, &c.Complex64)
	case nil:
		c.Valid = false
		return
	default:
		err = fmt.Errorf(errComplex64Unmarshal, reflect.TypeOf(v).Name())
	}
	c.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex64) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !c.Valid {
		data = []byte(nullString)
		return
	}
	data, err = json.Marshal(strconv.FormatComplex(complex128(c.Complex64), 'g', -1, 64))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (c *Complex64) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		c.Complex64, c.Valid = 0, false
		return
	case emptyString:
		c.Complex64, c.Valid = 0, true
		return
	default:
		var n complex128
		n, err = strconv.ParseComplex(str, 64)
		c.Complex64 = complex64(n)
	}
	c.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex64) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !c.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatComplex(complex128(c.Complex64), 'g', -1, 64))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (c *Complex64) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Complex64Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Complex64Wrapper)
	if err = dec.Decode(item); err == nil {
		c.Complex64, c.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (c Complex64) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Complex64Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Complex64Wrapper{Value: c.Complex64, Valid: c.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
