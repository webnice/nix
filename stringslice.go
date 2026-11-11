package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/webnice/nix/v2/wrapper"
)

// StringSlice Объект []string, допускающий значение nil.
// Значение совместимо с типом данных set баз данных: слайс строк при записи
// в базу данных кодируется строкой с перечислением элементов через запятую,
// пустой слайс — пустой строкой, нулевой объект — NULL. Чтение выполняется
// из строкового представления set, элементы разделяются запятой, запятая
// и обратная косая черта в элементах экранируются обратной косой чертой
// и восстанавливаются при чтении.
type StringSlice struct {
	StringSlice []string // Значение объекта.
	Valid       bool     // Valid равен true если значение не nil.
}

// NewStringSlice Создание нового объекта StringSlice.
func NewStringSlice() StringSlice {
	return StringSlice{
		StringSlice: []string{},
		Valid:       false,
	}
}

// NewStringSliceValue Создание нового действительного объекта StringSlice из значения.
func NewStringSliceValue(value []string) StringSlice {
	return StringSlice{
		StringSlice: append([]string{}, value...),
		Valid:       true,
	}
}

// NewStringSlicePointerValue Создание нового действительного объекта StringSlice из ссылки на значение.
func NewStringSlicePointerValue(ptr *[]string) StringSlice {
	if ptr == nil {
		return NewStringSlice()
	}
	return NewStringSliceValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (ss *StringSlice) SetValid(value []string) {
	ss.StringSlice, ss.Valid = append([]string{}, value...), true
	ss.NullIfDefault()
}

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (ss *StringSlice) Set(value []string) {
	ss.StringSlice, ss.Valid = append([]string{}, value...), true
}

// Reset Сброс значения и установка флага не действительного значения.
func (ss *StringSlice) Reset() {
	ss.StringSlice, ss.Valid = ss.StringSlice[:0], false
}

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (ss *StringSlice) NullIfDefault() StringSlice {
	if len(ss.StringSlice) == 0 {
		ss.Reset()
	}
	return *ss
}

// MustValue Возвращает значение в любом случае.
func (ss *StringSlice) MustValue() []string {
	if !ss.Valid {
		return []string{}
	}
	return ss.StringSlice
}

// Pointer Возвращает ссылку на значение.
func (ss *StringSlice) Pointer() *[]string {
	if !ss.Valid {
		return nil
	}
	return &ss.StringSlice
}

// Scan Реализация интерфейса Scanner.
func (ss *StringSlice) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		ss.Reset()
		return
	case []byte:
		ss.parse(string(x))
	case string:
		ss.parse(x)
	default:
		err = fmt.Errorf(errStringSliceScan, x, value)
	}
	ss.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (ss StringSlice) Value() (driver.Value, error) {
	if !ss.Valid {
		return nil, nil
	}
	return encodeSet(ss.StringSlice), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (ss *StringSlice) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		ss.Valid = false
		return
	case []any:
		err = ss.parseSlice(x)
	case map[string]any:
		return ss.parseObject(x)
	default:
		err = fmt.Errorf(errStringSliceUnmarshal, v)
	}
	ss.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (ss StringSlice) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !ss.Valid {
		data = []byte(nullString)
		return
	}
	data, err = json.Marshal(ss.StringSlice)

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (ss *StringSlice) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)

	switch str := string(text); str {
	case nullString:
		ss.Reset()
	case emptyString:
		ss.StringSlice, ss.Valid = []string{}, true
	default:
		ss.parse(str)
		ss.Valid = true
	}

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (ss StringSlice) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !ss.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(encodeSet(ss.StringSlice))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (ss *StringSlice) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.StringSliceWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.StringSliceWrapper)
	if err = dec.Decode(item); err == nil {
		ss.StringSlice, ss.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (ss StringSlice) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.StringSliceWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.StringSliceWrapper{Value: ss.StringSlice, Valid: ss.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}

// parse Разбор строки формата set базы данных в значение объекта.
// Пустая строка — пустой набор, элементы разделяются запятыми,
// экранированные символы восстанавливаются.
func (ss *StringSlice) parse(str string) {
	const emptyString = ""
	if str == emptyString {
		ss.StringSlice = []string{}
		return
	}
	ss.StringSlice = decodeSetString(str)
}

// parseSlice Разбор массива JSON строк в значение объекта.
func (ss *StringSlice) parseSlice(x []any) (err error) {
	var ok bool

	ss.StringSlice = make([]string, 0, len(x))
	for i := range x {
		if _, ok = x[i].(string); !ok {
			return fmt.Errorf(errStringSliceUnmarshal, x[i])
		}
		ss.StringSlice = append(ss.StringSlice, x[i].(string))
	}

	return
}

// parseObject Разбор объекта {"StringSlice":…,"Valid":…} в значение объекта.
func (ss *StringSlice) parseObject(x map[string]any) (err error) {
	var (
		value []any
		valid bool
		ok    bool
	)

	if value, ok = x["StringSlice"].([]any); !ok {
		return fmt.Errorf(errStringSliceObject, x["StringSlice"], x["Valid"])
	}
	if valid, ok = x["Valid"].(bool); !ok {
		return fmt.Errorf(errStringSliceObject, x["StringSlice"], x["Valid"])
	}
	if err = ss.parseSlice(value); err == nil {
		ss.Valid = valid
	}

	return
}

// encodeSet Кодирование слайса строк в строку формата set базы данных:
// элементы перечисляются через запятую, запятая и обратная косая черта
// в элементах экранируются обратной косой чертой.
func encodeSet(value []string) string {
	var encoded = make([]string, 0, len(value))
	for i := range value {
		encoded = append(encoded, encodeSetString(value[i]))
	}

	return strings.Join(encoded, ",")
}

// encodeSetString Экранирование элемента строки формата set базы данных.
func encodeSetString(str string) string {
	if !strings.ContainsAny(str, `,\`) {
		return str
	}
	str = strings.ReplaceAll(str, `\`, `\\`)
	str = strings.ReplaceAll(str, `,`, `\,`)

	return str
}

// decodeSetString Разбор строки формата set базы данных в слайс строк:
// элементы разделяются запятыми, экранированные запятая и обратная косая
// черта восстанавливаются, прочие последовательности сохраняются как есть.
func decodeSetString(str string) []string {
	var (
		items   = make([]string, 0, strings.Count(str, ",")+1)
		current strings.Builder
	)

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '\\':
			if i+1 < len(str) && (str[i+1] == '\\' || str[i+1] == ',') {
				i++
			}
			current.WriteByte(str[i])
		case ',':
			items = append(items, current.String())
			current.Reset()
		default:
			current.WriteByte(str[i])
		}
	}

	return append(items, current.String())
}
