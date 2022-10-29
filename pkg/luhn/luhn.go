package luhn

// Check - проверка номера на соответствие алгоритму Луна.
func Check(number string) bool {
	l := len([]rune(number))
	if l < 2 {
		// Номер должен содержать не менее 2 цифр
		return false
	}
	sum := 0
	for pos, chr := range number {
		dig := int(chr - '0')
		if dig < 0 || dig > 9 {
			// Если символ не цифра, возвращаем false.
			return false
		}
		if pos%2 == l%2 {
			// Если позиция четная и длина четная,
			// либо позиция нечетная и длина нечетная,
			// то умножаем цифру на 2.
			dig *= 2
			if dig > 9 {
				// Если результат превышает 9, то вычитаем 9 из него
				dig -= 9
			}
		}
		// Суммируем все цифры
		sum += dig
	}
	// Проверяем, что сумма делится на 10
	return sum%10 == 0
}
