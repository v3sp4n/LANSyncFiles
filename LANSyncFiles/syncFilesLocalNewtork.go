package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"net/url"
	"regexp"
	"time"
	"fmt"
	"os"
)

func EncodeBytesToBase64(data []byte) string {
    return base64.StdEncoding.EncodeToString(data)
}
func DecodeBase64ToBytes(encodedString string) ([]byte, error) {
    decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
    if err != nil {
        return nil, err
    }
    return decodedBytes, nil
}

func main() {
	ips := map[string][][]string{}
	if _,err := os.Stat("ip and folders for sync/"); err != nil {
		os.Mkdir("ip and folders for sync/", os.ModeDir)
	}
	files,_ := ioutil.ReadDir("ip and folders for sync/")
	for _,v := range files {
		if m := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+)$`).FindStringSubmatch(v.Name()); !v.IsDir() && len(m) >= 4 {
			ips[v.Name()] = [][]string{}
			content,_ := ioutil.ReadFile("ip and folders for sync/"+v.Name())	
			contentSplit := strings.Split(string(content),"\n")
			for i := 0; i < len(contentSplit); i++ {
				if path := regexp.MustCompile(`^(\S+)\s+(\S+)$`).FindStringSubmatch(contentSplit[i]); len(path) >= 2 {
					if IsComment := regexp.MustCompile(`^#(\S+)`).FindStringSubmatch(contentSplit[i]); len(IsComment) == 0 {
						
						ips[v.Name()] = append(ips[v.Name()],[]string{
							path[1],path[2],
						}) 
					}
				}
			}
		}
	}

	go server()

	for {
		for ip,v := range ips {
			for slice := 0; slice < len(v); slice++{
				r,err := http.PostForm("http://"+ip+":8080/getFiles",url.Values{
					"path": {ips[ip][slice][0]},
				})
				if err != nil {
					fmt.Println("[PostForm /getFiles error]",err)
				} else {
					text,_ := ioutil.ReadAll(r.Body)
					syncFiles := []string{}
					myFiles := []string{}
					json.Unmarshal(text,&syncFiles)

					myFilesDir,_ := ioutil.ReadDir(ips[ip][slice][1])
					for _,v := range myFilesDir {
						if !v.IsDir() {
							myFiles = append(myFiles,v.Name())
						}
					}

					for s := 0; s < len(syncFiles); s++ {
						find := false
						for i := 0; i < len(myFiles); i++ {
							if syncFiles[s] == myFiles[i] {
								find = true
							}
						}
						if !find {
							r,err := http.PostForm("http://"+ip+":8080/download", url.Values{
								"path": {fmt.Sprintf("%s/%s",ips[ip][slice][0],syncFiles[s])},
							})
							if err != nil {
								fmt.Println("[PostForm /download error]",err)
							} else {
								fmt.Println("[PostForm /download info]sync",syncFiles[s])
								body,_ := ioutil.ReadAll(r.Body)
								content,_ := DecodeBase64ToBytes(string(body))
								ioutil.WriteFile(fmt.Sprintf("%s/%s",ips[ip][slice][1],syncFiles[s]), content, 0644)
							}
						}
					}

				}
			}
			time.Sleep(1 * time.Second)
		}
		time.Sleep(5 * time.Second)
	}


}

func server() {
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		fmt.Println("[HandleFunc /download info]path",r.PostForm.Get("path"))

		content,_ := ioutil.ReadFile(r.PostForm.Get("path"))
		w.Write([]byte(EncodeBytesToBase64(content)))

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
	})
	http.HandleFunc("/getFiles", func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		fmt.Println("[HandleFunc /getFiles info]path",r.PostForm.Get("path"))
		path := r.PostForm.Get("path")

		if _,err := os.Stat(path); err != nil {
			http.Error(w,"path not exits!",404)
			return
		}

		myFiles := []string{}
		files,_ := ioutil.ReadDir(path)
		for _,v := range files {
			if !v.IsDir() {
				myFiles = append(myFiles,v.Name())
			}
		}
		j,_ := json.Marshal(myFiles)
		w.Write(j)

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
	})
	http.ListenAndServe(":8080", nil)
}