/*
* @Author: Hifun
* @Date: 2020/8/27 19:27
 */
package main

import (
    "archive/zip"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

/*NewBaseJsonBean用于创建一个struct对象:*/
type BaseJsonBean struct {

    Code    int         `json:"code"`

    Data    interface{} `json:"data"`

    Message string      `json:"message"`

}

func NewBaseJsonBean() *BaseJsonBean {
    return &BaseJsonBean{}
}

const (
    configName = "config.json"
    currentPath = "./path.txt"
)

var targetPath  = ""
func sayHello(w http.ResponseWriter, r *http.Request)  {
    w.Write([]byte("hello world!" + targetPath))
    return
}
func initCurrentPath()  {
    targetPath = loadConfig(currentPath)
}
func main()  {
    // 先初始化当前的路径
    initCurrentPath()
    fmt.Printf("server start success!\n Listening...\nCurrentPath:  "+targetPath)
    // 测试接口
    http.HandleFunc("/test",sayHello)
    //  注册上传压缩包的接口
    http.HandleFunc("/upload", uploadHandler)
    // 注册更新文件内容的接口
    http.HandleFunc("/updateContext",UpdateConfig)
    // 注册读取文件内容的接口
    http.HandleFunc("/getContext",getConfigContext)
    // 创建一个监听 使用默认的handler监听 8090端口
    err := http.ListenAndServe(":8090", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err.Error())
    }
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("\n enter func uploadHandler\n")
    w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
    w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
    w.Header().Set("content-type", "application/json")             //返回数据格式是json
    // 限制客户端上传文件的大小
    r.Body = http.MaxBytesReader(w, r.Body, 20*1024*1024)
    err := r.ParseMultipartForm(20 * 1024 * 1024)
    if err != nil {
       http.Error(w, err.Error(), http.StatusInternalServerError)
       return
    }
    // 获取上传的文件
    file, fileHeader, err := r.FormFile("uploadFile")
    // 检查文件类型
    ret := strings.HasSuffix(fileHeader.Filename, ".zip")
    if ret == false {
       http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    //r.ParseForm()
    appName := r.Form["appName"][0]
    targetPath := "./" + appName
    isExist,err := PathExists(targetPath)
    if err != nil {
        fmt.Printf("get dir error![%v]\n", err)
        return
    }
    if !isExist {
        // 创建文件夹
        err := os.Mkdir(targetPath, os.ModePerm)
        if err != nil {
            fmt.Printf("mkdir failed![%v]\n", err)
        } else {
            fmt.Printf("mkdir success!\n")
        }
    }
    // 写入文件
    dst, err := os.Create(targetPath + "/" + fileHeader.Filename)
    defer dst.Close()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer file.Close()
    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    result := NewBaseJsonBean()
    result.Code = 1000
    result.Message = "hifun upload success!"
    json, _ := json.Marshal(result)
    w.Write(json)
    return
}
func UpdateConfig(w http.ResponseWriter,r *http.Request)  {
    // 获取客户端POST方式传递的参数
    r.ParseForm()
    appName := r.Form["appName"]
    version := r.Form["version"]
    fmt.Println("hifun: getVersion:  ",version)
    updateConfig(appName[0],version[0])
    w.Header().Set("Access-Control-Allow-Origin", "*")
    result := NewBaseJsonBean()
    result.Code = 1000
    result.Message = "updateConfig success!"
    //result.Data = data
    w.Header().Set("Content-Type", "application/json")
    json, _ := json.Marshal(result)
    w.Write(json)
    return
}
func updateConfig(appName,version string)  {
    targetPath := "./" + appName + "/"
    // 解压缩文件
    zipFil,tarDir := targetPath + version + ".zip",targetPath
    Unzip(zipFil,tarDir)
    strTest := appName + "/" + version
    var strByte = []byte(strTest)
    err := ioutil.WriteFile(targetPath + configName, strByte, 0666)
    if err != nil {
        fmt.Println("write fail")
    }
    fmt.Println("write success")
}
func getConfigContext(w http.ResponseWriter,r *http.Request)  {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    // 获取客户端POST方式传递的参数
    r.ParseForm()
    appName := r.Form["appName"][0]
    targetConfig := "./" + appName + "/" + configName
    config := loadConfig(targetConfig)
    // 向客户端返回JSON数据
    data := make(map[string]interface{})
    data["url"] = targetPath + config+"/index.html"
    data["appName"] = appName
    temLen := len(config)
    appVersion := config[temLen-5: temLen]
    data["appVersion"] =appVersion
    result := NewBaseJsonBean()
    result.Code = 1000
    result.Message = "获取成功"
    result.Data = data
    w.Header().Set("Content-Type", "application/json")
    json, _ := json.Marshal(result)
    w.Write(json)
    return
}
//读取到file中，再利用ioutil将file直接读取到[]byte中, 这是最优
func loadConfig(targetConfig string) string {
    f, err := os.Open(targetConfig)
    if err != nil {
        fmt.Println("read file fail", err)
        return ""
    }
    defer f.Close()

    fd, err := ioutil.ReadAll(f)
    if err != nil {
        fmt.Println("read to fd fail", err)
        return ""
    }

    return string(fd)
}
// srcFile could be a single file or a directory
func Zip(srcFile string, destZip string) error {
    zipfile, err := os.Create(destZip)
    if err != nil {
        return err
    }
    defer zipfile.Close()

    archive := zip.NewWriter(zipfile)
    defer archive.Close()

    filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        header, err := zip.FileInfoHeader(info)
        if err != nil {
            return err
        }


        header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile) + "/")
        // header.Name = path
        if info.IsDir() {
            header.Name += "/"
        } else {
            header.Method = zip.Deflate
        }

        writer, err := archive.CreateHeader(header)
        if err != nil {
            return err
        }

        if ! info.IsDir() {
            file, err := os.Open(path)
            if err != nil {
                return err
            }
            defer file.Close()
            _, err = io.Copy(writer, file)
        }
        return err
    })

    return err
}
// unzip a zipFile to a directory
func Unzip(zipFile string, destDir string) error {
    zipReader, err := zip.OpenReader(zipFile)
    if err != nil {
        return err
    }
    defer zipReader.Close()

    for _, f := range zipReader.File {
        fpath := filepath.Join(destDir, f.Name)
        if f.FileInfo().IsDir() {
            os.MkdirAll(fpath, os.ModePerm)
        } else {
            if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
                return err
            }

            inFile, err := f.Open()
            if err != nil {
                return err
            }
            defer inFile.Close()

            outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer outFile.Close()

            _, err = io.Copy(outFile, inFile)
            if err != nil {
                return err
            }
        }
    }
    return nil
}
// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}