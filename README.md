Zmon is a host monitoring system that is extremely easy to deploy and configure.

Notifications are sent via SMTP or via pushover.net (alerts to mobile devices via Google Cloud Messaging).

Configuration
-------------

Zmon will eventually have its own minimal web interface for changing the configuration. In the meantime you have to write the config file manually into ~/.zmon. It is a URL-like string like the following:

    "disk=/&tcp=localhost:22&tcp=localhost:80&sa=&st=yves.junqueira%40gmail.com&sf=root%40cetico.org"

This creates a prober that checks if ports 22 and 80 are reachable and if the root filesystem has enough space, sending an email to `st` with a sender of `sf` in case of problems.

Example config string:


Installation
----------

Zmon should be run as an unprivileged user. It doesn't need to be installed like normal daemons. 

Download the binary version of a recent release for your platform and gunzip it, or build it from source. Copy the binary to the user's $HOME/bin directory and run it from there.

To ensure that zmon runs after boot, create a crontab entry for re-running Zmon. From the shell, type: 

```
$ crontab -e
```

And add a crontab line such as:

    @hourly nohup $HOME/bin/zmon "disk=/&tcp=localhost:22&tcp=localhost:80&sa=&st=email\%40example.com&sf=root\%40zmon.org" &

Notes on using it with crontab:
- it doesn't use special privileges
- the % symbols must be escaped by \
- it will silently exit if another copy of zmon is already running
- it's configured to run @hourly just in case it unexpectedly crashes -
  strictly speaking, @reboot should be enough.
- See the "Configuration" section above for instructions on creating the config string.

Limitations
-----------

If the hardware or network becomes offline or zmon stops working for whatever reason, there is currently no way for users to know. This could be solved by adding a central server that receives heartbeats from zmons, but this is currently not available.


Development
-----------

Goals and tasks are tracked in the [Zmon Trello board](https://trello.com/b/ulJljBwJ/zmon).
