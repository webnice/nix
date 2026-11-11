package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/webnice/nix/wrapper"
)

// Uint Объект uint, допускающий значение nil.
type Uint struct {
	Uint  uint // Значение объекта.
	Valid bool // Valid равен true если значение не nil.
}

// NewUint Создание нового объекта Uint.
func NewUint() Uint {
	return Uint{
		Uint:  0,
		Valid: false,
	}
}

// NewUintValue Создание нового действительного объекта Uint из значения.
func NewUintValue(value uint) Uint {
	return Uint{
		Uint:  value,
		Valid: true,
	}
}

// NewUintPointerValue Создание нового действительного объекта Uint из ссылки на значение.
func NewUintPointerValue(ptr *uint) Uint {
	if ptr == nil {
		return NewUint()
	}
	return NewUintValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (u *Uint) SetValid(value uint) { u.Uint, u.Valid = value, true; u.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (u *Uint) Set(value uint) { u.Uint, u.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (u *Uint) Reset() { u.Uint, u.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (u *Uint) NullIfDefault() Uint {
	if u.Uint == 0 {
		u.Reset()
	}
	return *u
}

// MustValue Возвращает значение в любом случае.
func (u *Uint) MustValue() uint {
	if !u.Valid {
		return 0
	}
	return u.Uint
}

// Pointer Возвращает ссылку на значение.
func (u *Uint) Pointer() *uint {
	if !u.Valid {
		return nil
	}
	return &u.Uint
}

// Scan Реализация интерфейса Scanner.
func (u *Uint) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		u.Uint, u.Valid = 0, false
		return
	case uint:
		u.Uint = value.(uint)
	case uint64:
		u.Uint = uint(value.(uint64))
	default:
		buf := asString(x)
		var n uint64
		n, err = strconv.ParseUint(buf, 10, 64)
		u.Uint = uint(n)
	}
	u.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return []byte(strconv.FormatUint(uint64(u.Uint), 10)), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *Uint) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		u.Uint, u.Valid = 0, false
		return
	case float64:
		err = json.Unmarshal(data, &u.Uint)
	case string:
		var str = x
		if len(str) == 0 {
			u.Valid = false
			return
		}
		var n uint64
		n, err = strconv.ParseUint(str, 10, 64)
		u.Uint = uint(n)
	case map[string]any:
		var rex *regexp.Regexp
		_, uiOK := x["Uint"].(float64)
		valid, validOK := x["Valid"].(bool)
		if !uiOK || !validOK {
			return fmt.Errorf(errUintObject, x["Uint"], x["Valid"])
		}
		rex = regexp.MustCompile(`"Uint"\s*:\s*([0-9]+)`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = u.UnmarshalText([]byte(arr[1]))
			u.Valid = valid
		}
		return
	default:
		err = fmt.Errorf(errUintUnmarshal, reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatUint(uint64(u.Uint), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (u *Uint) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		u.Uint, u.Valid = 0, false
		return
	case emptyString:
		u.Uint, u.Valid = 0, true
		return
	default:
		var n uint64
		n, err = strconv.ParseUint(str, 10, 64)
		u.Uint = uint(n)
	}
	u.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatUint(uint64(u.Uint), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *Uint) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.UintWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.UintWrapper)
	if err = dec.Decode(item); err == nil {
		u.Uint, u.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.UintWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.UintWrapper{Value: u.Uint, Valid: u.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
