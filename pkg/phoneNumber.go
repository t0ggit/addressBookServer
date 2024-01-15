package pkg

import (
	"errors"
	"fmt"
	"strings"
)

const (
	allowedChars = "1234567890 +-()"
	ignoredChars = " +-()"
	neededLength = 11
	maxLength    = 20
)

// NormalizePhoneNumber преобразует номер в формат 8XXXXXXXXXX
// В случае некорректных данных возвращает "" и error.
// Некорректными данными считаются те номера, у которых:
//   - Количество символов меньше, чем neededLength или больше, чем maxLength
//   - Первый символ не содержится в allowedFirstChars
//   - Есть символы, не содержащиеся в allowedChars
//   - Меньше, чем neededLength цифр
func NormalizePhoneNumber(phoneNumber string) (normalizedPhoneNumber string, err error) {
	wErr := NewWrappedError("NormalizePhoneNumber()")

	// Проверка на максимальное количество символов
	if len(phoneNumber) > maxLength {
		err = errors.New(fmt.Sprintf("phoneNumber too long (max %d characters): %s", maxLength, phoneNumber))
		wErr.Specify(err, "len(phoneNumber) > maxLength").LogError()
		return "", err
	}

	// Проверка на минимальное количество символов после удаления крайних пробелов
	phoneNumber = strings.TrimSpace(phoneNumber)
	if len(phoneNumber) < neededLength {
		err = errors.New(fmt.Sprintf("phoneNumber too short (min %d characters): %s", neededLength, phoneNumber))
		wErr.Specify(err, "len(phoneNumber) < neededLength").LogError()
		return "", err
	}

	// Начинаем собирать нормальный номер
	normalizedPhoneNumberBuilder := strings.Builder{}

	// Обработка country code
	switch phoneNumber[0] {
	case '+':
		// После "+" должна быть "7", иначе ошибка
		if phoneNumber[1] == '7' {
			normalizedPhoneNumberBuilder.WriteRune('8')
			phoneNumber = phoneNumber[2:]
		} else {
			err = errors.New("invalid country code: " + phoneNumber)
			wErr.Specify(err, "switch phoneNumber[0] { -> case '+'").LogError()
			return "", err
		}
	case '8', '7':
		normalizedPhoneNumberBuilder.WriteRune('8')
		phoneNumber = phoneNumber[1:]
	default:
		// Некорректный первый символ номера
		err = errors.New("invalid first character in phoneNumber")
		wErr.Specify(err, "switch phoneNumber[0] { -> default").LogError()
		return "", err
	}

	// Проходимся по всем остальным символам номера
	for _, char := range phoneNumber {
		// Случай недопустимого символа (возвращаем ошибку)
		if strings.IndexRune(allowedChars, char) == -1 {
			err = errors.New("invalid character in phoneNumber: " + string(char))
			wErr.Specify(err, "strings.IndexRune(allowedChars, char) == -1").LogError()
			return "", err
		}

		// Случай игнорируемого символа
		if strings.IndexRune(ignoredChars, char) != -1 {
			continue
		}

		// Случай значащего символа
		normalizedPhoneNumberBuilder.WriteRune(char)
	}

	normalizedPhoneNumber = normalizedPhoneNumberBuilder.String()

	if len(normalizedPhoneNumber) != neededLength {
		err = errors.New(fmt.Sprintf("invalid phoneNumber length (need %d characters): %s", neededLength, normalizedPhoneNumber))
		wErr.Specify(err, "len(normalizedPhoneNumber) != neededLength").LogError()
		return "", err
	}

	return normalizedPhoneNumber, nil
}
