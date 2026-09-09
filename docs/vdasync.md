# vdasync command

`vdasync` is the main CLI to synchronize target data from source data.

See also [operations logs](opelog.md) for large and long running data transfers or traceability requirements.

## Usage

Notes:

- plugins must be configured through a configuration file, see [configuration page](conf.md).
- TLS configuration is usually provided through a configuration file, see [TLS page](tls.md).
- concurrency is disabled by default, but increasing it is generally recommended
to gain better performance, as explained later.

Basic usage is

    vdasync [-dryrun] [-rm] [-check] -source <source DSS> -target <target DSS>

DSS stands for data storage system:
this can refer either to local files, to remote files accessed on a host running `vdaserver`,
or else to a plugin configured through a file,
see [DSS syntax](dssurl.md).

Source and target directories must exist in the case of files, their respective sub-trees will be synchronized.

For instance

    vdasync -dryrun -rm -source /path/to/dev -target /path/to/backup/for/dev
    vdasync -rm -source /path/to/dev -target /path/to/backup/for/dev
    vdasync -dryrun -check -source /path/to/dev -target /path/to/backup/for/dev

Example restoring local files from a remote backup server, see [vdaserver page](vdaserver.md) to run the latter:

    vdasync -rm -source dss://backup-server:9443/path/to/backup -target /path/to/dev

### Help listing all arguments

    $ vdasync -h
    Usage of vdasync:
    -ca string
            server or plugin TLS CA certificate
    etc.

### Synchronization features

    -check
        compute checksums
    -csal string
        comma separated list of hash algoritms to compute checksums: sha256 sha512 sha3_256 sha3_512 (default "sha256")
    -dryrun
        don't run operation, just report actions
    -excl string
        file containing regexps for paths to be excluded, defaults to none
    -incl string
        file containing regexps for paths to be included, defaults to all
    -nomtime
        don't set modification time, update if source changed later
    -nomtlink
        same as nomtime but only applies to symlinks
    -noperm
        neither check nor set permissions
    -rm
        remove files in sync target
    -source string
        source of the command
    -target string
        target of the command

### General purpose

    -conc int
        number of concurrent activities
    -config string
        configuration file, see documentation
    -cprof string
        cpu.prof file
    -level string
        log level, defaults to ERROR
    -log string
        log file, defaults to vdasync-<pid>.log in temp dir, "std[out|err]" are known keywords
    -out string
        file for output, defaults to stdout, "std[out|err]" are known keywords (default "stdout")
    -silent
        no output, or simple summary in case also verbose
    -verbose
        detailed output
    -version

### TLS configuration

    -cert string
        server or plugin TLS certificate
    -clientca string
        client TLS CA certificate
    -clientcert string
        client TLS certificate
    -clientkey string
        client TLS certificate key
    -insec
        don't check certificate when communicating with server
    -insecplugin
        don't check certificate when communicating with plugins
    -key string
        server or plugin TLS certificate key
    -notls
        insecure communication with servers over http
    -notlsplugin
        insecure communication with plugins over http

### Use of concurrency

As said above, increasing `vdasync` concurrency is generally recommended to gain better performance.
Its setting depends on the infrastructure and the plugins involved.

- As a default, the number of available CPU cores can be provided in many cases.
- Writing to slow devices should reduce it or even disable it (USB stick),
as parallel writes may even become counterproductive.
- Access to remote resources must take care of the target service capacity
that is often shared between many users.
- Using S3 and other HTTP-based services often benefits increasing it because related requests
involve network latency but may be run safely in parallel;
nevertheless this must be balanced with shared resources usage.
- Same remark applies in the case of network based storage like NFS or NAS.
- Client based encryption requires local compute resources,
therefore concurrency will be tuned according to related capacity.
