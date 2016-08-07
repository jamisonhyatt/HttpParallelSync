package HttpParallelSync

import (
    "net/http"
    "os"
    "io"
    "strconv"
    "fmt"
    "encoding/json"

)

//type MultiDownloadRequest struct {
//    URL
//}

type CaddyClient struct {
    Host        string
    Port        int
    Ssl         bool
    HttpClient  *http.Client
    BaseURI     string
    WorkingDir  string
    Parallelism int
}

//Flow
//Get Dir listing
//Create all the Dirs
//Download all the files in the current dir
//for each dir, recurse the above



func (caddy *CaddyClient)  ListDirectoryContents(directory string) ([]FileInfo, error) {

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/",caddy.BaseURI, directory),nil)
    if (err != nil) {
        return nil, err
    }
    req.Header.Add("Accept", "application/json")

    resp, err := caddy.HttpClient.Do(req)
    if (err != nil ) {
        return nil, err
    }
    if (resp.StatusCode != http.StatusOK ) {
        return nil, fmt.Errorf("Expected status OK, received Status: %s", resp.StatusCode)
    }

    var fileList []FileInfo
    err = json.NewDecoder(resp.Body).Decode(&fileList)
    if (err != nil ) {
        return nil, err
    }

    return fileList, nil
}




func (caddy *CaddyClient) GetFilePart(filePartRequest FilePartRequest ) (*os.File,  error){
    filePart, err := os.Create(filePartRequest.DestinationFileName)
    if (err != nil) {
        return nil, err
    }
    defer filePart.Close()

    reqURI := fmt.Sprintf("%s/%s",caddy.BaseURI, filePartRequest.CurrentDirectory + filePartRequest.FileInfo.URL[1:len(filePartRequest.FileInfo.URL)])
    fmt.Printf("Request URI: %s\n", reqURI)
    req, err := http.NewRequest("GET", reqURI,nil)

    AddRangeHeader(req, filePartRequest.StartByteRange, filePartRequest.EndByteRange)

    resp, err := caddy.HttpClient.Do(req)
    if (err != nil) {
        return nil, err
    }
    defer resp.Body.Close()

    if (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent) {
        err := fmt.Errorf("Expected OK or  PartialContent, received %v", resp.StatusCode)
        return nil, err
    }

    _, err = io.Copy(filePart, resp.Body)
    if (err != nil) {
        return nil, err
    }

    return filePart, nil
}

func AddRangeHeader (req *http.Request, startRange uint64, endRange uint64)  {
    var startRangeString string
    var endRangeString string

    startRangeString = strconv.FormatUint(startRange, 10)
    if endRange == 0 {
        //Get the entire file
        endRangeString = ""

    } else  {
        endRangeString = strconv.FormatUint(endRange,10)
    }

    req.Header.Add("Range", fmt.Sprintf("bytes=%s-%s",startRangeString, endRangeString))

}