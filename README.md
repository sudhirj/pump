# Pump
Transmit large files over lossy networks using fountain codes

```
tx := Tx{chunkSize: 8000000, packetSize: 1000}
fd := tx.add(string identifier, io.ReaderAt r, int64 fileSize)
sendToReceivers(fd)

tx.activate(fd, chunkIndex, int weightage)

while tx.active() {
  p := tx.packet()
  sendOrBroadcastOverLossyNetwork(p)
}

///

rx := Rx{}
rd := rx.expect(fd, receiveIntoFilePath)

while p := conn.receive() {
  rx.load(p)
}

rd.done?()
rd.stats()
```

