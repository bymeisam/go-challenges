package main

func GetString(v interface{}) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

func TypeSwitch(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case int:
		return "int"
	case bool:
		return "bool"
	default:
		return "unknown"
	}
}

func main() {}
