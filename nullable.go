package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// nullableValue Контракт скалярных типов значений, допустимых для Nullable.
type nullableValue interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string | ~bool |
		~complex64 | ~complex128
}

// nullableWire Двоичное представление Nullable для кодека gob.
type nullableWire[T nullableValue] struct {
	Value T
	Valid bool
}

// Nullable Объект произвольного скалярного типа T, допускающий значение nil.
type Nullable[T nullableValue] struct {
	Nullable T    // Значение объекта.
	Valid    bool // Valid равен true если значение не nil.
}

// NewNullable Создание нового объекта Nullable.
func NewNullable[T nullableValue]() Nullable[T] {
	var zero T
	return Nullable[T]{
		Nullable: zero,
		Valid:    false,
	}
}

// NewNullableValue Создание нового действительного объекта Nullable из значения.
func NewNullableValue[T nullableValue](value T) Nullable[T] {
	return Nullable[T]{
		Nullable: value,
		Valid:    true,
	}
}

// NewNullablePointerValue Создание нового действительного объекта Nullable из ссылки на значение.
func NewNullablePointerValue[T nullableValue](ptr *T) Nullable[T] {
	if ptr == nil {
		return NewNullable[T]()
	}
	return NewNullableValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
func (n *Nullable[T]) SetValid(value T) { n.Nullable, n.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (n *Nullable[T]) Reset() { var zero T; n.Nullable, n.Valid = zero, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (n *Nullable[T]) NullIfDefault() Nullable[T] {
	var zero T
	if n.Nullable == zero {
		n.Reset()
	}
	return *n
}

// MustValue Возвращает значение в любом случае.
func (n *Nullable[T]) MustValue() T {
	if !n.Valid {
		var zero T
		return zero
	}
	return n.Nullable
}

// Pointer Возвращает ссылку на значение.
func (n *Nullable[T]) Pointer() *T {
	if !n.Valid {
		return nil
	}
	return &n.Nullable
}

// Scan Реализация интерфейса Scanner.
func (n *Nullable[T]) Scan(value interface{}) (err error) {
	if value == nil {
		n.Reset()
		return
	}
	err = n.parse(asString(value))
	n.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (n Nullable[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	rv := reflect.ValueOf(n.Nullable)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return []byte(strconv.FormatUint(rv.Uint(), 10)), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.String:
		return rv.String(), nil
	default:
		// Комплексные типы database/sql не поддерживает, сохраняется строковая форма.
		return strconv.FormatComplex(rv.Complex(), 'g', -1, 128), nil
	}
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (n *Nullable[T]) UnmarshalJSON(data []byte) (err error) {
	const nullString = "null"

	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	if string(data) == nullString {
		n.Reset()
		return
	}
	var s string
	switch data[0] {
	case '"':
		if err = json.Unmarshal(data, &s); err != nil {
			return
		}
	default:
		s = string(data)
	}
	if len(s) == 0 {
		n.Valid = false
		return
	}
	if err = n.parse(s); err == nil {
		n.Valid = true
	}

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (n Nullable[T]) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !n.Valid {
		data = []byte(nullString)
		return
	}
	data, err = json.Marshal(n.Nullable)

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (n *Nullable[T]) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)

	switch str := string(text); str {
	case nullString:
		n.Reset()
		return
	case emptyString:
		var zero T
		n.Nullable, n.Valid = zero, true
		return
	default:
		if err = n.parse(str); err == nil {
			n.Valid = true
		}
	}

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (n Nullable[T]) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !n.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(fmt.Sprintf("%v", n.Nullable))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (n *Nullable[T]) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   nullableWire[T]
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	if err = dec.Decode(&item); err == nil {
		n.Nullable, n.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (n Nullable[T]) MarshalBinary() (data []byte, err error) {
	var (
		buf *bytes.Buffer
		enc *gob.Encoder
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	err = enc.Encode(nullableWire[T]{Value: n.Nullable, Valid: n.Valid})
	data = buf.Bytes()

	return
}

// parse Разбор строки в значение типа T по его базовому типу.
func (n *Nullable[T]) parse(s string) (err error) {
	rv := reflect.ValueOf(&n.Nullable).Elem()
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var v int64
		if v, err = strconv.ParseInt(s, 10, 64); err == nil {
			rv.SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var v uint64
		if v, err = strconv.ParseUint(s, 10, 64); err == nil {
			rv.SetUint(v)
		}
	case reflect.Float32, reflect.Float64:
		var v float64
		if v, err = strconv.ParseFloat(s, 64); err == nil {
			rv.SetFloat(v)
		}
	case reflect.Bool:
		var v bool
		if v, err = strconv.ParseBool(s); err == nil {
			rv.SetBool(v)
		}
	case reflect.String:
		rv.SetString(s)
	case reflect.Complex64, reflect.Complex128:
		var v complex128
		if v, err = strconv.ParseComplex(s, 128); err == nil {
			rv.SetComplex(v)
		}
	}

	return
}
