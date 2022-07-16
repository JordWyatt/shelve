package shelve

import "os"

var (
	STAGING_DIRECTORY = os.Getenv("SHELVE_STAGING_DIRECTORY")
	TARGET_DIRECTORY  = os.Getenv("SHELVE_TARGET_DIRECTORY")
)

type Directory struct {
	Name string
	Path string
}
