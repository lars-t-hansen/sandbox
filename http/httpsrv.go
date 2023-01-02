package main

import (
  "fmt"
  "io"
  "log"
  "os"
  "net/http"
)

func main() {
  http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
      fmt.Fprintf(w, "Bad method")
      return
    }
    ct, ok := r.Header["Content-Type"]
    if !ok || ct[0] != "application/json" {
      fmt.Fprintf(w, "Bad content-type")
      return
    }
    payload := make([]byte, r.ContentLength);
    _, err := r.Body.Read(payload)
    if err != nil && err != io.EOF {
      fmt.Fprintf(w, "Bad content")
    } else {
      fmt.Fprintf(w, "OK")
      f, err := os.OpenFile("snappy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
      if err == nil {
        f.Write(payload)
        f.Write([]byte{'\n'})
        f.Close()
      }
    }
  })
  log.Fatal(http.ListenAndServe(":8086", nil))
}


