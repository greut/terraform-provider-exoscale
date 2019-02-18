# Salt

Salt is one of the leading configuration management software. This example setups a mono-master configuration.

NB, the master is also a minion, as this enables managing it via salt.

```console
% terraform apply

master_ip = ubuntu@89.145.160.50
```

## Master

Accepting the minions in.

```console
% ssh ubuntu@89.145.160.50

$ sudo salt-key
Accepted Keys:
Denied Keys:
Unaccepted Keys:
salt-master-001
salt-minion-001
salt-minion-002
Rejected Keys:

$ sudo salt-key -a salt-minion-001
The following keys are going to be accepted:
Unaccepted Keys:
salt-minion-001
Proceed? [n/Y] y
Key for minion salt-minion-001 accepted.

...
```

Trying it.

```console
$ sudo salt '*' cmd.run 'hostname -a'
salt-master-001:
    salt-master-001
salt-minion-001:
    salt-minion-001
salt-minion-002:
    salt-minion-002
```

## Salt-API

```json
{
  "hosts": ["89.145.160.50.xip.io"],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [{
    "C": "CH",
    "L": "Lausanne",
    "O": "Akenes SA",
    "OU": "Exoscale",
    "ST": ""
  }]
}
```

```
$ cfssl genkey -initca csr.json | cfssljson -bare ca

$ cfssl gencert \
    -ca ca.pem \
    -ca-key ca-key.pem \
    -hostname 89.145.160.50.xip.io csr.json | cfssljson -bare

$ sudo mkdir -p /etc/pki/tls/certs
$ sudo mkdir -p /etc/pki/tls/private

$ sudo cp cert.pem /etc/pki/tls/certs/89.145.160.50.xip.io.pem
$ sudo cp cert-key.pem /etc/pki/tls/private/89.145.160.50.xip.io.pem
```

### `/etc/salt/master.d/api.conf`

The tornado server is the fanciest, maybe not the most stable one.

```yaml
rest_tornado:
    port: 8443
    ssl_crt: /etc/pki/tls/certs/89.145.160.50.xip.io.pem
    ssl_key: /etc/pki/tls/private/89.145.160.50.xip.io.pem
```

### `/etc/salt/master.d/eauth.conf`

```yaml
external_auth:
    pam:
        salt:
            - .*
            - '@wheel'
            - '@runner'
```



```console
$ echo salt:pa$$ | sudo chpasswd

$ sudo systemctl restart salt-api
$ sudo systemctl restart salt-master
$ sudo systemctl restart salt-minion
```

### Locally...


```console
% pip install salt-pepper

% export SALTAPI_USER=salt SALTAPI_PASS=pa$$ SALTAPI_EAUTH=pam
% export SALTAPI_URL=https://89.145.160.50.xip.io:8443/

% pepper '*' test.ping
{
    "return": [
        {
            "salt-master-001": true,
            "salt-minion-001": true,
            "salt-minion-002": true
        }
    ]
}

```
