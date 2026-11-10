package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"

	"github.com/webnice/nix/wrapper"
)

// RawMessage Объект json.RawMessage, допускающий значение nil.
type RawMessage struct {
	RawMessage json.RawMessage // Значение объекта.
	Valid      bool            // Valid равен true если значение не nil.
}

// NewRawMessage Создание нового объекта RawMessage.
func NewRawMessage() RawMessage {
	return RawMessage{
		RawMessage: nil,
		Valid:      false,
	}
}

// NewRawMessageValue Создание нового действительного объекта RawMessage из значения.
func NewRawMessageValue(value json.RawMessage) RawMessage {
	return RawMessage{
		RawMessage: append(json.RawMessage(nil), value...),
		Valid:      true,
	}
}

// NewRawMessagePointerValue Создание нового действительного объекта RawMessage из ссылки на значение.
func NewRawMessagePointerValue(ptr *json.RawMessage) RawMessage {
	if ptr == nil {
		return NewRawMessage()
	}
	return NewRawMessageValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
func (s *RawMessage) SetValid(value json.RawMessage) {
	s.RawMessage, s.Valid = append(s.RawMessage[:0], value...), true
}

// Reset Сброс значения и установка флага не действительного значения.
func (s *RawMessage) Reset() {
	s.RawMessage, s.Valid = nil, false
}

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (s *RawMessage) NullIfDefault() RawMessage {
	if len(s.RawMessage) == 0 {
		s.Reset()
	}
	return *s
}

// MustValue Возвращает значение в любом случае.
func (s *RawMessage) MustValue() json.RawMessage {
	if !s.Valid {
		return nil
	}
	return s.RawMessage
}

// Pointer Возвращает ссылку на значение.
func (s *RawMessage) Pointer() *json.RawMessage {
	if !s.Valid {
		return nil
	}
	return &s.RawMessage
}

// Scan Реализация интерфейса Scanner.
func (s *RawMessage) Scan(value interface{}) (err error) {
	if value == nil {
		s.Reset()
		return
	}
	s.RawMessage = json.RawMessage(asString(value))
	s.Valid = true

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (s RawMessage) Value() (driver.Value, error) {
	if !s.Valid {
		return nil, nil
	}
	return []byte(s.RawMessage), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (s *RawMessage) UnmarshalJSON(data []byte) (err error) {
	const nullString = "null"

	if len(data) == 0 || string(data) == nullString {
		s.RawMessage, s.Valid = nil, false
		return
	}
	s.RawMessage = append(s.RawMessage[:0], data...)
	s.Valid = true

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (s RawMessage) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !s.Valid || len(s.RawMessage) == 0 {
		data = []byte(nullString)
		return
	}
	data = s.RawMessage

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (s *RawMessage) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)

	switch str := string(text); str {
	case nullString:
		s.RawMessage, s.Valid = nil, false
		return
	case emptyString:
		s.RawMessage, s.Valid = json.RawMessage{}, true
		return
	default:
		s.RawMessage = append(s.RawMessage[:0], text...)
		s.Valid = true
	}

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (s RawMessage) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !s.Valid {
		text = []byte(nullString)
		return
	}
	text = s.RawMessage

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (s *RawMessage) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.RawMessageWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.RawMessageWrapper)
	if err = dec.Decode(item); err == nil {
		s.RawMessage, s.Valid = append(s.RawMessage[:0], item.Value...), item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (s RawMessage) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.RawMessageWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.RawMessageWrapper{Value: s.RawMessage, Valid: s.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
