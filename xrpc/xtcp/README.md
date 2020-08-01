# ztcp
zrpc implementation built on TCP.

Based on [goproc](https://github.com/valyala/gorpc) with these modifications:

1. Use method (uint8) replaces of function & service name (string).

2. Use magic number to establish connection.

3. Add read/write deadline on each read/write.

4. Use header to pass needed arguments for a Zai rpc.

5. Import zlog for logging

6. Replacing Gob encoding by any objects supporting marshal & unmarshal

7. Remove batch supports

8. Remove public Async API

## Timeout

There are two types of timeout in ztcp:

1. Request timeout

This timeout is copied from gorpc, and it's helpful to cancel request which
waiting too long in request queue.

2. Connection read/write deadline



## Performance Tuning

### Done

### TODO

### Given up

#### Conn Deadline

In Go standard library, net.Conn use Time.Until to get duration.

Which means if the Time has no monotonic time it will call time.Now() again,
so it's meaningless to call tsc.UnixNano() outside.

