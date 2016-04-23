package resource // import "engo.io/engo/resource"

type Resource interface {
	Kind() string
	Name() string
	URL() string
}
