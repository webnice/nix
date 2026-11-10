package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/webnice/nix/wrapper"
)

// Bool Объект boolean, допускающий значение nil.
type Bool struct {
	Bool  bool // Значение объекта.
	Valid bool // Valid равен true если значение не nil.
}

// NewBool Создание нового не действительного объекта Bool.
func NewBool() Bool {
	return Bool{
		Bool:  false,
		Valid: false,
	}
}

// NewBoolValue Создание нового действительного объекта Bool из значения.
func NewBoolValue(value bool) Bool {
	return Bool{
		Bool:  value,
		Valid: true,
	}
}

// NewBoolPointerValue Создание нового действительного объекта Bool из ссылки на значение.
func NewBoolPointerValue(ptr *bool) Bool {
	if ptr == nil {
		return NewBool()
	}
	return NewBoolValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (b *Bool) SetValid(value bool) { b.Bool, b.Valid = value, true; b.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (b *Bool) Set(value bool) { b.Bool, b.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (b *Bool) Reset() { b.Bool, b.Valid = false, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (b *Bool) NullIfDefault() Bool {
	if !b.Bool {
		b.Reset()
	}
	return *b
}

// MustValue Возвращает значение в любом случае.
func (b *Bool) MustValue() bool { return b.Valid && b.Bool }

// Pointer Возвращает ссылку на значение.
func (b *Bool) Pointer() *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

// Scan Реализация интерфейса Scanner.
func (b *Bool) Scan(value any) (err error) {
	var (
		v  any
		ok bool
	)

	b.Bool, b.Valid = false, false
	if value == nil {
		return
	}
	v, err = driver.Bool.ConvertValue(value)
	b.Valid = err == nil
	if b.Bool, ok = v.(bool); !ok {
		b.Reset()
	}

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (b Bool) Value() (driver.Value, error) {
	if !b.Valid {
		return nil, nil
	}
	return b.Bool, nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (b *Bool) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case bool:
		b.Bool = x
	case map[string]any:
		value, okValue := x["Bool"].(bool)
		valid, okValid := x["Valid"].(bool)
		if !okValue || !okValid {
			return fmt.Errorf("unmarshalling object into Go value of type nix.Bool requires key "+
				`"Bool" to be of type string and key "Valid" to be of type bool; `+
				"found %T and %T, respectively", x["Bool"], x["Valid"])
		}
		b.Bool = value
		b.Valid = valid
	case nil:
		b.Valid = false
		return
	}
	b.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (b Bool) MarshalJSON() (data []byte, err error) {
	const (
		nullString  = "null"
		trueString  = "true"
		falseString = "false"
	)

	if !b.Valid {
		data = []byte(nullString)
		return
	}
	if !b.Bool {
		data = []byte(falseString)
		return
	}
	data = []byte(trueString)

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (b *Bool) UnmarshalText(text []byte) (err error) {
	const (
		errInput    = "ошибка в значении: %q"
		emptyString = ""
		nullString  = "null"
		trueString  = "true"
		falseString = "false"
	)
	var str string

	switch str = string(text); str {
	case emptyString, nullString:
		b.Valid = false
		return
	case trueString:
		b.Bool = true
	case falseString:
		b.Bool = false
	default:
		b.Valid, err = false, fmt.Errorf(errInput, str)
		return
	}
	b.Valid = true

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (b Bool) MarshalText() (text []byte, err error) {
	const (
		trueString  = "true"
		falseString = "false"
		nullString  = "null"
	)

	if !b.Valid {
		text = []byte(nullString)
		return
	}
	if !b.Bool {
		text = []byte(falseString)
		return
	}
	text = []byte(trueString)

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (b *Bool) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.BoolWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.BoolWrapper)
	if err = dec.Decode(item); err == nil {
		b.Bool, b.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (b Bool) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.BoolWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.BoolWrapper{Value: b.Bool, Valid: b.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
