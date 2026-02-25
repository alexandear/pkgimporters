# pkgimporters

A command-line tool to fetch the number of known importers for Go packages from pkg.go.dev.

## Description

pkgimporters queries pkg.go.dev to retrieve importer counts for Go packages. It supports multiple ways to specify packages:

- Query specific packages via positional arguments
- Query packages using the `-pkgs` flag with comma-separated values (e.g., `-pkgs fmt,bufio`)
- Fetch all standard library packages with `-pkgs std`

Results can be sorted by package name (default) or by importer count in descending order.

## Installation

```sh
go install github.com/alexandear/pkgimporters@latest
```

## Usage

```sh
pkgimporters [-pkgs pkg1,pkg2,...|std] [-workers N] [-sort name|count] [package ...]
```

### Options

- `-pkgs` - Comma-separated list of packages to fetch (e.g., `-pkgs fmt,bufio`) or 'std' for all standard library packages
- `-workers N` - Number of concurrent requests (default: 5)
- `-sort` - Sort results by 'name' (default) or 'count' (descending importer count)

**Note:** Flags must be specified before positional arguments.

### Examples

Fetch importers for a single package:

```console
❯ pkgimporters os
os                   2,397,740
```

Fetch importers for multiple packages:

```console
❯ pkgimporters fmt bufio net/http golang.org/x/tools/go/analysis
bufio                          515,194
fmt                            5,485,422
golang.org/x/tools/go/analysis 6,136
net/http                       1,705,800
```

Use comma-separated list of packages:

```console
❯ pkgimporters -pkgs io,math/rand/v2
io                   1,533,321
math/rand/v2         4,986
```

Fetch importers for all standard library packages (takes a while):

```console
❯ pkgimporters std
archive/tar            27,061
archive/zip            22,244
bufio                  515,194
bytes                  1,204,809
cmp                    7,792
compress/bzip2         3,458
compress/flate         8,076
compress/gzip          77,599
compress/lzw           918
compress/zlib          12,336
container/heap         30,970
container/list         46,274
container/ring         4,077
context                1,967,772
crypto                 48,051
crypto/aes             43,798
crypto/cipher          49,170
crypto/des             6,441
crypto/dsa             5,046
crypto/ecdh            894
crypto/ecdsa           63,177
crypto/ed25519         8,319
crypto/elliptic        32,487
crypto/fips140         33
crypto/hkdf            47
crypto/hmac            54,631
crypto/hpke            0
crypto/md5             89,187
crypto/mlkem           57
crypto/mlkem/mlkemtest 0
crypto/pbkdf2          53
crypto/rand            199,629
crypto/rc4             4,819
crypto/rsa             55,943
crypto/sha1            78,079
crypto/sha256          133,589
crypto/sha3            122
crypto/sha512          31,041
crypto/subtle          23,649
crypto/tls             177,174
crypto/x509            115,983
crypto/x509/pkix       22,880
database/sql           192,619
database/sql/driver    33,356
debug/buildinfo        201
debug/dwarf            1,494
debug/elf              2,415
debug/gosym            417
debug/macho            903
debug/pe               951
debug/plan9obj         73
embed                  73,798
encoding               17,173
encoding/ascii85       537
encoding/asn1          17,469
encoding/base32        9,307
encoding/base64        231,634
encoding/binary        263,373
encoding/csv           43,453
encoding/gob           39,873
encoding/hex           236,380
encoding/json          1,456,575
encoding/pem           58,775
encoding/xml           75,937
errors                 1,580,116
expvar                 16,372
flag                   528,784
fmt                    5,485,422
go/ast                 39,682
go/build               19,289
go/build/constraint    163
go/constant            3,934
go/doc                 3,578
go/doc/comment         125
go/format              17,463
go/importer            2,172
go/parser              29,395
go/printer             7,151
go/scanner             4,376
go/token               41,812
go/types               15,525
go/version             77
hash                   67,614
hash/adler32           2,506
hash/crc32             21,411
hash/crc64             1,863
hash/fnv               21,297
hash/maphash           876
html                   15,318
html/template          116,219
image                  56,067
image/color            41,142
image/color/palette    1,083
image/draw             11,722
image/gif              9,201
image/jpeg             18,148
image/png              27,935
index/suffixarray      374
io                     1,533,321
io/fs                  41,598
io/ioutil              861,634
iter                   5,674
log                    1,359,920
log/slog               59,515
log/syslog             9,544
maps                   12,437
math                   589,822
math/big               221,940
math/bits              47,605
math/cmplx             3,712
math/rand              353,345
math/rand/v2           4,986
mime                   31,412
mime/multipart         45,161
mime/quotedprintable   1,289
net                    630,659
net/http               1,705,800
net/http/cgi           704
net/http/cookiejar     10,127
net/http/fcgi          1,107
net/http/httptest      17,540
net/http/httptrace     3,038
net/http/httputil      36,104
net/http/pprof         33,598
net/mail               25,070
net/netip              9,298
net/rpc                17,986
net/rpc/jsonrpc        2,254
net/smtp               12,041
net/textproto          17,188
net/url                557,415
os                     2,397,740
os/exec                256,170
os/signal              203,107
os/user                40,920
path                   296,255
path/filepath          663,112
plugin                 5,493
reflect                795,504
regexp                 487,451
regexp/syntax          1,410
runtime                405,611
runtime/cgo            1,004
runtime/coverage       27
runtime/debug          54,726
runtime/metrics        382
runtime/pprof          23,383
runtime/race           4
runtime/trace          3,493
slices                 59,822
sort                   551,464
strconv                1,435,979
strings                2,555,764
structs                123
sync                   1,296,984
sync/atomic            227,792
syscall                296,640
testing                107,228
testing/cryptotest     0
testing/fstest         193
testing/iotest         109
testing/quick          365
testing/slogtest       13
testing/synctest       3
text/scanner           4,090
text/tabwriter         31,186
text/template          110,065
text/template/parse    1,198
time                   2,721,686
time/tzdata            863
unicode                125,884
unicode/utf16          10,615
unicode/utf8           127,102
unique                 136
unsafe                 173,618
weak                   75
```

