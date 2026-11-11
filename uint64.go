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

// Uint64 Объект uint64, допускающий значение nil.
type Uint64 struct {
	Uint64 uint64 // Значение объекта.
	Valid  bool   // Valid равен true если значение не nil.
}

// NewUint64 Создание нового объекта Uint64.
func NewUint64() Uint64 {
	return Uint64{
		Uint64: 0,
		Valid:  false,
	}
}

// NewUint64Value Создание нового действительного объекта Uint64 из значения.
func NewUint64Value(value uint64) Uint64 {
	return Uint64{
		Uint64: value,
		Valid:  true,
	}
}

// NewUint64PointerValue Создание нового действительного объекта Uint64 из ссылки на значение.
func NewUint64PointerValue(ptr *uint64) Uint64 {
	if ptr == nil {
		return NewUint64()
	}
	return NewUint64Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (u *Uint64) SetValid(value uint64) { u.Uint64, u.Valid = value, true; u.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (u *Uint64) Set(value uint64) { u.Uint64, u.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (u *Uint64) Reset() { u.Uint64, u.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (u *Uint64) NullIfDefault() Uint64 {
	if u.Uint64 == 0 {
		u.Reset()
	}
	return *u
}

// MustValue Возвращает значение в любом случае.
func (u *Uint64) MustValue() uint64 {
	if !u.Valid {
		return 0
	}
	return u.Uint64
}

// Pointer Возвращает ссылку на значение.
func (u *Uint64) Pointer() *uint64 {
	if !u.Valid {
		return nil
	}
	return &u.Uint64
}

// Scan Реализация интерфейса Scanner.
func (u *Uint64) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		u.Uint64, u.Valid = 0, false
		return
	case uint64:
		u.Uint64 = value.(uint64)
	default:
		buf := asString(x)
		u.Uint64, err = strconv.ParseUint(buf, 10, 64)
	}
	u.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint64) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return []byte(strconv.FormatUint(u.Uint64, 10)), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *Uint64) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		u.Uint64, u.Valid = 0, false
		return
	case float64:
		err = json.Unmarshal(data, &u.Uint64)
	case string:
		var str = x
		if len(str) == 0 {
			u.Valid = false
			return
		}
		u.Uint64, err = strconv.ParseUint(str, 10, 64)
	case map[string]any:
		var rex *regexp.Regexp
		_, uiOK := x["Uint64"].(float64)
		valid, validOK := x["Valid"].(bool)
		if !uiOK || !validOK {
			return fmt.Errorf(errUint64Object, x["Uint64"], x["Valid"])
		}
		// TODO: Надо подумать как сделать лучше...
		rex = regexp.MustCompile(`"Uint64"\s*:\s*([0-9]+)`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = u.UnmarshalText([]byte(arr[1]))
			u.Valid = valid
		}
		return
	default:
		err = fmt.Errorf(errUint64Unmarshal, reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint64) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatUint(u.Uint64, 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (u *Uint64) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		u.Uint64, u.Valid = 0, false
		return
	case emptyString:
		u.Uint64, u.Valid = 0, true
		return
	default:
		u.Uint64, err = strconv.ParseUint(str, 10, 64)
	}
	u.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint64) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatUint(u.Uint64, 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *Uint64) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Uint64Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Uint64Wrapper)
	if err = dec.Decode(item); err == nil {
		u.Uint64, u.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint64) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Uint64Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Uint64Wrapper{Value: u.Uint64, Valid: u.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
