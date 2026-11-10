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

// Uint8 Объект uint8, допускающий значение nil.
type Uint8 struct {
	Uint8 uint8 // Значение объекта.
	Valid bool  // Valid равен true если значение не nil.
}

// NewUint8 Создание нового объекта Uint8.
func NewUint8() Uint8 {
	return Uint8{
		Uint8: 0,
		Valid: false,
	}
}

// NewUint8Value Создание нового действительного объекта Uint8 из значения.
func NewUint8Value(value uint8) Uint8 {
	return Uint8{
		Uint8: value,
		Valid: true,
	}
}

// NewUint8PointerValue Создание нового действительного объекта Uint8 из ссылки на значение.
func NewUint8PointerValue(ptr *uint8) Uint8 {
	if ptr == nil {
		return NewUint8()
	}
	return NewUint8Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
func (u *Uint8) SetValid(value uint8) { u.Uint8, u.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (u *Uint8) Reset() { u.Uint8, u.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (u *Uint8) NullIfDefault() Uint8 {
	if u.Uint8 == 0 {
		u.Reset()
	}
	return *u
}

// MustValue Возвращает значение в любом случае.
func (u *Uint8) MustValue() uint8 {
	if !u.Valid {
		return 0
	}
	return u.Uint8
}

// Pointer Возвращает ссылку на значение.
func (u *Uint8) Pointer() *uint8 {
	if !u.Valid {
		return nil
	}
	return &u.Uint8
}

// Scan Реализация интерфейса Scanner.
func (u *Uint8) Scan(value interface{}) (err error) {
	switch x := value.(type) {
	case nil:
		u.Uint8, u.Valid = 0, false
		return
	case uint8:
		u.Uint8 = value.(uint8)
	default:
		buf := asString(x)
		var n uint64
		n, err = strconv.ParseUint(buf, 10, 8)
		u.Uint8 = uint8(n)
	}
	u.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint8) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return []byte(strconv.FormatUint(uint64(u.Uint8), 10)), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *Uint8) UnmarshalJSON(data []byte) (err error) {
	var v interface{}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		u.Uint8, u.Valid = 0, false
		return
	case float64:
		err = json.Unmarshal(data, &u.Uint8)
	case string:
		var str = x
		if len(str) == 0 {
			u.Valid = false
			return
		}
		var n uint64
		n, err = strconv.ParseUint(str, 10, 8)
		u.Uint8 = uint8(n)
	case map[string]interface{}:
		var rex *regexp.Regexp
		_, uiOK := x["Uint8"].(float64)
		valid, validOK := x["Valid"].(bool)
		if !uiOK || !validOK {
			return fmt.Errorf(`unmarshalling object into Go value of type nix.Uint8 requires key "Uint8" to be of type `+
				`string and key "Valid" to be of type bool; `+
				`found %T and %T, respectively`, x["Uint8"], x["Valid"])
		}
		rex = regexp.MustCompile(`"Uint8"\s*:\s*([0-9]+)`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = u.UnmarshalText([]byte(arr[1]))
			u.Valid = valid
		}
		return
	default:
		err = fmt.Errorf("Can't unmarshal %v into Go value of type nix.Uint8", reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint8) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatUint(uint64(u.Uint8), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (u *Uint8) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		u.Uint8, u.Valid = 0, false
		return
	case emptyString:
		u.Uint8, u.Valid = 0, true
		return
	default:
		var n uint64
		n, err = strconv.ParseUint(str, 10, 8)
		u.Uint8 = uint8(n)
	}
	u.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint8) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatUint(uint64(u.Uint8), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *Uint8) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Uint8Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Uint8Wrapper)
	if err = dec.Decode(item); err == nil {
		u.Uint8, u.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint8) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Uint8Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Uint8Wrapper{Value: u.Uint8, Valid: u.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
