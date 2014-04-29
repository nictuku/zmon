Zmon is a host monitoring system that aims to be extremely easy to deploy and configure. It is not there yet.

Notifications are sent via SMTP or via pushover.net (alerts to mobile devices via Google Cloud Messaging).

Configuration
-------------

Zmon will eventually have its own minimal web interface for changing the configuration. In the meantime you have to write the config file manually into $HOME/.zmon/zmon.json. Example:

    {
      "Probes": [
        {
          "Type": "disk",
          "Target": "/",
          "IntervalSeconds": 5
        },
        {
          "Type": "tcp",
          "Target": "localhost:22",
          "IntervalSeconds": 5
        },
        {
          "Type": "http",
          "Target": "http://localhost:4040",
          "IntervalSeconds": 5
        }
      ],
      "Notification": [
        {
          "Type": "pushover",
          "Destination": "userdestination"
        },
        {
          "Type": "smtp",
          "Destination": "user@example.com",
          "From": "zmon@example.com"
        }
      ]
    }


Installation
----------

Download the binary version of a recent release for your platform and gunzip it, or build it from source. Copy the binary to the user's $HOME/bin directory and run it from there.

To ensure that zmon runs after boot,  a convenient method is to create a crontab entry for re-running Zmon. From the shell, type: 

```
$ crontab -e
```

Add the following lines:

    MAILTO=""
    @reboot nohup $HOME/bin/zmon &

Notes on using it with crontab:
- it doesn't use special privileges
- the MAILTO="" prevents cron from sending you all the output of zmon.
- See the "Configuration" section above for instructions on creating the config file.

Limitations
-----------

If the hardware or network becomes offline or zmon stops working for whatever reason, users won't know. Work is being done to make Zmons frequently contact a _mothership_ that will be responsible for finding missing or dead agents.
