# What?
A tool to capture and later verify/report/update checksums. 

The aim of this tool is not to replace tools like `tripwire` in a security context but rather to record checksums of files as part of a backup/replication strategy for long term storage.

# Background?
I store my photos on a QNAP NAS at home.  This QNAP replicates its data to itself, to another QNAP and eventually to a machine in another country.

The problem is should there be Bit Rot or corruption on the source QNAP within a month that rot will have spread to all corners of the storage strategy.

I need a way to detect and correct this rot whenever it happens, this tool lets me detect that allowing me to retrieve files from other copies to recover.

# Usage?

First you have to capture checksums on all your files, do this the first time using the tool:

```
rotrep add --path /data
```

You'll now have `.checksums.json` files in every directory, you now have to go into a cycle of verify/update regularly:

```
rotrep verify --path /data
```

Should this report any issue, verify if this is expected changes to your files, if not restore from other backups.

You should also check the status of files on every one of your replicas regularly to ensure they are in a good state, this ensures that you'll always be able to recover files should the source fail or rot.

Finally update your checksums - this will add new files and update checksums of any changed files:

If you had no changes you wish to update use the `add` command else use the `update` command which will be slower as it also does a full verify.

```
rotrep add --path /data # or update
```

All commands take `--verbose`, `--debug`, `--workers` and other options, see `--help`.

# Who?
R.I.Pienaar / www.devco.net / @ripienaar
