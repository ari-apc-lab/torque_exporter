# torque_exporter
Allow a server to collect metrics from TORQUE and expose them in Prometheus format. The exporter accesses TORQUE via ssh to a remote machine that can perform `qstat` and `showq` commands.

## Install

> Requires Go >=1.8

```
go get github.com/ari-apc-lab/torque_exporter
$GOPATH/src/github.com/ari-apc-lab/torque_exporter/utils/install.sh
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

## License

Apache 2.0

By downloading this software, the downloader agrees with the specified terms and conditions of the License Agreement and the particularities of the license provided.

