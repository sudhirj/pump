# chunkasaur
Transmit large files over lossy networks using Raptor fountain codes

```
tx := Tx{chunkSize: 8000000, packetSize: 1000}
fd := tx.add(string identifier, io.ReaderAt r, int64 fileSize)
sendToReceivers(fd)

fd.activate(int weight)
tx.activate(fd, int weight)

fd.activate()
tx.activate(fd)

fd.deactivate()
tx.deactivate(fd)

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

