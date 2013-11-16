ObamaD is a host monitoring system that is extremely easy to deploy and configure.

Configuration
-------------

TODO: create a web UI for generating the config string based on a form.

Example input string:

    "disk=/&tcp=localhost:22&sa=&st=yves.junqueira%40gmail.com&sf=root%40cetico.org"

This creates a prober that checks if port 22 is reachable and if the root
filesystem has enough space, sending an email in case of problems.

Deployment
----------

ObamaD should be run as an unprivileged user. It doesn't need to be installed like normal daemons. Copy the binary to the user's $HOME/bin directory and run it from there.

To ensure that it's run after boot, create a crontab entry for re-running ObamaD. From the shell, type: 

$ crontab -e

And add a crontab line such as:

    @hourly nohup $HOME/bin/obamad "disk=/&tcp=localhost:22&sa=&st=email\%40example.com&sf=root\%40obamad.com" &

Notes on using it with crontab:
- it doesn't use special privileges
- the % symbols must be escaped by \
- it will silently exit if another copy of obamad is already running
- it's configured to run @hourly just in case it unexpectedly crashes -
  strictly speaking, @reboot should be enough.
- See the "Configuration" section above for instructions on creating the config string.