package main

type DataProcessor interface {
	ReadData() string
	ProcessData(data string) string
	SaveData(data string) string
}

type BaseProcessor struct {
	processor DataProcessor
}

func (b *BaseProcessor) Execute() string {
	data := b.processor.ReadData()
	processed := b.processor.ProcessData(data)
	return b.processor.SaveData(processed)
}

type CSVProcessor struct{}

func (c CSVProcessor) ReadData() string {
	return "CSV data"
}

func (c CSVProcessor) ProcessData(data string) string {
	return "Processed " + data
}

func (c CSVProcessor) SaveData(data string) string {
	return "Saved " + data
}

func main() {}
