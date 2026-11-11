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

// Uintptr Объект uintptr, допускающий значение nil.
type Uintptr struct {
	Uintptr uintptr // Значение объекта.
	Valid   bool    // Valid равен true если значение не nil.
}

// NewUintptr Создание нового объекта Uintptr.
func NewUintptr() Uintptr {
	return Uintptr{
		Uintptr: 0,
		Valid:   false,
	}
}

// NewUintptrValue Создание нового действительного объекта Uintptr из значения.
func NewUintptrValue(value uintptr) Uintptr {
	return Uintptr{
		Uintptr: value,
		Valid:   true,
	}
}

// NewUintptrPointerValue Создание нового действительного объекта Uintptr из ссылки на значение.
func NewUintptrPointerValue(ptr *uintptr) Uintptr {
	if ptr == nil {
		return NewUintptr()
	}
	return NewUintptrValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (u *Uintptr) SetValid(value uintptr) { u.Uintptr, u.Valid = value, true; u.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (u *Uintptr) Set(value uintptr) { u.Uintptr, u.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (u *Uintptr) Reset() { u.Uintptr, u.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (u *Uintptr) NullIfDefault() Uintptr {
	if u.Uintptr == 0 {
		u.Reset()
	}
	return *u
}

// MustValue Возвращает значение в любом случае.
func (u *Uintptr) MustValue() uintptr {
	if !u.Valid {
		return 0
	}
	return u.Uintptr
}

// Pointer Возвращает ссылку на значение.
func (u *Uintptr) Pointer() *uintptr {
	if !u.Valid {
		return nil
	}
	return &u.Uintptr
}

// Scan Реализация интерфейса Scanner.
func (u *Uintptr) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		u.Uintptr, u.Valid = 0, false
		return
	case uintptr:
		u.Uintptr = value.(uintptr)
	default:
		buf := asString(x)
		var n uint64
		n, err = strconv.ParseUint(buf, 10, 64)
		u.Uintptr = uintptr(n)
	}
	u.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uintptr) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return []byte(strconv.FormatUint(uint64(u.Uintptr), 10)), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *Uintptr) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		u.Uintptr, u.Valid = 0, false
		return
	case float64:
		err = json.Unmarshal(data, &u.Uintptr)
	case string:
		var str = x
		if len(str) == 0 {
			u.Valid = false
			return
		}
		var n uint64
		n, err = strconv.ParseUint(str, 10, 64)
		u.Uintptr = uintptr(n)
	case map[string]any:
		var rex *regexp.Regexp
		_, uiOK := x["Uintptr"].(float64)
		valid, validOK := x["Valid"].(bool)
		if !uiOK || !validOK {
			return fmt.Errorf(errUintptrObject, x["Uintptr"], x["Valid"])
		}
		rex = regexp.MustCompile(`"Uintptr"\s*:\s*([0-9]+)`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = u.UnmarshalText([]byte(arr[1]))
			u.Valid = valid
		}
		return
	default:
		err = fmt.Errorf(errUintptrUnmarshal, reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uintptr) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatUint(uint64(u.Uintptr), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (u *Uintptr) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		u.Uintptr, u.Valid = 0, false
		return
	case emptyString:
		u.Uintptr, u.Valid = 0, true
		return
	default:
		var n uint64
		n, err = strconv.ParseUint(str, 10, 64)
		u.Uintptr = uintptr(n)
	}
	u.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uintptr) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatUint(uint64(u.Uintptr), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *Uintptr) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.UintptrWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.UintptrWrapper)
	if err = dec.Decode(item); err == nil {
		u.Uintptr, u.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uintptr) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.UintptrWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.UintptrWrapper{Value: u.Uintptr, Valid: u.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
