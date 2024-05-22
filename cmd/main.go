package main

import (
	"UrlShortener/internal/pkg"
)

func main() {
	pkg.InitServer()
	//shortener := Handelfunctions.URLShortener{}
	//
	//// Обработчик для корневого URL
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "Welcome to the URL Shortener! Use /shorten to shorten a URL and /short/ to redirect.")
	//})
	//
	//http.HandleFunc("/shorten", shortener.HandleShorten)
	//http.HandleFunc("/short/", shortener.HandleRedirect)
	//
	//fmt.Println("URL Shortener is running on :8080")
	//http.ListenAndServe(":8080", nil)
}
