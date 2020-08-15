package marketprocessor

type Processor interface {
	Process() (done <-chan bool, err error)
	Recover()
}
