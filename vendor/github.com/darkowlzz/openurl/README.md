# openurl

[![Build Status](https://travis-ci.org/darkowlzz/openurl.svg?branch=master)](https://travis-ci.org/darkowlzz/openurl)
[![codecov](https://codecov.io/gh/darkowlzz/openurl/branch/master/graph/badge.svg)](https://codecov.io/gh/darkowlzz/openurl)

golang package for opening URLs in default web browser.


## Usage

```
import (
  ...

  "github.com/darkowlzz/openurl"
)

if err := openurl.Open("http://example.com"); err != nil {
  log.Fatal(err)
}
```  