Sort by importer count (descending):

```console
❯ time pkgimporters -pkgs std -sort count
fmt                    5,485,422
time                   2,721,686
strings                2,555,764
os                     2,397,740
context                1,967,772
net/http               1,705,800
errors                 1,580,116
io                     1,533,321
encoding/json          1,456,575
strconv                1,435,979
log                    1,359,920
sync                   1,296,984
bytes                  1,204,809
io/ioutil              861,634
reflect                795,504
path/filepath          663,112
net                    630,659
math                   589,822
net/url                557,415
sort                   551,464
flag                   528,784
bufio                  515,194
regexp                 487,451
runtime                405,611
math/rand              353,345
syscall                296,640
path                   296,255
encoding/binary        263,373
os/exec                256,170
encoding/hex           236,380
encoding/base64        231,634
sync/atomic            227,792
math/big               221,940
os/signal              203,107
crypto/rand            199,629
database/sql           192,619
crypto/tls             177,174
unsafe                 173,618
crypto/sha256          133,589
unicode/utf8           127,102
unicode                125,884
html/template          116,219
crypto/x509            115,983
text/template          110,065
testing                107,228
crypto/md5             89,187
crypto/sha1            78,079
compress/gzip          77,599
encoding/xml           75,937
embed                  73,798
hash                   67,614
crypto/ecdsa           63,177
slices                 59,822
log/slog               59,515
encoding/pem           58,775
image                  56,067
crypto/rsa             55,943
runtime/debug          54,726
crypto/hmac            54,631
crypto/cipher          49,170
crypto                 48,051
math/bits              47,605
container/list         46,274
mime/multipart         45,161
crypto/aes             43,798
encoding/csv           43,453
go/token               41,812
io/fs                  41,598
image/color            41,142
os/user                40,920
encoding/gob           39,873
go/ast                 39,682
net/http/httputil      36,104
net/http/pprof         33,598
database/sql/driver    33,356
crypto/elliptic        32,487
mime                   31,412
text/tabwriter         31,186
crypto/sha512          31,041
container/heap         30,970
go/parser              29,395
image/png              27,935
archive/tar            27,061
net/mail               25,070
crypto/subtle          23,649
runtime/pprof          23,383
crypto/x509/pkix       22,880
archive/zip            22,244
hash/crc32             21,411
hash/fnv               21,297
go/build               19,289
image/jpeg             18,148
net/rpc                17,986
net/http/httptest      17,540
encoding/asn1          17,469
go/format              17,463
net/textproto          17,188
encoding               17,173
expvar                 16,372
go/types               15,525
html                   15,318
maps                   12,437
compress/zlib          12,336
net/smtp               12,041
image/draw             11,722
unicode/utf16          10,615
net/http/cookiejar     10,127
log/syslog             9,544
encoding/base32        9,307
net/netip              9,298
image/gif              9,201
crypto/ed25519         8,319
compress/flate         8,076
cmp                    7,792
go/printer             7,151
crypto/des             6,441
iter                   5,674
plugin                 5,493
crypto/dsa             5,046
math/rand/v2           4,986
crypto/rc4             4,819
go/scanner             4,376
text/scanner           4,090
container/ring         4,077
go/constant            3,934
math/cmplx             3,712
go/doc                 3,578
runtime/trace          3,493
compress/bzip2         3,458
net/http/httptrace     3,038
hash/adler32           2,506
debug/elf              2,415
net/rpc/jsonrpc        2,254
go/importer            2,172
hash/crc64             1,863
debug/dwarf            1,494
regexp/syntax          1,410
mime/quotedprintable   1,289
text/template/parse    1,198
net/http/fcgi          1,107
image/color/palette    1,083
runtime/cgo            1,004
debug/pe               951
compress/lzw           918
debug/macho            903
crypto/ecdh            894
hash/maphash           876
time/tzdata            863
net/http/cgi           704
encoding/ascii85       537
debug/gosym            417
runtime/metrics        382
index/suffixarray      374
testing/quick          365
debug/buildinfo        201
testing/fstest         193
go/build/constraint    163
unique                 136
go/doc/comment         125
structs                123
crypto/sha3            122
testing/iotest         109
go/version             77
weak                   75
debug/plan9obj         73
crypto/mlkem           57
crypto/pbkdf2          53
crypto/hkdf            47
crypto/fips140         33
runtime/coverage       27
testing/slogtest       13
runtime/race           4
testing/synctest       3
crypto/hpke            0
crypto/mlkem/mlkemtest 0
testing/cryptotest     0
pkgimporters -pkgs std -sort count  0.74s user 1.44s system 1% cpu 2:53.79 total
```

Sort multiple packages by importer count:

```console
❯ pkgimporters -sort count bufio fmt
fmt                  5,485,422
bufio                515,194
```

Use 20 concurrent requests:

```sh
pkgimporters -workers 20 -pkgs std
```
