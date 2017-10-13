package main

import (
	"io/ioutil"
	"fmt"
	"strings"
	"github.com/fatih/color"
	"sort"
)

const DEFAULT_VALUE = "Input your value here!"

func main() {

	var allConfigs []Config
	allConFiles := searchConfFiles()
	allConFilesNum := len(allConFiles)
	for i := 0; i < allConFilesNum; i++ {
		fileContent := readConfFile(allConFiles[i])
		config := splitConfFileToKeysAndValues(allConFiles[i], fileContent)
		config.CheckValues()
		allConfigs = append(allConfigs, config)
	}

	compareKeysInDiffFiles(allConfigs)

}

/**
获取当前目录下的所有配置文件
 */
func searchConfFiles() []string{
	var autoFileNamePrefix string
	var allFileNamesBeforeFoundConfFile []string
	var allConFiles []string
	fileInfo, err := ioutil.ReadDir(".")
	checkError(err)
	fileNum := len(fileInfo)
	for i := 0; i < fileNum; i++ {
		eachFile := fileInfo[i]
		if eachFile.IsDir() {
			continue
		}
		eachFileName := eachFile.Name()
		fmt.Println("eachFileName:", eachFileName)
		if len(autoFileNamePrefix) == 0 {
			allFileNamesBeforeFoundConfFile = append(allFileNamesBeforeFoundConfFile, eachFileName)
			if strings.Contains(eachFileName, ".auto") {
				splitFileName := strings.Split(eachFileName, ".")
				if len(splitFileName[0]) == 0 {
					autoFileNamePrefix = splitFileName[1]
				} else {
					autoFileNamePrefix = splitFileName[0]
				}

			}
		} else {
			if strings.Contains(eachFileName, autoFileNamePrefix) {
				allConFiles = append(allConFiles, eachFileName)
			}
		}

		fmt.Println("config prefix:", string(autoFileNamePrefix))
	}

	if len(autoFileNamePrefix) == 0 {
		panic("Auto Files Not Found")
	}

	missedFileNum := len(allFileNamesBeforeFoundConfFile)
	for i := 0; i < missedFileNum; i++ {
		missedFileName := allFileNamesBeforeFoundConfFile[i]
		if strings.Contains(missedFileName, autoFileNamePrefix) {
			allConFiles = append(allConFiles, missedFileName)
		}
	}


	sort.Stable(sort.StringSlice(allConFiles))

	allConfFileNum := len(allConFiles)
	for i := 0; i < allConfFileNum; i++ {
		fmt.Println("config file: ", allConFiles[i])
	}

	return allConFiles
}

/**
读取配置文件内容.
 */
func readConfFile(filename string) []byte {
	confFile, err := ioutil.ReadFile(filename)
	checkError(err)
	return confFile
}

/**
把配置文件内容分成Key-Value.
 */
func splitConfFileToKeysAndValues(filename string, fileContent []byte) Config{
	var keys []string
	var values []string
	fileContentStr := string(fileContent)
	splitStr := ""
	if strings.Index(fileContentStr, "\r\n") != -1 {
		splitStr = "\r\n"
	} else if strings.Index(fileContentStr, "\n") != -1 {
		splitStr = "\n"
	} else {
		splitStr = "\r"
	}
	fileContentInLines := strings.Split(fileContentStr, splitStr)
	for i := 0; i < len(fileContentInLines); i++ {
		eachLineContent := fileContentInLines[i]
		// 去掉每行两边的空字符
		eachLineContent = strings.TrimSpace(eachLineContent)
		if len(eachLineContent) == 0 {
			continue
		}

		// 如果是注释
		if isComment(eachLineContent) {
			continue
		}
		
		// 如果不存在=
		if strings.Index(eachLineContent, "=") == -1 {
			continue
		}
		keyValue := strings.Split(eachLineContent, "=")
		// 去掉key两边的空格
		keyValue[0] = strings.TrimSpace(keyValue[0])
		// 确保Key中间的空格只有一个
		keySplit := strings.Split(keyValue[0], " ")
		keySplitLen := len(keySplit)
		var realKeyParts []string
		for j := 0; j < keySplitLen; j++ {
			if keySplit[j] != "" {
				realKeyParts = append(realKeyParts, keySplit[j])
			}
		}
		realKey := strings.Join(realKeyParts, " ")

		keys = append(keys, realKey)
		values = append(values, keyValue[1])
	}
	return Config{filename,keys, values}
	
}

type Config struct {
	FILENAME string
	KEYS []string
	VALUES []string
}

/**
检查单个配置的Key的值是否为空或者是默认值.
 */
func (config Config) CheckValues() {
	values := config.VALUES
	lenOfValues := len(values)
	for i := 0; i < lenOfValues; i++ {
		if len(values[i]) == 0 {
			echoWarning(config.KEYS[i], "seems doesn't have a value in", config.FILENAME)
		} else if values[i] == DEFAULT_VALUE {
			echoErr(config.KEYS[i], "is set with default value, fix it!")
		}
	}
}

/**
比较所有文件的Key是否一样
 */
func compareKeysInDiffFiles(confs []Config) {

	configCount := len(confs)
	for i := 0; i < configCount; i++ {
		for j := i + 1; j < configCount; j++ {
			compareKeys(confs[i], confs[j])
		}
	}
}

/**
比较两个配置的KEY是否一致，不一致的则输出提示.
 */
func compareKeys(conf1, conf2 Config)  {
	compare1To2 := arrayDiff(conf1.KEYS, conf2.KEYS)
	if len(compare1To2) > 0 {
		echoErr(color.CyanString(strings.Join(compare1To2, ", ")), "configured in", conf1.FILENAME, ", but not found in", conf2.FILENAME)
	}

	compare2To1 := arrayDiff(conf2.KEYS, conf1.KEYS)
	if len(compare2To1) > 0 {
		echoErr(color.CyanString(strings.Join(compare2To1, ", ")), "configured in", conf2.FILENAME, ", but not found in", conf1.FILENAME)
	}
}


/**
找出存在于a1而不在a2的元素
 */
func arrayDiff(a1, a2 []string) []string {
	var notIn2 []string
	for _, value := range a1 {
		notFound := true
		for _, value2 := range a2  {
			if value == value2 {
				notFound = false
				break
			}
		}
		if notFound {
			notIn2 = append(notIn2, value)
		}
	}
	return notIn2
}

/**
标准输出Warning
 */
func echoWarning(a ...interface{})  {
	fmt.Println(color.YellowString("Warning:"), a)
}

/**
标准输出Error
 */
func echoErr(a ...interface{})  {
	fmt.Println(color.RedString("Error:"), a)
}

/**
判断Error是否存在的方法
 */
func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

/**
判断一行内容是否是注释.
 */
func isComment(content string) bool{
	if strings.Index(content, "//") == 0 || strings.Index(content, "*") == 0 || strings.Index(content, "#") == 0 {
		return true
	} else {
		return false
	}
}
