package main

func Hello(hello string) string {
	if hello == "123" {
		return "456"
	}

	return hello
}

func main() {
	Hello("abc")
	Hello("123")
}
