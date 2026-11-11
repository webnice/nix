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

	"github.com/webnice/nix/v2/wrapper"
)

// Uint16 Объект uint16, допускающий значение nil.
type Uint16 struct {
	Uint16 uint16 // Значение объекта.
	Valid  bool   // Valid равен true если значение не nil.
}

// NewUint16 Создание нового объекта Uint16.
func NewUint16() Uint16 {
	return Uint16{
		Uint16: 0,
		Valid:  false,
	}
}

// NewUint16Value Создание нового действительного объекта Uint16 из значения.
func NewUint16Value(value uint16) Uint16 {
	return Uint16{
		Uint16: value,
		Valid:  true,
	}
}

// NewUint16PointerValue Создание нового действительного объекта Uint16 из ссылки на значение.
func NewUint16PointerValue(ptr *uint16) Uint16 {
	if ptr == nil {
		return NewUint16()
	}
	return NewUint16Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (u *Uint16) SetValid(value uint16) { u.Uint16, u.Valid = value, true; u.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (u *Uint16) Set(value uint16) { u.Uint16, u.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (u *Uint16) Reset() { u.Uint16, u.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (u *Uint16) NullIfDefault() Uint16 {
	if u.Uint16 == 0 {
		u.Reset()
	}
	return *u
}

// MustValue Возвращает значение в любом случае.
func (u *Uint16) MustValue() uint16 {
	if !u.Valid {
		return 0
	}
	return u.Uint16
}

// Pointer Возвращает ссылку на значение.
func (u *Uint16) Pointer() *uint16 {
	if !u.Valid {
		return nil
	}
	return &u.Uint16
}

// Scan Реализация интерфейса Scanner.
func (u *Uint16) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		u.Uint16, u.Valid = 0, false
		return
	case uint16:
		u.Uint16 = value.(uint16)
	default:
		buf := asString(x)
		var n uint64
		n, err = strconv.ParseUint(buf, 10, 16)
		u.Uint16 = uint16(n)
	}
	u.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint16) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return []byte(strconv.FormatUint(uint64(u.Uint16), 10)), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *Uint16) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		u.Uint16, u.Valid = 0, false
		return
	case float64:
		err = json.Unmarshal(data, &u.Uint16)
	case string:
		var str = x
		if len(str) == 0 {
			u.Valid = false
			return
		}
		var n uint64
		n, err = strconv.ParseUint(str, 10, 16)
		u.Uint16 = uint16(n)
	case map[string]any:
		var rex *regexp.Regexp
		_, uiOK := x["Uint16"].(float64)
		valid, validOK := x["Valid"].(bool)
		if !uiOK || !validOK {
			return fmt.Errorf(errUint16Object, x["Uint16"], x["Valid"])
		}
		rex = regexp.MustCompile(`"Uint16"\s*:\s*([0-9]+)`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = u.UnmarshalText([]byte(arr[1]))
			u.Valid = valid
		}
		return
	default:
		err = fmt.Errorf(errUint16Unmarshal, reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint16) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatUint(uint64(u.Uint16), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (u *Uint16) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		u.Uint16, u.Valid = 0, false
		return
	case emptyString:
		u.Uint16, u.Valid = 0, true
		return
	default:
		var n uint64
		n, err = strconv.ParseUint(str, 10, 16)
		u.Uint16 = uint16(n)
	}
	u.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint16) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatUint(uint64(u.Uint16), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *Uint16) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Uint16Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Uint16Wrapper)
	if err = dec.Decode(item); err == nil {
		u.Uint16, u.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint16) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Uint16Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Uint16Wrapper{Value: u.Uint16, Valid: u.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
