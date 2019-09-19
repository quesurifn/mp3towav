package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/foolin/gin-template"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func deleteFiles(paths [2]string) {
	time.Sleep(5 * time.Second)
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			return
		}
	}
	return
}

type jsonResponse struct {
	URL string `json:"url"`
}

func convert(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	filenameWithoutExtension := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}
	upload := "uploads/" + file.Filename

	if err := c.SaveUploadedFile(file, upload); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	download := "public/downloads/" + filenameWithoutExtension + ".wav"
	cmd := exec.Command("ffmpeg", "-i", upload, "-vn", "-c:a", "copy", download)
	cmd.Run()

	response := &jsonResponse{
		// URL: "https://wavtomp3.io/" + filenameWithoutExtension + ".wav",
		URL: "/file/" + filenameWithoutExtension + ".wav",
	}

	c.JSON(200, response)
}

func sendFileAndDelete(c *gin.Context) {

	name := c.Param("name")

	println("name")
	println(name)
	filenameWithoutExtension := strings.TrimSuffix(name, filepath.Ext(name))
	println("Filename")
	println(filenameWithoutExtension)

	upload := "uploads/" + name
	download := "public/downloads/" + filenameWithoutExtension + ".wav"
	pathsToDelete := [2]string{download, upload}
	go deleteFiles(pathsToDelete)
	println("Download")
	println(download, filenameWithoutExtension)
	c.FileAttachment(download, filenameWithoutExtension)
}

func main() {
	router := gin.Default()
	router.Static("/public", "./public")
	router.HTMLRender = gintemplate.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.GET("/file/:name", sendFileAndDelete)
	router.POST("/convert", convert)
	router.Run(":4000")
}
