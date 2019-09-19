# torque_exporter
Allow a server to collect metrics from TORQUE and expose them in Prometheus format. The exporter accesses TORQUE via ssh to a remote machine that can perform `qstat` and `showq` commands.

## Install

> Requires Go >=1.8

```
go get github.com/ari-apc-lab/torque_exporter
go install github.com/ari-apc-lab/torque_exporter
```

## Usage

```
torque_exporter -host=<HOST> -ssh-user=<USER> -ssh-password=<PASSWD> [-listen-address=:<PORT>] [-countrytz=<TZ>] [-log.level=<LOGLEVEL>]
```
### Defaults

\<PORT\>: `:9100`  
\<HOST\>: `localhost`, not supported  
\<TZ\>: `Europe/Madrid`  
\<LOGLEVEL\>: `error`  

## Debug

delve works nicely:
https://github.com/go-delve/delve/blob/master/Documentation/cli/getting_started.md

Once installed, debug it by:

```
dlv debug github.com/ari-apc-lab/torque_exporter -- -host=<HOST> -ssh-user=<USER> -ssh-password=<PASSWD>
```
Set a breakpoint like this (for example on my machine):

(dlv) b C:\dev\gopath\src\github.com\spiros-atos\torque_exporter\qstat.go:111

Run the program (continue) like this:

(dlv) c

Then in another terminal window (or browser of course) "visit" localhost:9100/metrics. This will break at the above set breakpoint from where you can then 'n' (execute next line), 's' (step into function), ... etc. (see gdb, pdb, or delve's documentation).

## License

Apache 2.0

By downloading this software, the downloader agrees with the specified terms and conditions of the License Agreement and the particularities of the license provided.

