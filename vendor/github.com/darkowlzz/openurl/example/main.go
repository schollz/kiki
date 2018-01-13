package main

import (
  "log"

  "github.com/darkowlzz/openurl"
)

func main() {
  if err := openurl.Open("http://example.com"); err != nil {
    log.Fatal(err)
  }
}
