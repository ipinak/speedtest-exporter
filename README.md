# SpeedTest Exporter

A small prometheus exporter monitoring your network speed.

The code for speed test checking is copied bv this: [speedtest-go](https://github.com/showwin/speedtest-go)

# Build

The following command will generate a binary speedtest-exporter which opens a server on
port 9100

```bash
make
```

# Documentation

This programs runs a goroutine on the background and a http server on port 8080
exposing the `/metrics` path.

## Metrics

`speed_test_dl_speed` - your download speed in MBits

`speed_test_ul_speed` - your upload speed in MBits

# License

MIT

