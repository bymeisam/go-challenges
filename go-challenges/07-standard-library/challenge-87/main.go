package main

import "encoding/xml"

type Book struct {
	XMLName xml.Name `xml:"book"`
	Title   string   `xml:"title"`
	Author  string   `xml:"author"`
}

func MarshalXML(book Book) ([]byte, error) {
	return xml.Marshal(book)
}

func UnmarshalXML(data []byte) (Book, error) {
	var book Book
	err := xml.Unmarshal(data, &book)
	return book, err
}

func main() {}
