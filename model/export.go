package model

type Export interface {
	Export() (interface{}, error)
}
