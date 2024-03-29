package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/mafredri/goodspeaker"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	// Music Flow
	// key = "4efgvbn m546Uy7kolKrftgbn =-0u&~"
	// iv  = "54eRty@hkL,;/y9U"
	// LG Sound Bar
	key = "T^&*J%^7tr~4^%^&I(o%^!jIJ__+a0 k"
	iv  = "'%^Ur7gy$~t+f)%@"
)

func main() {
	pcapFile := flag.String("pcap", "", "Read from pcap file (default stdin)")
	flag.StringVar(&key, "key", key, "AES key for encryption")
	flag.StringVar(&iv, "iv", iv, "IV for encryption")
	flag.Parse()

	withAES, err := goodspeaker.WithAES([]byte(key), []byte(iv))
	if err != nil {
		panic(err)
	}

	f := os.Stdin
	if *pcapFile != "" {
		f, err = os.Open(*pcapFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}

	handle, err := pcap.OpenOfflineFile(f)
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if packet.ApplicationLayer() != nil {
			payload := bytes.NewReader(packet.ApplicationLayer().Payload())
			r := goodspeaker.NewReader(payload, withAES)

			b, err := ioutil.ReadAll(r)
			if err != nil && !errors.Is(err, io.EOF) {
				continue
			}

			var v interface{}
			if err = json.Unmarshal(b, &v); err != nil {
				continue
			}

			b, err = json.MarshalIndent(v, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(os.Stdout, "%s,\n", b)
		}
	}
}
