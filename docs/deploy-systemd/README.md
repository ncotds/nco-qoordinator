### Running NCO-Qoordinator API as systemd service

#### Prerequisites

Install freetds-dev package using your distribution pkg-manager.

For example:
```
# Ubuntu
apt update && apt install freetds-dev
```

#### Installation

* download pre-build tgz-archive from [release page](https://github.com/ncotds/nco-qoordinator/releases)
* unpack:
  ```shell
  tar xzvf ncoq-api-linux-amd64.tgz
  ```
* copy binary and configs:
  ```shell
  cd ncoq-api-linux-amd64
  mkdir -p ~/.local/bin && mv ncoq-api ~/.local/bin/
  mkdir -p ~/.config/systemd/user && mv ncoq-api.service ~/.config/systemd/user/
  mkdir -p ~/.config/ncoq-api && mv config.yml ~/.config/ncoq-api/
  ```
* update config with actual values (see [config file comments](../../config/example.yml) for details)
* enable service:
  ```shell
  systemctl --user daemon-reload
  systemctl --user enable ncoq-api
  systemctl --user start ncoq-api
  systemctl --user status ncoq-api
  ```
