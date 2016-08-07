package HttpParallelSync

import (
    "os"
    "io"
    "log"
    "fmt"
    "net/http"
    "encoding/json"
)

func Sync(currentDirectory string) error {
    log.Printf("Starting sync on %s\n", currentDirectory)
    files, err := ListDirectoryContents(currentDirectory)

    if (err != nil) {
        return err
    }
    err = CreateDirectories(currentDirectory, files)
    if (err != nil) {
        return err
    }
    err = DownloadFiles(currentDirectory, files)
    if (err != nil) {
        return err
    }

    for i:=0; i<len(files); i++ {
        if (files[i]).IsDir {
            var nextSyncDir string
            if (currentDirectory == "") {
                nextSyncDir = fmt.Sprintf("%s", files[i].Name)
            } else {
                nextSyncDir = fmt.Sprintf("%s/%s", currentDirectory, files[i].Name)
            }
            caddy.Sync(nextSyncDir)
        }
    }

    return nil
}



func CreateDirectories(baseDir string, files []FileInfo) error {

    for i := 0; i < len(files); i++ {
        f := files[i]
        if (f.IsDir) {
            var dirToCreate string
            if (baseDir == "") {
                dirToCreate = fmt.Sprintf("%s", f.Name)
            } else {
                dirToCreate = fmt.Sprintf("%s/%s", baseDir, f.Name)
            }
            workDir, _ := os.Getwd()

            if _, err := os.Stat(dirToCreate); os.IsNotExist(err) {
                log.Printf("Working directory = %s, creating Dir: %s\n",workDir , dirToCreate)
                err := os.MkdirAll(dirToCreate, 0755)
                if (err != nil) {
                    return err
                }
            }
        }
    }
    return nil
}

func DownloadFiles(currentDirectory string, files []FileInfo) error {
    for i:=0;i<len(files);i++ {
        file := files[i]
        if (!file.IsDir) {
            fileRequest := FilePartRequest{
                CurrentDirectory: currentDirectory,
                FileInfo: file,
                StartByteRange: 0,
                EndByteRange: 0,
                DestinationFileName: fmt.Sprintf("%s/%s", currentDirectory, file.Name),
            }
            caddy.GetFilePart(fileRequest )
        }

    }
    return nil
}

func Combine(destinationFile string, files []string) error {
    destination, err := os.Create(destinationFile)
    if (err != nil) {
        return err
    }
    defer destination.Close()

    for i := 0; i < len(files); i++ {
        log.Printf("reading %s", files[i])
        reader, err := os.Open(files[i])
        if (err != nil) {
            return err
        }
        io.Copy(destination, reader)
        os.Remove(files[i])
    }


    return nil
}