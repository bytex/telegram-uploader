package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"strconv"
	"fmt"
	"log"
	"github.com/tucnak/telebot"
)

var bot *telebot.Bot

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

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10485760)
	var content string
	if file, handler, err := r.FormFile("file"); err == nil {
		defer file.Close()
		image := telebot.File{}
		bot.SendPhoto()

		fmt.Fprintf(w, "%v", handler.Header)
		content = `        <h2>Upload complete</h2>
        <ul>
        <li></li>
        </ul>`;
	} else {
		content = fmt.Printf("<h2>Error uploading file: %s</h2>", err)
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

	if mybot, err := telebot.NewBot(token); err != nil {
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