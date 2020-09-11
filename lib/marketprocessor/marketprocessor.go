package marketprocessor

type Processor interface {
	Process() <-chan bool
}
