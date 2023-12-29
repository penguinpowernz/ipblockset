# ipblockset

Daemon that downloads a list of bad IPs from [github.com/scriptzteam/IP-BlockList-v4](https://github.com/scriptzteam/IP-BlockList-v4),
adds them to an IP set and blocks them on the `INPUT` chain.  By default it happens every 7 days, but cron can be configured to run
at the schedule you like.

## Usage

Oneshot mode to pull, parse and set.

    $ sudo ipblockset
    2023/12/30 02:04:02 Pulling IPs from https://raw.githubusercontent.com/scriptzteam/IP-BlockList-v4/master/ips.txt
    2023/12/30 02:04:02 Found 4570 lines
    2023/12/30 02:04:02 Omitted 619 IPs due to low level
    2023/12/30 02:04:02 Found 3944 IPs to block
    2023/12/30 02:04:02 Flushing and recreating IP set blocklist
    2023/12/30 02:04:02 Adding 3944 IPs to the IP set
    2023/12/30 02:04:06 Done

Daemon mode to run in the background:

    $ sudo ipblockset -d

Or use the systemd service:

    $ sudo systemctl start ipblockset
    $ sudo systemctl enable ipblockset

## File Format

At this stage it just uses the default the file format used by [scriptzteam](https://github.com/scriptzteam). It alse uses
only levels 3 and above, as per the default script on [IP-BlockList-v4](https://github.com/scriptzteam/IP-BlockList-v4).

You can specify your own URL to use instead of the default, but it should have the same format.