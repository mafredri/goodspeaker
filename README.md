# goodspeaker

This library speaks the LG speaker protocol which is used by both LG Music Flow Player and LG Wi-Fi Speaker, perhaps others as well. The name comes from Go, LG (Life's Good) and well, speakers.

Both encrypted and unencrypted communication are supported, but encryption IV and key may need to be adjusted depending on what type of speaker we're communicating with.

See [github.com/mafredri/musicflow](https://github.com/mafredri/musicflow) for an actual implementation using this protocol.

## Usage

```console
go get -u github.com/mafredri/goodspeaker
```

For decoding captured packets, see `cmd/pcap-decode`.
