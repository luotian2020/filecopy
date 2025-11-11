package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	SourceDir  string `json:"sourceDir"`
	TargetDir  string `json:"targetDir"`
	SourceDate string `json:"sourceDate"`
}

func main() {
	// 读取配置
	config, err := readConfig("config.json")
	if err != nil {
		fmt.Println("读取配置失败:", err)
		return
	}

	// 确保目标目录存在
	err = os.MkdirAll(config.TargetDir, os.ModePerm)
	if err != nil {
		fmt.Println("无法创建目标目录:", err)
		return
	}

	// 解析日期
	targetDate, err := time.Parse("2006-01-02", config.SourceDate)
	if err != nil {
		fmt.Println("日期格式错误:", err)
		return
	}

	// 遍历源目录
	err = filepath.Walk(config.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// 获取文件修改时间（按本地时间比较日期）
		modTime := info.ModTime()
		if sameDay(modTime, targetDate) {
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
			if ext == "" {
				ext = "unknown"
			}

			targetSubDir := filepath.Join(config.TargetDir, ext)
			err := os.MkdirAll(targetSubDir, os.ModePerm)
			if err != nil {
				fmt.Println("创建子目录失败:", err)
				return nil
			}

			targetPath := filepath.Join(targetSubDir, info.Name())
			if err := copyFile(path, targetPath); err != nil {
				fmt.Println("复制失败:", err)
			} else {
				fmt.Println("已复制:", path, "=>", targetPath)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("遍历出错:", err)
	}
}

func readConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func sameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}
