# dos2unix

A very "clever" way to convert `crlf` streams into just `cr` using the integer unit as much as possible.

The actual filter is implemented in the [filter.go](filter.go) and [constants.go](constants.go) files. Everything else is in support of the command-line handler.

## why did you do this?

I came across a in inefficient implementation that bothered me. There is a simple implementation that uses only 2 bytes of state and works as well as the [`dos2unix` pacakge](https://formulae.brew.sh/formula/dos2unix) I installed from Homebrew. However, I am working on a 64-bit CPU and it seemed silly to let 48 of those bits go fallow every cycle. This implementation is my answer to the problem.

* Crazy fast
* Constant allocations regardless of source file size
* Ignore encoding

OK, ignoring the encoding isn't a feature. But it is true that this implementation doesn't decode the input. I assume ASCII or a compatible encoding. I don't think it's possible for a UTF-8 codepoint to exist that has `0x0d 0x0a`, so I think we're safe for any encodings we're likely to encounter ([UTF-8](https://en.wikipedia.org/wiki/UTF-8), 7- or 8-bit [ASCII](https://en.wikipedia.org/wiki/ASCII), [CP1252](https://en.wikipedia.org/wiki/Windows-1252), any [ISO-8859](https://en.wikipedia.org/wiki/ISO/IEC_8859) variants like latin-1, and the old [Mac](https://en.wikipedia.org/wiki/Mac_OS_Roman) codepage). This program will definitely not work with [UTF-16](https://en.wikipedia.org/wiki/UTF-16) source files or [Shift-JIS](https://en.wikipedia.org/wiki/Shift_JIS).

## why did you make this compatible with the [linux command](https://linux.die.net/man/1/dos2unix)

Because I'm weird and I'm not going to apologize for that.

## algorithm

This works by handling only 8 bytes at a time, so buffered reader and writer
are recommended. The method here is to read 8 bytes and convert it into an
interger, big-endian. From there, we can AND with byte masks and see if that
equals an integer that is CRLF in the same byte position. For example:

```text
"012\r\n567" -> uint64(0x3031320d0a353637)
   0x3031320d0a353637
 & 0x000000ffff000000
---------------------
   0x0000000d0a000000
```

We know that `0x00000d0a000000` means that the 4th and 5th bytes
(offsets 3,4) must be CRLF. So we can take the left 3 bytes by ANDing
with `0xffffff0000000000`. We skip the we can take bytes 5, 6, 7, and 8 by
ANDing with `0x0000000fffffffff`, which we left-shift by 8 to consume the
4th byte.

```text
0x3031320d0a353637 & 0xffffff0000000000 -> 0x3031320000000000
0x3031320d0a353637 & 0x0000000fffffffff -> 0x000000000a353637
0x000000000a353637 << 8 -> 0x0000000a35363700

  0x3031320000000000
| 0x0000000a35363700
--------------------
  0x3031320a35363700
```

Since we consumed one byte, we know we should only write the first 7 bytes,
not the full 8.

The only question is, what to do if `\r\n` is split across an 8-byte boundary?
We know that this can only happen if the `\r` appears as the last byte. In that
case, we can make sure to ignore the last byte. To not lose that byte, we can
write that byte into the first position and then read 7 bytes starting at
position 1 the next time the loop goes around. That makes this case just like
reading with `\r\n` at the head of the array, so it is no longer special.

## Performance

On my laptop I was able to push ~600 MiB/s on a file with `crlf` every ~80 chars ([Tolstoy's __War and Peace__ from the Guttenberg project](https://www.gutenberg.org/ebooks/2600), end-to-end for 8GB):

```shell
$ cat tmp/huge.dos.txt | ./dos2unix | pv -rtb | xxh64sum
8.07GiB 0:00:13 [ 612MiB/s]
04d23dc04e854803  stdin
```

On an nearly equally-sized file with no `crlf`, it was over 700 MiB/s:

```shell
$ cat tmp/huge.unix.txt | ./dos2unix | pv -rtb | xxh64sum
8.07GiB 0:00:11 [ 732MiB/s]
1538cfccfecd201c  stdin
```

For comparison, the `dos2unix` command from Homebrew was a lot slower. It also removed the BOM from the source file, which is why the checksum is different.

```shell
$ cat tmp/huge.dos.txt | dos2unix | pv -rtb | xxh64sum
8.07GiB 0:04:55 [28.0MiB/s]
560a7c2d8aca5c96  stdin
```

As proof that my pipeline is not the limiting factor, I removed the processing and just went straight to `pv` and `xxh64sum`. I use `xxh64sum` because it is roughly 4x faster than `wc -c`. When there is nothing at the end of the pipe, it appears that some optimizations are happening and the speed becomes unrealistically fast.

```shell
$ cat tmp/huge.dos.txt | pv -rtb | xxh64sum
8.23GiB 0:00:03 [2.22GiB/s]
a586065aa7ef4582  stdin
```

## Author's note

This was something I threw together in my free time while working at [Bluecore](https://www.bluecore.com). If you like writing Go and want a fun place to do it, check 'em out.
