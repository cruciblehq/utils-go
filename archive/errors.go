package archive

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrWriteFailed         = crex.New("archive write failed")
	ErrExtractFailed       = crex.New("archive extraction failed")
	ErrReadFailed          = crex.New("archive read failed")
	ErrInvalidPath         = crex.New("invalid archive path")
	ErrUnsupportedFileType = crex.New("unsupported file type")
	ErrUnsupportedFormat   = crex.New("unsupported archive format")
)
