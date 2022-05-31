package get

import (
	"Storage/V4/apiServer/locate"
	"Storage/lib/objectstream"
	"fmt"
	"io"
)

func getStream(object string) (io.Reader, error) {
	server := locate.Locate(object)

	if server == "" {
		return nil, fmt.Errorf("object %S locate fail", object)
	}

	return objectstream.NewGetStream(server, object)

}
