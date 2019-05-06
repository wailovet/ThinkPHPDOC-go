package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/issue9/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var mainDir = "Application"

func pathToUrlPre(path string) string {
	path = strings.ReplaceAll(path, `\`, "/")
	item := strings.Split(path, mainDir)
	if len(item) == 2 {
		item = strings.Split(item[1], "/")
		if len(item) == 4 {
			return "/" + item[1] + "/" + strings.ReplaceAll(strings.ReplaceAll(item[3], ".php", ""), "Controller.class", "") + "/"
		}
	}
	return ""
}

type fileDocInfo struct {
	Name  string
	Api   string
	Query string
	Doc   string
}

type phpDocMenuValue struct {
	Name        string     `json:"name"`
	File        string     `json:"file"`
	FileContent phpDocFile `json:"-"`
}
type phpDocMenu struct {
	Name  string            `json:"name"`
	Key   string            `json:"key"`
	Value []phpDocMenuValue `json:"value"`
}
type phpDocFile struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	DocUrl  string `json:"doc_url"`
	Desc    string `json:"desc"`
	Example string `json:"example"`
}

func pathToUrlEnd(path string) (string, []fileDocInfo, error) {
	tmp, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("pathToUrlEnd:< ", path, "> err:", err.Error())
	}
	if strings.Index(string(tmp), "@api") == -1 {
		return "", nil, errors.New("略过")
	}
	data := strings.ReplaceAll(string(tmp), "\r", "")
	dataList := strings.Split(data, "\n")
	var result []fileDocInfo
	mark := ""

	var classDoc = ""
	for e := range dataList {
		if strings.Index(dataList[e], "*") > -1 {
			mark += dataList[e]
		}

		if strings.Index(dataList[e], "@api") > -1 {
			apitmp := strings.Split(dataList[e], "@api")
			classDoc = strings.ReplaceAll(apitmp[1], " ", "")
		}
		if strings.Index(dataList[e], "class") > -1 {
			if strings.Index(dataList[e], "extends") > -1 {
				mark = ""
			}
		}
		if strings.Index(dataList[e], "private") > -1 {
			if strings.Index(dataList[e], "function") > -1 {
				mark = ""
			}
		}
		if strings.Index(dataList[e], "public") > -1 {
			dataList[e] = strings.ReplaceAll(dataList[e], "public", "")
			if strings.Index(dataList[e], "function") > -1 {

				dataList[e] = strings.ReplaceAll(dataList[e], "function", "")
				dataList[e] = strings.ReplaceAll(dataList[e], "{", "")
				dataList[e] = strings.ReplaceAll(dataList[e], "}", "")
				dataList[e] = strings.ReplaceAll(dataList[e], " ", "")
				dataList[e] = strings.ReplaceAll(dataList[e], "($", "?")
				dataList[e] = strings.ReplaceAll(dataList[e], "(", "")
				dataList[e] = strings.ReplaceAll(dataList[e], ")", "")
				dataList[e] = strings.ReplaceAll(dataList[e], ",$", "&")
				dataListAll := strings.Split(dataList[e], "?")
				query := ""
				if len(dataListAll) > 1 {
					query = dataListAll[1]
				}
				mark = strings.ReplaceAll(mark, "/*", "")
				mark = strings.ReplaceAll(mark, "*/", "")
				mark = strings.ReplaceAll(mark, "*", "\n")
				markLines := strings.Split(mark, "\n")

				for i := len(markLines) - 1; i >= 0; i-- {
					if strings.ReplaceAll(markLines[i], " ", "") == "" {
						markLines = append(markLines[:i], markLines[i+1:]...)
					}
				}
				chname := ""
				if len(markLines) > 0 {
					chname = strings.ReplaceAll(markLines[0], " ", "")
				}
				mark = strings.Join(markLines, "\n")
				result = append(result, fileDocInfo{
					Api:   dataListAll[0],
					Name:  chname,
					Query: query,
					Doc:   mark,
				})
				mark = ""
			}
		}
	}
	return classDoc, result, nil
}

func main() {
	root := "./"
	if len(os.Args) > 1 {
		log.Println("运行目录:", os.Args[1])
		root, _ = filepath.Abs(os.Args[1])
		workdir, _ := filepath.Abs(root + "/ApiOx/doc/")
		err := os.Chdir(workdir)
		if err != nil {
			log.Fatal("Chdir:", err.Error())
		}
	} else {
		root, _ = filepath.Abs("./")
		workdir, _ := filepath.Abs(root + "/ApiOx/doc/")
		err := os.Chdir(workdir)
		if err != nil {
			log.Fatal("Chdir:", err.Error())
		}
	}

	dir, err := filepath.Abs(root + "/" + mainDir + "/")
	if err != nil {
		log.Println("err:", err.Error())
	}

	var result []phpDocMenu

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		name := strings.ReplaceAll(path, ".class", "")
		name = strings.ReplaceAll(name, `/`, "-")
		name = strings.ReplaceAll(name, `\`, "-")

		if strings.Index(strings.ToLower(name), "base") > 0 {
			return nil
		}

		ok, err := filepath.Match(`*-Controller-*.php`, name)
		if ok {
			pre := pathToUrlPre(path)
			className, ends, err := pathToUrlEnd(path)
			if err != nil {
				log.Println("pathToUrlEndErr:", path, err.Error())

			} else {
				log.Println("组:", className)
				key := fmt.Sprintf("%x", md5.Sum([]byte(className)))
				resultItem := phpDocMenu{
					Name: className,
					Key:  key,
				}
				for e := range ends {
					log.Println("API名:", ends[e].Name)
					log.Println("API地址:", pre+ends[e].Api)
					log.Println("参数:", ends[e].Query)
					log.Println("---------------------")

					fileKey := fmt.Sprintf("%x", md5.Sum([]byte(ends[e].Name)))
					resultItem.Value = append(resultItem.Value, phpDocMenuValue{
						Name: ends[e].Name,
						File: key + "/" + fileKey + ".json",
						FileContent: phpDocFile{
							Name: ends[e].Name,
							Url:  pre + ends[e].Api,
							Desc: `参数:` + ends[e].Query + "\n注释:" + ends[e].Doc,
						},
					})

					//log.Println("详情:", ends[e].Doc)
				}
				result = append(result, resultItem)
				//name := strings.ReplaceAll(info.Name(), ".class", "")
			}
		}
		return err
	})

	for e := range result {
		for k := range result[e].Value {
			dataItem, _ := json.MarshalIndent(&result[e].Value[k].FileContent, "", "  ")
			os.Mkdir(filepath.Dir(result[e].Value[k].File), os.ModeDir)
			if !utils.FileExists(result[e].Value[k].File) {
				ioutil.WriteFile(result[e].Value[k].File, dataItem, 755)
			}
		}
	}

	data, _ := json.MarshalIndent(&result, "", "  ")
	ioutil.WriteFile("index.json", data, 755)

	if err != nil {
		log.Println("filepath.Walk err:", err.Error())
	}
}
