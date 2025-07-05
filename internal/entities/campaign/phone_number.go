package campaign

import (
	"regexp"
)

// PhoneNumber представляет номер телефона как value object
type PhoneNumber struct {
	value string
}

// NewPhoneNumber создает новый номер телефона после валидации
func NewPhoneNumber(phone string) (*PhoneNumber, error) {
	normalized := normalizePhone(phone)
	if !isValidPhone(normalized) {
		return nil, ErrInvalidPhoneNumber
	}
	return &PhoneNumber{value: normalized}, nil
}

// Equal проверяет равенство двух номеров телефонов
func (p *PhoneNumber) Equal(other *PhoneNumber) bool {
	if other == nil {
		return false
	}
	return p.value == other.value
}

// Value возвращает строковое представление номера
func (p *PhoneNumber) Value() string {
	return p.value
}

// String возвращает строковое представление номера (для fmt.Stringer)
func (p *PhoneNumber) String() string {
	return p.value
}

// normalizePhone удаляет все нецифровые символы из строки
func normalizePhone(s string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(s, "")
}

var phoneRegexp = regexp.MustCompile(`^7\d{10}$`)

// isValidPhone проверяет, что строка — валидный российский номер (11 цифр, начинается с 7)
func isValidPhone(s string) bool {
	return phoneRegexp.MatchString(s)
}
