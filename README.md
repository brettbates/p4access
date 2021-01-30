# p4 access

**p4access** is a Perforce broker filter program that responds to requests for access to an area with the group(s) that will give them that access and who the owners are to contact.

# Running the command
```
p4 access <read|write> <path>

Read will find any read or open groups, write finds only write groups. The more specific you are with a path, the better the results. For example:

p4 access read //depot/Jam/MAIN/...

Will give you any read/open group that has a protection entry for //depot/Jam/MAIN/... if we can't find one for MAIN, look for //depot/Jam/..., failing that //depot/... etc.
```

# Setup

> Please note I have only tested this on Linux, I don't know if it works on Windows, but I don't see any reason it wouldn't.

First clone the repository to a disk that your p4broker can access and build the binary:

```
git clone git@github.com:brettbates/p4access.git
cd p4access
go build
```

This produces a 'p4access' binary.

Add a stanza to the brokers .conf file like so:

```
command: access
{
    action=filter;
    execute="/path/to/p4access/p4access";
}
```

Where p4access is the binary created by building this module.


# Environment Variables
These environment variables are required, they must be readable by the running p4broker, so you will probably need to restart your broker. If you are using the sdp, you can add them to /p4/common/bin/p4_vars, if you do this, change the execute from p4broker.conf to something like:

> execute="/p4/common/bin/p4master_run 1 -c /p4/common/bin/p4access/p4access";

```
P4ACCESS_P4PORT
    The url:port of the target server.
P4ACCESS_P4USER
    The running perforce user, must be a super user to get the required information
P4ACCESS_P4CLIENT
    Not used currently, but the name of a p4 client/workspace


Paths:
All paths are from the perspective of the p4broker, so to avoid confusion, use the full path to the file. 
P4ACCESS_RESPONSE
    The template to use for a non-error response
    './io/results.go.tpl'
P4ACCESS_HELP
    The help text file
    './io/help.txt'
P4ACCESS_LOG
    The log file
    'p4access.log'
```
