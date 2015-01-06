## How to use

kenko -c /etc/hello/kenko.conf

```
2015/01/05 22:26:55 [kenko] Configuration loaded from: kenko.conf
2015/01/05 22:26:55 [kenko] Kenko is configured to report metrics to: 127.0.0.1 on port 127.0.0.1 at a 5s interval.
...
```

Example configuration:

```toml
ec2 = false

[riemann]
host = "127.0.0.1"
port = 5555
exit_on_send_error = true

[event]
interval = 5
ttl = 10

[load]
warning = 10
critical = 100
check = true
```

## TODO:

Bundle with upstart script for easy start/stop/restart

Based on github.com/ippontech/goshin published with the following license

License
-------

Copyright 2012-2014 [Ippon Technologies](http://www.ippon.fr) and [Ippon Hosting](http://www.ippon-hosting.com/)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this application except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.




