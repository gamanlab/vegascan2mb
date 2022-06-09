package main

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/annlumia/mbserver"
)

func MbServer() {
	serv := mbserver.NewServer()
	err := serv.ListenTCP(*mbAddress)
	if err != nil {
		log.Printf("%v\n", err)
	}
	defer serv.Close()

	serv.RegisterFunctionHandler(3, func(s *mbserver.Server, f mbserver.Framer) ([]byte, *mbserver.Exception) {
		data := f.GetData()
		register := int(binary.BigEndian.Uint16(data[0:2]))
		numRegs := int(binary.BigEndian.Uint16(data[2:4]))
		endRegister := register + numRegs
		if endRegister > 64 {
			return []byte{}, &mbserver.IllegalDataAddress
		}

		return append([]byte{byte(numRegs * 2)}, mbserver.Uint16ToBytes(s.HoldingRegisters[register:endRegister])...), &mbserver.Success
	})

	// Wait forever
	for {
		time.Sleep(time.Second)
		res, err := FetchVegaData(*vegaAddress)
		if err == nil {
			strCsv = res
		}

		valBytes, err := CsvStringToBytes(res, *LE)
		if err == nil {
			for i := 0; i < len(valBytes)/2; i++ {
				if *LE {
					serv.HoldingRegisters[i] = binary.LittleEndian.Uint16(valBytes[2*i : 2*i+2])
				} else {
					serv.HoldingRegisters[i] = binary.BigEndian.Uint16(valBytes[2*i : 2*i+2])
				}
			}
		}

	}

}
