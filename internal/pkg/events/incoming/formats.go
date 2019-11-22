package incoming

type EventMetricDataFormat interface {
	GetIndexName() string
	ParseMessage(message string) error
}
