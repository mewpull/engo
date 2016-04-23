package audio // import "engo.io/engo/assets/audio"

import "io"

// ReadSeekCloser is an io.ReadSeeker and io.Closer.
type ReadSeekCloser interface {
	io.ReadSeeker
	io.Closer
}
