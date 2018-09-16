package common

type Source rune

const (
	BCpEvtSrcWUI Source = 'w'
)

type BCpEvent struct {
	Source Source
	Data   interface{}
}
