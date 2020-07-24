# ztcp
zrpc implementation built on TCP.

## Performance Tuning

### Conn Deadline

In Go standard library, conn use Time.Until to get duration.

Which means if the Time has no monotonic time it will call time.Now() again,
so it's meaningless to call tsc.UnixNano() outside.

