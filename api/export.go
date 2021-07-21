// todo: #20 move this to api package
package api

type Export interface {
	Export() (interface{}, error)
}
