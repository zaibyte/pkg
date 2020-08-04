# ztcp

zrpc implementation built on TCP.

Based on [goproc](https://github.com/valyala/gorpc) with these modifications:

1. No multi-methods supports, only has three methods: Put Object, Get Object, Delete Object.

2. Implement End-to-End checksum.

3. Add read/write deadline on each read/write.

4. Add header.

5. Import xlog for logging.

6. Replacing Gob encoding by binary encoding.

7. Remove batch supports

8. Remove public Async API

9. Client/Server reader will wait for a certain time to get header, if timeout it'll retry, avoiding hang.

## Performance Tuning

The origin has tried its best to make things non-blocking.

It's hard to compare directly, because the features of xtcp is limited. But use almost "same" benchmark test,
xtcp gets 3x better than gorpc. (Both of them are using same configs, including buffer size, client connections, flush delay)

### Done

1. Binary encoding/decoding

2. Reuse memory (pool for tiny bytes slice)

...

### TODO

### Given up

#### Conn Deadline

In Go standard library, net.Conn use Time.Until to get duration.

Which means if the Time has no monotonic time it will call time.Now() again,
so it's meaningless to call tsc.UnixNano() outside.

