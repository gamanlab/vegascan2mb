package main

import (
	"flag"
	"log"
)

var (
	LE          = flag.Bool("le", true, "Use little endian bytes order")
	mbAddress   = flag.String("addr", "0.0.0.0:502", "IP address & port for modbus slave")
	vegaAddress = flag.String("url", "http://127.0.0.1/val.csv", "URL of vegascan")
)

var strCsv = `#System: VEGASCAN 693
#Version: 1.98/01
#TAG: JMR
#SNR: 56222393
#Date: 08.06.22
#Time: 16:07:41
#Ontime: 0 00:09:39

#PLS;TAG;VALUE;UNIT
001;"FRKTI-102 FLOW ";83;mł/h
002;"FRKTI-102 TOTAL";44614;mł
003;"FRR-01    FLOW ";59;mł/h
004;"FRR-01    TOTAL";33153;mł
005;"FRR-01    SUHU ";20;°C
006;"FRDM01    FLOW ";40;mł/h
007;"FRDM01    TOTAL";32657;mł
008;"FRDM01    SUHU ";20;°C
009;"FTHCL01   FLOW ";0;mł/h
010;"FTHCL01   TOTAL";9;mł
011;"FTNAOH01  FLOW ";0;mł/h
012;"FTNAOH01  TOTAL";7;mł
013;"FTCA01    FLOW ";0;l/h
014;"FTCA01    TOTAL";148;mł
015;"FTCA01    SUHU ";32;°C`

func main() {
	flag.Parse()

	log.Println("Start vegascan to modbus converter.")
	log.Println("For help, run with flag -h")

	MbServer()
}
