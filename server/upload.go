package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 文件上传
// floder 文件所要保存的目录
// maxLen 文件最大长度
// 允许的扩展名[正则表达式]，例如：(.png|.jpeg|.jpg|.gif)
func (c *Context) UploadFiles(folder string, maxLen int, allowExt string) ([]string, error) {
	result := make([]string, 0)
	// 允许的扩展名正则匹配
	regExt, err := regexp.Compile(allowExt)
	if err != nil {
		return nil, err
	}

	// 文件长度限制
	length, err := strconv.Atoi(c.Request.Header.Get("Content-Length"))
	if err != nil || (length > maxLen) {
		return nil, errors.New("Upload file is too big!")
	}

	// get the multipart reader for the request.
	reader, err := c.Request.MultipartReader()
	if err != nil {
		return nil, err
	}

	// 文件全路径
	var fullFilePath string
	// 读取每个文件,保存。
	for {
		// 通过boundary，获取每个文件数据。
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		// 检查文件名
		fileName := part.FileName()
		if fileName == "" {
			continue
		}
		// 检查文件类型
		fileExt := strings.ToLower(path.Ext(fileName))
		if !regExt.MatchString(fileExt) {
			continue
			//return nil, errors.New("Invalid upload file type!")
		}
		// 构造新的文件名和全路径
		subFilePath := fmt.Sprintf("%s/%s", time.Now().Format("2006-01-02"), fileName)
		fullFilePath = path.Clean(folder + subFilePath)

		// 新建文件夹,0777
		os.MkdirAll(path.Dir(fullFilePath), os.ModePerm)
		// 建立目标文件
		dst, err := os.Create(fullFilePath)
		defer dst.Close()
		if err != nil {
			return nil, err
		}

		// 拷贝数据流到文件
		if _, err = io.Copy(dst, part); err != nil {
			return nil, err
		}

		// 将成功写入的文件路径信息写入结果集
		result = append(result, subFilePath)
	}

	// 检查结果
	if len(result) == 0 {
		return nil, errors.New("Invalid upload file!")
	}

	return result, nil
}
