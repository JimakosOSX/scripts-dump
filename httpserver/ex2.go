// :) 
package main

import (
    "net/http"
    "fmt"
    "gorilla/mux"
)

func handleGet(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "get\n")
}

func handlePost(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Post\n")
}

func main() {
    r := mux.NewRouter()
    srv := &http.Server{
        Addr: ":80",
    }
    srv.ListenAndServe()
}
