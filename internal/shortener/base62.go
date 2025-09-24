package shortener

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"


func ToBase62(num int64) string {
	encoded := ""
	for num > 0 {
		encoded = string(base62Chars[num%62]) + encoded
		num /= 62
	}

	if encoded == "" {
		return "0"
	}
	return encoded
}
