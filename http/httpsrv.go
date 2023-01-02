package main

import (
  "fmt"
  "io"
  "log"
  "os"
  "net/http"
  "strconv"
  "time"
)

func main() {
  http.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("/time")
    if r.Method != "GET" {
      w.WriteHeader(403)
      fmt.Fprintf(w, "Bad method")
      return
    }
    t := time.Now()
    s := fmt.Sprintf("%d\r\n", t.Unix())
    w.Header()["Content-Length"] = []string{strconv.Itoa(len(s))}
    w.Header()["Content-Type"] = []string{"text/plain"}
    w.WriteHeader(200)
    fmt.Fprintf(w, "%d\r\n", t.Unix())
  })
  http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("/data")
    if r.Method != "POST" {
      w.WriteHeader(403)
      fmt.Fprintf(w, "Bad method")
      return
    }
    ct, ok := r.Header["Content-Type"]
    if !ok || ct[0] != "application/json" {
      w.WriteHeader(400)
      fmt.Fprintf(w, "Bad content-type")
      return
    }
    payload := make([]byte, r.ContentLength);
    _, err := r.Body.Read(payload)
    if err != nil && err != io.EOF {
      w.WriteHeader(400)
      fmt.Fprintf(w, "Bad content")
    } else {
      w.WriteHeader(202)
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


