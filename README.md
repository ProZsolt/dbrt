# dbrt

Report water meter readings for [Fővárosi Vízművek][1] via [Díjbeszedő Holding][2] website with Go

Download it using

```go get github.com/prozsolt/dbrt```

## Usage

```go
package main

import (
  "github.com/prozsolt/dbrt"
)
func main(){
  err := dbrt.ReportMeterReading("FIZETO_AZONOSITO", 1337)
  if err != nil {
    // handle error
  }
}
```

## Limitation

Currently only support one meter per fizetési azonositó

[1]: https://ugyfelszolgalat.vizmuvek.hu/
[2]: https://www.dbrt.hu/kezdolap