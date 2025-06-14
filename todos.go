package indigo

// TODO 1: be able to decompress incoming deflate and gzip. Maybe even some other compressions, e.g. lz4 or zstd

// TODO 2: be able to compress outgoing stream with deflate, gzip, lz4 and zstd. Must also keep track of client's
// TODO 2: supported compression algorithms

// TODO 3: make a unified interface for JSON marshall/unmarshall and make it able, to use different implementation
// TODO 3: via config
