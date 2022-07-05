package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// HomeDir 获取home目录
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// CgroupDir 获取cgroup目录，暂时使用固定值
func CgroupDir() string {
	return "/sys/fs/cgroup"
}

// JoinPath 拼接目录
func JoinPath(DirPrefix, id, DirSuffix string) string {
	return strings.Join([]string{DirPrefix, id, DirSuffix}, "")
}

// IsDirOrFileExist 判断目录或文件是否存在
func IsDirOrFileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("check file: %s failed by %v", path, err)
		}
	}
	return true, nil
}

// WriteToFile 写int64到文件, 并返回原值
func WriteIntToFile(path string, data int64) (int64, error) {
	oldDataB, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	oldData, err := strconv.ParseInt(strings.Replace(string(oldDataB), "\n", "", -1), 10, 64)
	if err != nil {
		return 0, err
	}
	err = ioutil.WriteFile(path, []byte(strconv.FormatInt(data, 10)), 0644)
	if err != nil {
		return 0, err
	}
	return oldData, nil
}
