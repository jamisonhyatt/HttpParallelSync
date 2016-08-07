package HttpParallelSync

import (
    "os"
    "io"
    "log"
    "fmt"
    "github.com/sdming/gosnow"
    "strconv"
    "golang.org/x/sync/errgroup"
)

//Flow
//Get Dir listing
//Create all the Dirs
//Download all the files in the current dir
//for each dir, recurse the above

const ParallelismSizeMinimum = 10 * 1024 * 1024 //10MB

func Sync(caddy *CaddyClient, currentDirectory string, parallelism int) error {
    workDir, _ := os.Getwd()
    log.Printf("Starting sync on %s into %s/%s\n", currentDirectory, workDir, currentDirectory )
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
        existingFileInfo, err := os.Stat(fileRelative)

        if (err == nil) {
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
        //reset err, as file doesn't exist threw err above
        err = nil

        if (file.Size <= ParallelismSizeMinimum) {
            err = DownloadFile(fileRelative, caddy,currentDirectory, file)

        } else {
            err = ParallelDownloadFile(fileRelative, caddy,currentDirectory, file, 3)
        }
        if (err != nil) {
            return err
        }
    }

    return nil
}


func DownloadFile(destinationFile string, caddy *CaddyClient, currentURIPath string, file FileInfo) error {
    fileRequest := FilePartRequest{
        CurrentURIPath: currentURIPath,
        FileInfo: file,
        StartByteRange: 0,
        EndByteRange: 0,
        DestinationFile: destinationFile,
    }
    return caddy.GetFilePart(fileRequest )
}

func ParallelDownloadFile (destinationFile string, caddy *CaddyClient, currentURIPath string, file FileInfo, parallelism int) error {
    v, err := gosnow.Default()
    if (err != nil) {
        return err
    }

    chunkSize := uint64(file.Size) / uint64(parallelism)
    fileRequests := make([]FilePartRequest, parallelism)

    var group errgroup.Group

    chunkStart := uint64(0)
    chunkThrough := chunkSize
    for i := 0; i < parallelism; i++ {
        tempFileNameFlake, err := v.Next()
        if err != nil {
            return err
        }


        if (i == len(fileRequests) -1) {
            chunkThrough = 0
        }
        var tempFile string
        if (currentURIPath == "") {
            tempFile = strconv.FormatUint(tempFileNameFlake, 10)
        } else {
            tempFile = fmt.Sprintf("%s/%s", currentURIPath, strconv.FormatUint(tempFileNameFlake, 10))
        }

        request := FilePartRequest{
            CurrentURIPath: currentURIPath,
            FileInfo: file,
            StartByteRange: chunkStart,
            EndByteRange: chunkThrough,
            DestinationFile: tempFile,
        }
        fileRequests[i] = request

        chunkStart += chunkSize + uint64(1)
        chunkThrough = chunkStart + chunkSize

        group.Go(func() error {
            return caddy.GetFilePart(request)
        })
    }

    //wait for goroutines
    err = group.Wait()
    if (err != nil ){
        DeleteTempFiles(fileRequests)
        return err
    }

    err = Combine(destinationFile, fileRequests)
    if (err != nil) {
        return err
    }

    return nil
}

func DeleteTempFiles(requests []FilePartRequest) {

    for _, request := range requests {
        _, err := os.Stat(request.DestinationFile)
        if (err != nil) {
            os.Remove(request.DestinationFile)
        }
    }
}

func Combine(destinationFile string, files []FilePartRequest) error {
    destination, err := os.Create(destinationFile)
    if (err != nil) {
        return err
    }
    defer destination.Close()

    for _, file := range(files) {
        log.Printf("reading %s", file.DestinationFile)
        reader, err := os.Open(file.DestinationFile)
        if (err != nil) {
            return err
        }
        _, err = io.Copy(destination, reader)
        if (err != nil) {
            return err
        }
        err  = os.Remove(file.DestinationFile)
        if (err != nil) {
            return err
        }
    }


    return nil
}