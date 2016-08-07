package HttpParallelSync

import (
    "os"
    "io"
    "log"
    "fmt"
)

//Flow
//Get Dir listing
//Create all the Dirs
//Download all the files in the current dir
//for each dir, recurse the above

const ParallelismSizeMinimum = 10 * 1024 * 1024 //10MB

func Sync(caddy *CaddyClient, currentDirectory string, parallelism int) error {
    log.Printf("Starting sync on %s\n", currentDirectory)
    files, err := caddy.ListDirectoryContents(currentDirectory)

    if (err != nil) {
        return err
    }
    err = CreateDirectories(currentDirectory, files)
    if (err != nil) {
        return err
    }
    err = DownloadFiles(caddy, currentDirectory, files)
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
            Sync(caddy, nextSyncDir, parallelism)
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

func DownloadFiles(caddy *CaddyClient, currentDirectory string, files []FileInfo) error {
    for i:=0;i<len(files);i++ {
        file := files[i]
        if file.IsDir { continue}

        var fileRelative string
        if (currentDirectory == "") {
            fileRelative = file.Name
        } else {
            fileRelative = fmt.Sprintf("%s/%s", currentDirectory, file.Name)
        }

        //if the file exists, and it's the same size, skip it
        //if the file exists and it's a different size, but our file has a later modified time, skip it.
        //otherwise, delete the existing file and re-download.
        if existingFileInfo, err := os.Stat(fileRelative); os.IsExist(err) {
            if (existingFileInfo.Size() == file.Size) {
                log.Printf("Skipping - same size: %s", fileRelative)
                continue;
            }
            if (existingFileInfo.ModTime().After(file.ModTime)){
                log.Printf("Skipping - local file mod time is after remote file: %s", fileRelative)
                continue;
            }

            err := os.Remove(fileRelative)
            if (err != nil) {
                return err
            }
        }

        var err error
        if (file.Size <= ParallelismSizeMinimum) {
            err = DownloadFile(caddy, fileRelative, file)

        } else {
            err = ParallelDownloadFile(caddy, currentDirectory, file, 3)
        }
        if (err != nil) {
            return err
        }
    }

    return nil
}


func DownloadFile(caddy *CaddyClient, destinationFile string, file FileInfo) error {
    fileRequest := FilePartRequest{
        FileInfo: file,
        StartByteRange: 0,
        EndByteRange: 0,
        DestinationFile: destinationFile,
    }
    return caddy.GetFilePart(fileRequest )
}

func ParallelDownloadFile (caddy *CaddyClient, destinationFile string, file FileInfo, parallelism int) error {

    return DownloadFile(caddy, destinationFile, file)
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