# ztcp

zrpc implementation built on TCP.

Based on [goproc](https://github.com/valyala/gorpc) with these modifications:

1. No multi-methods supports, only has three methods: Put Object, Get Object, Delete Object.

2. Implement End-to-End checksum.

3. Use magic number to verify request.

4. Add read/write deadline on each read/write.

5. Add header.

6. Import xlog for logging.

7. Replacing Gob encoding by binary encoding.

8. Remove batch supports

9. Remove public Async API

## Performance Tuning

The origin has tried its best to make things non-blocking.

### Done

### TODO

### Given up

#### Conn Deadline

In Go standard library, net.Conn use Time.Until to get duration.

Which means if the Time has no monotonic time it will call time.Now() again,
so it's meaningless to call tsc.UnixNano() outside.

