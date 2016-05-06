package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"strconv"
	"fmt"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/cast"
	"bytes"
)

const template = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="description" content="A tool which uploads files to Telegram">
    <meta name="author" content="Ruslan Balkin">
    <link rel="icon" href="../../favicon.ico">
    <title>Telegram uploader</title>
    <link href="https://yastatic.net/bootstrap/3.3.6/css/bootstrap.min.css" rel="stylesheet">
    <style>.app {padding:50px;}</style>
  </head>
  <body>
    <nav class="navbar navbar-inverse navbar-fixed-top">
      <div class="container">
        <div class="navbar-header">
          <a class="navbar-brand" href="https://github.com/bytex/telegram-uploader">Telegram Uploader</a>
        </div>
      </div>
    </nav>

    <div class="container">

      <div class="app">
        <h1>Telegram uploader</h1>
        %s
      </div>

    </div><!-- /.container -->

    <script src="https://yastatic.net/jquery/2.2.3/jquery.min.js"></script>
    <script src="https://yastatic.net/bootstrap/3.3.6/js/bootstrap.min.js"></script>
  </body>
</html>
`

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	content := `        <form action="/upload" method="POST" enctype="multipart/form-data">
          <input type="file" name="file" />
          <input type="submit" name="submit" />
        </form>
`;
	fmt.Fprintf(w, template, content)
}

var bot *tgbotapi.BotAPI

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10485760)
	var content string
	if file, handler, err := r.FormFile("file"); err == nil {
		defer file.Close()
		buf := make([]byte, 10485760)
		if in, err := file.Read(buf); err == nil {
			log.Println("Read bytes", in)
			content = fmt.Sprintf("<h2>Error uploading file: %s</h2>", err)
			b := tgbotapi.FileBytes{Name: "image.jpg", Bytes: buf}
			upload := tgbotapi.NewPhotoUpload(cast.ToInt64(viper.Get("Chat")), b)
			if m, err := bot.Send(upload); err == nil {
				fileId, maxWidth, maxHeight := "", 0, 0
				buf := bytes.NewBufferString("        <h2>Upload complete</h2>\n<ul>\n")
				for _, entry := range *m.Photo {
					if entry.Width > maxWidth || entry.Height > maxHeight {
						maxWidth = entry.Width
						maxHeight = entry.Height
						fileId = entry.FileID
						fmt.Fprintf(buf, "<li><strong>%s</strong>: %d x %d</li>", entry.FileID, entry.Width, entry.Height)
					}
				}
				fmt.Fprintf(buf, "</ul>\n<p>Best: %s</p>\n", fileId)
				fmt.Fprintf(buf, "<p>File: %s</p>", handler.Filename)
				content = buf.String()
			} else {
				content = fmt.Sprintf("<h2>Error uploading file: %s</h2>", err)
			}
		} else {
			content = fmt.Sprintf("<h2>Error reading uploaded file: %s</h2>", err)
		}
	} else {
		content = fmt.Sprintf("<h2>Error uploading file: %s</h2>", err)
	}
	fmt.Fprintf(w, template, content)
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/telegram-uploader/")
	viper.AddConfigPath("$HOME/.telegram-uploader")
	viper.AddConfigPath("$HOME/.config/telegram-uploader")
	viper.AddConfigPath(".")
	viper.SetDefault("Port", "8090")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("No configuration file loaded - using defaults: %s", err)
	}

	token, port := viper.GetString("Token"), viper.GetInt("Port")

	log.Printf("Initializing bot with %s", token)

	if mybot, err := tgbotapi.NewBotAPI(token); err != nil {
		log.Printf("Failed to initialize bot: %s", err)
		return
	} else {
		bot = mybot
	}

	log.Printf("Launching webserver on port %d", port)

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/upload", UploadHandler)
	http.ListenAndServe(":" + strconv.Itoa(port), r)
}