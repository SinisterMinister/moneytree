package marketprocessor

type Processor interface {
	Process(stop <-chan bool) (done <-chan bool, err error)
}
