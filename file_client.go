package HttpParallelSync

import (
    "os"
    "time"
)

type FileInfo struct {
    Name    string
    IsDir   bool
    Size    int64
    URL     string
    ModTime time.Time
    Mode int
}

type FilePartRequest struct {
    CurrentURIPath string
    FileInfo FileInfo
    StartByteRange uint64
    EndByteRange uint64
    DestinationFile string

}

type IFileClient interface {
    ListDirectoryContents(directory string) ([]FileInfo, error)
    GetFilePart(item FileInfo, startRange uint64, endRange uint64, destinationFile string ) (*os.File,  error)
}
