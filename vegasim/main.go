package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

var port = flag.Uint("port", 80, "Port number for http server")

var dataTemplate = `#System: VEGASCAN 693
#Version: 1.98/01
#TAG: JMR
#SNR: 56222393
#Date: %s
#Time: %s
#Ontime: %s

#PLS;TAG;VALUE;UNIT
001;"FRKTI-102 FLOW ";%v;mł/h
002;"FRKTI-102 TOTAL";%v;mł
003;"FRR-01    FLOW ";%v;mł/h
004;"FRR-01    TOTAL";%v;mł
005;"FRR-01    SUHU ";%v;°C
006;"FRDM01    FLOW ";%v;mł/h
007;"FRDM01    TOTAL";%v;mł
008;"FRDM01    SUHU ";%v;°C
009;"FTHCL01   FLOW ";%v;mł/h
010;"FTHCL01   TOTAL";%v;mł
011;"FTNAOH01  FLOW ";%v;mł/h
012;"FTNAOH01  TOTAL";%v;mł
013;"FTCA01    FLOW ";%v;l/h
014;"FTCA01    TOTAL";%v;mł
015;"FTCA01    SUHU ";%v;°C`

type Timespan time.Duration

func (t Timespan) Format(format string) string {
	z := time.Unix(0, 0).UTC()
	return z.Add(time.Duration(t)).Format(format)
}

func main() {
	flag.Parse()

	log.Println("Starting VegaScan simulator on port ", *port)
	log.Println("Available route is: GET /val.csv")

	var data = ""

	start := time.Now()

	go func() {

		incrVal := uint32(20000)
		for {

			values := make([]uint32, 3)
			for index := range values {
				values[index] = rand.Uint32()
			}

			ts := time.Now()
			data = fmt.Sprintf(dataTemplate,
				ts.Format("02.01.06"),
				ts.Format("15:04:05"),
				Timespan(time.Since(start)).Format("2 15.04.05"),
				values[0],
				values[1],
				values[2],
				100,
				200,
				300,
				400,
				incrVal,
				60000,
				70000,
				100000,
				110000,
				120000,
				140000,
				200000,
			)
			incrVal += 100

			if incrVal > 25000 {
				incrVal = 20000
			}

			time.Sleep(time.Millisecond * 200)
		}
	}()

	e := echo.New()
	e.GET("/val.csv", func(c echo.Context) error {
		return c.String(http.StatusOK, data)
	})

	e.HideBanner = true
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", *port)))
}
