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
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
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
	filenameWithoutExtension := strings.TrimSuffix(name, filepath.Ext(name))
	println(filenameWithoutExtension)

	upload := "uploads/" + name
	download := "public/downloads/" + filenameWithoutExtension + ".wav"
	pathsToDelete := [2]string{download, upload}
	go deleteFiles(pathsToDelete)
	c.FileAttachment(download, filenameWithoutExtension+".wav")
}

func main() {
	r := gin.Default()
	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			secureMiddleware := secure.New(secure.Options{
				SSLRedirect: true,
				SSLHost:     "mp3towav.io:443",
			})
			err := secureMiddleware.Process(c.Writer, c.Request)

			// If there was an error, do not continue.
			if err != nil {
				return
			}

			c.Next()
		}
	}()
	r.Use(secureFunc)

	router := gin.Default()
	r.Use(static.Serve("/", static.LocalFile("/public", false)))
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
	router.GET("/terms", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.GET("/uploaded", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.GET("/heartbeat", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "beat")
	})
	router.GET("/file/:name", sendFileAndDelete)
	router.POST("/convert", convert)
	router.NoRoute(func(ctx *gin.Context) {
		ctx.HTML(http.StatusNotFound, "404.html", gin.H{})
	})

	r.GET("/", func(c *gin.Context) {
		c.String(200, "X-Frame-Options header is now `DENY`.")
	})
	go r.Run(":80")
	router.RunTLS(":443", "/etc/letsencrypt/live/mp3towav.io/fullchain.pem", "/etc/letsencrypt/live/mp3towav.io/privkey.pem")
}
