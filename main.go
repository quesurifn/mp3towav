package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/foolin/gin-template"
	"github.com/gin-gonic/gin"
)

func deleteFiles(paths [2]string) {
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			return
		}
	}
	return
}

func convert(c *gin.Context) {

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}
	upload := "uploads/" + file.Filename

	if err := c.SaveUploadedFile(file, upload); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	name := strings.Replace(upload, ".mp3", ".wav", -1)
	name = strings.Replace(name, "uploads/", "", -1)
	download := "downloads/" + name
	cmd := exec.Command("ffmpeg", "-i", upload, "-vn", "-c:a", "copy", download)
	cmd.Run()
	pathsToRemove := [2]string{download, upload}
	deleteFiles(pathsToRemove)
	c.File(download)
}

func main() {
	router := gin.Default()

	//new template engine
	router.HTMLRender = gintemplate.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "page.html", gin.H{"title": "Page file title!!"})
	})
	router.POST("/convert", convert)
	router.Run(":3000")
}
